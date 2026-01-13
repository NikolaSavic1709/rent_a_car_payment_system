package model

import "time"

type Payment struct {
	ID       int       `json:"id"`
	Deadline time.Time `json:"deadline"`
	Cost     float64   `json:"cost"`
	Product  Product   `json:"product"`
}
