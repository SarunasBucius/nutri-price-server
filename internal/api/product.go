package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/SarunasBucius/nutri-price-server/internal/model"
	"github.com/SarunasBucius/nutri-price-server/internal/utils/uerror"
)

type ProductAPI struct {
	Service IProductService
}

func NewProductAPI(productService IProductService) *ProductAPI {
	return &ProductAPI{Service: productService}
}

type IProductService interface {
	ConfirmPurchasedProducts(ctx context.Context, retailer, receiptDate string, products []model.PurchasedProductNew) error
}

func (p *ProductAPI) ConfirmPurchasedProducts(w http.ResponseWriter, r *http.Request) {
	var purchasedProducts model.ConfirmPurchasedProductsRequest
	if err := json.NewDecoder(r.Body).Decode(&purchasedProducts); err != nil {
		errorResponse(r.Context(), w, uerror.NewBadRequest("invalid request body", err))
		return
	}

	if err := p.Service.ConfirmPurchasedProducts(r.Context(), purchasedProducts.Retailer, purchasedProducts.Date, purchasedProducts.Products); err != nil {
		errorResponse(r.Context(), w, err)
		return
	}

	successResponse(r.Context(), w, newSuccessMessage("successfully confirmed products"))
}
