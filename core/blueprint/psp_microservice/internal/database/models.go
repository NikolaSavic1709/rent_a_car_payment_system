package database

import (
	"time"

	"github.com/google/uuid"
)

type Transaction struct {
	TransactionId     uuid.UUID         `json:"transactionId" gorm:"primaryKey"`
	MerchantId        uint              `json:"merchantId" binding:"required"`
	MerchantOrderId   uuid.UUID         `json:"merchantOrderId" binding:"required"`
	Status            TransactionStatus `json:"status"`
	Timestamp         time.Time         `json:"timestamp"`
	MerchantTimestamp time.Time         `json:"merchantTimestamp"`
	Amount            float32           `json:"amount" binding:"required"`
	Currency          string            `json:"currency" binding:"required"`
	PaymentMethod		string  `json:"paymentMethod" binding:"required"`
	QRRef             uint64            `json:"qrRef" gorm:"uniqueIndex"`
}

type WebShopPaymentRequest struct {
	PaymentDeadline   time.Time `json:"paymentDeadline" binding:"required"`
	Currency          string    `json:"currency" binding:"required"`
	Amount            float32   `json:"amount" binding:"required"`
	MerchantId        uint      `json:"merchantId" binding:"required"`
	MerchantPassword  string    `json:"merchantPassword" binding:"required"`
	MerchantOrderId   uuid.UUID `json:"merchantOrderId" binding:"required"`
	MerchantTimestamp time.Time `json:"merchantTimestamp" binding:"required"`
	PaymentMethod		string  `json:"paymentMethod" binding:"required"`
	// SuccessURL        string    `json:"successURL"`
	// FailURL           string    `json:"failURL"`
	// ErrorURL          string    `json:"errorURL"`
}

type PaymentRequest struct {
	ExpDate         time.Time `json:"expDate"`
	CardNumber      string    `json:"cardNumber"`
	Currency        string    `json:"currency" binding:"required"`
	Amount          float32   `json:"amount" binding:"required"`
	MerchantId      uint      `json:"merchantId" binding:"required"`
	MerchantOrderId uuid.UUID `json:"merchantOrderId"`
	TransactionId   uuid.UUID `json:"transactionId" binding:"required"`
	Timestamp       time.Time `json:"timestamp" binding:"required"`
}

type PaymentStartResponse struct {
	PaymentURL string    `json:"paymentURL"`
	TokenId    uuid.UUID `json:"tokenId"`
	Token      string    `json:"token"`
	TokenExp   time.Time `json:"tokenExp"`
	QRRef	 uint64    `json:"qrRef"`
}

type CardDetailsRequest struct {
	CardNumber           string    `json:"cardNumber" binding:"required"`
	MerchantOrderId      uuid.UUID `json:"merchantOrderId" binding:"required"`
	ExpDate              time.Time `json:"expDate" binding:"required"`
	CardVerificationCode uint      `json:"cardVerificationCode" binding:"required"`
}

type QRCodeRequest struct {
	CardNumber           string    `json:"cardNumber" binding:"required"`
	QRRef                uint64    `json:"qrRef" binding:"required"`
	ExpDate              time.Time `json:"expDate" binding:"required"`
	CardVerificationCode uint      `json:"cardVerificationCode" binding:"required"`
}


type TransactionResponse struct {
	AcquirerOrderId   uuid.UUID         `json:"acquirerOrderId" binding:"required"`
	AcquirerTimestamp time.Time         `json:"acquirerTimestamp" binding:"required"`
	MerchantOrderId   uuid.UUID         `json:"merchantOrderId" binding:"required"`
	TransactionId     uuid.UUID         `json:"transactionId" binding:"required"`
	Status            TransactionStatus `json:"status" binding:"required"`
}

type Merchant struct {
	Username 		string `json:"username"`
	MerchantId        uint   `json:"merchantId"`
	Password          string `json:"password"`
	Salt              string `json:"salt"`
	SuccessURL        string `json:"successURL"`
	FailURL           string `json:"failURL"`
	ErrorURL          string `json:"errorURL"`
}

type TransactionStatus int

const (
	Successful TransactionStatus = iota
	InProgress
	Failed
	Error
)

type Subscription struct {
	SubscriptionId uint `gorm:"primaryKey;autoIncrement"`
	MerchantId     uint
	Method         PaymenthMethod
}

type PaymenthMethod int

const (
	Card PaymenthMethod = iota
	Paypal
	Crypto
	QrCode
)

type NBSUploadResponse struct {
    S struct {
        Code int    `json:"code"` // 0 for OK 
        Desc string `json:"desc"`
    } `json:"s"`
    T string `json:"t"` // Raw text from QR [cite: 250]
    N struct {
        K  string `json:"K"`
        V  string `json:"V"`
        C  string `json:"C"`
        R  string `json:"R"`
        N  string `json:"N"`
        I  string `json:"I"`
        P  string `json:"P"`
        SF string `json:"SF"`
        S  string `json:"S"`
        RO string `json:"RO"` // This is your QRRef [cite: 247, 218]
    } `json:"n"`
}
