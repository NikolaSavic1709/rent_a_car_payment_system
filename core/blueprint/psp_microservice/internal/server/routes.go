package server

import (
	"net/http"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3001", "http://localhost:3002"}, // Add your frontend URLs
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
	r.GET("/crypto-payment-details", s.CryptoPaymentDetailsHandler)
	r.GET("/crypto-status", s.CryptoPaymentStatusHandler)

	// Login route
	r.POST("/login", s.LoginHandler)

	// Auth middleware for merchant endpoints.
	// If `Authorization: Bearer <token>` is present it is accepted as-is (token is stored
	// in the context under `auth_token`). Otherwise falls back to header username/password
	// check using X-Merchant-Username and X-Merchant-Password.
	authMiddleware := func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
			if token != "" {
				c.Set("auth_token", token)
				c.Next()
				return
			}
		}

		// Fallback to username/password headers
		username := c.GetHeader("X-Merchant-Username")
		password := c.GetHeader("X-Merchant-Password")
		if username == "" || password == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing credentials"})
			return
		}
		merchant, err := s.db.CheckMerchantByUsername(username, password)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if merchant == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		c.Set("merchant", merchant)
		c.Next()
	}

	// Protect subscription routes
	auth := r.Group("/", authMiddleware)
	r.POST("/subscription/url", s.SendSubscriptionUrlsHandler)
	auth.POST("/subscription", s.SaveSubscriptionForMarchantHandler)
	auth.GET("/subscription/:merchantId", s.GetSubscriptionsForMarchantHandler)
	auth.GET("/merchants", s.GetAllMerchantsHandler) // Protected route to get all merchant usernames

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
