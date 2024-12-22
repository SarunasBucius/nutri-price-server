package receipt

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/SarunasBucius/nutri-price-server/internal/model"
	"github.com/SarunasBucius/nutri-price-server/internal/service/receipt/retailer"
)

type Service struct {
	ReceiptRepo IReceiptRepository
	ProductRepo IProductRepository
}

func NewReceiptService(receiptRepo IReceiptRepository, productRepo IProductRepository) *Service {
	return &Service{
		ReceiptRepo: receiptRepo,
		ProductRepo: productRepo,
	}
}

type IReceiptRepository interface {
	InsertRawReceipt(ctx context.Context, date time.Time, receiptLines string, products model.ReceiptProducts) error
}

type IProductRepository interface {
	GetDistinctProductsByNames(ctx context.Context, productNames []string) (map[string]model.PurchasedProduct, error)
}

func (s *Service) ProcessReceipt(ctx context.Context, receipt string) (model.ParseReceiptFromTextResponse, error) {
	receiptParser, err := retailer.NewReceiptParser(receipt)
	if err != nil {
		return model.ParseReceiptFromTextResponse{}, fmt.Errorf("create receipt parser: %w", err)
	}

	date, err := receiptParser.ParseDate()
	if err != nil {
		return model.ParseReceiptFromTextResponse{}, fmt.Errorf("parse date: %w", err)
	}

	products, err := receiptParser.ParseProducts()
	if err != nil {
		return model.ParseReceiptFromTextResponse{}, fmt.Errorf("parse products: %w", err)
	}

	productsByName, err := s.ProductRepo.GetDistinctProductsByNames(ctx, products.GetNames())
	if err != nil {
		return model.ParseReceiptFromTextResponse{}, fmt.Errorf("find products by names: %w", err)
	}

	products.FillCategoriesAndNotes(productsByName)

	if err := s.ReceiptRepo.InsertRawReceipt(ctx, date, receipt, products); err != nil {
		slog.ErrorContext(ctx, "insert raw receipt", "error", err)
	}

	return model.ParseReceiptFromTextResponse{
		Date:     date.Format(time.DateOnly),
		Retailer: receiptParser.GetRetailer(),
		Products: products,
	}, nil
}
