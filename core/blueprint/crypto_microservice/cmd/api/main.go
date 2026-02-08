package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"crypto_microservice/internal/blockchain"
	"crypto_microservice/internal/database"
	"crypto_microservice/internal/server"
)

func main() {
	// Initialize database
	dbService := database.New()
	defer dbService.Close()

	// Initialize blockchain monitor
	monitor := blockchain.NewMonitor(dbService)
	go monitor.Start()

	// Create server
	srv := server.NewServer(dbService, monitor)

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down gracefully...")
		monitor.Stop()
	}()

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting crypto payment service on port %s", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}
}
