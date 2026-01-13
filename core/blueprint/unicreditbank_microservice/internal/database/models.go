package database

import (
	"github.com/google/uuid"
	"time"
)

type BankClient struct {
	ID          uint      `gorm:"primaryKey"`
	Name        string    `json:"name"`
	Surname     string    `json:"surname"`
	Birthday    time.Time `json:"birthday"`
	Email       string    `json:"email"`
	PhoneNumber string    `json:"phoneNumber"`
}

type BankAccount struct {
	ID            uint          `gorm:"primaryKey"`
	AccountNumber string        `json:"accountNumber"`
	UserID        uint          `json:"userId"`
	User          BankClient    `gorm:"foreignKey:UserID"`
	Balance       float32       `json:"balance"`
	Currency      string        `json:"currency"`
	DateCreated   time.Time     `json:"dateCreated"`
	Status        AccountStatus `json:"status"`
}

type Card struct {
	ID            uint        `gorm:"primaryKey"`
	CardId        string      `json:"cardId"`
	BankAccountID uint        `json:"bankAccountID"`
	BankAccount   BankAccount `gorm:"foreignKey:BankAccountID"`
	CardNumber    string      `json:"cardNumber"`
	ExpiryDate    time.Time   `json:"expiryDate"`
	CardType      CardType    `json:"cardType"`
	IsTokenized   bool        `json:"isTokenized"`
}

type Transaction struct {
	ID                uint              `gorm:"primaryKey"`
	TransactionId     uuid.UUID         `json:"transactionId"`
	AcquirerOrderId   uuid.UUID         `json:"acquirerOrderId"`
	MerchantId        uint              `json:"merchantId"`
	Status            TransactionStatus `json:"status"`
	Amount            float32           `json:"amount"`
	Currency          string            `json:"currency"`
	Timestamp         time.Time         `json:"timestamp"`
	PartialCardNumber string            `json:"partialCardNumber"`
}

type TransactionStatus int

const (
	Successful TransactionStatus = iota
	InProgress
	Failed
	Error
)

type AccountStatus int

const (
	Active AccountStatus = iota
	Closed
	Blocked
)

type CardType int

const (
	Debit CardType = iota
	Credit
	Prepaid
)
