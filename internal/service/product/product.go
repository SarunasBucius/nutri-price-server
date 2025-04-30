package product

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/SarunasBucius/nutri-price-server/internal/model"
)

type Service struct {
	ProductRepo          IProductRepository
	ReceiptRepo          IReceiptRepository
	NutritionalValueRepo INutritionalValueRepository
}

func NewProductService(productRepo IProductRepository, receiptRepo IReceiptRepository, nutritionalValueRepo INutritionalValueRepository) *Service {
	return &Service{
		ProductRepo:          productRepo,
		ReceiptRepo:          receiptRepo,
		NutritionalValueRepo: nutritionalValueRepo,
	}
}

type IProductRepository interface {
	InsertPurchases(ctx context.Context, retailer string, receiptDate time.Time, products []model.PurchasedProductNew) error
	InsertProducts(ctx context.Context, productNames []string) error
	GetProductIDsByName(ctx context.Context, productNames []string) (map[string]string, error)
}

type IReceiptRepository interface {
	SetRawReceiptSubmittedProducts(ctx context.Context, receiptDate time.Time, submittedProducts []model.PurchasedProductNew) error
	UpdateProductNameAlias(ctx context.Context, editedNameByParsedName map[string]model.ProductAndVarietyName) error
	ConfirmReceipts(ctx context.Context, retailer, date string) error
}

type INutritionalValueRepository interface {
	InsertEmptyProducts(ctx context.Context, products []string) error
}

func (s *Service) InsertProducts(ctx context.Context, retailer, receiptDate string, purchases []model.PurchasedProductNew) error {
	date, err := time.Parse(time.DateOnly, receiptDate)
	if err != nil {
		return err
	}

	productNames := make([]string, 0, len(purchases))
	for _, product := range purchases {
		productNames = append(productNames, product.Name)
	}

	err = s.ProductRepo.InsertProducts(ctx, productNames)
	if err != nil {
		return fmt.Errorf("insert products: %w", err)
	}

	productIDs, err := s.ProductRepo.GetProductIDsByName(ctx, productNames)
	if err != nil {
		return fmt.Errorf("get product IDs by name: %w", err)
	}

	for i, product := range purchases {
		if id, ok := productIDs[product.Name]; ok {
			purchases[i].ProductID = id
			continue
		}
		return fmt.Errorf("product ID not found for name %s", product.Name)

	}

	if err := s.ProductRepo.InsertPurchases(ctx, retailer, date, purchases); err != nil {
		return fmt.Errorf("insert purchases: %w", err)
	}

	if err := s.ReceiptRepo.SetRawReceiptSubmittedProducts(ctx, date, purchases); err != nil {
		slog.ErrorContext(ctx, "set raw receipt submitted products", "error", err)
	}

	return nil
}

func (s *Service) ConfirmPurchasedProducts(ctx context.Context, retailer, receiptDate string, products []model.PurchasedProductNew) error {
	editedNameByParsedName := make(map[string]model.ProductAndVarietyName, len(products))
	for _, product := range products {
		editedNameByParsedName[product.ParsedName] = model.ProductAndVarietyName{
			Name:        product.Name,
			VarietyName: product.VarietyName,
		}
	}

	if err := s.InsertProducts(ctx, retailer, receiptDate, products); err != nil {
		return fmt.Errorf("insert products: %w", err)
	}

	if err := s.ReceiptRepo.UpdateProductNameAlias(ctx, editedNameByParsedName); err != nil {
		slog.ErrorContext(ctx, "update product name alias", "error", err)
	}

	if err := s.ReceiptRepo.ConfirmReceipts(ctx, retailer, receiptDate); err != nil {
		return fmt.Errorf("confirm receipt: %w", err)
	}
	return nil
}
