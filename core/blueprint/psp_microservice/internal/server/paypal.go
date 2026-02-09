package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"psp_microservice/internal/database"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PayPalPaymentHandler initiates a PayPal payment flow
// Flow: PSP -> PayPal Service -> PayPal Sandbox -> Callback -> PSP
func (s *Server) PayPalPaymentHandler(c *gin.Context, merchantOrderId uuid.UUID, transactionId uuid.UUID) {

	// Get the transaction details
	transaction, err := s.db.GetTransactionByMerchantOrderId(merchantOrderId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	// Prepare request for PayPal microservice
	paypalRequest := PayPalPaymentRequest{
		TransactionID:   transactionId,
		MerchantOrderID: merchantOrderId,
		MerchantID:      transaction.MerchantId,
		Amount:          float64(transaction.Amount),
		Currency:        transaction.Currency,
		Description:     fmt.Sprintf("Order %s", merchantOrderId.String()),
	}

	// Forward to PayPal microservice
	s.ForwardPaymentToPayPalService(paypalRequest, c)
}

// ForwardPaymentToPayPalService sends payment request to PayPal microservice
func (s *Server) ForwardPaymentToPayPalService(paymentRequest PayPalPaymentRequest, c *gin.Context) {
	paypalServiceURL := os.Getenv("PAYPAL_SERVICE_URL")
	if paypalServiceURL == "" {
		paypalServiceURL = "http://paypal_service:8080/payment"
	}

	// Marshal payment request
	reqBody, err := json.Marshal(paymentRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payment request"})
		return
	}

	// Create HTTP request to PayPal service
	req, err := http.NewRequest("POST", paypalServiceURL, bytes.NewBuffer(reqBody))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to PayPal service"})
		return
	}
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "PayPal service unavailable"})
		return
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read PayPal response"})
		return
	}

	// Check status
	if resp.StatusCode != http.StatusOK {
		c.JSON(resp.StatusCode, gin.H{"error": "PayPal payment creation failed", "details": string(body)})
		return
	}

	// Parse PayPal response
	var paypalResponse PayPalPaymentResponse
	if err := json.Unmarshal(body, &paypalResponse); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid PayPal response"})
		return
	}

	// Return response to client with approval URL
	response := database.PaymentStartResponse{
		PaymentURL: paypalResponse.ApprovalURL,
		TokenId:    paymentRequest.TransactionID,
		Token:      paypalResponse.PaymentID.String(),
		TokenExp:   time.Now().Add(30 * time.Minute),
	}

	c.JSON(http.StatusOK, response)
}

// PayPalPaymentRequest for PayPal microservice
type PayPalPaymentRequest struct {
	TransactionID   uuid.UUID `json:"transactionId" binding:"required"`
	MerchantOrderID uuid.UUID `json:"merchantOrderId" binding:"required"`
	MerchantID      uint      `json:"merchantId" binding:"required"`
	Amount          float64   `json:"amount" binding:"required,gt=0"`
	Currency        string    `json:"currency" binding:"required"`
	Description     string    `json:"description"`
}

// PayPalPaymentResponse from PayPal microservice
type PayPalPaymentResponse struct {
	PaymentID     uuid.UUID `json:"paymentId"`
	PayPalOrderID string    `json:"paypalOrderId"`
	ApprovalURL   string    `json:"approvalUrl"`
	Status        string    `json:"status"`
}
