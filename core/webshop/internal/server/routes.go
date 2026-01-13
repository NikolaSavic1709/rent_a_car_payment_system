package server

import (
	"fmt"
	"webshop/internal/model"

	"github.com/google/uuid"

	//"golang.org/x/crypto/bcrypt"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:5173",
			"http://localhost:3000",
			"http://localhost",
			"http://localhost:80",
			"http://nginx",
			"http://nginx:80",
			"https://localhost",
			"https://nginx",
		}, // Add your frontend URL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true, // Enable cookies/auth
	}))

	r.GET("/", s.HelloWorldHandler)
	r.GET("/health", s.healthHandler)

	r.POST("/login", s.LoginHandler)
	r.POST("/register", s.RegisterHandler)

	r.POST("/purchase-status", s.CreatePurchaseStatusHandler)

	authorized := r.Group("/")
	fmt.Println(authorized)
	authorized.Use(s.AuthMiddleware())
	{
		authorized.GET("/protected", s.ProtectedHandler)
		authorized.GET("/admin", Authorize("admin"), s.AdminEndpoint)

		// Route for vehicles
		authorized.GET("/vehicles", s.ListVehicles)                       // List vehicles
		authorized.GET("/vehicles/:id", s.GetVehicleByID)                 // Get vehicle by ID
		authorized.POST("/vehicles", Authorize("admin"), s.CreateVehicle) // Create vehicle (only admin)
		authorized.POST("/vehicles/purchase", s.PurchaseVehicle)          // Rent a vehicle

		// Route for user info
		authorized.GET("/user/info", s.GetUserInfo)

		authorized.POST("/purchase-status/check", s.CheckPurchaseStatusHandler) // Check Purchase Status

		authorized.GET("/subscription", Authorize("admin"), s.GetSubscriptionUrl)

	}

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

func (s *Server) LoginHandler(c *gin.Context) {

	var loginData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&loginData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	//usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_.-]{3,20}$`)
	//if !usernameRegex.MatchString(loginData.Username) {
	//	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid username. It must be 3-20 characters long and may only contain letters, numbers, and the characters _ . -"})
	//	return
	//}
	//
	//// Validate password: 8-20 characters, allows a-zA-Z0-9 and _-.
	//passwordRegex := regexp.MustCompile(`^[a-zA-Z0-9_.-]{8,20}$`)
	//if !passwordRegex.MatchString(loginData.Password) {
	//	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid password. It must be 8-20 characters long and may only contain letters, numbers, and the characters _ . -"})
	//	return
	//}
	authenticated, err := s.authService.Authenticate(loginData.Username, loginData.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !authenticated {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	user, err := s.db.GetUserByUsername(loginData.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	token, err := s.authService.GenerateToken(loginData.Username, user.Role, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (s *Server) RegisterHandler(c *gin.Context) {
	var registerData struct {
		Fullname string `json:"fullname"`
		Email    string `json:"email"`
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&registerData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	existingUser, err := s.db.GetUserByUsername(registerData.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}
	if existingUser != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	}
	var role = "customer"
	err = s.db.CreateUser(registerData.Fullname, registerData.Email, registerData.Username, registerData.Password, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
}

func (s *Server) AdminEndpoint(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Welcome, Admin!"})
}

func (s *Server) CreatePurchaseStatusHandler(c *gin.Context) {
	var status model.PurchaseStatus
	if err := c.ShouldBindJSON(&status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}
	fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
	err := s.db.InsertPurchaseStatus(status)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store purchase status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "PurchaseStatus stored successfully",
		"merchantOrderId": status.MerchantOrderId,
	})
}

func (s *Server) CheckPurchaseStatusHandler(c *gin.Context) {
	var data struct {
		MerchantOrderId string `json:"merchantOrderId"`
	}

	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	merchantOrderId, err := uuid.Parse(data.MerchantOrderId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid MerchantOrderId format"})
		return
	}

	purchaseStatus, err := s.db.GetPurchaseStatusByMerchantOrderId(merchantOrderId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check purchase status"})
		return
	}

	if purchaseStatus == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "MerchantOrderId not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"merchantOrderId": purchaseStatus.MerchantOrderId,
		"url":             purchaseStatus.URL,
	})
}

func (s *Server) GetSubscriptionUrl(c *gin.Context) {
	request := map[string]interface{}{
		"merchantId":       12345,
		"merchantPassword": "webshop",
	}

	pspResponse, err := s.getSubscriptionUrlFromPSP(request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to communicate with PSP"})
		return
	}

	c.JSON(http.StatusOK, pspResponse)
}
