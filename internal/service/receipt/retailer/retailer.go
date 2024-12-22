package retailer

import (
	"fmt"
	"strings"
	"time"

	"github.com/SarunasBucius/nutri-price-server/internal/model"
	"github.com/SarunasBucius/nutri-price-server/internal/service/receipt/retailer/norfa"
)

func NewReceiptParser(receipt string) (ReceiptParser, error) {
	if strings.Contains(receipt, "UAB NORFOS MAŽMENA") {
		return norfa.NewParser(strings.Split(receipt, "\n")), nil
	}
	return nil, fmt.Errorf("unknown retailer")
}

type ReceiptParser interface {
	ParseDate() (time.Time, error)
	ParseProducts() (model.ReceiptProducts, error)
	GetRetailer() string
}
