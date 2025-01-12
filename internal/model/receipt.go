package model

type ReceiptProducts []PurchasedProductNew

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

	for i := range *p {
		product := *p
		dbProduct := productsByName[product[i].Name]
		product[i].Group = dbProduct.Group
		product[i].Notes = dbProduct.Notes
	}
}

type ParseReceiptFromTextResponse struct {
	Date     string          `json:"date"`
	Retailer string          `json:"retailer"`
	Products ReceiptProducts `json:"products"`
}
