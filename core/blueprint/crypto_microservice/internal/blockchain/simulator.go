package blockchain

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// Simulator provides testnet blockchain simulation
type Simulator struct {
	// In-memory blockchain simulation
	transactions map[string]*SimulatedTransaction
}

type SimulatedTransaction struct {
	TxHash        string
	FromAddress   string
	ToAddress     string
	Amount        float64
	Currency      string
	Confirmations int
	BlockHeight   int64
	Status        string
}

// NewSimulator creates a new blockchain simulator
func NewSimulator() *Simulator {
	return &Simulator{
		transactions: make(map[string]*SimulatedTransaction),
	}
}

// GenerateTxHash generates a random transaction hash
func (s *Simulator) GenerateTxHash() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return "0x" + hex.EncodeToString(bytes)
}

// CreateTransaction simulates creating a blockchain transaction
func (s *Simulator) CreateTransaction(fromAddr, toAddr string, amount float64, currency string) *SimulatedTransaction {
	tx := &SimulatedTransaction{
		TxHash:        s.GenerateTxHash(),
		FromAddress:   fromAddr,
		ToAddress:     toAddr,
		Amount:        amount,
		Currency:      currency,
		Confirmations: 0,
		BlockHeight:   0,
		Status:        "pending",
	}

	s.transactions[tx.TxHash] = tx
	return tx
}

// GetTransaction retrieves a simulated transaction
func (s *Simulator) GetTransaction(txHash string) (*SimulatedTransaction, error) {
	tx, exists := s.transactions[txHash]
	if !exists {
		return nil, fmt.Errorf("transaction not found")
	}
	return tx, nil
}

// AddConfirmation adds a confirmation to a transaction
func (s *Simulator) AddConfirmation(txHash string) error {
	tx, exists := s.transactions[txHash]
	if !exists {
		return fmt.Errorf("transaction not found")
	}

	tx.Confirmations++
	if tx.Confirmations >= 1 && tx.Status == "pending" {
		tx.Status = "confirming"
	}

	return nil
}

// ConfirmTransaction marks a transaction as confirmed
func (s *Simulator) ConfirmTransaction(txHash string) error {
	tx, exists := s.transactions[txHash]
	if !exists {
		return fmt.Errorf("transaction not found")
	}

	tx.Status = "confirmed"
	return nil
}
