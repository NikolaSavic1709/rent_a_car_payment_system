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

func (s *Server) ListVehicles(c *gin.Context) {
	userRole := c.GetString("role") // JWT middleware adds "role"
	userID := c.GetInt("userID")    // JWT middleware adds "userID"

	categoryFilter := c.Query("category") // Optional filter

	// Get all vehicles and process potential errors
	vehicles, err := s.db.GetAllVehicles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch vehicles"})
		return
	}

	if userRole == "customer" {
		// Get active vehicles for the user
		activeVehicles, err := s.db.GetActiveVehiclesByUser(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch active vehicles"})
			return
		}

		activeVehicleIDs := make(map[int]struct{})
		for _, payment := range activeVehicles {
			activeVehicleIDs[payment.Vehicle.ID] = struct{}{}
		}

		// Filter out active vehicles that user already has
		filtered := []model.Vehicle{}
		for _, vehicle := range vehicles {
			if _, exists := activeVehicleIDs[vehicle.ID]; !exists {
				filtered = append(filtered, vehicle)
			}
		}
		vehicles = filtered
	}

	// Change filtering based on category if provided
	if categoryFilter != "" {
		filtered := []model.Vehicle{}
		for _, vehicle := range vehicles {
			if string(vehicle.Category) == categoryFilter {
				filtered = append(filtered, vehicle)
			}
		}
		vehicles = filtered
	}

	c.JSON(http.StatusOK, vehicles)
}

func (s *Server) GetVehicleByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	vehicle, err := s.db.GetVehicleByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Vehicle not found"})
		return
	}
	c.JSON(http.StatusOK, vehicle)
}

func (s *Server) PurchaseVehicle(c *gin.Context) {
	var req struct {
		VehicleID     int    `json:"vehicleId"`
		Days          int    `json:"days"`
		PaymentMethod string `json:"method"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	userID := c.GetInt("userId")
	vehicle, err := s.db.GetVehicleByID(req.VehicleID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Vehicle not found"})
		return
	}

	// Calculate payment details
	deadline := time.Now().AddDate(0, 0, req.Days)
	paymentDeadline := time.Now().Add(5 * time.Minute) // Payment must be completed in 5 minutes
	payment := model.Payment{
		Deadline: deadline,
		Cost:     vehicle.Price * float64(req.Days),
		Vehicle:  *vehicle,
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
	if err := s.db.CreatePayment(userID, payment, req.VehicleID, token); err != nil {
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

func (s *Server) CreateVehicle(c *gin.Context) {
	var vehicle model.Vehicle
	if err := c.ShouldBindJSON(&vehicle); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if err := s.db.CreateVehicle(vehicle); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create vehicle"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Vehicle created successfully"})
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
		p := payment.Vehicle
		fmt.Println("Vehicle payment")
		userInfo.Payments = append(userInfo.Payments, struct {
			ID          int                   `json:"id"`
			Category    model.VehicleCategory `json:"category"`
			Name        string                `json:"name"`
			Description string                `json:"description"`
			Deadline    time.Time             `json:"deadline"`
		}{
			ID:          p.ID,
			Category:    p.Category,
			Name:        p.Name,
			Description: p.Description,
			Deadline:    payment.Deadline,
		})
	}

	c.JSON(http.StatusOK, userInfo)
}
