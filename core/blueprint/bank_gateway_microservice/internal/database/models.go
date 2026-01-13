package database

import (
	"github.com/google/uuid"
	"time"
)

type Transaction struct {
	TransactionId   uuid.UUID `json:"transactionId" gorm:"primaryKey"`
	MerchantId      uint      `json:"merchantId"`
	MerchantOrderId uuid.UUID `json:"merchantOrderId"`
	RoutedBankId    uint      `json:"routedBank"`
	Timestamp       time.Time `json:"timestamp"`
	Amount          float32   `json:"amount" binding:"required"`
	Currency        string    `json:"currency" binding:"required"`
}
type PaymentRequest struct {
	ExpDate         time.Time `json:"expDate" binding:"required"`
	CardNumber      string    `json:"cardNumber" binding:"required"`
	Currency        string    `json:"currency" binding:"required"`
	Amount          float32   `json:"amount" binding:"required"`
	MerchantId      uint      `json:"merchantId" binding:"required"`
	MerchantOrderId uuid.UUID `json:"merchantOrderId" binding:"required"`
	TransactionId   uuid.UUID `json:"transactionId" binding:"required"`
	Timestamp       time.Time `json:"timestamp" binding:"required"`
}

type TransactionResponse struct {
	AcquirerOrderId   uuid.UUID         `json:"acquirerOrderId" binding:"required"`
	AcquirerTimestamp time.Time         `json:"acquirerTimestamp" binding:"required"`
	MerchantOrderId   uuid.UUID         `json:"merchantOrderId" binding:"required"`
	TransactionId     uuid.UUID         `json:"transactionId" binding:"required"`
	Status            TransactionStatus `json:"status" binding:"required"`
}

type MerchantInfo struct {
	MerchantId uint `json:"merchantId" gorm:"primaryKey"`
	BankId     uint `json:"bankId"`
}

type TransactionStatus int

const (
	Successful TransactionStatus = iota
	InProgress
	Failed
	Error
)
