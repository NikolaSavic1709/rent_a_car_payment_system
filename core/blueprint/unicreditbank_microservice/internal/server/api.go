package server

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"unicreditbank_microservice/internal/database"
)

func (s *Server) NewTransactionHandler(c *gin.Context) {
	var req database.Transaction

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := s.db.WriteTransaction(req)
	if err != nil {
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Transaction created", "transaction": req})
}
