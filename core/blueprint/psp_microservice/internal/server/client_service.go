package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"psp_microservice/internal/database"
)

func (s *Server) ForwardPaymentToBankGateway(paymentRequest database.PaymentRequest) {
	fmt.Println("usao u forward")
	go func() {
		bankGatewayServiceURL := "http://bank_gateway_service:8080/payment"
		reqBody, err := json.Marshal(paymentRequest)
		if err != nil {
			// Handle error, perhaps log and return an error response to the client
			return
		}
		fmt.Println(paymentRequest.CardNumber)
		resp, err := http.Post(bankGatewayServiceURL, "application/json", bytes.NewBuffer(reqBody))
		if err != nil {
			// Handle error, perhaps log and retry or notify an admin
			return
		}
		defer resp.Body.Close()
		//dobija OK od gateway-a
	}()
}

func (s *Server) SendURLToWebShop(url string, merchantOrderId uuid.UUID) {

	go func() {
		webShopURL := "http://webshop_service:8080/purchase-status"
		payload := struct {
			URL             string    `json:"url"`
			MerchantOrderID uuid.UUID `json:"merchantOrderId"`
		}{
			URL:             url,
			MerchantOrderID: merchantOrderId,
		}

		// Marshal the payload into JSON
		reqBody, err := json.Marshal(payload)
		if err != nil {
			// Handle error, perhaps log and return an error response to the client
			return
		}

		resp, err := http.Post(webShopURL, "application/json", bytes.NewBuffer(reqBody))
		if err != nil {
			// Handle error, perhaps log and retry or notify an admin
			return
		}
		defer resp.Body.Close()
		//dobija OK od webShop-a
	}()
}
