package server

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3001", "http://localhost:3002"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	r.GET("/", s.HelloWorldHandler)
	r.GET("/health", s.healthHandler)

	// Payment endpoints
	r.POST("/payment", s.InitiatePaymentHandler)
	r.GET("/payment-status/:paymentId", s.GetPaymentStatusHandler)
	r.POST("/verify-transaction", s.VerifyTransactionHandler)

	// Wallet management
	r.POST("/wallet/generate", s.GenerateWalletHandler)
	r.GET("/wallet/:merchantId/:currency", s.GetWalletHandler)

	// Admin/testing endpoints
	r.POST("/simulate-payment", s.SimulatePaymentHandler) // For testnet simulation
	r.POST("/simulate-confirmation", s.SimulateConfirmationHandler)

	return r
}

func (s *Server) HelloWorldHandler(c *gin.Context) {
	resp := make(map[string]string)
	resp["message"] = "Crypto Payment Service"
	resp["version"] = "1.0.0"
	c.JSON(http.StatusOK, resp)
}

func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.db.Health())
}
