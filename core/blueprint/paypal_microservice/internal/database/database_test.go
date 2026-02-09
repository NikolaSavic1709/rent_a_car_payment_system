package database

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(t *testing.T) (Service, func()) {
	ctx := context.Background()

	// Start PostgreSQL container
	pgContainer, err := postgres.Run(
		ctx,
		"postgres:latest",
		postgres.WithDatabase("test_paypal_db"),
		postgres.WithUsername("test_user"),
		postgres.WithPassword("test_password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}

	// Get connection string
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to get connection string: %v", err)
	}

	// Set environment variables
	t.Setenv("DB_CONNECTION", connStr)

	// Initialize database service
	db := New()

	cleanup := func() {
		db.Close()
		pgContainer.Terminate(ctx)
	}

	return db, cleanup
}

func TestCreatePayment(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	payment := &PayPalPayment{
		PaymentID:       uuid.New(),
		TransactionID:   uuid.New(),
		MerchantOrderID: uuid.New(),
		MerchantID:      12345,
		PayPalOrderID:   "PAYPAL-ORDER-123",
		Amount:          100.00,
		Currency:        "USD",
		Status:          Pending,
	}

	err := db.CreatePayment(payment)
	if err != nil {
		t.Fatalf("Failed to create payment: %v", err)
	}

	if payment.ID == 0 {
		t.Error("Payment ID should be set after creation")
	}
}

func TestGetPaymentByID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create a payment first
	paymentID := uuid.New()
	payment := &PayPalPayment{
		PaymentID:       paymentID,
		TransactionID:   uuid.New(),
		MerchantOrderID: uuid.New(),
		MerchantID:      12345,
		PayPalOrderID:   "PAYPAL-ORDER-456",
		Amount:          200.00,
		Currency:        "EUR",
		Status:          Pending,
	}

	err := db.CreatePayment(payment)
	if err != nil {
		t.Fatalf("Failed to create payment: %v", err)
	}

	// Retrieve the payment
	retrieved, err := db.GetPaymentByID(paymentID.String())
	if err != nil {
		t.Fatalf("Failed to get payment: %v", err)
	}

	if retrieved.PaymentID != paymentID {
		t.Errorf("Expected payment ID %s, got %s", paymentID, retrieved.PaymentID)
	}

	if retrieved.Amount != 200.00 {
		t.Errorf("Expected amount 200.00, got %.2f", retrieved.Amount)
	}
}

func TestUpdatePaymentStatus(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create a payment
	paymentID := uuid.New()
	payment := &PayPalPayment{
		PaymentID:       paymentID,
		TransactionID:   uuid.New(),
		MerchantOrderID: uuid.New(),
		MerchantID:      12345,
		PayPalOrderID:   "PAYPAL-ORDER-789",
		Amount:          300.00,
		Currency:        "USD",
		Status:          Pending,
	}

	err := db.CreatePayment(payment)
	if err != nil {
		t.Fatalf("Failed to create payment: %v", err)
	}

	// Update status to Completed
	err = db.UpdatePaymentStatus(paymentID.String(), Completed, "CAPTURE-123")
	if err != nil {
		t.Fatalf("Failed to update payment status: %v", err)
	}

	// Verify update
	retrieved, err := db.GetPaymentByID(paymentID.String())
	if err != nil {
		t.Fatalf("Failed to get payment: %v", err)
	}

	if retrieved.Status != Completed {
		t.Errorf("Expected status Completed, got %s", retrieved.Status.String())
	}
}
