package model

import "time"

type PurchasedProductNew struct {
	Name       string   `json:"name"`
	Price      Price    `json:"price"`
	Quantity   Quantity `json:"quantity"`
	Group      string   `json:"group"`
	Notes      string   `json:"notes"`
	ParsedName string   `json:"parsedName"`
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

type PurchasedProductsNew struct {
	Date              string                `json:"date"`
	Retailer          string                `json:"retailer"`
	PurchasedProducts []PurchasedProductNew `json:"products"`
}

type UnconfirmedReceiptSummary struct {
	Retailer string `json:"retailer"`
	Date     string `json:"date"`
}

type ConfirmPurchasedProductsRequest struct {
	Date     string                `json:"date"`
	Retailer string                `json:"retailer"`
	Products []PurchasedProductNew `json:"products"`
}
