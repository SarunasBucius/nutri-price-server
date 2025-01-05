package ustrconv

import (
	"fmt"
	"strconv"
	"strings"
)

func StringToPositiveFloat(num string) (float64, error) {
	num = strings.TrimLeft(num, "-")
	numWithoutComma := strings.Replace(num, ",", ".", 1)
	parsedNum, err := strconv.ParseFloat(numWithoutComma, 32)
	if err != nil {
		return 0, fmt.Errorf("parse string price to float: %w", err)
	}
	return parsedNum, nil
}
