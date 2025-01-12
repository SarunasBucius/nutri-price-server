package barbora

import (
	"fmt"
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

type unparsedProduct struct {
	product       string
	hasDeposit    bool
	discount      string
	isHalf        bool
	dynamicWeight []string
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

	discountedProducts := extractDiscountedProducts(p.ReceiptLines)

	discountsByProduct, err := getDiscountsByProduct(discountedProducts)
	if err != nil {
		return nil, fmt.Errorf("get discounts by product: %w", err)
	}

	parsedProducts := make([]model.PurchasedProductNew, 0, len(unparsedProducts))
	for _, product := range unparsedProducts {
		parsedProduct, err := parseProduct(product, discountsByProduct)
		if err != nil {
			return nil, fmt.Errorf("parse product %+v: %w", product, err)
		}
		parsedProducts = append(parsedProducts, parsedProduct)
	}
	return parsedProducts, nil
}

func (p BarboraParser) GetRetailer() string { return retailer }

func parseProduct(product unparsedProduct, discountsByProduct map[string]string) (model.PurchasedProductNew, error) {
	productSplitBySpace := strings.Split(product.product, " ")
	if len(productSplitBySpace) < 9 {
		return model.PurchasedProductNew{}, fmt.Errorf("unexpected product line: %s", product)
	}

	productName := strings.Join(productSplitBySpace[1:len(productSplitBySpace)-7], " ")

	unparsedPrice := strings.TrimPrefix(productSplitBySpace[len(productSplitBySpace)-1], "€")
	price, err := parsePrice(unparsedPrice, discountsByProduct[productName])
	if err != nil {
		return model.PurchasedProductNew{}, fmt.Errorf("parse price: %w", err)
	}

	amount := productSplitBySpace[len(productSplitBySpace)-7]
	unit := productSplitBySpace[len(productSplitBySpace)-6]
	quantity, err := getQuantity(amount, unit)
	if err != nil {
		return model.PurchasedProductNew{}, fmt.Errorf("get quantity: %w", err)
	}

	return model.PurchasedProductNew{
			Name:     productName,
			Price:    price,
			Quantity: quantity,
			// Group and notes will be filled from DB later.
			Group: "",
			Notes: "",
	}, nil
}

func getQuantity(amount, unit string) (model.Quantity, error) {
	amountFloat, err := ustrconv.StringToPositiveFloat(amount)
	if err != nil {
		return model.Quantity{}, fmt.Errorf("parse product amount: %w", err)
	}

	switch strings.ToLower(unit) {
	case "vnt.":
		return model.Quantity{
			Amount: amountFloat,
			Unit:   model.Pieces,
		}, nil
	case "kg":
		return model.Quantity{
			Unit:   model.Grams,
			Amount: umath.RoundFloat(amountFloat*1000, 0),
		}, nil
	default:
		return model.Quantity{}, nil
	}
}

func parsePrice(unparsedPrice, unparsedDiscount string) (model.Price, error) {
	paid, err := ustrconv.StringToPositiveFloat(unparsedPrice)
	if err != nil {
		return model.Price{}, fmt.Errorf("parse product price: %w", err)
	}

	discount, err := parseDiscount(unparsedDiscount)
	if err != nil {
		return model.Price{}, fmt.Errorf("parse product discount: %w", err)
	}

	return model.Price{
		Full:     umath.RoundFloat(paid+discount, 2),
		Discount: umath.RoundFloat(discount, 2),
		Paid:     umath.RoundFloat(paid, 2),
	}, nil
}

func parseDiscount(discountLine string) (float64, error) {
	if len(discountLine) == 0 {
		return 0, nil
	}

	discount := strings.TrimPrefix(discountLine, "-€")
	return ustrconv.StringToPositiveFloat(discount)
}

func extractProductLines(receiptLines []string) ([]unparsedProduct, error) {
	const productsEndSeparator = "Pritaikytos nuolaidos"
	const productsListStart = 2
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

func extractProduct(line string, products []unparsedProduct) []unparsedProduct {
	if len(products) == 0 {
		return append(products, unparsedProduct{product: line})
	}

	productsNum := len(products)

	if strings.HasPrefix(line, strconv.Itoa(productsNum+1)) {
		return append(products, unparsedProduct{product: line})
	}

	products[productsNum-1].product += " " + line

	return products
}
