package database

import (
	"context"
	"crypto/sha512"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

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
	CheckMerchant(merchantId uint, password string) (*Merchant, error)
	GetMerchantRedirectURL(merchantId uint, status TransactionStatus) (string, error)
	WriteTransaction(transaction Transaction) error
	GetTransactionByMerchantOrderId(merchantOrderId uuid.UUID) (PaymentRequest, error)
	GetTransactionByQRRef(qrRef uint64) (PaymentRequest, error)
	ChangeTransactionStatus(transactionId uuid.UUID, status TransactionStatus) (uint, error)
	DeletePreviousSubscription(merchantId uint) error
	SaveSubscription(merchantId uint, method uint) error
	GetSubscriptionsForMerchant(merchantId uint) ([]int, error)
}

type service struct {
	db *sql.DB
}

func (s *service) CheckMerchant(merchantId uint, password string) (*Merchant, error) {
	query := `SELECT merchant_id, password, salt FROM merchants WHERE merchant_id = $1`
	row := s.db.QueryRow(query, merchantId)

	var merchant Merchant
	err := row.Scan(&merchant.MerchantId, &merchant.Password, &merchant.Salt)
	fmt.Println(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	hashedPassword := sha512.Sum512([]byte(password + merchant.Salt))
	hashedPasswordHex := hex.EncodeToString(hashedPassword[:])
	fmt.Println("AAAA " + hashedPasswordHex)
	fmt.Println("AAAA " + merchant.Password)
	if merchant.Password != hashedPasswordHex {
		return nil, nil
	}

	return &merchant, nil
}
func (s *service) GetMerchantRedirectURL(merchantId uint, status TransactionStatus) (string, error) {
	var urlField string
	var redirectURL string

	if status == Error {
		urlField = "error_url"
	} else if status == Failed {
		urlField = "fail_url"
	} else if status == Successful {
		urlField = "success_url"
	} else {
		return "", fmt.Errorf("invalid status: %v", status)
	}

	query := fmt.Sprintf(
		"SELECT %s FROM merchants WHERE merchant_id = $1",
		urlField,
	)

	err := s.db.QueryRow(query, merchantId).Scan(&redirectURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("failed to fetch merchant redirect url: %w", err)
	}

	return redirectURL, nil
}


func (s *service) WriteTransaction(transaction Transaction) error {

	query := `INSERT INTO transactions (transaction_id, merchant_id, merchant_order_id, status, timestamp, merchant_timestamp, amount, currency, qr_ref) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	// Use the database connection to execute the query.
	_, err := s.db.Exec(query, transaction.TransactionId, transaction.MerchantId, transaction.MerchantOrderId, transaction.Status, transaction.Timestamp, transaction.MerchantTimestamp, transaction.Amount, transaction.Currency, transaction.QRRef)
	// Handle any errors from the database operation.
	if err != nil {
		return err
	}
	return nil
}
func (s *service) GetTransactionByMerchantOrderId(merchantOrderId uuid.UUID) (PaymentRequest, error) {
	query := `SELECT currency, amount, merchant_id, timestamp, transaction_id FROM transactions WHERE merchant_order_id = $1`
	row := s.db.QueryRow(query, merchantOrderId.String())
	fmt.Println(row)
	var paymentRequest PaymentRequest
	err := row.Scan(&paymentRequest.Currency, &paymentRequest.Amount, &paymentRequest.MerchantId, &paymentRequest.Timestamp, &paymentRequest.TransactionId)

	if err != nil {
		fmt.Println(err)
		return paymentRequest, err
	}
	paymentRequest.MerchantOrderId = merchantOrderId

	return paymentRequest, nil
}
func (s *service) GetTransactionByQRRef(qrRef uint64) (PaymentRequest, error) {
	query := `SELECT currency, amount, merchant_id, timestamp, transaction_id, merchant_order_id FROM transactions WHERE qr_ref = $1`
	row := s.db.QueryRow(query, qrRef)
	fmt.Println(row)
	var paymentRequest PaymentRequest
	err := row.Scan(&paymentRequest.Currency, &paymentRequest.Amount, &paymentRequest.MerchantId, &paymentRequest.Timestamp, &paymentRequest.TransactionId, &paymentRequest.MerchantOrderId)
	if err != nil {
		fmt.Println(err)
		return paymentRequest, err
	}
	return paymentRequest, nil
}

func (s *service) ChangeTransactionStatus(transactionId uuid.UUID, status TransactionStatus) (uint, error) {
	var merchantID uint

	// Determine the field to fetch based on the status
	if status != Error && status != Failed && status != Successful {
		return 0, fmt.Errorf("invalid status: %v", status)
	}

	// Fetch merchant_id
	query := `SELECT merchant_id FROM transactions WHERE transaction_id = $1`
	err := s.db.QueryRow(query, transactionId).Scan(&merchantID)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch merchant_id for transaction: %v", err)
	}

	// Update the transaction's status
	updateQuery := `UPDATE transactions SET status = $1 WHERE transaction_id = $2`
	_, err = s.db.Exec(updateQuery, status, transactionId)
	if err != nil {
		return 0, fmt.Errorf("failed to update transaction status: %v", err)
	}

	return merchantID, nil
}

func (s *service) DeletePreviousSubscription(merchantId uint) error {
	query := `DELETE FROM subscriptions WHERE merchant_id = $1;`
	_, err := s.db.Exec(query, merchantId)
	if err != nil {
		return fmt.Errorf("failed to delete subscriptions: %w", err)
	}
	return nil
}

func (s *service) SaveSubscription(merchantId uint, method uint) error {
	query := `INSERT INTO subscriptions (merchant_id, method) VALUES ($1, $2);`
	_, err := s.db.Exec(query, merchantId, method)
	if err != nil {
		return fmt.Errorf("failed to add subscription: %w", err)
	}
	return nil
}

func (s *service) GetSubscriptionsForMerchant(merchantId uint) ([]int, error) {
	query := `SELECT method FROM subscriptions WHERE merchant_id = $1;`

	var methods []int
	rows, err := s.db.Query(query, merchantId)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch subscriptions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var method int
		if err := rows.Scan(&method); err != nil {
			return nil, fmt.Errorf("failed to scan subscription method: %w", err)
		}
		methods = append(methods, method)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating through subscription rows: %w", err)
	}

	return methods, nil
}

var (
	database   = os.Getenv("DB_DATABASE")
	password   = os.Getenv("DB_PASSWORD")
	username   = os.Getenv("DB_USERNAME")
	port       = os.Getenv("DB_HOST_PORT")
	host       = os.Getenv("DB_HOST")
	schema     = os.Getenv("DB_SCHEMA")
	dbInstance *service
)

func New() Service {
	// Reuse Connection
	if dbInstance != nil {
		return dbInstance
	}
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&search_path=%s", username, password, host, port, database, schema)
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
		log.Fatalf("db down: %v", err) // Log the error and terminate the program
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
	if dbStats.OpenConnections > 40 { // Assuming 50 is the max for this example
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

func Connect() {
	var s = fmt.Sprintf("postgres://%s:%s@%s:%s/%s", username, password, host, port, database)
	fmt.Println(s)
	db, err := gorm.Open(postgres.Open(s), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	err1 := db.AutoMigrate(&Transaction{})
	err2 := db.AutoMigrate(&Merchant{})
	err3 := db.AutoMigrate(&Subscription{})
	if err1 != nil && err2 != nil && err3 != nil {
		return
	}
	//DB = db
}
