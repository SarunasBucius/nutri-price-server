package model

type Quantity struct {
	Unit   Unit    `json:"unit"`
	Amount float64 `json:"amount"`
}

type Unit string

const (
	Pieces      Unit = "pieces"
	Grams       Unit = "grams"
	Milliliters Unit = "milliliters"
)
