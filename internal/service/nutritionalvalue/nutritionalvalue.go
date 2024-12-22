package nutritionalvalue

import (
	"context"
	"fmt"

	"github.com/SarunasBucius/nutri-price-server/internal/model"
)

type Service struct {
	NutritionalValueRepo INutritionalValueRepository
}

func NewNutritionalValueService(nutritionalValueRepo INutritionalValueRepository) *Service {
	return &Service{
		NutritionalValueRepo: nutritionalValueRepo,
	}
}

type INutritionalValueRepository interface {
	InsertProductNutritionalValue(ctx context.Context, product, measurementUnit string, nv model.NutritionalValue) error
	GetProductsNutritionalValue(ctx context.Context) ([]model.ProductNutritionalValue, error)
	GetProductsNutritionalValueByProductNames(ctx context.Context, productNames []string) ([]model.ProductNutritionalValue, error)
	GetProductNutritionalValue(ctx context.Context, nutritionalValueID int) (model.ProductNutritionalValue, error)
	UpdateProductNutritionalValue(ctx context.Context, productNV model.ProductNutritionalValue) error
	DeleteProductNutritionalValue(ctx context.Context, id int) error
}

func (s *Service) InsertNutritionalValue(ctx context.Context, pnv model.ProductNutritionalValueNew) error {
	if err := s.NutritionalValueRepo.InsertProductNutritionalValue(ctx, pnv.Product, pnv.Unit, pnv.NutritionalValue); err != nil {
		return fmt.Errorf("insert product nutritional value: %w", err)
	}
	return nil
}

func (s *Service) GetProductsNutritionalValue(ctx context.Context, products []string) ([]model.ProductNutritionalValue, error) {
	if len(products) == 0 {
		productsNV, err := s.NutritionalValueRepo.GetProductsNutritionalValue(ctx)
		if err != nil {
			return nil, fmt.Errorf("get products nutritional value: %w", err)
		}
		return productsNV, nil
	}
	productsNVByName, err := s.NutritionalValueRepo.GetProductsNutritionalValueByProductNames(ctx, products)
	if err != nil {
		return nil, fmt.Errorf("get products nutritional value by product names: %w", err)
	}

	return productsNVByName, nil
}

func (s *Service) GetProductNutritionalValue(ctx context.Context, nvID int) (model.ProductNutritionalValue, error) {
	productNV, err := s.NutritionalValueRepo.GetProductNutritionalValue(ctx, nvID)
	if err != nil {
		return model.ProductNutritionalValue{}, fmt.Errorf("get product nutritional value by: %w", err)
	}
	return productNV, nil
}

func (s *Service) UpdateProductNutritionalValue(ctx context.Context, productNV model.ProductNutritionalValue) error {
	if err := s.NutritionalValueRepo.UpdateProductNutritionalValue(ctx, productNV); err != nil {
		return fmt.Errorf("update product nutritional value: %w", err)
	}
	return nil
}

func (s *Service) DeleteProductNutritionalValue(ctx context.Context, nvID int) error {
	if err := s.NutritionalValueRepo.DeleteProductNutritionalValue(ctx, nvID); err != nil {
		return fmt.Errorf("delete product nutritional value: %w", err)
	}
	return nil
}
