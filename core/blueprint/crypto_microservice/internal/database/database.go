package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
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

	// Payment operations
	CreatePayment(payment *CryptoPayment) error
	GetPaymentByPaymentId(paymentId uuid.UUID) (*CryptoPayment, error)
	UpdatePayment(payment *CryptoPayment) error

	// Wallet operations
	GetOrCreateMerchantWallet(merchantId uint, currency string) (*MerchantWallet, error)
	GetWallet(merchantId uint, currency string) (*MerchantWallet, error)
}

type service struct {
	db *sql.DB
}

var (
	database   = os.Getenv("DB_DATABASE")
	password   = os.Getenv("DB_PASSWORD")
	username   = os.Getenv("DB_USERNAME")
	port       = os.Getenv("DB_PORT")
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

// CreatePayment inserts a new crypto payment into the database
func (s *service) CreatePayment(payment *CryptoPayment) error {
	query := `
		INSERT INTO crypto_payments (
			payment_id, transaction_id, merchant_order_id, merchant_id,
			amount, currency, status, destination_address,
			required_confirmations, created_at, expiry_time, is_testnet
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id
	`

	err := s.db.QueryRow(
		query,
		payment.PaymentId, payment.TransactionId, payment.MerchantOrderId,
		payment.MerchantId, payment.Amount, payment.Currency, payment.Status,
		payment.DestinationAddress, payment.RequiredConfirmations,
		payment.CreatedAt, payment.ExpiryTime, payment.IsTestnet,
	).Scan(&payment.ID)

	return err
}

// GetPaymentByPaymentId retrieves a payment by its payment ID
func (s *service) GetPaymentByPaymentId(paymentId uuid.UUID) (*CryptoPayment, error) {
	query := `
		SELECT 
			id, payment_id, transaction_id, merchant_order_id, merchant_id,
			amount, currency, status, destination_address, source_address,
			tx_hash, block_height, confirmations, required_confirmations,
			created_at, expiry_time, confirmed_at, is_testnet
		FROM crypto_payments
		WHERE payment_id = $1
	`

	var payment CryptoPayment
	var confirmedAt sql.NullTime
	var sourceAddr, txHash sql.NullString
	var blockHeight sql.NullInt64

	err := s.db.QueryRow(query, paymentId).Scan(
		&payment.ID, &payment.PaymentId, &payment.TransactionId,
		&payment.MerchantOrderId, &payment.MerchantId, &payment.Amount,
		&payment.Currency, &payment.Status, &payment.DestinationAddress,
		&sourceAddr, &txHash, &blockHeight, &payment.Confirmations,
		&payment.RequiredConfirmations, &payment.CreatedAt,
		&payment.ExpiryTime, &confirmedAt, &payment.IsTestnet,
	)

	if err != nil {
		return nil, err
	}

	if sourceAddr.Valid {
		payment.SourceAddress = sourceAddr.String
	}
	if txHash.Valid {
		payment.TxHash = txHash.String
	}
	if blockHeight.Valid {
		payment.BlockHeight = blockHeight.Int64
	}
	if confirmedAt.Valid {
		payment.ConfirmedAt = &confirmedAt.Time
	}

	return &payment, nil
}

// UpdatePayment updates an existing payment
func (s *service) UpdatePayment(payment *CryptoPayment) error {
	query := `
		UPDATE crypto_payments
		SET status = $1, source_address = $2, tx_hash = $3,
			block_height = $4, confirmations = $5, confirmed_at = $6
		WHERE payment_id = $7
	`

	_, err := s.db.Exec(
		query,
		payment.Status, payment.SourceAddress, payment.TxHash,
		payment.BlockHeight, payment.Confirmations, payment.ConfirmedAt,
		payment.PaymentId,
	)

	return err
}

// GetOrCreateMerchantWallet gets or creates a merchant wallet for a specific currency
func (s *service) GetOrCreateMerchantWallet(merchantId uint, currency string) (*MerchantWallet, error) {
	// Try to get existing wallet
	query := `
		SELECT id, merchant_id, currency, wallet_address, public_key,
			balance, is_testnet, created_at, updated_at
		FROM merchant_wallets
		WHERE merchant_id = $1 AND currency = $2
	`

	var wallet MerchantWallet
	err := s.db.QueryRow(query, merchantId, currency).Scan(
		&wallet.ID, &wallet.MerchantId, &wallet.Currency,
		&wallet.WalletAddress, &wallet.PublicKey, &wallet.Balance,
		&wallet.IsTestnet, &wallet.CreatedAt, &wallet.UpdatedAt,
	)

	if err == nil {
		return &wallet, nil
	}

	if err != sql.ErrNoRows {
		return nil, err
	}

	// Create new wallet
	walletAddress, publicKey := generateTestnetWallet(currency, merchantId)

	insertQuery := `
		INSERT INTO merchant_wallets (
			merchant_id, currency, wallet_address, public_key,
			balance, is_testnet, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	now := time.Now()
	err = s.db.QueryRow(
		insertQuery,
		merchantId, currency, walletAddress, publicKey,
		0.0, true, now, now,
	).Scan(&wallet.ID)

	if err != nil {
		return nil, err
	}

	wallet.MerchantId = merchantId
	wallet.Currency = currency
	wallet.WalletAddress = walletAddress
	wallet.PublicKey = publicKey
	wallet.Balance = 0.0
	wallet.IsTestnet = true
	wallet.CreatedAt = now
	wallet.UpdatedAt = now

	return &wallet, nil
}

// generateTestnetWallet generates a testnet wallet address for the given currency
func generateTestnetWallet(currency string, merchantId uint) (address string, publicKey string) {
	// For testnet simulation, generate deterministic addresses
	switch currency {
	case "BTC":
		address = fmt.Sprintf("tb1q%s%d", uuid.New().String()[:20], merchantId)
		publicKey = fmt.Sprintf("02%s", uuid.New().String()[:62])
	case "ETH", "USDT":
		address = fmt.Sprintf("0x%s%d", uuid.New().String()[:38], merchantId)
		publicKey = fmt.Sprintf("04%s", uuid.New().String()[:126])
	default:
		address = fmt.Sprintf("test_%s_%d", currency, merchantId)
		publicKey = uuid.New().String()
	}
	return
}

// GetWallet retrieves a merchant wallet
func (s *service) GetWallet(merchantId uint, currency string) (*MerchantWallet, error) {
	query := `
		SELECT id, merchant_id, currency, wallet_address, public_key,
			balance, is_testnet, created_at, updated_at
		FROM merchant_wallets
		WHERE merchant_id = $1 AND currency = $2
	`

	var wallet MerchantWallet
	err := s.db.QueryRow(query, merchantId, currency).Scan(
		&wallet.ID, &wallet.MerchantId, &wallet.Currency,
		&wallet.WalletAddress, &wallet.PublicKey, &wallet.Balance,
		&wallet.IsTestnet, &wallet.CreatedAt, &wallet.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &wallet, nil
}
