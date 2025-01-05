package lidl

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/SarunasBucius/nutri-price-server/internal/model"
	"github.com/SarunasBucius/nutri-price-server/internal/utils/umath"
)

const retailer = "lidl"

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
	product       string
	hasDeposit    bool
	discount      string
	isHalf        bool
	dynamicWeight []string
}

func (p Parser) ParseDate() (time.Time, error) {
	const datePositionFromEnd = 2

	if len(p.ReceiptLines) < 1 {
		return time.Time{}, fmt.Errorf("unexpected receipt length")
	}

	dateLine := p.ReceiptLines[len(p.ReceiptLines)-1]

	dateLineSplitBySpace := strings.Split(dateLine, " ")

	if len(dateLineSplitBySpace) < datePositionFromEnd {
		return time.Time{}, fmt.Errorf("unexpected date line contents: %s", dateLine)
	}

	parsedDate, err := time.Parse(time.DateOnly, dateLineSplitBySpace[len(dateLineSplitBySpace)-datePositionFromEnd])
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
	const productsEndSeparator = "------------------------------------------------------"
	const linesBeforeProductsList = 4
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

func extractProduct(line string, products []unparsedProduct) []unparsedProduct {
	if len(products) == 0 {
		return appendProduct(line, products)
	}

	lastProduct := len(products) - 1

	if isDeposit(line) {
		products[lastProduct].hasDeposit = true
		return products
	}

	if products[lastProduct].isHalf {
		lineSplitBySpace := strings.Split(line, " ")
		lineSplitBySpace = slices.DeleteFunc(lineSplitBySpace, func(l string) bool {
			return l == ""
		})
		isDynamic := len(lineSplitBySpace) == 6 && lineSplitBySpace[1] == "X"
		if isDynamic {
			products[lastProduct].dynamicWeight = lineSplitBySpace
			products[lastProduct].isHalf = false
			return products
		}
		products[lastProduct].product += " " + line
		products[lastProduct].isHalf = !strings.HasSuffix(line, "A")
		return products
	}

	if startsWithNumericCode(line) {
		return appendProduct(line, products)
	}

	if isDiscount(line) {
		products[lastProduct].discount = line
		return products
	}

	return products
}

func isDeposit(line string) bool {
	return strings.Contains(line, "Užstatas")
}

func startsWithNumericCode(line string) bool {
	if len(line) < 7 {
		return false
	}
	_, err := strconv.Atoi(line[:7])
	if err == nil {
		return true
	}
	return false
}

func appendProduct(productLine string, products []unparsedProduct) []unparsedProduct {
	if strings.HasSuffix(productLine, "A") {
		return append(products, unparsedProduct{product: productLine})
	}

	return append(products, unparsedProduct{product: productLine, isHalf: true})
}

func isDiscount(product string) bool {
	lowerCaseProduct := strings.ToLower(product)
	return strings.Contains(lowerCaseProduct, "nuolaida") && strings.Contains(lowerCaseProduct, "a")
}

func parseProduct(product unparsedProduct) (model.ReceiptProduct, error) {
	const irrelevantProductPrefixLength = 7
	product.product = product.product[irrelevantProductPrefixLength:]

	unparsedPrice := getUnparsedPrice(product)
	price, err := parsePrice(product, unparsedPrice)
	if err != nil {
		return model.ReceiptProduct{}, err
	}

	productName := trimPriceInfoFromProductName(product.product, unparsedPrice)

	quantity, err := getQuantity(product)
	if err != nil {
		return model.ReceiptProduct{}, fmt.Errorf("extract quantity info: %w", err)
	}

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

func getQuantity(product unparsedProduct) (model.Quantity, error) {
	if len(product.dynamicWeight) != 6 {
		return model.Quantity{}, nil
	}

	amount, err := strconv.ParseFloat(strings.Replace(product.dynamicWeight[2], ",", ".", 1), 64)
	if err != nil {
		return model.Quantity{}, fmt.Errorf("parse product amount: %w", err)
	}

	switch product.dynamicWeight[3] {
	case "vnt.":
		return model.Quantity{
			Amount: amount,
			Unit:   model.Pieces,
		}, nil
	case "KG":
		return model.Quantity{
			Unit:   model.Grams,
			Amount: amount * 1000,
		}, nil
	default:
		return model.Quantity{}, nil
	}
}

func getUnparsedPrice(product unparsedProduct) string {
	if len(product.dynamicWeight) >= 2 {
		return product.dynamicWeight[len(product.dynamicWeight)-2]
	}
	productSplitBySpaces := strings.Split(product.product, " ")
	return productSplitBySpaces[len(productSplitBySpaces)-2]
}

func parsePrice(product unparsedProduct, unparsedPrice string) (model.Price, error) {
	fullPrice, err := stringToPositiveFloat(unparsedPrice)
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
	return stringToPositiveFloat(discountSplitBySpace[len(discountSplitBySpace)-2])
}

func trimPriceInfoFromProductName(product, price string) string {
	productWithoutPrice, _, _ := strings.Cut(product, price)
	return productWithoutPrice
}

// TODO move to utils
func stringToPositiveFloat(num string) (float64, error) {
	num = strings.TrimLeft(num, "-")
	numWithoutComma := strings.Replace(num, ",", ".", 1)
	parsedNum, err := strconv.ParseFloat(numWithoutComma, 32)
	if err != nil {
		return 0, fmt.Errorf("parse string price to float: %w", err)
	}
	return parsedNum, nil
}
