package blockchain

import (
	"crypto_microservice/internal/database"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// CallbackSender interface for sending callbacks
type CallbackSender interface {
	SendCallbackToPSP(payment *database.CryptoPayment)
}

type Monitor struct {
	db             database.Service
	activePayments map[uuid.UUID]bool
	mu             sync.RWMutex
	stopChan       chan struct{}
	callbackSender CallbackSender
}

func NewMonitor(db database.Service) *Monitor {
	return &Monitor{
		db:             db,
		activePayments: make(map[uuid.UUID]bool),
		stopChan:       make(chan struct{}),
	}
}

// SetCallbackSender sets the callback sender (server instance)
func (m *Monitor) SetCallbackSender(sender CallbackSender) {
	m.callbackSender = sender
}

func (m *Monitor) Start() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.checkAllActivePayments()
		case <-m.stopChan:
			return
		}
	}
}

func (m *Monitor) Stop() {
	close(m.stopChan)
}

func (m *Monitor) MonitorPayment(paymentId uuid.UUID) {
	m.mu.Lock()
	m.activePayments[paymentId] = true
	m.mu.Unlock()
}

func (m *Monitor) checkAllActivePayments() {
	m.mu.RLock()
	paymentIds := make([]uuid.UUID, 0, len(m.activePayments))
	for id := range m.activePayments {
		paymentIds = append(paymentIds, id)
	}
	m.mu.RUnlock()

	for _, paymentId := range paymentIds {
		m.checkPayment(paymentId)
	}
}

func (m *Monitor) checkPayment(paymentId uuid.UUID) {
	payment, err := m.db.GetPaymentByPaymentId(paymentId)
	if err != nil {
		return
	}

	// Skip if already confirmed or failed
	if payment.Status == database.Confirmed ||
		payment.Status == database.PaymentFailed ||
		payment.Status == database.Expired {
		m.mu.Lock()
		delete(m.activePayments, paymentId)
		m.mu.Unlock()
		return
	}

	// Check if expired
	if time.Now().After(payment.ExpiryTime) {
		payment.Status = database.Expired
		m.db.UpdatePayment(payment)
		m.sendCallback(payment)

		m.mu.Lock()
		delete(m.activePayments, paymentId)
		m.mu.Unlock()
		return
	}

	// For confirming payments, increment confirmations (simulation)
	if payment.Status == database.Confirming {
		payment.Confirmations++

		if payment.Confirmations >= payment.RequiredConfirmations {
			payment.Status = database.Confirmed
			now := time.Now()
			payment.ConfirmedAt = &now
			m.sendCallback(payment)

			m.mu.Lock()
			delete(m.activePayments, paymentId)
			m.mu.Unlock()
		}

		m.db.UpdatePayment(payment)
	}
}

func (m *Monitor) sendCallback(payment *database.CryptoPayment) {
	fmt.Printf("Payment %s status changed to: %s\n",
		payment.PaymentId, payment.Status.String())

	// Send callback to PSP if sender is configured
	if m.callbackSender != nil {
		m.callbackSender.SendCallbackToPSP(payment)
	}
}
