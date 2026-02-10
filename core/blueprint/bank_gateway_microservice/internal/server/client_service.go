package server

import (
	"bank_gateway_microservice/internal/database"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

func (s *Server) ForwardPaymentToBank(bankId uint, transaction database.PaymentRequest) {

	go func() {
		fmt.Println("salje banci")
		bankServiceURL := "http://erstebank_service:8080/payment"
		reqBody, err := json.Marshal(transaction)
		fmt.Println("transaction")
		fmt.Println(string(reqBody))
		if err != nil {
			// Handle error, perhaps log and return an error response to the client
			return
		}

		resp, err := http.Post(bankServiceURL, "application/json", bytes.NewBuffer(reqBody))
		if err != nil {
			// Handle error, perhaps log and retry or notify an admin
			return
		}

		defer resp.Body.Close()
		var bankResponse struct {
			Message     string                       `json:"message"`
			Transaction database.TransactionResponse `json:"transaction"`
		}
		// Check for non-OK HTTP status codes
		if resp.StatusCode != http.StatusOK {
			fmt.Println("Received non-OK status:", resp.Status)
			fmt.Println("AAAAAAAAAA")
			go processBankResponseForPSP(bankResponse.Transaction, errors.New("this is a basic error"))
			return
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			go processBankResponseForPSP(bankResponse.Transaction, errors.New("error reading response body"))
			return
		}
		err = json.Unmarshal(body, &bankResponse)
		if err != nil {
			fmt.Println("Error unmarshalling response:", err)
			go processBankResponseForPSP(bankResponse.Transaction, errors.New("error decoding JSON response"))
			return
		}

		fmt.Println("Received response:", bankResponse.Message)
		fmt.Println("Transaction:", bankResponse.Transaction)

		go processBankResponseForPSP(bankResponse.Transaction, nil) // Launch a separate goroutine for asynchronous processing
	}()
}

func processBankResponseForPSP(response database.TransactionResponse, bankError error) {

	var responseBytes []byte
	fmt.Println("uso u process bank response")
	fmt.Println(bankError)
	if bankError != nil {
		responseBytes, _ = json.Marshal(nil)
	} else {
		var err error
		responseBytes, err = json.Marshal(response)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	fmt.Println(string(responseBytes))

	pspURL := "https://nginx/payment-callback"
	fmt.Println("sss")
	req, err := http.NewRequest("PUT", pspURL, bytes.NewBuffer(responseBytes))
	if err != nil {
		fmt.Println(err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
}
