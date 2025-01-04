package api

import (
	"context"
	"encoding/json"
	"io"
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
	ProcessReceiptFromDB(ctx context.Context, receiptDate string) (model.ParseReceiptFromTextResponse, error)
}

func (rc *ReceiptAPI) ParseReceiptFromText(w http.ResponseWriter, r *http.Request) {
	receipt, err := getReceiptFromBody(r)
	if err != nil {
		errorResponse(r.Context(), w, err)
		return
	}

	processedReceipt, err := rc.Service.ProcessReceipt(r.Context(), receipt)
	if err != nil {
		errorResponse(r.Context(), w, err)
		return
	}

	successResponse(r.Context(), w, processedReceipt)
}

func (rc *ReceiptAPI) ParseReceiptInDB(w http.ResponseWriter, r *http.Request) {
	receiptDate := r.URL.Query().Get("receiptDate")

	processedReceipt, err := rc.Service.ProcessReceiptFromDB(r.Context(), receiptDate)
	if err != nil {
		errorResponse(r.Context(), w, err)
		return
	}

	successResponse(r.Context(), w, processedReceipt)
}

func getReceiptFromBody(r *http.Request) (string, error) {
	if r.Header.Get("Content-Type") == "text/plain" {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return "", uerror.NewBadRequest("unable to read text request body", err)
		}
		return string(body), nil
	}

	var receipt struct {
		Receipt string `json:"receipt"`
	}
	if err := json.NewDecoder(r.Body).Decode(&receipt); err != nil {
		return "", uerror.NewBadRequest("invalid json request body", err)
	}

	return receipt.Receipt, nil
}
