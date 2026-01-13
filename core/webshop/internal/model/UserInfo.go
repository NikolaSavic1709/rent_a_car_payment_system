package model

import "time"

type Payment struct {
	ID       int       `json:"id"`
	Deadline time.Time `json:"deadline"`
	Cost     float64   `json:"cost"`
	Vehicle  Vehicle   `json:"vehicle"`
}
