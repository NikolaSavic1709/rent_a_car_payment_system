package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// Transaction represents a blockchain transaction
type Transaction struct {
	TxHash      string
	FromAddress string
	ToAddress   string
	Amount      float64
	Currency    string
	Timestamp   time.Time
	Signature   string
	Status      string
}

// TransactionBuilder helps build blockchain transactions
type TransactionBuilder struct {
	transaction *Transaction
}

// NewTransactionBuilder creates a new transaction builder
func NewTransactionBuilder() *TransactionBuilder {
	return &TransactionBuilder{
		transaction: &Transaction{
			Timestamp: time.Now(),
			Status:    "pending",
		},
	}
}

// SetFromAddress sets the sender address
func (tb *TransactionBuilder) SetFromAddress(address string) *TransactionBuilder {
	tb.transaction.FromAddress = address
	return tb
}

// SetToAddress sets the recipient address
func (tb *TransactionBuilder) SetToAddress(address string) *TransactionBuilder {
	tb.transaction.ToAddress = address
	return tb
}

// SetAmount sets the transaction amount
func (tb *TransactionBuilder) SetAmount(amount float64) *TransactionBuilder {
	tb.transaction.Amount = amount
	return tb
}

// SetCurrency sets the currency
func (tb *TransactionBuilder) SetCurrency(currency string) *TransactionBuilder {
	tb.transaction.Currency = currency
	return tb
}

// Build finalizes and returns the transaction
func (tb *TransactionBuilder) Build() (*Transaction, error) {
	if tb.transaction.FromAddress == "" {
		return nil, fmt.Errorf("from address is required")
	}
	if tb.transaction.ToAddress == "" {
		return nil, fmt.Errorf("to address is required")
	}
	if tb.transaction.Amount <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}
	if tb.transaction.Currency == "" {
		return nil, fmt.Errorf("currency is required")
	}

	// Generate transaction hash
	tb.transaction.TxHash = tb.generateTxHash()

	return tb.transaction, nil
}

// generateTxHash generates a unique transaction hash
func (tb *TransactionBuilder) generateTxHash() string {
	data := fmt.Sprintf("%s:%s:%.8f:%s:%d",
		tb.transaction.FromAddress,
		tb.transaction.ToAddress,
		tb.transaction.Amount,
		tb.transaction.Currency,
		tb.transaction.Timestamp.Unix(),
	)

	hash := sha256.Sum256([]byte(data))
	return "0x" + hex.EncodeToString(hash[:])
}

// Sign simulates signing a transaction
func (tb *TransactionBuilder) Sign(privateKey string) error {
	// In a real implementation, this would use proper cryptographic signing
	// For testnet simulation, we just create a mock signature
	signature := sha256.Sum256([]byte(tb.transaction.TxHash + privateKey))
	tb.transaction.Signature = hex.EncodeToString(signature[:])
	return nil
}

// VerifyTransaction verifies a transaction's signature and validity
func VerifyTransaction(tx *Transaction) error {
	if tx.TxHash == "" {
		return fmt.Errorf("transaction hash is empty")
	}
	if tx.FromAddress == "" || tx.ToAddress == "" {
		return fmt.Errorf("addresses cannot be empty")
	}
	if tx.Amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}
	if tx.Signature == "" {
		return fmt.Errorf("transaction is not signed")
	}

	return nil
}

// BroadcastTransaction simulates broadcasting a transaction to the network
func BroadcastTransaction(tx *Transaction) (string, error) {
	if err := VerifyTransaction(tx); err != nil {
		return "", fmt.Errorf("invalid transaction: %w", err)
	}

	// In real implementation, this would broadcast to actual blockchain network
	// For simulation, we just return the transaction hash
	tx.Status = "broadcasted"

	return tx.TxHash, nil
}
