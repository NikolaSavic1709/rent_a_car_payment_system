package server

import (
	"crypto_microservice/internal/database"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GenerateWalletHandler generates a new wallet for a merchant
func (s *Server) GenerateWalletHandler(c *gin.Context) {
	var req database.WalletGenerateRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate currency
	if _, exists := database.SupportedCurrencies[req.Currency]; !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported currency"})
		return
	}

	// Generate or get existing wallet
	wallet, err := s.db.GetOrCreateMerchantWallet(req.MerchantId, req.Currency)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate wallet"})
		return
	}

	response := database.WalletGenerateResponse{
		WalletAddress: wallet.WalletAddress,
		Currency:      wallet.Currency,
		PublicKey:     wallet.PublicKey,
	}

	c.JSON(http.StatusOK, response)
}

// GetWalletHandler retrieves a merchant's wallet
func (s *Server) GetWalletHandler(c *gin.Context) {
	merchantIdStr := c.Param("merchantId")
	currency := c.Param("currency")

	merchantId, err := strconv.ParseUint(merchantIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid merchant ID"})
		return
	}

	// Validate currency
	if _, exists := database.SupportedCurrencies[currency]; !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported currency"})
		return
	}

	wallet, err := s.db.GetWallet(uint(merchantId), currency)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Wallet not found"})
		return
	}

	c.JSON(http.StatusOK, wallet)
}
