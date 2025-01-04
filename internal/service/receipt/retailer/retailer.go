package retailer

import (
	"strings"
	"time"

	"github.com/SarunasBucius/nutri-price-server/internal/model"
	"github.com/SarunasBucius/nutri-price-server/internal/service/receipt/retailer/lidl"
	"github.com/SarunasBucius/nutri-price-server/internal/service/receipt/retailer/norfa"
	"github.com/SarunasBucius/nutri-price-server/internal/utils/uerror"
)

func NewReceiptParser(receipt string) (ReceiptParser, error) {
	switch {
	case strings.Contains(receipt, "UAB NORFOS MAŽMENA"):
		return norfa.NewParser(strings.Split(receipt, "\n")), nil
	case strings.Contains(receipt, "Lidl Lietuva"):
		return lidl.NewParser(strings.Split(receipt, "\n")), nil
	default:
		return nil, uerror.NewBadRequest("unknown retailer", nil)
	}
}

type ReceiptParser interface {
	ParseDate() (time.Time, error)
	ParseProducts() (model.ReceiptProducts, error)
	GetRetailer() string
}
