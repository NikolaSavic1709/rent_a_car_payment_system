package server

import (
	"context"
	"fmt"
	"os"

	"github.com/plutov/paypal/v4"
)

// PayPalClient interface for PayPal operations
type PayPalClient interface {
	CreateOrder(amount float64, currency string, description string) (*paypal.Order, error)
	CaptureOrder(orderID string) (*paypal.CaptureOrderResponse, error)
	GetOrder(orderID string) (*paypal.Order, error)
}

type paypalClient struct {
	client     *paypal.Client
	successURL string
	cancelURL  string
}

// NewPayPalClient creates a new PayPal client
func NewPayPalClient() (PayPalClient, error) {
	clientID := os.Getenv("PAYPAL_CLIENT_ID")
	secret := os.Getenv("PAYPAL_SECRET")
	mode := os.Getenv("PAYPAL_MODE") // "sandbox" or "live"

	if clientID == "" || secret == "" {
		return nil, fmt.Errorf("PayPal credentials not configured")
	}

	// Default to sandbox if not specified
	if mode == "" {
		mode = "sandbox"
	}

	// Create PayPal client
	var client *paypal.Client
	var err error

	if mode == "live" {
		client, err = paypal.NewClient(clientID, secret, paypal.APIBaseLive)
	} else {
		client, err = paypal.NewClient(clientID, secret, paypal.APIBaseSandBox)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create PayPal client: %w", err)
	}

	// Get access token
	_, err = client.GetAccessToken(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get PayPal access token: %w", err)
	}

	successURL := os.Getenv("SUCCESS_URL")
	if successURL == "" {
		successURL = "http://localhost:3001/paypal?status=success"
	}

	cancelURL := os.Getenv("CANCEL_URL")
	if cancelURL == "" {
		cancelURL = "http://localhost:3001/paypal?status=cancel"
	}

	return &paypalClient{
		client:     client,
		successURL: successURL,
		cancelURL:  cancelURL,
	}, nil
}

// CreateOrder creates a new PayPal order
func (p *paypalClient) CreateOrder(amount float64, currency string, description string) (*paypal.Order, error) {
	// Default to USD if currency not specified
	if currency == "" {
		currency = "USD"
	}

	if description == "" {
		description = "Rent-a-Car Payment"
	}

	// Determine return URL based on current host
	returnURL := os.Getenv("PAYPAL_RETURN_URL")
	if returnURL == "" {
		returnURL = "http://localhost:8088/payment-success"
	}

	cancelReturnURL := os.Getenv("PAYPAL_CANCEL_URL")
	if cancelReturnURL == "" {
		cancelReturnURL = "http://localhost:8088/payment-cancel"
	}

	// Create purchase units
	purchaseUnits := []paypal.PurchaseUnitRequest{
		{
			Amount: &paypal.PurchaseUnitAmount{
				Currency: currency,
				Value:    fmt.Sprintf("%.2f", amount),
			},
			Description: description,
		},
	}

	// Create order request
	order, err := p.client.CreateOrder(
		context.Background(),
		paypal.OrderIntentCapture,
		purchaseUnits,
		nil,
		&paypal.ApplicationContext{
			ReturnURL: returnURL,
			CancelURL: cancelReturnURL,
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create PayPal order: %w", err)
	}

	return order, nil
}

// CaptureOrder captures an approved PayPal order
func (p *paypalClient) CaptureOrder(orderID string) (*paypal.CaptureOrderResponse, error) {
	capture, err := p.client.CaptureOrder(
		context.Background(),
		orderID,
		paypal.CaptureOrderRequest{},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to capture PayPal order: %w", err)
	}

	return capture, nil
}

// GetOrder retrieves order details
func (p *paypalClient) GetOrder(orderID string) (*paypal.Order, error) {
	order, err := p.client.GetOrder(context.Background(), orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get PayPal order: %w", err)
	}

	return order, nil
}
