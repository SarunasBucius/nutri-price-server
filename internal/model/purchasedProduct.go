package model

import "time"

type PurchasedProductNew struct {
	Name     string   `json:"name"`
	Price    Price    `json:"price"`
	Quantity Quantity `json:"quantity"`
	Group    string   `json:"group"`
	Notes    string   `json:"notes"`
}

type Price struct {
	Discount float64 `json:"discount"`
	Paid     float64 `json:"paid"`
	Full     float64 `json:"full"`
}

type PurchasedProduct struct {
	ID       int       `json:"id"`
	Name     string    `json:"name"`
	Retailer string    `json:"retailer"`
	Price    Price     `json:"price"`
	Quantity Quantity  `json:"quantity"`
	Group    string    `json:"group"`
	Date     time.Time `json:"date"`
	Notes    string    `json:"notes"`
}

type PostPurchasedProductsRequest struct {
	Date              string                `json:"date"`
	Retailer          string                `json:"retailer"`
	PurchasedProducts []PurchasedProductNew `json:"products"`
}
