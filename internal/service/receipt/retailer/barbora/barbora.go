package barbora

import (
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/SarunasBucius/nutri-price-server/internal/model"
	"github.com/SarunasBucius/nutri-price-server/internal/utils/umath"
	"github.com/SarunasBucius/nutri-price-server/internal/utils/ustrconv"
)

const retailer = "barbora"

type BarboraParser struct {
	ReceiptLines []string
	Retailer     string
}

func NewParser(receiptLines []string) BarboraParser {
	return BarboraParser{
		ReceiptLines: receiptLines,
		Retailer:     retailer,
	}
}

func (p BarboraParser) ParseDate() (time.Time, error) {
	const datePosition = 1
	if len(p.ReceiptLines) < 2 {
		return time.Time{}, fmt.Errorf("invalid receipt")
	}

	dateLine := p.ReceiptLines[datePosition]

	parsedDate, err := time.Parse(time.DateOnly, dateLine)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse receipt date: %w", err)
	}
	return parsedDate, nil
}

func (p BarboraParser) ParseProducts() (model.ReceiptProducts, error) {
	unparsedProducts, err := extractProductLines(p.ReceiptLines)
	if err != nil {
		return nil, fmt.Errorf("extract product lines: %w", err)
	}

	parsedProducts := make([]model.PurchasedProductNew, 0, len(unparsedProducts))
	for _, product := range unparsedProducts {
		if isDeposit(product) {
			parsedProducts[len(parsedProducts)-1].Price += 0.1
			continue
		}
		parsedProduct, err := parseProduct(product)
		if err != nil {
			return nil, fmt.Errorf("parse product %+v: %w", product, err)
		}
		parsedProducts = append(parsedProducts, parsedProduct)
	}
	return parsedProducts, nil
}

func (p BarboraParser) GetRetailer() string { return retailer }

func parseProduct(product string) (model.PurchasedProductNew, error) {
	productSplitBySpace := strings.Split(product, " ")
	if len(productSplitBySpace) < 9 {
		return model.PurchasedProductNew{}, fmt.Errorf("unexpected product line: %v", product)
	}

	productName := strings.Join(productSplitBySpace[1:len(productSplitBySpace)-7], " ")

	unparsedPrice := strings.TrimPrefix(productSplitBySpace[len(productSplitBySpace)-1], "€")
	price, err := parsePrice(unparsedPrice)
	if err != nil {
		return model.PurchasedProductNew{}, fmt.Errorf("parse price: %w", err)
	}

	amount := productSplitBySpace[len(productSplitBySpace)-7]
	unit := productSplitBySpace[len(productSplitBySpace)-6]
	quantity, err := getQuantity(amount, unit, productName)
	if err != nil {
		return model.PurchasedProductNew{}, fmt.Errorf("get quantity: %w", err)
	}

	return model.PurchasedProductNew{
		Name:     productName,
		Price:    price,
		Quantity: quantity,
	}, nil
}

func getQuantity(amount, unit, product string) (model.Quantity, error) {
	amountFloat, err := ustrconv.StringToPositiveFloat(amount)
	if err != nil {
		return model.Quantity{}, fmt.Errorf("parse product amount: %w", err)
	}

	switch strings.ToLower(unit) {
	case "vnt.":
		return parseMetricQuantity(amountFloat, unit, product), nil
	case "kg":
		return model.Quantity{
			Unit:   model.Grams,
			Amount: umath.RoundFloat(amountFloat*1000, 0),
		}, nil
	default:
		return model.Quantity{}, nil
	}
}

var numberAndQuantityUnit = regexp.MustCompile(`(\d+[.,]?\d*)(kg|g|ml|l)`)

