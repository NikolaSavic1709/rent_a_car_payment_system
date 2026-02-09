package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
)

// Service represents a service that interacts with a database.
type Service interface {
	// Health returns a map of health status information.
	// The keys and values in the map are service-specific.
	Health() map[string]string

	// Close terminates the database connection.
	// It returns an error if the connection cannot be closed.
	Close() error

	// PayPal payment operations
	CreatePayment(payment *PayPalPayment) error
	GetPaymentByID(paymentID string) (*PayPalPayment, error)
	GetPaymentByOrderID(orderID string) (*PayPalPayment, error)
	UpdatePaymentStatus(paymentID string, status PaymentStatus, captureID string) error
	UpdatePaymentWithPayer(paymentID, payerEmail, payerID, payerName string) error
}

type service struct {
	db *sql.DB
}

var (
	database   = os.Getenv("DB_NAME")
	password   = os.Getenv("DB_PASSWORD")
	username   = os.Getenv("DB_USER")
	port       = os.Getenv("DB_PORT")
	host       = os.Getenv("DB_HOST")
	dbInstance *service
)

func New() Service {
	// Reuse Connection
	if dbInstance != nil {
		return dbInstance
	}

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", username, password, host, port, database)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatal(err)
	}

	dbInstance = &service{
		db: db,
	}
	return dbInstance
}

// Health checks the health of the database connection by pinging the database.
// It returns a map with keys indicating various health statistics.
func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	// Ping the database
	err := s.db.PingContext(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		log.Fatalf("db down: %v", err)
		return stats
	}

	// Database is up, add more statistics
	stats["status"] = "up"
	stats["message"] = "It's healthy"

	// Get database stats (like open connections, in use, idle, etc.)
	dbStats := s.db.Stats()
	stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
	stats["in_use"] = strconv.Itoa(dbStats.InUse)
	stats["idle"] = strconv.Itoa(dbStats.Idle)
	stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
	stats["wait_duration"] = dbStats.WaitDuration.String()
	stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
	stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeClosed, 10)

	// Evaluate stats to provide a health message
	if dbStats.OpenConnections > 40 {
		stats["message"] = "The database is experiencing heavy load."
	}

	if dbStats.WaitCount > 1000 {
		stats["message"] = "The database has a high number of wait events, indicating potential bottlenecks."
	}

	if dbStats.MaxIdleClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many idle connections are being closed, consider revising the connection pool settings."
	}

	if dbStats.MaxLifetimeClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern."
	}

	return stats
}

// Close closes the database connection.
// It logs a message indicating the disconnection from the specific database.
// If the connection is successfully closed, it returns nil.
// If an error occurs while closing the connection, it returns the error.
func (s *service) Close() error {
	log.Printf("Disconnected from database: %s", database)
	return s.db.Close()
}

// CreatePayment creates a new PayPal payment in the database
func (s *service) CreatePayment(payment *PayPalPayment) error {
	query := `
		INSERT INTO paypal_payments (
			payment_id, transaction_id, merchant_order_id, merchant_id,
			paypal_order_id, amount, currency, status, approval_url, description
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at
	`

	err := s.db.QueryRow(
		query,
		payment.PaymentID,
		payment.TransactionID,
		payment.MerchantOrderID,
		payment.MerchantID,
		payment.PayPalOrderID,
		payment.Amount,
		payment.Currency,
		payment.Status,
		payment.ApprovalURL,
		payment.Description,
	).Scan(&payment.ID, &payment.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create payment: %w", err)
	}

	return nil
}

