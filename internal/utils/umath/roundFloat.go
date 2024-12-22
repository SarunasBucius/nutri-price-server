package umath

import "math"

func RoundFloat(num float64, decimalPlaces int) float64 {
	decimalMultiplier := math.Pow10(decimalPlaces)
	return math.Round(num*decimalMultiplier) / decimalMultiplier
}
