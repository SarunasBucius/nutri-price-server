package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/SarunasBucius/nutri-price-server/internal/model"
	"github.com/SarunasBucius/nutri-price-server/internal/utils/uerror"
	"github.com/go-chi/chi/v5"
)

type NutritionalValueAPI struct {
	Service INutritionalValueService
}

func NewNutritionalValuesAPI(nutritionalValuesService INutritionalValueService) *NutritionalValueAPI {
	return &NutritionalValueAPI{Service: nutritionalValuesService}
}

type INutritionalValueService interface {
	InsertNutritionalValue(ctx context.Context, productNV model.ProductNutritionalValueNew) error
	GetProductsNutritionalValue(ctx context.Context, products []string) ([]model.ProductNutritionalValue, error)
	GetProductNutritionalValue(ctx context.Context, nvID int) (model.ProductNutritionalValue, error)
	UpdateProductNutritionalValue(ctx context.Context, productNV model.ProductNutritionalValue) error
	DeleteProductNutritionalValue(ctx context.Context, nvID int) error
	GetNutritionalValuesUnits(ctx context.Context) ([]model.NutritionalValueUnits, error)
}

func (n *NutritionalValueAPI) InsertNutritionalValues(w http.ResponseWriter, r *http.Request) {
	var productNV model.ProductNutritionalValueNew
	if err := json.NewDecoder(r.Body).Decode(&productNV); err != nil {
		errorResponse(r.Context(), w, uerror.NewBadRequest("invalid request body", err))
		return
	}

	if err := n.Service.InsertNutritionalValue(r.Context(), productNV); err != nil {
		errorResponse(r.Context(), w, err)
		return
	}

	successResponse(r.Context(), w, newSuccessMessage("successfully inserted product nutritional value"))
}

func (n *NutritionalValueAPI) GetNutritionalValues(w http.ResponseWriter, r *http.Request) {
	productNamesQuery := r.URL.Query().Get("productNames")

	var productNames []string
	if len(productNamesQuery) != 0 {
		productNames = strings.Split(productNamesQuery, ",")
	}

	nutritionalValues, err := n.Service.GetProductsNutritionalValue(r.Context(), productNames)
	if err != nil {
		errorResponse(r.Context(), w, err)
		return
	}

	successResponse(r.Context(), w, map[string]any{"nutritionalValues": emptyIfNil(nutritionalValues)})
}

func (n *NutritionalValueAPI) GetNutritionalValuesUnits(w http.ResponseWriter, r *http.Request) {
	nvUnits, err := n.Service.GetNutritionalValuesUnits(r.Context())
	if err != nil {
		errorResponse(r.Context(), w, err)
		return
	}

	successResponse(r.Context(), w, map[string]any{"filledProducts": emptyIfNil(nvUnits)})
}

func (n *NutritionalValueAPI) GetNutritionalValue(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "nutritionalValueID")

	id, err := strconv.Atoi(idParam)
	if err != nil {
		errorResponse(r.Context(), w, uerror.NewBadRequest("invalid id", err))
		return
	}

	nutritionalValue, err := n.Service.GetProductNutritionalValue(r.Context(), id)
	if err != nil {
		errorResponse(r.Context(), w, err)
		return
	}

	successResponse(r.Context(), w, nutritionalValue)
}

func (n *NutritionalValueAPI) UpdateNutritionalValue(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "nutritionalValueID")

	id, err := strconv.Atoi(idParam)
	if err != nil {
		errorResponse(r.Context(), w, uerror.NewBadRequest("invalid id", err))
		return
	}

	var productNV model.ProductNutritionalValue
	if err := json.NewDecoder(r.Body).Decode(&productNV); err != nil {
		errorResponse(r.Context(), w, uerror.NewBadRequest("invalid request body", err))
		return
	}
	productNV.ID = id

	if err := n.Service.UpdateProductNutritionalValue(r.Context(), productNV); err != nil {
		errorResponse(r.Context(), w, err)
		return
	}

	successResponse(r.Context(), w, newSuccessMessage("successfully updated product nutritional value"))
}

func (n *NutritionalValueAPI) DeleteNutritionalValues(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "nutritionalValueID")

	id, err := strconv.Atoi(idParam)
	if err != nil {
		errorResponse(r.Context(), w, uerror.NewBadRequest("invalid id", err))
		return
	}

	if err := n.Service.DeleteProductNutritionalValue(r.Context(), id); err != nil {
		errorResponse(r.Context(), w, err)
		return
	}

	successResponse(r.Context(), w, newSuccessMessage("successfully deleted product nutritional value"))
}
