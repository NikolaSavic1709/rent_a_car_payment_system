package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"psp_microservice/internal/database"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *Server) CryptoPaymentHandler(c *gin.Context, merchantOrderId uuid.UUID) {
	tokenId := uuid.New()
	// Use merchantOrderId in URL so frontend can fetch payment details
	paymentURL := fmt.Sprintf("http://localhost:3002/payment?merchantOrderId=%s&tokenId=%s", merchantOrderId, tokenId)

	response := database.PaymentStartResponse{
		PaymentURL: paymentURL,
		TokenId:    tokenId,
		Token:      "token",
		TokenExp:   time.Now().Add(30 * time.Minute),
	}
	c.JSON(http.StatusOK, response)
}

// CryptoPaymentDetailsHandler handles requests from the crypto payment page
func (s *Server) CryptoPaymentDetailsHandler(c *gin.Context) {
	merchantOrderIdStr := c.Query("merchantOrderId")
	if merchantOrderIdStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "merchantOrderId is required"})
		return
	}

	merchantOrderId, err := uuid.Parse(merchantOrderIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid merchantOrderId format"})
		return
	}

	// Get the transaction associated with this merchantOrderId
	transaction, err := s.db.GetTransactionByMerchantOrderId(merchantOrderId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	// Convert fiat currency to crypto
	// Default to BTC for crypto payments
	// TODO: Allow user to select crypto currency
	cryptoCurrency := "BTC"
	cryptoAmount := convertToCrypto(float64(transaction.Amount), transaction.Currency, cryptoCurrency)

	// Forward the payment request to crypto service
	paymentRequest := database.PaymentRequest{
		Currency:        cryptoCurrency,
		Amount:          float32(cryptoAmount),
		MerchantId:      transaction.MerchantId,
		MerchantOrderId: merchantOrderId,
		TransactionId:   transaction.TransactionId,
		Timestamp:       transaction.Timestamp,
	}

	s.ForwardPaymentToCryptoService(paymentRequest, c)
}

// convertToCrypto converts fiat amount to cryptocurrency amount
// This is a simplified version - in production, use real-time exchange rates
func convertToCrypto(fiatAmount float64, fiatCurrency string, cryptoCurrency string) float64 {
	// Simplified conversion rates (example rates, not real-time)
	// In production, fetch from an API like CoinGecko or Binance
	rates := map[string]map[string]float64{
		"RSD": { // Serbian Dinar
			"BTC":  0.000000095, // ~1 RSD = 0.000000095 BTC (example rate)
			"ETH":  0.0000015,   // ~1 RSD = 0.0000015 ETH
			"USDT": 0.0093,      // ~1 RSD = 0.0093 USDT
		},
		"USD": {
			"BTC":  0.000010,
			"ETH":  0.00017,
			"USDT": 1.0,
		},
		"EUR": {
			"BTC":  0.000011,
			"ETH":  0.00019,
			"USDT": 1.08,
		},
	}

	rate, exists := rates[fiatCurrency][cryptoCurrency]
	if !exists {
		// Default rate if not found
		rate = rates["RSD"][cryptoCurrency]
	}

	return fiatAmount * rate
}

// ForwardPaymentToCryptoService forwards the payment to the crypto microservice
func (s *Server) ForwardPaymentToCryptoService(paymentRequest database.PaymentRequest, c *gin.Context) {
	fmt.Println("Forwarding payment to crypto service")

	cryptoServiceURL := "http://crypto_service:8080/payment"

	// Create the request body matching crypto service's CryptoPaymentRequest
	cryptoReq := map[string]interface{}{
		"transactionId":   paymentRequest.TransactionId,
		"merchantOrderId": paymentRequest.MerchantOrderId,
		"amount":          paymentRequest.Amount,
		"currency":        paymentRequest.Currency,
		"timestamp":       paymentRequest.Timestamp,
		"merchantId":      paymentRequest.MerchantId,
	}

	reqBody, err := json.Marshal(cryptoReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	resp, err := http.Post(cryptoServiceURL, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Crypto service unavailable"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(resp.StatusCode, gin.H{"error": "Crypto service error"})
		return
	}

	// Parse crypto service response
	var cryptoResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&cryptoResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response"})
		return
	}

	c.JSON(http.StatusOK, cryptoResp)
}

// CryptoPaymentStatusHandler checks the status of a crypto payment
func (s *Server) CryptoPaymentStatusHandler(c *gin.Context) {
	paymentIdStr := c.Query("paymentId")
	if paymentIdStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "paymentId is required"})
		return
	}

	paymentId, err := uuid.Parse(paymentIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid paymentId format"})
		return
	}

	// Query crypto service for payment status
	cryptoServiceURL := fmt.Sprintf("http://crypto_service:8080/payment-status/%s", paymentId)
	resp, err := http.Get(cryptoServiceURL)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Crypto service unavailable"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(resp.StatusCode, gin.H{"error": "Failed to get payment status"})
		return
	}

	var statusResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&statusResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse status"})
		return
	}

	c.JSON(http.StatusOK, statusResp)
}
