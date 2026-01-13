package server

import (
	"bank_gateway_microservice/internal/database"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (s *Server) NewTransactionHandler(c *gin.Context) {
	var req database.Transaction

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	//err := s.db.WriteTransaction(req)
	//if err != nil {
	//	return
	//}
	c.JSON(http.StatusCreated, gin.H{"message": "Transaction created", "transaction": req})
}
func (s *Server) PaymentHandler(c *gin.Context) {
	fmt.Println("usao u bank gateway")
	var req database.PaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Println("greska bind json")

		c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	bankId, err := s.db.GetBankByMerchantId(req.MerchantId)
	fmt.Println("bank id")
	fmt.Println(bankId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Bank not recognized"})

	}

	transaction := database.Transaction{
		RoutedBankId:    bankId,
		MerchantId:      req.MerchantId,
		MerchantOrderId: req.MerchantOrderId,
		Amount:          req.Amount,
		Timestamp:       req.Timestamp,
		TransactionId:   req.TransactionId,
		Currency:        req.Currency,
	}
	fmt.Println("before write")
	err = s.db.WriteTransaction(transaction)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Error"})

	}
	s.ForwardPaymentToBank(bankId, req)
	c.JSON(http.StatusOK, gin.H{"message": "Payment request forwarded to bank"})

}
func (s *Server) PaymentCallbackHandler(c *gin.Context) {

}
