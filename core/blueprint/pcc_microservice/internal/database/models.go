package database

import (
	"github.com/google/uuid"
	"time"
)

type Transaction struct {
	ID                uint      `gorm:"primaryKey"`
	TransactionId     uuid.UUID `json:"transactionId"`
	Status            Status    `json:"status"`
	AcquirerTimestamp time.Time `json:"acquirerTimestamp"`
	Timestamp         time.Time `json:"timestamp"`
	AcquirerId        uint      `json:"acquirerId"`
	IssuerId          uint      `json:"issuerId"`
}

type Bank struct {
	ID                       uint
	BankId                   uint
	Name                     string
	BankIdentificationNumber string
}

type Status int

const (
	Successful Status = iota
	InProgress
	Failed
)
