package model

type ReceiptProduct struct {
	PurchasedProductNew
	ProductLineInReceipt string `json:"productLineInReceipt"`
}

type ReceiptProducts []ReceiptProduct

func (p *ReceiptProducts) GetNames() []string {
	if p == nil {
		return nil
	}

	productNames := make([]string, 0, len(*p))
	for _, product := range *p {
		productNames = append(productNames, product.Name)
	}

	return productNames
}

func (p *ReceiptProducts) FillCategoriesAndNotes(productsByName map[string]PurchasedProduct) {
	if p == nil || len(productsByName) == 0 {
		return
	}

	for _, product := range *p {
		dbProduct := productsByName[product.Name]
		product.Group = dbProduct.Group
		product.Notes = dbProduct.Notes
	}
}

type ParseReceiptFromTextResponse struct {
	Date     string          `json:"date"`
	Retailer string          `json:"retailer"`
	Products ReceiptProducts `json:"products"`
}
