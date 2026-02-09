package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"paypal_microservice/internal/database"
)

// SendCallbackToPSP sends payment status callback to PSP
func (s *Server) SendCallbackToPSP(payment *database.PayPalPayment, status database.PaymentStatus) {
	go func() {
		pspURL := os.Getenv("PSP_CALLBACK_URL")
		if pspURL == "" {
			pspURL = "http://psp_service:8080/payment-callback"
		}

		callback := database.PayPalCallback{
			TransactionID:   payment.TransactionID,
			MerchantOrderID: payment.MerchantOrderID,
			Status:          mapPayPalStatusToPSPStatus(status),
			PayPalOrderID:   payment.PayPalOrderID,
			Amount:          payment.Amount,
			Currency:        payment.Currency,
			PayPalTimestamp: time.Now(),
		}

		reqBody, err := json.Marshal(callback)
		if err != nil {
			log.Printf("Error marshaling callback: %v", err)
			return
		}

		req, err := http.NewRequest("PUT", pspURL, bytes.NewBuffer(reqBody))
		if err != nil {
			log.Printf("Error creating request: %v", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Error sending callback to PSP: %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("PSP callback failed with status: %d", resp.StatusCode)
		} else {
			log.Printf("PSP callback successful for payment: %s", payment.PaymentID)
		}
	}()
}

// mapPayPalStatusToPSPStatus maps PayPal payment status to PSP transaction status
func mapPayPalStatusToPSPStatus(status database.PaymentStatus) database.TransactionStatus {
	switch status {
	case database.Completed:
		return database.Successful
	case database.Cancelled, database.Failed:
		return database.TransactionFailed
	case database.Approved:
		return database.InProgress
	default:
		return database.InProgress
	}
}

// getSuccessRedirectURL returns the success redirect URL
func (s *Server) getSuccessRedirectURL(paymentID string) string {
	successURL := os.Getenv("SUCCESS_URL")
	if successURL == "" {
		successURL = "http://localhost:3001/paypal?status=success"
	}
	return fmt.Sprintf("%s&paymentId=%s", successURL, paymentID)
}

// getCancelRedirectURL returns the cancel redirect URL
func (s *Server) getCancelRedirectURL(reason string) string {
	cancelURL := os.Getenv("CANCEL_URL")
	if cancelURL == "" {
		cancelURL = "http://localhost:3001/paypal?status=cancel"
	}
	if reason != "" {
		return fmt.Sprintf("%s&reason=%s", cancelURL, reason)
	}
	return cancelURL
}