// GetPaymentByID retrieves a payment by its payment ID
func (s *service) GetPaymentByID(paymentID string) (*PayPalPayment, error) {
	query := `
		SELECT id, payment_id, transaction_id, merchant_order_id, merchant_id,
			   paypal_order_id, paypal_capture_id, amount, currency, status,
			   payer_email, payer_id, payer_name, created_at, approved_at,
			   completed_at, cancelled_at, approval_url, description, failure_reason
		FROM paypal_payments
		WHERE payment_id = $1
	`

	payment := &PayPalPayment{}
	err := s.db.QueryRow(query, paymentID).Scan(
		&payment.ID,
		&payment.PaymentID,
		&payment.TransactionID,
		&payment.MerchantOrderID,
		&payment.MerchantID,
		&payment.PayPalOrderID,
		&payment.PayPalCaptureID,
		&payment.Amount,
		&payment.Currency,
		&payment.Status,
		&payment.PayerEmail,
		&payment.PayerID,
		&payment.PayerName,
		&payment.CreatedAt,
		&payment.ApprovedAt,
		&payment.CompletedAt,
		&payment.CancelledAt,
		&payment.ApprovalURL,
		&payment.Description,
		&payment.FailureReason,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("payment not found")
		}
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	return payment, nil
}

// GetPaymentByOrderID retrieves a payment by PayPal order ID
func (s *service) GetPaymentByOrderID(orderID string) (*PayPalPayment, error) {
	query := `
		SELECT id, payment_id, transaction_id, merchant_order_id, merchant_id,
			   paypal_order_id, paypal_capture_id, amount, currency, status,
			   payer_email, payer_id, payer_name, created_at, approved_at,
			   completed_at, cancelled_at, approval_url, description, failure_reason
		FROM paypal_payments
		WHERE paypal_order_id = $1
	`

	payment := &PayPalPayment{}
	err := s.db.QueryRow(query, orderID).Scan(
		&payment.ID,
		&payment.PaymentID,
		&payment.TransactionID,
		&payment.MerchantOrderID,
		&payment.MerchantID,
		&payment.PayPalOrderID,
		&payment.PayPalCaptureID,
		&payment.Amount,
		&payment.Currency,
		&payment.Status,
		&payment.PayerEmail,
		&payment.PayerID,
		&payment.PayerName,
		&payment.CreatedAt,
		&payment.ApprovedAt,
		&payment.CompletedAt,
		&payment.CancelledAt,
		&payment.ApprovalURL,
		&payment.Description,
		&payment.FailureReason,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("payment not found")
		}
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	return payment, nil
}

// UpdatePaymentStatus updates the payment status
func (s *service) UpdatePaymentStatus(paymentID string, status PaymentStatus, captureID string) error {
	var query string
	var err error

	switch status {
	case Approved:
		query = `
			UPDATE paypal_payments
			SET status = $1, approved_at = NOW()
			WHERE payment_id = $2
		`
		_, err = s.db.Exec(query, status, paymentID)
	case Completed:
		query = `
			UPDATE paypal_payments
			SET status = $1, paypal_capture_id = $2, completed_at = NOW()
			WHERE payment_id = $3
		`
		_, err = s.db.Exec(query, status, captureID, paymentID)
	case Cancelled, Failed:
		timestampField := "cancelled_at"
		if status == Failed {
			timestampField = "completed_at"
		}
		query = fmt.Sprintf(`
			UPDATE paypal_payments
			SET status = $1, %s = NOW()
			WHERE payment_id = $2
		`, timestampField)
		_, err = s.db.Exec(query, status, paymentID)
	default:
		query = `
			UPDATE paypal_payments
			SET status = $1
			WHERE payment_id = $2
		`
		_, err = s.db.Exec(query, status, paymentID)
	}

	if err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	return nil
}

// UpdatePaymentWithPayer updates payment with payer information
func (s *service) UpdatePaymentWithPayer(paymentID, payerEmail, payerID, payerName string) error {
	query := `
		UPDATE paypal_payments
		SET payer_email = $1, payer_id = $2, payer_name = $3
		WHERE payment_id = $4
	`

	_, err := s.db.Exec(query, payerEmail, payerID, payerName, paymentID)
	if err != nil {
		return fmt.Errorf("failed to update payer info: %w", err)
	}

	return nil
}
