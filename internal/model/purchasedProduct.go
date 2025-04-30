package model

import "time"

type PurchasedProductNew struct {
	ProductID   string   `json:"productId"`
	Name        string   `json:"name"`
	VarietyName string   `json:"varietyName"`
	Price       float64  `json:"price"`
	Quantity    Quantity `json:"quantity"`
	Notes       string   `json:"notes"`
	ParsedName  string   `json:"parsedName"`
}

type ProductAndVarietyName struct {
	Name        string `json:"name"`
	VarietyName string `json:"varietyName"`
}

type PurchasedProduct struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	VarietyName string    `json:"varietyName"`
	Retailer    string    `json:"retailer"`
	Price       float64   `json:"price"`
	Quantity    Quantity  `json:"quantity"`
	Date        time.Time `json:"date"`
	Notes       string    `json:"notes"`
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
