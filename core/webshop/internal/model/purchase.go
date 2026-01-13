package model

import "github.com/google/uuid"

type PurchaseStatus struct {
	ID              int       `json:"id"`
	URL             string    `json:"url"`
	MerchantOrderId uuid.UUID `json:"merchantOrderId"`
}
