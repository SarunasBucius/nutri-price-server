package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/SarunasBucius/nutri-price-server/internal/model"
	"github.com/SarunasBucius/nutri-price-server/internal/utils/uerror"
	"github.com/go-chi/chi/v5"
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
	GetUnconfirmedReceiptSummaries(ctx context.Context) ([]model.UnconfirmedReceiptSummary, error)
	GetUnconfirmedReceipt(ctx context.Context, retailer, date string) ([]model.PurchasedProductNew, error)
	GetLastReceiptDates(ctx context.Context) ([]model.LastReceiptDate, error)
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

func (rc *ReceiptAPI) GetUnconfirmedReceiptSummaries(w http.ResponseWriter, r *http.Request) {
	summaries, err := rc.Service.GetUnconfirmedReceiptSummaries(r.Context())
	if err != nil {
		errorResponse(r.Context(), w, err)
		return
	}

	successResponse(r.Context(), w, emptyIfNil(summaries))
}

func (rc *ReceiptAPI) GetUnconfirmedReceipt(w http.ResponseWriter, r *http.Request) {
	retailerAndDate := chi.URLParam(r, "retailerAndDate")
	retailerAndDateSplit := strings.Split(retailerAndDate, "_")
	if len(retailerAndDateSplit) != 2 {
		errorResponse(r.Context(), w, uerror.NewBadRequest("invalid retailerAndDate query parameter", nil))
		return
	}

	products, err := rc.Service.GetUnconfirmedReceipt(r.Context(), retailerAndDateSplit[0], retailerAndDateSplit[1])
	if err != nil {
		errorResponse(r.Context(), w, err)
		return
	}

	successResponse(r.Context(), w, products)
}

func (rc *ReceiptAPI) GetLastReceiptDates(w http.ResponseWriter, r *http.Request) {
	dates, err := rc.Service.GetLastReceiptDates(r.Context())
	if err != nil {
		errorResponse(r.Context(), w, err)
		return
	}

	successResponse(r.Context(), w, emptyIfNil(dates))
}
