package database

import (
	"time"

	"github.com/google/uuid"
)

// CryptoPayment represents a crypto payment transaction
type CryptoPayment struct {
	ID              uint      `gorm:"primaryKey"`
	PaymentId       uuid.UUID `gorm:"uniqueIndex" json:"paymentId"`
	TransactionId   uuid.UUID `json:"transactionId"` // From PSP
	MerchantOrderId uuid.UUID `json:"merchantOrderId"`
	MerchantId      uint      `json:"merchantId"`

	// Payment details
	Amount   float64       `json:"amount"`
	Currency string        `json:"currency"` // BTC, ETH, USDT
	Status   PaymentStatus `json:"status"`

	// Wallet addresses
	DestinationAddress string `json:"destinationAddress"` // Merchant wallet
	SourceAddress      string `json:"sourceAddress"`      // Customer wallet (once detected)

	// Blockchain details
	TxHash                string `json:"txHash"`
	BlockHeight           int64  `json:"blockHeight"`
	Confirmations         int    `json:"confirmations"`
	RequiredConfirmations int    `json:"requiredConfirmations"`

	// Timestamps
	CreatedAt   time.Time  `json:"createdAt"`
	ExpiryTime  time.Time  `json:"expiryTime"`
	ConfirmedAt *time.Time `json:"confirmedAt,omitempty"`

	// Testnet flag
	IsTestnet bool `json:"isTestnet"`
}

// MerchantWallet represents a merchant's crypto wallet
type MerchantWallet struct {
	ID            uint      `gorm:"primaryKey"`
	MerchantId    uint      `gorm:"uniqueIndex:idx_merchant_currency" json:"merchantId"`
	Currency      string    `gorm:"uniqueIndex:idx_merchant_currency" json:"currency"`
	WalletAddress string    `json:"walletAddress"`
	PublicKey     string    `json:"publicKey"`
	PrivateKey    string    `json:"-"` // Encrypted, never expose in JSON
	Balance       float64   `json:"balance"`
	IsTestnet     bool      `json:"isTestnet"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// BlockchainTransaction tracks all blockchain transactions
type BlockchainTransaction struct {
	ID            uint       `gorm:"primaryKey"`
	TxHash        string     `gorm:"uniqueIndex" json:"txHash"`
	PaymentId     uuid.UUID  `json:"paymentId"`
	FromAddress   string     `json:"fromAddress"`
	ToAddress     string     `json:"toAddress"`
	Amount        float64    `json:"amount"`
	Currency      string     `json:"currency"`
	BlockHeight   int64      `json:"blockHeight"`
	Confirmations int        `json:"confirmations"`
	Status        string     `json:"status"` // "pending", "confirmed", "failed"
	DetectedAt    time.Time  `json:"detectedAt"`
	ConfirmedAt   *time.Time `json:"confirmedAt,omitempty"`
}

// PaymentStatus enum
type PaymentStatus int

const (
	Pending PaymentStatus = iota
	Confirming
	Confirmed
	Expired
	PaymentFailed
)

func (s PaymentStatus) String() string {
	return [...]string{"pending", "confirming", "confirmed", "expired", "failed"}[s]
}

// Currency configuration
type CryptoConfig struct {
	Currency              string
	RequiredConfirmations int
	PaymentWindow         time.Duration // How long to wait for payment
	TestnetRPC            string
	MainnetRPC            string
}

var SupportedCurrencies = map[string]CryptoConfig{
	"BTC": {
		Currency:              "BTC",
		RequiredConfirmations: 3,
		PaymentWindow:         30 * time.Minute,
		TestnetRPC:            "https://blockstream.info/testnet/api",
	},
	"ETH": {
		Currency:              "ETH",
		RequiredConfirmations: 12,
		PaymentWindow:         30 * time.Minute,
		TestnetRPC:            "https://sepolia.infura.io/v3/YOUR_KEY",
	},
	"USDT": {
		Currency:              "USDT",
		RequiredConfirmations: 12,
		PaymentWindow:         30 * time.Minute,
		TestnetRPC:            "https://sepolia.infura.io/v3/YOUR_KEY", // ERC-20
	},
}

// Request/Response models

// CryptoPaymentRequest is the request from PSP to initiate a crypto payment
type CryptoPaymentRequest struct {
	TransactionId   uuid.UUID `json:"transactionId" binding:"required"`
	MerchantOrderId uuid.UUID `json:"merchantOrderId" binding:"required"`
	Amount          float64   `json:"amount" binding:"required"`
	Currency        string    `json:"currency" binding:"required"` // "BTC", "ETH", "USDT"
	Timestamp       time.Time `json:"timestamp" binding:"required"`
	MerchantId      uint      `json:"merchantId" binding:"required"`
}

// CryptoPaymentResponse is the response sent back to PSP
type CryptoPaymentResponse struct {
	PaymentId             uuid.UUID `json:"paymentId"`
	DestinationAddress    string    `json:"destinationAddress"`
	Amount                float64   `json:"amount"`
	Currency              string    `json:"currency"`
	ExpiryTime            time.Time `json:"expiryTime"`
	RequiredConfirmations int       `json:"requiredConfirmations"`
	Status                string    `json:"status"`
	QRCode                string    `json:"qrCode,omitempty"`
}

// PaymentStatusResponse is the response for payment status queries
type PaymentStatusResponse struct {
	PaymentId     uuid.UUID `json:"paymentId"`
	Status        string    `json:"status"`
	Confirmations int       `json:"confirmations"`
	TxHash        string    `json:"txHash,omitempty"`
	BlockHeight   int64     `json:"blockHeight,omitempty"`
	ConfirmedAt   time.Time `json:"confirmedAt,omitempty"`
}

// WalletGenerateRequest is the request to generate a new wallet
type WalletGenerateRequest struct {
	MerchantId uint   `json:"merchantId" binding:"required"`
	Currency   string `json:"currency" binding:"required"`
}

// WalletGenerateResponse is the response with wallet details
type WalletGenerateResponse struct {
	WalletAddress string `json:"walletAddress"`
	Currency      string `json:"currency"`
	PublicKey     string `json:"publicKey,omitempty"`
}

// TransactionVerifyRequest is for verifying a transaction
type TransactionVerifyRequest struct {
	PaymentId uuid.UUID `json:"paymentId" binding:"required"`
	TxHash    string    `json:"txHash" binding:"required"`
}

// TransactionVerifyResponse is the verification result
type TransactionVerifyResponse struct {
	Valid         bool      `json:"valid"`
	Amount        float64   `json:"amount"`
	Confirmations int       `json:"confirmations"`
	Timestamp     time.Time `json:"timestamp"`
}

// CryptoPaymentCallback is sent to PSP when payment status changes
type CryptoPaymentCallback struct {
	TransactionId   uuid.UUID         `json:"transactionId" binding:"required"`
	MerchantOrderId uuid.UUID         `json:"merchantOrderId" binding:"required"`
	Status          TransactionStatus `json:"status" binding:"required"`
	TxHash          string            `json:"txHash"`
	Amount          float64           `json:"amount"`
	Currency        string            `json:"currency"`
	Confirmations   int               `json:"confirmations"`
	CryptoTimestamp time.Time         `json:"cryptoTimestamp" binding:"required"`
}

// TransactionStatus enum (matches PSP's status)
type TransactionStatus int

const (
	Successful TransactionStatus = iota
	InProgress
	Failed
	Error
)
