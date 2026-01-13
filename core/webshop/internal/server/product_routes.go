package server

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
	"webshop/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *Server) ListProducts(c *gin.Context) {
	userRole := c.GetString("role") // JWT middleware dodeljuje "role"
	userID := c.GetInt("userID")    // JWT middleware dodeljuje "userID"

	categoryFilter := c.Query("category") // Opcioni filter

	// Preuzmi sve proizvode i obradi potencijalne greške
	products, err := s.db.GetAllProducts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
		return
	}

	if userRole == "customer" {
		// Preuzmi aktivne proizvode za korisnika i obradi potencijalne greške
		activeProducts, err := s.db.GetActiveProductsByUser(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch active products"})
			return
		}

		activeProductIDs := make(map[int]struct{})
		for _, payment := range activeProducts {
			activeProductIDs[payment.Product.ID] = struct{}{}
		}

		// Filtriraj proizvode koje korisnik već ima
		filtered := []model.Product{}
		for _, product := range products {
			if _, exists := activeProductIDs[product.ID]; !exists {
				filtered = append(filtered, product)
			}
		}
		products = filtered
	}

	// Primeni opcioni filter po kategoriji
	if categoryFilter != "" {
		filtered := []model.Product{}
		for _, product := range products {
			if string(product.Category) == categoryFilter {
				filtered = append(filtered, product)
			}
		}
		products = filtered
	}

	c.JSON(http.StatusOK, products)
}

func (s *Server) GetProductByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	product, err := s.db.GetProductByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}
	c.JSON(http.StatusOK, product)
}

func (s *Server) PurchaseProduct(c *gin.Context) {
	var req struct {
		ProductID     int    `json:"productId"`
		Years         int    `json:"years"`
		PaymentMethod string `json:"method"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	userID := c.GetInt("userId")
	product, err := s.db.GetProductByID(req.ProductID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// Calculate payment details
	deadline := time.Now().AddDate(req.Years, 0, 0)
	paymentDeadline := time.Now().Add(5 * time.Minute) // Payment must be completed in 5 minutes
	payment := model.Payment{
		Deadline: deadline,
		Cost:     product.Price * float64(req.Years),
		Product:  *product,
	}

	// Generate unique order details
	merchantOrderID, err := uuid.NewUUID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate uuid"})
		return
	}

	c.Set("merchantOrderID", merchantOrderID)
	merchantTimestamp := time.Now() //.Format("2006-01-02T15:04:05")

	// PSP request payload
	pspRequest := map[string]interface{}{
		"paymentDeadline":   paymentDeadline,
		"amount":            payment.Cost,
		"currency":          "RSD",
		"successURL":        "http://localhost:3000/payment/success",
		"failURL":           "http://localhost:3000/payment/fail",
		"errorURL":          "http://localhost:3000/payment/error",
		"merchantId":        12345,
		"merchantPassword":  "webshop",
		"merchantOrderId":   merchantOrderID,
		"merchantTimestamp": merchantTimestamp,
	}
	fmt.Println(merchantOrderID)
	// Send request to PSP
	pspResponse, err := s.sendPSPRequest(pspRequest)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to communicate with PSP"})
		return
	}

	// Extract token from PSP response
	token, ok := pspResponse["tokenId"].(string)
	fmt.Println(pspResponse)
	fmt.Println(token)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid response from PSP"})
		return
	}

	// Store payment and PSP token
	if err := s.db.CreatePayment(userID, payment, req.ProductID, token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process payment"})
		return
	}

	// Retrieve merchantOrderID from context
	//if value, exists := c.Get("merchantOrderID"); exists {
	//	if retrievedOrderID, ok := value.(string); ok {
	//		fmt.Println("Successfully retrieved merchantOrderID:", retrievedOrderID)
	//	} else {
	//		fmt.Println("Failed to cast merchantOrderID to string")
	//	}
	//}

	paymentURL, ok := pspResponse["paymentURL"].(string)
	merchantPaymentURL := fmt.Sprintf("%s&merchantOrderId=%s", paymentURL, merchantOrderID)
	fmt.Println(paymentURL)

	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid response from PSP"})
		return
	}

	// Return token and redirect URL to frontend
	c.JSON(http.StatusOK, gin.H{
		"message":         "Payment initialized successfully",
		"pspToken":        token,
		"merchantOrderId": merchantOrderID,
		"redirectUrl":     merchantPaymentURL,
	})
}

func (s *Server) CreateProduct(c *gin.Context) {
	var product model.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if product.Category == model.Package && len(product.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Package must contain items"})
		return
	}

	if err := s.db.CreateProduct(product); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Product created successfully"})
}

func (s *Server) GetUserInfo(c *gin.Context) {
	userID := c.GetInt("userId")
	fmt.Println(userID)

	user, err := s.db.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	fmt.Println("HH")
	// Formatiraj odgovore sa datumima važenja aktivacija
	userInfo := struct {
		FullName string        `json:"fullname"`
		Payments []interface{} `json:"payments"`
	}{
		FullName: user.Fullname,
		Payments: []interface{}{},
	}

	for _, payment := range user.Payments {
		p := payment.Product
		fmt.Println("HH")
		userInfo.Payments = append(userInfo.Payments, struct {
			ID          int                   `json:"id"`
			Category    model.ProductCategory `json:"category"`
			Title       string                `json:"title"`
			Description string                `json:"description"`
			Deadline    time.Time             `json:"deadline"`
		}{
			ID:          p.ID,
			Category:    p.Category,
			Title:       p.Title,
			Description: p.Description,
			Deadline:    payment.Deadline,
		})
	}

	c.JSON(http.StatusOK, userInfo)
}
