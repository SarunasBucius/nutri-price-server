package norfa

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/SarunasBucius/nutri-price-server/internal/model"
	"github.com/SarunasBucius/nutri-price-server/internal/utils/umath"
)

var dynamicWeightRegexp = regexp.MustCompile(`(\d+(?:[.,]\d+)?)x(\d+(?:[.,]\d+)?)`)
var numberAndQuantityUnit = regexp.MustCompile(`(\d+[.,]?\d*)(kg|g|ml|l)`)

type weightParser interface {
	getQuantity() (model.Quantity, error)
	// trimProductName trims info related to dynamic weight from product name.
	trimProductName() string
}

func newWeightParser(product string) weightParser {
	if weight, found := hasDynamicWeight(product); found {
		return dynamicWeight{
			weight:  weight,
			product: product,
		}
	}
	return staticWeight{
		product: product,
	}
}

type dynamicWeight struct {
	weight  string
	product string
}

func (w dynamicWeight) getQuantity() (model.Quantity, error) {
	weightSeparatedByDot := strings.ReplaceAll(w.weight, ",", ".")

	amount, err := strconv.ParseFloat(weightSeparatedByDot, 32)
	if err != nil {
		return model.Quantity{}, fmt.Errorf("parse dynamic weight to float: %w", err)
	}

	return model.Quantity{
		Unit:   model.Grams,
		Amount: umath.RoundFloat(amount*1000, 0),
	}, nil
}

func (w dynamicWeight) trimProductName() string {
	product, _, _ := strings.Cut(w.product, w.weight)
	product, _, _ = strings.Cut(product, "1kg")
	product, _, _ = strings.Cut(product, "1 kg")
	product = strings.TrimSpace(product)
	product, _ = strings.CutSuffix(product, ",")
	return product
}

type staticWeight struct {
	product string
}

func (w staticWeight) getQuantity() (model.Quantity, error) {
	productNameWithoutSpaces := strings.ReplaceAll(w.product, " ", "")
	match := numberAndQuantityUnit.FindStringSubmatch(productNameWithoutSpaces)

	if len(match) != 3 {
		return model.Quantity{
			Unit:   model.Pieces,
			Amount: 1,
		}, nil
	}

	match[1] = strings.ReplaceAll(match[1], ",", ".")

	amount, err := strconv.ParseFloat(match[1], 32)
	if err != nil {
		return model.Quantity{}, err
	}

	switch match[2] {
	case "kg":
		return model.Quantity{
			Unit:   model.Grams,
			Amount: amount * 1000,
		}, nil
	case "g":
		return model.Quantity{
			Unit:   model.Grams,
			Amount: amount,
		}, nil
	case "ml":
		return model.Quantity{
			Unit:   model.Milliliters,
			Amount: amount,
		}, nil
	case "l":
		return model.Quantity{
			Unit:   model.Milliliters,
			Amount: amount * 1000,
		}, nil
	default:
		return model.Quantity{}, fmt.Errorf("unexpected regex result %q", match[0])
	}
}

func (w staticWeight) trimProductName() string { return w.product }

func hasDynamicWeight(product string) (string, bool) {
	match := dynamicWeightRegexp.FindStringSubmatch(product)
	if len(match) == 0 {
		return "", false
	}
	return match[1], true
}
