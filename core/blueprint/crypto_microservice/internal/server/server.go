package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"crypto_microservice/internal/blockchain"
	"crypto_microservice/internal/database"
)

type Server struct {
	port    int
	db      database.Service
	monitor *blockchain.Monitor
}

func NewServer(db database.Service, monitor *blockchain.Monitor) *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	newServer := &Server{
		port:    port,
		db:      db,
		monitor: monitor,
	}

	// Set the callback sender so monitor can send callbacks
	monitor.SetCallbackSender(newServer)

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
