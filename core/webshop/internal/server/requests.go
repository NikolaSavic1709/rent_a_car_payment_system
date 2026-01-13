package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (s *Server) sendPSPRequest(data map[string]interface{}) (map[string]interface{}, error) {
	pspURL := "http://psp_service:8080/payment" //  PSP URL
	fmt.Println("AAAAA")
	// Create JSON payload
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %v", err)
	}
	fmt.Println("BBBB")
	// Send HTTP request
	resp, err := http.Post(pspURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to send request to PSP: %v", err)
	}
	defer resp.Body.Close()
	fmt.Println("CCCC")
	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read PSP response: %v", err)
	}

	fmt.Println("DDDD")
	// Parse response as JSON
	var pspResponse map[string]interface{}
	if err := json.Unmarshal(body, &pspResponse); err != nil {
		return nil, fmt.Errorf("failed to parse PSP response: %v", err)
	}
	fmt.Println("GGGGG")
	return pspResponse, nil
}

func (s *Server) getSubscriptionUrlFromPSP(data map[string]interface{}) (map[string]interface{}, error) {
	pspURL := "http://psp_service:8080/subscription/url"

	payload, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %v", err)
	}

	resp, err := http.Post(pspURL, "application/json", bytes.NewBuffer(payload))
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
