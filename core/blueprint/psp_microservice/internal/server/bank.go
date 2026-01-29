package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *Server) CardPaymentHandler(c *gin.Context) {
	tokenId := uuid.New()
	paymentURL := fmt.Sprintf("http://localhost:3001/card?tokenId=%s", tokenId) //:TODO for other payments

	response := database.PaymentStartResponse{
		PaymentURL: paymentURL,
		TokenId:    tokenId,
		Token:      "token",
		TokenExp:   time.Now().Add(15 * time.Minute),
	}
	c.JSON(http.StatusOK, response)
}

func (s *Server) QrCodePaymentHandler(c *gin.Context) {
	// Generate QR code URL (stubbed here)
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