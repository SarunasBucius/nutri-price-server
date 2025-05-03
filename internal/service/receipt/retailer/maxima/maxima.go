package maxima

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/SarunasBucius/nutri-price-server/internal/model"
	"github.com/SarunasBucius/nutri-price-server/internal/utils/umath"
	"github.com/SarunasBucius/nutri-price-server/internal/utils/ustrconv"
)

const retailer = "maxima"

type MaximaParser struct {
	ReceiptLines []string
	Retailer     string
}

func NewParser(receiptLines []string) MaximaParser {
	return MaximaParser{
		ReceiptLines: receiptLines,
		Retailer:     retailer,
	}
}

type unparsedProduct struct {
	product       string
	hasDeposit    bool
	discount      string
	isHalf        bool
	dynamicWeight []string
}

func (p MaximaParser) ParseDate() (time.Time, error) {
	const datePosition = 1

	dateLine := getDateLine(p.ReceiptLines)

	dateLineSplitBySpace := slices.DeleteFunc(strings.Split(dateLine, " "), func(word string) bool {
		return word == ""
	})

	if len(dateLineSplitBySpace) <= datePosition {
		return time.Time{}, fmt.Errorf("unexpected date line contents: %s", dateLine)
	}

	parsedDate, err := time.Parse(time.DateOnly, dateLineSplitBySpace[datePosition])
	if err != nil {
		return time.Time{}, fmt.Errorf("parse receipt date: %w", err)
	}
	return parsedDate, nil
}

func (p MaximaParser) ParseProducts() (model.ReceiptProducts, error) {
	unparsedProducts, err := extractProductLines(p.ReceiptLines)
	if err != nil {
		return nil, fmt.Errorf("extract product lines: %w", err)
	}

	parsedProducts := make([]model.PurchasedProductNew, 0, len(unparsedProducts))
	for _, product := range unparsedProducts {
		parsedProduct, err := parseProduct(product)
		if err != nil {
			return nil, fmt.Errorf("parse product %+v: %w", product, err)
		}
		parsedProducts = append(parsedProducts, parsedProduct)
	}
	return parsedProducts, nil
}

func (p MaximaParser) GetRetailer() string { return retailer }

func getDateLine(receiptLines []string) string {
	for i := len(receiptLines) - 1; i >= 0; i-- {
		if strings.Contains(strings.ToLower(receiptLines[i]), "laikas") {
			return receiptLines[i]
		}
	}
	return ""
}

func extractProductLines(receiptLines []string) ([]unparsedProduct, error) {
	const productsEndSeparator = "========================"
	productsListStart := findProductsListStart(receiptLines)
	if len(receiptLines) <= productsListStart {
		return nil, fmt.Errorf("too short receipt")
	}
	receiptLines = receiptLines[productsListStart:]

	var products []unparsedProduct
	for i := range receiptLines {
		if strings.HasPrefix(receiptLines[i], productsEndSeparator) {
			break
		}

		products = extractProduct(receiptLines[i], products)
	}

	return products, nil
}

func findProductsListStart(receiptLines []string) int {
	for i, line := range receiptLines {
		if strings.Contains(strings.ToLower(line), "kvitas") {
			return i + 1
		}
	}
	return -1
}

func extractProduct(line string, products []unparsedProduct) []unparsedProduct {
	if len(products) == 0 {
		return appendProduct(line, products)
	}

	lastProduct := len(products) - 1

	if isDeposit(line) {
		products[lastProduct].hasDeposit = true
		return products
	}

	lineSplitBySpace := strings.Split(line, " ")
	lineSplitBySpace = slices.DeleteFunc(lineSplitBySpace, func(l string) bool {
		return l == ""
	})
	isDynamic := len(lineSplitBySpace) >= 4 && strings.ToLower(lineSplitBySpace[1]) == "x"
	if isDynamic {
		products[lastProduct].dynamicWeight = lineSplitBySpace
		products[lastProduct].isHalf = false
		return products
	}

	if products[lastProduct].isHalf && products[lastProduct].discount == "" {
		products[lastProduct].product += " " + line
		products[lastProduct].isHalf = !strings.HasSuffix(line, "A")
		return products
	}

	if products[lastProduct].isHalf {
		products[lastProduct].discount += " " + line
		products[lastProduct].isHalf = false
		return products
	}

	if isDiscount(line) {
		products[lastProduct].discount = line
		products[lastProduct].isHalf = !strings.HasSuffix(line, "A")
		return products
	}

	return appendProduct(line, products)
}

