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
	InsertProducts(ctx context.Context, retailer string, receiptDate time.Time, products []model.PurchasedProductNew) error
	GetProductGroups(ctx context.Context) ([]string, error)
	GetProductsByGroup(ctx context.Context, productGroups []string) ([]model.PurchasedProduct, error)
	GetProduct(ctx context.Context, productID int) (model.PurchasedProduct, error)
	UpdateProduct(ctx context.Context, product model.PurchasedProduct) error
	DeleteProduct(ctx context.Context, productID int) error
}

type IReceiptRepository interface {
	SetRawReceiptSubmittedProducts(ctx context.Context, receiptDate time.Time, submittedProducts []model.PurchasedProductNew) error
	UpdateProductNameAlias(ctx context.Context, editedNameByParsedName map[string]string) error
	ConfirmReceipts(ctx context.Context, retailer, date string) error
}

type INutritionalValueRepository interface {
	InsertEmptyProducts(ctx context.Context, products []string) error
}

func (s *Service) InsertProducts(ctx context.Context, retailer, receiptDate string, products []model.PurchasedProductNew) error {
	if err := s.insertEmptyProducts(ctx, products); err != nil {
		return fmt.Errorf("insert empty products: %w", err)
	}

	date, err := time.Parse(time.DateOnly, receiptDate)
	if err != nil {
		return err
	}

	if err := s.ProductRepo.InsertProducts(ctx, retailer, date, products); err != nil {
		return fmt.Errorf("insert products: %w", err)
	}

	if err := s.ReceiptRepo.SetRawReceiptSubmittedProducts(ctx, date, products); err != nil {
		slog.ErrorContext(ctx, "set raw receipt submitted products", "error", err)
	}

	return nil
}

func (s *Service) GetProductGroups(ctx context.Context) ([]string, error) {
	productGroups, err := s.ProductRepo.GetProductGroups(ctx)
	if err != nil {
		return nil, fmt.Errorf("get product groups: %w", err)
	}

	return productGroups, nil
}

func (s *Service) GetProducts(ctx context.Context, productGroups []string) ([]model.PurchasedProduct, error) {
	if len(productGroups) == 0 {
		return nil, nil
	}
	products, err := s.ProductRepo.GetProductsByGroup(ctx, productGroups)
	if err != nil {
		return nil, fmt.Errorf("get products by group: %w", err)
	}
	return products, nil
}

func (s *Service) GetProduct(ctx context.Context, productID int) (model.PurchasedProduct, error) {
	product, err := s.ProductRepo.GetProduct(ctx, productID)
	if err != nil {
		return model.PurchasedProduct{}, fmt.Errorf("get product: %w", err)
	}
	return product, nil
}

func (s *Service) UpdateProduct(ctx context.Context, product model.PurchasedProduct) error {
	if err := s.NutritionalValueRepo.InsertEmptyProducts(ctx, []string{product.Name}); err != nil {
		return fmt.Errorf("insert empty products: %w", err)
	}

	if err := s.ProductRepo.UpdateProduct(ctx, product); err != nil {
		return fmt.Errorf("update product: %w", err)
	}
	return nil
}

func (s *Service) DeleteProduct(ctx context.Context, productID int) error {
	if err := s.ProductRepo.DeleteProduct(ctx, productID); err != nil {
		return fmt.Errorf("delete product: %w", err)
	}
	return nil
}

func (s *Service) ConfirmPurchasedProducts(ctx context.Context, retailer, receiptDate string, products []model.PurchasedProductNew) error {
	editedNameByParsedName := make(map[string]string, len(products))
	for _, product := range products {
		editedNameByParsedName[product.ParsedName] = product.Name
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

func (s *Service) insertEmptyProducts(ctx context.Context, products []model.PurchasedProductNew) error {
	productNames := make([]string, 0, len(products))
	for _, product := range products {
		productNames = append(productNames, product.Name)
	}

	if err := s.NutritionalValueRepo.InsertEmptyProducts(ctx, productNames); err != nil {
		return fmt.Errorf("insert empty products: %w", err)
	}
	return nil
}
