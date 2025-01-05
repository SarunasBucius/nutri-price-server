package norfa

import (
	"fmt"
	"strings"
	"time"

	"github.com/SarunasBucius/nutri-price-server/internal/model"
	"github.com/SarunasBucius/nutri-price-server/internal/utils/umath"
	"github.com/SarunasBucius/nutri-price-server/internal/utils/ustrconv"
)

const retailer = "norfa"

type Parser struct {
	ReceiptLines []string
	Retailer     string
}

func NewParser(receiptLines []string) Parser {
	return Parser{
		ReceiptLines: receiptLines,
		Retailer:     retailer,
	}
}

type unparsedProduct struct {
	product    string
	hasDeposit bool
	discount   string
	isHalf     bool
}

func (p Parser) ParseDate() (time.Time, error) {
	const dateCharactersNum = len(time.DateOnly)

	if len(p.ReceiptLines) < 2 {
		return time.Time{}, fmt.Errorf("unexpected receipt length")
	}

	dateLine := p.ReceiptLines[len(p.ReceiptLines)-2]
	if len(dateLine) < dateCharactersNum {
		return time.Time{}, fmt.Errorf("unexpected date line length")
	}
	receiptDate := dateLine[:dateCharactersNum]

	parsedDate, err := time.Parse(time.DateOnly, receiptDate)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse receipt date: %w", err)
	}
	return parsedDate, nil
}

func (p Parser) ParseProducts() (model.ReceiptProducts, error) {
	unparsedProducts, err := extractProductLines(p.ReceiptLines)
	if err != nil {
		return nil, fmt.Errorf("extract product lines: %w", err)
	}

	parsedProducts := make([]model.ReceiptProduct, 0, len(unparsedProducts))
	for _, product := range unparsedProducts {
		parsedProduct, err := parseProduct(product)
		if err != nil {
			return nil, fmt.Errorf("parse product %+v: %w", product, err)
		}
		parsedProducts = append(parsedProducts, parsedProduct)
	}
	return parsedProducts, nil
}

func (p Parser) GetRetailer() string { return retailer }

func extractProductLines(receiptLines []string) ([]unparsedProduct, error) {
	const productsEndSeparator = "#"
	const linesBeforeProductsList = 6
	if len(receiptLines) <= linesBeforeProductsList {
		return nil, fmt.Errorf("too short receipt")
	}
	receiptLines = receiptLines[linesBeforeProductsList:]

	var products []unparsedProduct
	for i := range receiptLines {
		receiptLineWithoutCarriage := strings.ReplaceAll(receiptLines[i], "\r", "")

		if strings.HasSuffix(receiptLineWithoutCarriage, productsEndSeparator) {
			break
		}

		products = extractProduct(receiptLineWithoutCarriage, products)
	}

	return products, nil
}

func parseProduct(product unparsedProduct) (model.ReceiptProduct, error) {
	unparsedPrice := getUnparsedPrice(product.product)
	price, err := parsePrice(product, unparsedPrice)
	if err != nil {
		return model.ReceiptProduct{}, err
	}

	productName := trimPriceInfoFromProductName(product.product, unparsedPrice)

	weightParser := newWeightParser(productName)
	quantity, err := weightParser.getQuantity()
	if err != nil {
		return model.ReceiptProduct{}, fmt.Errorf("extract quantity info: %w", err)
	}
	productName = weightParser.trimProductName()

	return model.ReceiptProduct{
		ProductLineInReceipt: product.product,
		PurchasedProductNew: model.PurchasedProductNew{
			Name:     strings.TrimSpace(productName),
			Price:    price,
			Quantity: quantity,
			// Group and notes will be filled from DB later.
			Group: "",
			Notes: "",
		},
	}, nil
}

func getUnparsedPrice(product string) string {
	productSplitBySpaces := strings.Split(product, " ")
	return productSplitBySpaces[len(productSplitBySpaces)-2]
}

func parsePrice(product unparsedProduct, unparsedPrice string) (model.Price, error) {
	fullPrice, err := ustrconv.StringToPositiveFloat(unparsedPrice)
	if err != nil {
		return model.Price{}, fmt.Errorf("parse product price: %w", err)
	}
	if product.hasDeposit {
		fullPrice -= 0.10
	}

	discount, err := parseDiscount(product.discount)
	if err != nil {
		return model.Price{}, fmt.Errorf("parse product discount: %w", err)
	}

	return model.Price{
		Full:     umath.RoundFloat(fullPrice, 2),
		Discount: umath.RoundFloat(discount, 2),
		Paid:     umath.RoundFloat(fullPrice-discount, 2),
	}, nil
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

func extractProduct(line string, products []unparsedProduct) []unparsedProduct {
	line = strings.TrimLeft(line, "*")

	if len(products) == 0 {
		return appendProduct(line, products)
	}

	lastProduct := len(products) - 1

	if isDeposit(line) {
		products[lastProduct].hasDeposit = true
		return products
	}

	if isDiscount(line) {
		products[lastProduct].discount = line
		return products
	}

	if products[lastProduct].isHalf {
		products[lastProduct].product += " " + line
		products[lastProduct].isHalf = false
		return products
	}

	return appendProduct(line, products)
}

func appendProduct(productLine string, products []unparsedProduct) []unparsedProduct {
	if strings.HasSuffix(productLine, "M1") {
		return append(products, unparsedProduct{product: productLine})
	}

	return append(products, unparsedProduct{product: productLine, isHalf: true})
}

func isDeposit(product string) bool {
	lowerCaseProduct := strings.ToLower(product)
	return strings.Contains(lowerCaseProduct, "užstatas už pakuotę")
}

func isDiscount(product string) bool {
	lowerCaseProduct := strings.ToLower(product)
	return strings.Contains(lowerCaseProduct, "nuolaida") && strings.Contains(lowerCaseProduct, "eur")
}