func isDeposit(line string) bool {
	return strings.Contains(line, "depozitinė")
}

func appendProduct(productLine string, products []unparsedProduct) []unparsedProduct {
	if strings.HasSuffix(productLine, "A") {
		return append(products, unparsedProduct{product: productLine})
	}

	return append(products, unparsedProduct{product: productLine, isHalf: true})
}

func isDiscount(product string) bool {
	lowerCaseProduct := strings.ToLower(product)
	return strings.Contains(lowerCaseProduct, "nuolaida")
}

func parseProduct(product unparsedProduct) (model.PurchasedProductNew, error) {
	unparsedPrice := getUnparsedPrice(product)
	price, err := parsePrice(product, unparsedPrice)
	if err != nil {
		return model.PurchasedProductNew{}, err
	}

	productName := trimPriceInfoFromProductName(product.product, unparsedPrice)

	quantity, err := getQuantity(product)
	if err != nil {
		return model.PurchasedProductNew{}, fmt.Errorf("extract quantity info: %w", err)
	}

	return model.PurchasedProductNew{
		VarietyName: strings.TrimSpace(productName),
		Price:       price,
		Quantity:    quantity,
	}, nil
}

func getQuantity(product unparsedProduct) (model.Quantity, error) {
	if len(product.dynamicWeight) != 4 && len(product.dynamicWeight) != 6 {
		return model.Quantity{}, nil
	}

	amount, err := strconv.ParseFloat(strings.Replace(product.dynamicWeight[2], ",", ".", 1), 64)
	if err != nil {
		return model.Quantity{}, fmt.Errorf("parse product amount: %w", err)
	}

	switch strings.ToLower(product.dynamicWeight[3]) {
	case "vnt.":
		return model.Quantity{
			Amount: amount,
			Unit:   model.Pieces,
		}, nil
	case "kg":
		return model.Quantity{
			Unit:   model.Grams,
			Amount: amount * 1000,
		}, nil
	default:
		return model.Quantity{}, nil
	}
}

func getUnparsedPrice(product unparsedProduct) string {
	if len(product.dynamicWeight) == 6 {
		return product.dynamicWeight[len(product.dynamicWeight)-2]
	}
	productSplitBySpaces := strings.Split(product.product, " ")
	return productSplitBySpaces[len(productSplitBySpaces)-2]
}

func parsePrice(product unparsedProduct, unparsedPrice string) (float64, error) {
	fullPrice, err := ustrconv.StringToPositiveFloat(unparsedPrice)
	if err != nil {
		return 0, fmt.Errorf("parse product price: %w", err)
	}
	if product.hasDeposit {
		fullPrice += 0.10
	}

	discount, err := parseDiscount(product.discount)
	if err != nil {
		return 0, fmt.Errorf("parse product discount: %w", err)
	}

	return umath.RoundFloat(fullPrice-discount, 2), nil
}

func parseDiscount(discountLine string) (float64, error) {
	if len(discountLine) == 0 {
		return 0, nil
	}

	discountSplitBySpace := strings.Split(discountLine, " ")
	if len(discountSplitBySpace) < 2 {
		return 0, fmt.Errorf("too short discount line")
	}
	return ustrconv.StringToPositiveFloat(discountSplitBySpace[len(discountSplitBySpace)-2])
}

func trimPriceInfoFromProductName(product, price string) string {
	productWithoutPrice, _, _ := strings.Cut(product, price)
	return productWithoutPrice
}
