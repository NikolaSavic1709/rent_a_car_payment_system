package server
import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)
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