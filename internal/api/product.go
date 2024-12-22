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

type ProductAPI struct {
	Service IProductService
}

func NewProductAPI(productService IProductService) *ProductAPI {
	return &ProductAPI{Service: productService}
}

type IProductService interface {
	InsertProducts(ctx context.Context, retailer, receiptDate string, products []model.PurchasedProductNew) error
	GetProductGroups(ctx context.Context) ([]string, error)
	GetProducts(ctx context.Context, productGroups []string) ([]model.PurchasedProduct, error)
	GetProduct(ctx context.Context, productID int) (model.PurchasedProduct, error)
	UpdateProduct(ctx context.Context, product model.PurchasedProduct) error
	DeleteProduct(ctx context.Context, productID int) error
}

func (p *ProductAPI) InsertProducts(w http.ResponseWriter, r *http.Request) {
	var purchasedProducts model.PostPurchasedProductsRequest
	if err := json.NewDecoder(r.Body).Decode(&purchasedProducts); err != nil {
		errorResponse(r.Context(), w, uerror.NewBadRequest("invalid request body", err))
		return
	}

	if err := p.Service.InsertProducts(r.Context(), purchasedProducts.Retailer, purchasedProducts.Date, purchasedProducts.PurchasedProducts); err != nil {
		errorResponse(r.Context(), w, err)
		return
	}

	successResponse(r.Context(), w, newSuccessMessage("successfully inserted products"))
}

func (p *ProductAPI) GetProductGroups(w http.ResponseWriter, r *http.Request) {
	productGroups, err := p.Service.GetProductGroups(r.Context())
	if err != nil {
		errorResponse(r.Context(), w, err)
		return
	}

	successResponse(r.Context(), w, map[string]any{"productGroups": emptyIfNil(productGroups)})
}

func (p *ProductAPI) GetProducts(w http.ResponseWriter, r *http.Request) {
	productGroupsQuery := r.URL.Query().Get("productGroups")
	if len(productGroupsQuery) == 0 {
		errorResponse(r.Context(), w, uerror.NewBadRequest("specify productGroups query parameter", nil))
		return
	}
	productGroups := strings.Split(productGroupsQuery, ",")

	products, err := p.Service.GetProducts(r.Context(), productGroups)
	if err != nil {
		errorResponse(r.Context(), w, err)
		return
	}

	successResponse(r.Context(), w, map[string]any{"products": emptyIfNil(products)})
}

func (p *ProductAPI) GetProduct(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "productID")

	id, err := strconv.Atoi(idParam)
	if err != nil {
		errorResponse(r.Context(), w, uerror.NewBadRequest("invalid id", err))
		return
	}

	product, err := p.Service.GetProduct(r.Context(), id)
	if err != nil {
		errorResponse(r.Context(), w, err)
		return
	}

	successResponse(r.Context(), w, product)
}

func (p *ProductAPI) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "productID")

	id, err := strconv.Atoi(idParam)
	if err != nil {
		errorResponse(r.Context(), w, uerror.NewBadRequest("invalid id", err))
		return
	}

	var purchasedProduct model.PurchasedProduct
	if err := json.NewDecoder(r.Body).Decode(&purchasedProduct); err != nil {
		errorResponse(r.Context(), w, uerror.NewBadRequest("invalid request body", err))
		return
	}

	purchasedProduct.ID = id
	if err := p.Service.UpdateProduct(r.Context(), purchasedProduct); err != nil {
		errorResponse(r.Context(), w, err)
		return
	}

	successResponse(r.Context(), w, newSuccessMessage("successfully updated product"))
}

func (p *ProductAPI) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "productID")

	id, err := strconv.Atoi(idParam)
	if err != nil {
		errorResponse(r.Context(), w, uerror.NewBadRequest("invalid id", err))
		return
	}

	if err := p.Service.DeleteProduct(r.Context(), id); err != nil {
		errorResponse(r.Context(), w, err)
		return
	}

	successResponse(r.Context(), w, newSuccessMessage("successfully deleted product"))
}
