package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"psp_microservice/internal/database"
	"strconv"
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

func (s *Server) CardDetailsHandler(c *gin.Context) {
	var req database.CardDetailsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}
	fmt.Println(req.MerchantOrderId)
	paymentRequest, err := s.db.GetTransactionByMerchantOrderId(req.MerchantOrderId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Error"})

	}
	fmt.Println("payment request")
	fmt.Println(paymentRequest.Amount)
	paymentRequest.CardNumber = req.CardNumber
	paymentRequest.ExpDate = req.ExpDate
	// s.SendURLToWebShop("http://localhost:3000/payment/success", req.MerchantOrderId)
	s.ForwardPaymentToBankGateway(paymentRequest)
	c.JSON(http.StatusOK, gin.H{"message": "Payment request forwarded"})
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

	tokenId := uuid.New()
	paymentURL := fmt.Sprintf("http://localhost:3001/card?tokenId=%s", tokenId) //:TODO for other payments

	response := database.PaymentStartResponse{
		PaymentURL: paymentURL,
		TokenId:    tokenId,
		Token:      "token",
		TokenExp:   time.Now().Add(15 * time.Minute),
	}
	fmt.Println("KRAJ")
	c.JSON(http.StatusOK, response)
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

func (s *Server) SendSubscriptionUrlsHandler(c *gin.Context) {
	type UrlSubscriptionRequest struct {
		MerchantId       uint   `json:"merchantId" binding:"required"`
		MerchantPassword string `json:"merchantPassword" binding:"required"`
	}

	var req UrlSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	merchant, err := s.db.CheckMerchant(req.MerchantId, req.MerchantPassword)
	if merchant == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid merchant"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	urlResponse := map[string]interface{}{
		"url": fmt.Sprintf("http://localhost:3001/subscription?merchantId=%s", strconv.Itoa(int(req.MerchantId))),
	}
	c.JSON(http.StatusOK, urlResponse)
}

func (s *Server) SaveSubscriptionForMarchantHandler(c *gin.Context) {
	type SubscriptionRequest struct {
		MerchantId       uint   `json:"merchantId" binding:"required"`
		MerchantPassword string `json:"merchantPassword" binding:"required"`
		Methods          []uint `json:"methods" binding:"required"`
	}

	var req SubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	merchant, err := s.db.CheckMerchant(req.MerchantId, req.MerchantPassword)
	if merchant == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid merchant"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = s.db.DeletePreviousSubscription(req.MerchantId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, method := range req.Methods {
		err = s.db.SaveSubscription(req.MerchantId, method)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, nil)
}

func (s *Server) GetSubscriptionsForMarchantHandler(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("merchantId"))
	methods, err := s.db.GetSubscriptionsForMerchant(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, methods)
}
