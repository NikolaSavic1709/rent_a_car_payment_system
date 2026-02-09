package server

import (
	"crypto_microservice/internal/database"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *Server) InitiatePaymentHandler(c *gin.Context) {
	var req database.CryptoPaymentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("Error binding JSON: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	fmt.Printf("Received payment request: MerchantId=%d, Amount=%f, Currency=%s\n", req.MerchantId, req.Amount, req.Currency)

	// Get or create merchant wallet for this currency
	wallet, err := s.db.GetOrCreateMerchantWallet(req.MerchantId, req.Currency)
	if err != nil {
		fmt.Printf("Error getting/creating merchant wallet: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get merchant wallet: %v", err)})
		return
	}

	fmt.Printf("Got wallet for merchant %d: %s\n", req.MerchantId, wallet.WalletAddress)

	// Get currency config
	config, exists := database.SupportedCurrencies[req.Currency]
	if !exists {
		fmt.Printf("Unsupported currency: %s\n", req.Currency)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported currency"})
		return
	}

	// Create payment record
	payment := database.CryptoPayment{
		PaymentId:             uuid.New(),
		TransactionId:         req.TransactionId,
		MerchantOrderId:       req.MerchantOrderId,
		MerchantId:            req.MerchantId,
		Amount:                req.Amount,
		Currency:              req.Currency,
		Status:                database.Pending,
		DestinationAddress:    wallet.WalletAddress,
		RequiredConfirmations: config.RequiredConfirmations,
		CreatedAt:             time.Now(),
		ExpiryTime:            time.Now().Add(config.PaymentWindow),
		IsTestnet:             true,
	}

	if err := s.db.CreatePayment(&payment); err != nil {
		fmt.Printf("Error creating payment: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create payment: %v", err)})
		return
	}

	fmt.Printf("Payment created successfully: %s\n", payment.PaymentId)

	// Start monitoring for this payment
	go s.monitor.MonitorPayment(payment.PaymentId)

	// Generate payment URI for QR code
	paymentURI := fmt.Sprintf("%s:%s?amount=%.8f&label=Payment_%s",
		req.Currency, wallet.WalletAddress, req.Amount, payment.PaymentId.String()[:8])

	response := database.CryptoPaymentResponse{
		PaymentId:             payment.PaymentId,
		DestinationAddress:    wallet.WalletAddress,
		Amount:                req.Amount,
		Currency:              req.Currency,
		ExpiryTime:            payment.ExpiryTime,
		RequiredConfirmations: config.RequiredConfirmations,
		Status:                payment.Status.String(),
		QRCode:                paymentURI,
	}

	c.JSON(http.StatusOK, response)
}

func (s *Server) GetPaymentStatusHandler(c *gin.Context) {
	paymentIdStr := c.Param("paymentId")
	paymentId, err := uuid.Parse(paymentIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment ID"})
		return
	}

	payment, err := s.db.GetPaymentByPaymentId(paymentId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
		return
	}

	// Check if expired
	if payment.Status == database.Pending && time.Now().After(payment.ExpiryTime) {
		payment.Status = database.Expired
		s.db.UpdatePayment(payment)
	}

	response := database.PaymentStatusResponse{
		PaymentId:     payment.PaymentId,
		Status:        payment.Status.String(),
		Confirmations: payment.Confirmations,
		TxHash:        payment.TxHash,
		BlockHeight:   payment.BlockHeight,
	}

	if payment.ConfirmedAt != nil {
		response.ConfirmedAt = *payment.ConfirmedAt
	}

	c.JSON(http.StatusOK, response)
}

// Simulate payment for testing (testnet only)
func (s *Server) SimulatePaymentHandler(c *gin.Context) {
	type SimulateRequest struct {
		PaymentId     uuid.UUID `json:"paymentId" binding:"required"`
		SourceAddress string    `json:"sourceAddress" binding:"required"`
	}

	var req SimulateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payment, err := s.db.GetPaymentByPaymentId(req.PaymentId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
		return
	}

	// Simulate blockchain transaction
	// Generate a 64-character hex hash (like Bitcoin/Ethereum)
	txHash := fmt.Sprintf("0x%s%s", 
		uuid.New().String()[:32], 
		uuid.New().String()[:32])

	payment.Status = database.Confirming
	payment.SourceAddress = req.SourceAddress
	payment.TxHash = txHash
	payment.Confirmations = 1
	payment.BlockHeight = time.Now().Unix()

	if err := s.db.UpdatePayment(payment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update payment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Payment simulated",
		"txHash":  txHash,
		"status":  "confirming",
	})
}

// SimulateConfirmationHandler simulates adding confirmations to a payment
func (s *Server) SimulateConfirmationHandler(c *gin.Context) {
	type ConfirmRequest struct {
		PaymentId uuid.UUID `json:"paymentId" binding:"required"`
	}

	var req ConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payment, err := s.db.GetPaymentByPaymentId(req.PaymentId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
		return
	}

	if payment.Status != database.Confirming {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payment is not in confirming status"})
		return
	}

	// Add one more confirmation
	payment.Confirmations++

	if payment.Confirmations >= payment.RequiredConfirmations {
		payment.Status = database.Confirmed
		now := time.Now()
		payment.ConfirmedAt = &now
	}

	if err := s.db.UpdatePayment(payment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update payment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Confirmation added",
		"confirmations": payment.Confirmations,
		"status":        payment.Status.String(),
	})
}

// VerifyTransactionHandler verifies a transaction on the blockchain
func (s *Server) VerifyTransactionHandler(c *gin.Context) {
	var req database.TransactionVerifyRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payment, err := s.db.GetPaymentByPaymentId(req.PaymentId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
		return
	}

	// For testnet simulation, we just verify the txHash matches
	valid := payment.TxHash == req.TxHash

	response := database.TransactionVerifyResponse{
		Valid:         valid,
		Amount:        payment.Amount,
		Confirmations: payment.Confirmations,
		Timestamp:     payment.CreatedAt,
	}

	c.JSON(http.StatusOK, response)
}
