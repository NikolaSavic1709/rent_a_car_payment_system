package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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

	WriteTransaction(transaction Transaction) error

	Pay(acquirerOrderId uuid.UUID, currency string, amount float32, cardNumber string, expiryDate time.Time, merchantId uint) (TransactionStatus, error)
}

type service struct {
	db *sql.DB
}

func (s *service) WriteTransaction(transaction Transaction) error {

	query := `INSERT INTO transactions (transaction_id, acquirer_order_id, acquirer_timestamp, merchant_id, merchant_order_id, status, amount, currency, timestamp, partial_card_number) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	// Use the database connection to execute the query.
	_, err := s.db.Exec(query, transaction.TransactionId, transaction.AcquirerOrderId, transaction.AcquirerTimestamp, transaction.MerchantId, transaction.MerchantOrderId, transaction.Status, transaction.Amount, transaction.Currency, transaction.Timestamp, transaction.PartialCardNumber)

	// Handle any errors from the database operation.
	if err != nil {
		return err
	}

	return nil
}

func (s *service) Pay(acquirerOrderId uuid.UUID, currency string, amount float32, cardNumber string, expiryDate time.Time, merchantId uint) (TransactionStatus, error) {

	updateTransactionStatus := func(status TransactionStatus) {
		queryUpdateStatus := `UPDATE transactions SET status = $1 WHERE acquirer_order_id = $2`
		_, err := s.db.Exec(queryUpdateStatus, status, acquirerOrderId)
		if err != nil {
			fmt.Printf("failed to update transaction status: %v\n", err)
		}
	}

	if !isValidCardNumber(cardNumber) {
		updateTransactionStatus(Failed)
		return Failed, fmt.Errorf("invalid card number")
	}

	queryCard := `SELECT id, bank_account_id, encrypted_pan, expiry_date, card_type, is_tokenized 
	              FROM cards`

	rows, err := s.db.Query(queryCard)
	if err != nil {
		updateTransactionStatus(Error)
		return Error, fmt.Errorf("failed to fetch cards: %w", err)
	}
	defer rows.Close()
	encrypted_pan, err := Encrypt(cardNumber)
	fmt.Printf("Encrypted PAN: %s\n", encrypted_pan)
	var card Card
	var foundCard bool
	for rows.Next() {
		err := rows.Scan(
			&card.ID,
			&card.BankAccountID,
			&card.EncryptedPAN,
			&card.ExpiryDate,
			&card.CardType,
			&card.IsTokenized,
		)
		if err != nil {
			continue
		}

		// Decrypt the PAN
		decryptedPAN, err := Decrypt(card.EncryptedPAN)
		if err != nil {
			fmt.Printf("Failed to decrypt PAN for card pan %d: %v\n", card.EncryptedPAN, err)
			continue
		}
		fmt.Printf("Decrypted PAN: %s\n", decryptedPAN)
		// Compare with input card number
		if decryptedPAN == cardNumber {
			fmt.Printf("Card found: ID=%d, BankAccountID=%d\n", card.ID, card.BankAccountID)
			foundCard = true
			break
		}
	}

	if !foundCard {
		fmt.Println("Card not found")
		updateTransactionStatus(Error)
		return Error, fmt.Errorf("failed to fetch card: card not found")
	}

	if card.ExpiryDate.Year() != expiryDate.Year() || card.ExpiryDate.Month() != expiryDate.Month() {
		fmt.Printf("Card expired: card expiry=%s, provided expiry=%s\n", card.ExpiryDate.Format("01/2006"), expiryDate.Format("01/2006"))
		updateTransactionStatus(Failed)
		return Failed, nil
	}

	queryBankAccount := `SELECT id, balance, currency FROM bank_accounts WHERE id = $1`

	var bankAccount struct {
		ID       uint
		Balance  float32
		Currency string
	}
	err = s.db.QueryRow(queryBankAccount, card.BankAccountID).Scan(&bankAccount.ID, &bankAccount.Balance, &bankAccount.Currency)
	if err != nil {
		fmt.Println("Failed to fetch bank account: ", err)
		updateTransactionStatus(Error)
		return Error, fmt.Errorf("failed to fetch bank account: %w", err)
	}

	if bankAccount.Balance < amount || bankAccount.Currency != currency {
		updateTransactionStatus(Failed)
		fmt.Printf("Insufficient funds or currency mismatch: balance=%f, required=%f, account currency=%s, required currency=%s\n", bankAccount.Balance, amount, bankAccount.Currency, currency)
		return Failed, nil
	}
	var merchantBankAccountID uint
	merchantQuery := `SELECT bank_account_id FROM merchants WHERE merchant_id = $1`
	err = s.db.QueryRow(merchantQuery, merchantId).Scan(&merchantBankAccountID)
	if err != nil {
		fmt.Println("Failed to fetch merchant bank account: ", err)
		updateTransactionStatus(Failed)
		return Failed, fmt.Errorf("fail, merchant does not exist: %w", err)
	}

	bankAccountQuery := `SELECT 1 FROM bank_accounts WHERE id = $1`
	err = s.db.QueryRow(bankAccountQuery, merchantBankAccountID).Scan(new(int)) // Scan into a dummy variable
	if err != nil {
		fmt.Println("Failed to fetch merchant bank account: ", err)
		updateTransactionStatus(Error)
		if errors.Is(err, sql.ErrNoRows) {
			return Error, fmt.Errorf("fail, bank account does not exist")
		}
		return Error, fmt.Errorf("fail, error checking bank account existence: %w", err)
	}

	tx, err := s.db.Begin() // Start a transaction
	if err != nil {
		fmt.Println("Failed to start transaction: ", err)
		return Error, fmt.Errorf("failed to start transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback() // Rollback in case of panic
			panic(p)
		} else if err != nil {
			tx.Rollback() // Rollback in case of error
		} else {
			err = tx.Commit() // Commit if all is well
		}
	}()
	var currentBalance float32
	err = tx.QueryRow(`SELECT balance FROM bank_accounts WHERE id = $1 FOR UPDATE`, bankAccount.ID).Scan(&currentBalance)
	if err != nil {
		fmt.Println("Failed to fetch balance: ", err)
		updateTransactionStatus(Error)
		return Error, fmt.Errorf("failed to fetch balance: %w", err)
	}

	if currentBalance < amount {
		fmt.Printf("Insufficient funds: balance=%f, required=%f\n", currentBalance, amount)
		updateTransactionStatus(Failed)
		return Failed, fmt.Errorf("insufficient funds: balance=%f, required=%f", currentBalance, amount)
	}	
	updateBalanceQuery1 := `UPDATE bank_accounts SET balance = balance - $1 WHERE id = $2`
	_, err = tx.Exec(updateBalanceQuery1, amount, bankAccount.ID)
	if err != nil {
		fmt.Println("Failed to update balance for account 1: ", err)
		updateTransactionStatus(Error)
		return Error, fmt.Errorf("failed to update balance for account 1: %w", err)
	}

	updateBalanceQuery2 := `UPDATE bank_accounts SET balance = balance + $1 WHERE id = $2`
	_, err = tx.Exec(updateBalanceQuery2, amount, merchantBankAccountID)
	if err != nil {
		fmt.Println("Failed to update balance for account 2: ", err)
		updateTransactionStatus(Error)
		return Error, fmt.Errorf("failed to update balance for account 2: %w", err)
	}

	updateTransactionStatus(Successful)
	return Successful, nil
}
func isValidCardNumber(cardNumber string) bool {
	sum := 0
	nDigits := len(cardNumber)
	parity := nDigits % 2

	for i := 0; i < nDigits; i++ {
		digit := int(cardNumber[i] - '0')
		if digit < 0 || digit > 9 {
			return false // invalid character
		}

		// Double every second digit from the left
		if i%2 == parity {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
	}

	return sum%10 == 0
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
	db, err := gorm.Open(postgres.Open(fmt.Sprintf("postgres://%s:%s@%s:%s/%s", username, password, host, port, database)), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	err1 := db.AutoMigrate(&BankClient{})
	err2 := db.AutoMigrate(&BankAccount{})
	err3 := db.AutoMigrate(&Card{})
	err4 := db.AutoMigrate(&Transaction{})
	err5 := db.AutoMigrate(&Merchant{})
	if err1 != nil && err2 != nil && err3 != nil && err4 != nil && err5 != nil {
		return
	}
	//DB = db
}
