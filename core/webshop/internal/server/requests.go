package server

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (s *Server) sendPSPRequest(data map[string]interface{}) (map[string]interface{}, error) {
	pspURL := "https://nginx/payment" //  PSP URL (HTTPS for nginx)
	fmt.Println("Sending request to PSP...")
	// Create JSON payload
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %v", err)
	}
	fmt.Println("Debug: to be sent to PSP")

	// Create HTTP client with TLS config to skip verification (for self-signed certs)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	// Send HTTP request
	resp, err := client.Post(pspURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to send request to PSP: %v", err)
	}
	defer resp.Body.Close()
	fmt.Println("Received response from PSP")
	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read PSP response: %v", err)
	}

	fmt.Println("Response body:", string(body))
	// Parse response as JSON
	var pspResponse map[string]interface{}
	if err := json.Unmarshal(body, &pspResponse); err != nil {
		return nil, fmt.Errorf("failed to parse PSP response: %v", err)
	}
	fmt.Println("Parsed PSP response:", pspResponse)
	return pspResponse, nil
}

func (s *Server) getSubscriptionUrlFromPSP(data map[string]interface{}) (map[string]interface{}, error) {
	pspURL := "https://nginx/subscription/url"

	payload, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %v", err)
	}

	// Create HTTP client with TLS config
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Post(pspURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to send request to PSP: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read PSP response: %v", err)
	}

	var pspResponse map[string]interface{}
	if err := json.Unmarshal(body, &pspResponse); err != nil {
		return nil, fmt.Errorf("failed to parse PSP response: %v", err)
	}

	return pspResponse, nil
}
