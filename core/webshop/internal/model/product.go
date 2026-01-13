package model

type ProductCategory string

const (
	Package       ProductCategory = "PACKAGE"
	TV            ProductCategory = "TV"
	Internet      ProductCategory = "INTERNET"
	MobilePhone   ProductCategory = "MOBILE_PHONE"
	LandlinePhone ProductCategory = "LANDLINE_PHONE"
)

type Product struct {
	ID          int             `json:"id"`
	Category    ProductCategory `json:"category"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Price       float64         `json:"price"`
	Items       []Item          `json:"items,omitempty"` // For packages
}
