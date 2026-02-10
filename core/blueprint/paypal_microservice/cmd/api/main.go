package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"paypal_microservice/internal/database"
	"paypal_microservice/internal/server"
)

func main() {
	// Initialize database
	dbService := database.New()
	defer dbService.Close()

	// Create server
	srv := server.NewServer(dbService)

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down gracefully...")
	}()

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting PayPal payment service on port %s", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}
}