func parseMetricQuantity(amountPcs float64, unit, product string) model.Quantity {
	productNameWithoutSpaces := strings.ReplaceAll(product, " ", "")
	match := numberAndQuantityUnit.FindStringSubmatch(productNameWithoutSpaces)

	if len(match) != 3 {
		return model.Quantity{
			Unit:   unit,
			Amount: amountPcs,
		}
	}

	match[1] = strings.ReplaceAll(match[1], ",", ".")

	amount, err := strconv.ParseFloat(match[1], 32)
	if err != nil {
		slog.Error("parse float", "error", err)
		return model.Quantity{
			Unit:   unit,
			Amount: amountPcs,
		}
	}

	switch match[2] {
	case "kg":
		return model.Quantity{
			Unit:   model.Grams,
			Amount: amount * 1000 * amountPcs,
		}
	case "g":
		return model.Quantity{
			Unit:   model.Grams,
			Amount: amount * amountPcs,
		}
	case "ml":
		return model.Quantity{
			Unit:   model.Milliliters,
			Amount: amount * amountPcs,
		}
	case "l":
		return model.Quantity{
			Unit:   model.Milliliters,
			Amount: amount * 1000 * amountPcs,
		}
	default:
		slog.Error("unexpected regex result", "match", match[0])
		return model.Quantity{
			Unit:   unit,
			Amount: amountPcs,
		}
	}
}

func parsePrice(unparsedPrice string) (float64, error) {
	paid, err := ustrconv.StringToPositiveFloat(unparsedPrice)
	if err != nil {
		return 0, fmt.Errorf("parse product price: %w", err)
	}

	return umath.RoundFloat(paid, 2), nil
}

func extractProductLines(receiptLines []string) ([]string, error) {
	const productsEndSeparator = "Pritaikytos nuolaidos"
	const productsListStart = 2
	if len(receiptLines) <= productsListStart {
		return nil, fmt.Errorf("too short receipt")
	}
	receiptLines = receiptLines[productsListStart:]

	var products []string
	for i := range receiptLines {
		if strings.HasPrefix(receiptLines[i], productsEndSeparator) {
			break
		}

		products = extractProduct(receiptLines[i], products)
	}
	return products, nil
}

func extractDiscountedProducts(receiptLines []string) []string {
	const discountStartMarker = "Pritaikytos nuolaidos"
	discountsListStart := 0
	for i := range receiptLines {
		if strings.HasPrefix(receiptLines[i], discountStartMarker) {
			discountsListStart = i + 1
			break
		}
	}
	if discountsListStart == 0 || len(receiptLines) <= discountsListStart {
		return nil
	}
	discountedProducts := make([]string, 0, len(receiptLines))
	for i := discountsListStart; i < len(receiptLines); i++ {
		if len(discountedProducts) == 0 {
			discountedProducts = append(discountedProducts, receiptLines[i])
			continue
		}
		lastProduct := len(discountedProducts) - 1
		productAndDiscount := strings.Split(discountedProducts[lastProduct], " -€")
		if len(productAndDiscount) != 2 {
			discountedProducts[lastProduct] += " " + receiptLines[i]
			continue
		}
		discountedProducts = append(discountedProducts, receiptLines[i])
	}
	return discountedProducts
}

func getDiscountsByProduct(discountedProducts []string) (map[string]string, error) {
	discountsByProduct := make(map[string]string, len(discountedProducts))
	for i := range discountedProducts {
		productAndDiscount := strings.Split(discountedProducts[i], " -€")
		if len(productAndDiscount) != 2 {
			return nil, fmt.Errorf("invalid discount line: %s", discountedProducts[i])
		}
		discountsByProduct[productAndDiscount[0]] = productAndDiscount[1]
	}

	return discountsByProduct, nil
}

func extractProduct(line string, products []string) []string {
	if len(products) == 0 {
		return append(products, line)
	}

	productsNum := len(products)

	if strings.HasPrefix(line, strconv.Itoa(productsNum+1)) {
		return append(products, line)
	}

	products[productsNum-1] += " " + line

	return products
}

func isDeposit(product string) bool {
	return strings.Contains(product, "(depozitinis)")
}
