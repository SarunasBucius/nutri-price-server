package model

type Quantity struct {
	Unit   string  `json:"unit"`
	Amount float64 `json:"amount"`
}

const (
	Pieces      = "pieces"
	Grams       = "grams"
	Milliliters = "milliliters"
)
