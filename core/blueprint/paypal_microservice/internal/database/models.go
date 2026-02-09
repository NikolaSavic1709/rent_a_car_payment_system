package database

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// PayPalPayment represents a PayPal payment transaction
type PayPalPayment struct {
	ID              uint      `json:"id"`
	PaymentID       uuid.UUID `json:"paymentId"`
	TransactionID   uuid.UUID `json:"transactionId"`   // From PSP
	MerchantOrderID uuid.UUID `json:"merchantOrderId"` // From webshop
	MerchantID      uint      `json:"merchantId"`

	// PayPal specific
	PayPalOrderID   string         `json:"paypalOrderId"`
	PayPalCaptureID sql.NullString `json:"paypalCaptureId,omitempty"`

	// Payment details
	Amount   float64       `json:"amount"`
	Currency string        `json:"currency"`
	Status   PaymentStatus `json:"status"`

	// Payer information
	PayerEmail sql.NullString `json:"payerEmail,omitempty"`
	PayerID    sql.NullString `json:"payerId,omitempty"`
	PayerName  sql.NullString `json:"payerName,omitempty"`

	// Timestamps
	CreatedAt   time.Time    `json:"createdAt"`
	ApprovedAt  sql.NullTime `json:"approvedAt,omitempty"`
	CompletedAt sql.NullTime `json:"completedAt,omitempty"`
	CancelledAt sql.NullTime `json:"cancelledAt,omitempty"`

	// URLs
	ApprovalURL sql.NullString `json:"approvalUrl,omitempty"`

	// Additional info
	Description   sql.NullString `json:"description,omitempty"`
	FailureReason sql.NullString `json:"failureReason,omitempty"`
}

// PaymentStatus enum
type PaymentStatus int

const (
	Pending   PaymentStatus = iota // Order created, awaiting approval
	Approved                       // User approved, awaiting capture
	Completed                      // Payment captured successfully
	Cancelled                      // User cancelled payment
	Failed                         // Payment processing failed
)

func (s PaymentStatus) String() string {
	return [...]string{"pending", "approved", "completed", "cancelled", "failed"}[s]
}

// PaymentRequest represents the request to create a PayPal payment
type PaymentRequest struct {
	TransactionID   uuid.UUID `json:"transactionId" binding:"required"`
	MerchantOrderID uuid.UUID `json:"merchantOrderId" binding:"required"`
	MerchantID      uint      `json:"merchantId" binding:"required"`
	Amount          float64   `json:"amount" binding:"required,gt=0"`
	Currency        string    `json:"currency" binding:"required"`
	Description     string    `json:"description"`
}

// PaymentResponse represents the response after creating a payment
type PaymentResponse struct {
	PaymentID     uuid.UUID `json:"paymentId"`
	PayPalOrderID string    `json:"paypalOrderId"`
	ApprovalURL   string    `json:"approvalUrl"`
	Status        string    `json:"status"`
}

// PaymentStatusResponse represents the payment status response
type PaymentStatusResponse struct {
	PaymentID     uuid.UUID  `json:"paymentId"`
	TransactionID uuid.UUID  `json:"transactionId"`
	PayPalOrderID string     `json:"paypalOrderId"`
	Status        string     `json:"status"`
	Amount        float64    `json:"amount"`
	Currency      string     `json:"currency"`
	PayerEmail    string     `json:"payerEmail,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
	CompletedAt   *time.Time `json:"completedAt,omitempty"`
}

// PayPalCallback represents the callback data sent to PSP
type PayPalCallback struct {
	TransactionID   uuid.UUID         `json:"transactionId"`
	MerchantOrderID uuid.UUID         `json:"merchantOrderId"`
	Status          TransactionStatus `json:"status"`
	PayPalOrderID   string            `json:"paypalOrderId"`
	Amount          float64           `json:"amount"`
	Currency        string            `json:"currency"`
	PayPalTimestamp time.Time         `json:"paypalTimestamp"`
}

// TransactionStatus matches PSP's status enum
type TransactionStatus int

const (
	Successful TransactionStatus = iota
	InProgress
	TransactionFailed
	Error
)

func (s TransactionStatus) String() string {
	return [...]string{"successful", "inProgress", "failed", "error"}[s]
}
