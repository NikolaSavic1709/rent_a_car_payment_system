package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"psp_microservice/internal/database"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *Server) NewTransactionHandler(c *gin.Context) {
	var req database.Transaction

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := s.db.WriteTransaction(req)
	if err != nil {
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Transaction created", "transaction": req})
}

// LoginHandler checks merchant credentials by username and password.
func (s *Server) LoginHandler(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	merchant, err := s.db.CheckMerchantByUsername(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if merchant == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

// GetAllMerchantsHandler returns list of merchants (username and merchant_id).
func (s *Server) GetAllMerchantsHandler(c *gin.Context) {
	merchants, err := s.db.GetAllMerchants()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Return only username and merchantId for each merchant
	var out []map[string]interface{}
	for _, m := range merchants {
		out = append(out, map[string]interface{}{"username": m.Username, "merchantId": m.MerchantId})
	}
	c.JSON(http.StatusOK, gin.H{"merchants": out})
}

func (s *Server) PaymentHandler(c *gin.Context) {
	fmt.Println("USAO")
	var req database.WebShopPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	transaction := database.Transaction{
		MerchantId:        req.MerchantId,
		MerchantOrderId:   req.MerchantOrderId,
		MerchantTimestamp: req.MerchantTimestamp,
		Amount:            req.Amount,
		Timestamp:         time.Now(),
		TransactionId:     uuid.New(),
		Status:            database.InProgress,
		Currency:          req.Currency,
		PaymentMethod:     req.PaymentMethod,
		QRRef:             generateQRRefFromTimestamp(),
	}
	//var merchant database.Merchant
	merchant, err := s.db.CheckMerchant(req.MerchantId, req.MerchantPassword)
	if merchant == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid merchant"})
		return

	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	fmt.Println(transaction.MerchantOrderId)
	err = s.db.WriteTransaction(transaction)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Error"})

	}
	fmt.Println(req.PaymentMethod)
	if req.PaymentMethod == "CREDIT_CARD" {
		s.CardPaymentHandler(c)
	} else if req.PaymentMethod == "QR" {
		s.QrCodePaymentHandler(c, transaction.QRRef)
	} else if req.PaymentMethod == "PAYPAL" {
		s.PayPalPaymentHandler(c, transaction.MerchantOrderId, transaction.TransactionId)
	} else if req.PaymentMethod == "CRYPTO" {
		s.CryptoPaymentHandler(c, transaction.MerchantOrderId)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Unsupported payment method"})
	}
	//todo add cryptocurrency

}

func (s *Server) PaymentCallbackHandler(c *gin.Context) {
	fmt.Println("uso")
	var req database.TransactionResponse

	body, _ := io.ReadAll(c.Request.Body)
	if err := json.Unmarshal(body, &req); err != nil {
		fmt.Println("Unmarshal Error:", err)
	}
	fmt.Println("Request Body:", string(body))

	merchant_id, err := s.db.ChangeTransactionStatus(req.TransactionId, req.Status)
	url, err := s.db.GetMerchantRedirectURL(merchant_id, req.Status)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Error"})
	}
	fmt.Println(req.Status)
	s.SendURLToWebShop(url, req.MerchantOrderId)
	c.JSON(http.StatusOK, gin.H{"message": "Payment response forwarded"})
}

// TransactionStatusHandler allows frontend to poll transaction status by merchantOrderId.
// Returns an empty `url` when transaction is still InProgress (pending). For final
// statuses it returns the merchant redirect URL (success/fail/error) which may be empty
// if the merchant hasn't configured it.
func (s *Server) TransactionStatusHandler(c *gin.Context) {
	var req struct {
		MerchantOrderId uuid.UUID `json:"merchantOrderId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	url, err := s.db.GetRedirectURLByMerchantOrderId(req.MerchantOrderId)
	if err != nil {
		// Not found or DB error
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": url})
}

func generateQRRefFromTimestamp() uint64 {
	now := time.Now()
	qrRef := uint64(now.Year()*1e10 + int(now.Month())*1e8 + now.Day()*1e6 +
		now.Hour()*1e4 + now.Minute()*1e2 + now.Second())
	qrRef = qrRef*1000 + uint64(now.Nanosecond()/1e6)
	return qrRef
}
