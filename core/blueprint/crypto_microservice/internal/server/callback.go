package server

import (
	"bytes"
	"crypto_microservice/internal/database"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func (s *Server) SendCallbackToPSP(payment *database.CryptoPayment) {
	go func() {
		pspURL := "http://psp_service:8080/payment-callback"

		callback := database.CryptoPaymentCallback{
			TransactionId:   payment.TransactionId,
			MerchantOrderId: payment.MerchantOrderId,
			Status:          mapCryptoStatusToPSPStatus(payment.Status),
			TxHash:          payment.TxHash,
			Amount:          payment.Amount,
			Currency:        payment.Currency,
			Confirmations:   payment.Confirmations,
			CryptoTimestamp: time.Now(),
		}

		reqBody, err := json.Marshal(callback)
		if err != nil {
			fmt.Printf("Error marshaling callback: %v\n", err)
			return
		}

		req, err := http.NewRequest("PUT", pspURL, bytes.NewBuffer(reqBody))
		if err != nil {
			fmt.Printf("Error creating request: %v\n", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error sending callback to PSP: %v\n", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("PSP callback failed with status: %d\n", resp.StatusCode)
		} else {
			fmt.Printf("PSP callback successful for payment: %s\n", payment.PaymentId)
		}
	}()
}

func mapCryptoStatusToPSPStatus(status database.PaymentStatus) database.TransactionStatus {
	switch status {
	case database.Confirmed:
		return database.Successful
	case database.PaymentFailed, database.Expired:
		return database.Failed
	default:
		return database.InProgress
	}
}

// TransactionStatus matches PSP's status enum
type TransactionStatus int

const (
	Successful TransactionStatus = iota
	InProgress
	Failed
	Error
)
