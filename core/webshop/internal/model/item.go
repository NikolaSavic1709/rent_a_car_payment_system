package model

type Item struct {
	Category    ProductCategory `json:"category"`
	Description string          `json:"description"`
}
