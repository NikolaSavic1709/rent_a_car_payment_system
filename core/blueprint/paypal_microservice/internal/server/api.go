package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"paypal_microservice/internal/database"
)

// CreatePaymentHandler creates a new PayPal order
func (s *Server) CreatePaymentHandler(c *gin.Context) {
	var req database.PaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Invalid payment request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	log.Printf("Creating PayPal payment: TransactionID=%s, Amount=%.2f %s",
		req.TransactionID, req.Amount, req.Currency)

	// Create PayPal order
	order, err := s.paypal.CreateOrder(req.Amount, req.Currency, req.Description)
	if err != nil {
		log.Printf("Failed to create PayPal order: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create PayPal order"})
		return
	}

	// Generate payment ID
	paymentID := uuid.New()

	// Find approval URL from PayPal order
	var approvalURL string
	for _, link := range order.Links {
		if link.Rel == "approve" {
			approvalURL = link.Href
			break
		}
	}

	if approvalURL == "" {
		log.Printf("No approval URL in PayPal order response")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid PayPal order response"})
		return
	}

	// Store payment in database
	payment := &database.PayPalPayment{
		PaymentID:       paymentID,
		TransactionID:   req.TransactionID,
		MerchantOrderID: req.MerchantOrderID,
		MerchantID:      req.MerchantID,
		PayPalOrderID:   order.ID,
		Amount:          req.Amount,
		Currency:        req.Currency,
		Status:          database.Pending,
	}

	// Set optional fields
	if approvalURL != "" {
		payment.ApprovalURL.String = approvalURL
		payment.ApprovalURL.Valid = true
	}
	if req.Description != "" {
		payment.Description.String = req.Description
		payment.Description.Valid = true
	}

	if err := s.db.CreatePayment(payment); err != nil {
		log.Printf("Failed to store payment: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store payment"})
		return
	}

	log.Printf("PayPal payment created: PaymentID=%s, OrderID=%s", paymentID, order.ID)

	// Return response
	response := database.PaymentResponse{
		PaymentID:     paymentID,
		PayPalOrderID: order.ID,
		ApprovalURL:   approvalURL,
		Status:        database.Pending.String(),
	}

	c.JSON(http.StatusOK, response)
}

// GetPaymentStatusHandler retrieves payment status
func (s *Server) GetPaymentStatusHandler(c *gin.Context) {
	paymentID := c.Param("paymentId")

	log.Printf("Getting payment status for: %s", paymentID)

	payment, err := s.db.GetPaymentByID(paymentID)
	if err != nil {
		log.Printf("Payment not found: %s", paymentID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
		return
	}

	response := database.PaymentStatusResponse{
		PaymentID:     payment.PaymentID,
		TransactionID: payment.TransactionID,
		PayPalOrderID: payment.PayPalOrderID,
		Status:        payment.Status.String(),
		Amount:        payment.Amount,
		Currency:      payment.Currency,
		CreatedAt:     payment.CreatedAt,
	}

	if payment.PayerEmail.Valid {
		response.PayerEmail = payment.PayerEmail.String
	}

	if payment.CompletedAt.Valid {
		response.CompletedAt = &payment.CompletedAt.Time
	}

	c.JSON(http.StatusOK, response)
}

// PaymentSuccessHandler handles successful PayPal payment callback
func (s *Server) PaymentSuccessHandler(c *gin.Context) {
	orderID := c.Query("token") // PayPal sends order ID as 'token' parameter
	payerID := c.Query("PayerID")

	log.Printf("PayPal success callback: OrderID=%s, PayerID=%s", orderID, payerID)

	if orderID == "" {
		log.Printf("Missing order ID in success callback")
		c.Redirect(http.StatusFound, s.getCancelRedirectURL("missing_order_id"))
		return
	}

	// Get payment from database
	payment, err := s.db.GetPaymentByOrderID(orderID)
	if err != nil {
		log.Printf("Payment not found for order: %s", orderID)
		c.Redirect(http.StatusFound, s.getCancelRedirectURL("payment_not_found"))
		return
	}

	// Capture the payment
	capture, err := s.paypal.CaptureOrder(orderID)
	if err != nil {
		log.Printf("Failed to capture PayPal order %s: %v", orderID, err)

		// Update payment as failed
		s.db.UpdatePaymentStatus(payment.PaymentID.String(), database.Failed, "")

		// Send callback to PSP
		s.SendCallbackToPSP(payment, database.Failed)

		c.Redirect(http.StatusFound, s.getCancelRedirectURL("capture_failed"))
		return
	}

	// Extract capture ID
	captureID := ""
	if len(capture.PurchaseUnits) > 0 && len(capture.PurchaseUnits[0].Payments.Captures) > 0 {
		captureID = capture.PurchaseUnits[0].Payments.Captures[0].ID
	}

	log.Printf("PayPal payment captured: OrderID=%s, CaptureID=%s", orderID, captureID)

	// Update payment status
	if err := s.db.UpdatePaymentStatus(payment.PaymentID.String(), database.Completed, captureID); err != nil {
		log.Printf("Failed to update payment status: %v", err)
	}

	// Update payer information if available
	if capture.Payer != nil {
		payerEmail := ""
		if capture.Payer.EmailAddress != "" {
			payerEmail = capture.Payer.EmailAddress
		}

		payerName := ""
		if capture.Payer.Name != nil {
			payerName = fmt.Sprintf("%s %s", capture.Payer.Name.GivenName, capture.Payer.Name.Surname)
		}

		s.db.UpdatePaymentWithPayer(
			payment.PaymentID.String(),
			payerEmail,
			capture.Payer.PayerID,
			payerName,
		)
	}

	// Send callback to PSP
	s.SendCallbackToPSP(payment, database.Completed)

	// Redirect to success page with paymentId and merchantOrderId
	c.Redirect(http.StatusFound, s.getSuccessRedirectURL(payment.PaymentID.String(), payment.MerchantOrderID.String()))
}

// PaymentCancelHandler handles cancelled PayPal payment
func (s *Server) PaymentCancelHandler(c *gin.Context) {
	orderID := c.Query("token")

	log.Printf("PayPal cancel callback: OrderID=%s", orderID)

	if orderID == "" {
		c.Redirect(http.StatusFound, s.getCancelRedirectURL("missing_order_id"))
		return
	}

	// Get payment from database
	payment, err := s.db.GetPaymentByOrderID(orderID)
	if err != nil {
		log.Printf("Payment not found for order: %s", orderID)
		c.Redirect(http.StatusFound, s.getCancelRedirectURL("payment_not_found"))
		return
	}

	// Update payment status
	if err := s.db.UpdatePaymentStatus(payment.PaymentID.String(), database.Cancelled, ""); err != nil {
		log.Printf("Failed to update payment status: %v", err)
	}

	// Send callback to PSP
	s.SendCallbackToPSP(payment, database.Cancelled)

	// Redirect to cancel page
	c.Redirect(http.StatusFound, s.getCancelRedirectURL(payment.PaymentID.String()))
}

// getSuccessRedirectURL generates redirect URL for successful payment
