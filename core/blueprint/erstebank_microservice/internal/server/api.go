package server

import (
	"erstebank_microservice/internal/database"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"time"
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
func (s *Server) PaymentHandler(c *gin.Context) {
	fmt.Println("USO U BANKU")
	type TempRequest struct {
		ExpDate         time.Time `json:"expDate" binding:"required"`
		CardNumber      string    `json:"cardNumber" binding:"required"`
		Currency        string    `json:"currency" binding:"required"`
		Amount          float32   `json:"amount" binding:"required"`
		MerchantId      uint      `json:"merchantId" binding:"required"`
		MerchantOrderId uuid.UUID `json:"merchantOrderId" binding:"required"`
		TransactionId   uuid.UUID `json:"transactionId" binding:"required"`
		Timestamp       time.Time `json:"timestamp" binding:"required"`
	}
	type PaymentRequest struct {
		ExpDate    time.Time `json:"expDate" binding:"required"`
		CardNumber string    `json:"cardNumber" binding:"required"`
		Currency   string    `json:"currency" binding:"required"`
		Amount     float32   `json:"amount" binding:"required"`
	}
	type Transaction struct {
		MerchantId      uint      `json:"merchantId" binding:"required"`
		MerchantOrderId uuid.UUID `json:"merchantOrderId" binding:"required"`
		Amount          float32   `json:"amount" binding:"required"`
		Timestamp       time.Time `json:"timestamp" binding:"required"`
		CardNumberLast  string    `json:"cardNumberLast" binding:"required,len=4"`
		TransactionId   uuid.UUID `json:"transactionId" binding:"required"`
		Currency        string    `json:"currency" binding:"required"`
	}
	type TransactionResponse struct {
		AcquirerOrderId   uuid.UUID                  `json:"acquirerOrderId" binding:"required"`
		AcquirerTimestamp time.Time                  `json:"acquirerTimestamp" binding:"required"`
		MerchantOrderId   uuid.UUID                  `json:"merchantOrderId" binding:"required"`
		TransactionId     uuid.UUID                  `json:"transactionId" binding:"required"`
		Status            database.TransactionStatus `json:"status" binding:"required"`
	}
	var tempReq TempRequest

	// Bind the incoming JSON to the TempRequest struct
	if err := c.ShouldBindJSON(&tempReq); err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	paymentReq := PaymentRequest{
		ExpDate:    tempReq.ExpDate,
		CardNumber: tempReq.CardNumber,
		Currency:   tempReq.Currency,
		Amount:     tempReq.Amount,
	}

	transaction := database.Transaction{
		AcquirerOrderId:   uuid.New(),
		AcquirerTimestamp: time.Now(),
		Status:            database.InProgress,
		MerchantId:        tempReq.MerchantId,
		MerchantOrderId:   tempReq.MerchantOrderId,
		Amount:            tempReq.Amount,
		Timestamp:         tempReq.Timestamp,
		PartialCardNumber: tempReq.CardNumber[len(tempReq.CardNumber)-4:], // Extract last 4 digits
		TransactionId:     tempReq.TransactionId,
		Currency:          tempReq.Currency,
	}

	err := s.db.WriteTransaction(transaction)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Error"})

	}

	var status database.TransactionStatus
	
	fmt.Println("Before payment")
	status, err = s.db.Pay(transaction.AcquirerOrderId, paymentReq.Currency, paymentReq.Amount, paymentReq.CardNumber, paymentReq.ExpDate, transaction.MerchantId)
	fmt.Println("After payment")

	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Error"})

	}
	response := TransactionResponse{
		AcquirerOrderId:   transaction.AcquirerOrderId,
		AcquirerTimestamp: transaction.AcquirerTimestamp,
		MerchantOrderId:   transaction.MerchantOrderId,
		TransactionId:     transaction.TransactionId,
		Status:            status,
	}

	if status == database.Successful {
		fmt.Println("successful")
		c.JSON(http.StatusOK, gin.H{"message": "Transaction created", "transaction": response})

	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Failed payment", "transaction": response})
	}
}
