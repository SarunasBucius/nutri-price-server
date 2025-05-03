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

func (p *ReceiptProducts) UpdateProductNames(aliasByParsedName map[string]ProductAndVarietyName) {
	if p == nil || len(aliasByParsedName) == 0 {
		return
	}

	for i := range *p {
		product := *p
		product[i].ParsedName = product[i].VarietyName
		if alias, ok := aliasByParsedName[product[i].Name]; ok {
			product[i].Name = alias.Name
			product[i].VarietyName = alias.VarietyName
		}
	}
}

type ParseReceiptFromTextResponse struct {
	Date     string          `json:"date"`
	Retailer string          `json:"retailer"`
	Products ReceiptProducts `json:"products"`
}

type LastReceiptDate struct {
	Date     string `json:"date"`
	Retailer string `json:"retailer"`
}
