package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/SarunasBucius/nutri-price-server/internal/model"
	"github.com/SarunasBucius/nutri-price-server/internal/utils/uerror"
)

type ReceiptAPI struct {
	Service IReceiptService
}

func NewReceiptAPI(receiptService IReceiptService) *ReceiptAPI {
	return &ReceiptAPI{Service: receiptService}
}

type IReceiptService interface {
	ProcessReceipt(ctx context.Context, receipt string) (model.ParseReceiptFromTextResponse, error)
}

func (rc *ReceiptAPI) ParseReceiptFromText(w http.ResponseWriter, r *http.Request) {
	var receipt struct {
		Receipt string `json:"receipt"`
	}
	if err := json.NewDecoder(r.Body).Decode(&receipt); err != nil {
		errorResponse(r.Context(), w, uerror.NewBadRequest("invalid request body", err))
		return
	}

	processedReceipt, err := rc.Service.ProcessReceipt(r.Context(), receipt.Receipt)
	if err != nil {
		errorResponse(r.Context(), w, err)
		return
	}

	successResponse(r.Context(), w, processedReceipt)
}
