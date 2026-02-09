package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"paypal_microservice/internal/database"
)

type Server struct {
	port   int
	db     database.Service
	paypal PayPalClient
}

func NewServer(db database.Service) *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))

	// Initialize PayPal client
	paypalClient, err := NewPayPalClient()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize PayPal client: %v", err))
	}

	newServer := &Server{
		port:   port,
		db:     db,
		paypal: paypalClient,
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", newServer.port),
		Handler:      newServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
