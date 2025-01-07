package retailer

import (
	"slices"
	"strings"
	"time"

	"github.com/SarunasBucius/nutri-price-server/internal/model"
	"github.com/SarunasBucius/nutri-price-server/internal/service/receipt/retailer/lidl"
	"github.com/SarunasBucius/nutri-price-server/internal/service/receipt/retailer/norfa"
	"github.com/SarunasBucius/nutri-price-server/internal/utils/uerror"
)

func NewReceiptParser(receipt string) (ReceiptParser, error) {
	receiptLines := strings.Split(receipt, "\n")
	receiptLines = slices.DeleteFunc(receiptLines, func(l string) bool {
		return l == "" || l == "\r"
	})
	switch {
	case strings.Contains(receipt, "UAB NORFOS MAŽMENA"):
		return norfa.NewParser(receiptLines), nil
	case strings.Contains(receipt, "Lidl Lietuva"):
		return lidl.NewParser(receiptLines), nil
	default:
		return nil, uerror.NewBadRequest("unknown retailer", nil)
	}
}

type ReceiptParser interface {
	ParseDate() (time.Time, error)
	ParseProducts() (model.ReceiptProducts, error)
	GetRetailer() string
}
