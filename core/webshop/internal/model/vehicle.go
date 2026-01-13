package model

type VehicleCategory string

const (
	SUV       VehicleCategory = "SUV"
	Sedan     VehicleCategory = "Sedan"
	Hatchback VehicleCategory = "Hatchback"
	Coupe     VehicleCategory = "Coupe"
	Pickup    VehicleCategory = "Pickup"
)

type Vehicle struct {
	ID          int             `json:"id"`
	Category    VehicleCategory `json:"category"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Price       float64         `json:"price"`
}
