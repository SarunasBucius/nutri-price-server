package model

type NutritionalValue struct {
	EnergyValueKCAL    float64 `json:"energyValueKcal"`
	Fat                float64 `json:"fat"`
	SaturatedFat       float64 `json:"saturatedFat"`
	Carbohydrate       float64 `json:"carbohydrate"`
	CarbohydrateSugars float64 `json:"carbohydrateSugars"`
	Fibre              float64 `json:"fibre"`
	SolubleFibre       float64 `json:"solubleFibre"`
	InsolubleFibre     float64 `json:"insolubleFibre"`
	Protein            float64 `json:"protein"`
	Salt               float64 `json:"salt"`
}

type ProductNutritionalValueNew struct {
	Product          string           `json:"product"`
	Unit             string           `json:"unit"`
	NutritionalValue NutritionalValue `json:"nutritionalValue"`
}

type ProductNutritionalValue struct {
	ID               int              `json:"id"`
	Product          string           `json:"product"`
	Unit             Unit             `json:"unit"`
	NutritionalValue NutritionalValue `json:"nutritionalValue"`
}
