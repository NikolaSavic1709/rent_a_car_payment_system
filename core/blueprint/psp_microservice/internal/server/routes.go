package server

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3001"}, // Add your frontend URL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true, // Enable cookies/auth
	}))

	r.GET("/", s.HelloWorldHandler)

	r.GET("/health", s.healthHandler)

	r.POST("/test-postgre", s.NewTransactionHandler)
	r.POST("/payment", s.PaymentHandler)
	r.POST("/card-details", s.CardDetailsHandler)
	r.POST("/qr-scan", s.QRCodeScanningHandler)
	r.POST("/transaction-status", s.TransactionStatusHandler)
	r.PUT("/payment-callback", s.PaymentCallbackHandler)

	r.POST("/subscription/url", s.SendSubscriptionUrlsHandler)
	r.POST("/subscription", s.SaveSubscriptionForMarchantHandler)
	r.GET("/subscription/:merchantId", s.GetSubscriptionsForMarchantHandler)

	return r
}

func (s *Server) HelloWorldHandler(c *gin.Context) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"

	c.JSON(http.StatusOK, resp)
}

func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.db.Health())
}
