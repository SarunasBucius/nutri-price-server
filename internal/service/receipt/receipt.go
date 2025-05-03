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
}

func NewReceiptService(receiptRepo IReceiptRepository) *Service {
	return &Service{
		ReceiptRepo: receiptRepo,
	}
}

type IReceiptRepository interface {
	InsertRawReceipt(ctx context.Context, date time.Time, receiptLines, retailer string, products model.ReceiptProducts) error
	GetUnprocessedReceipt(ctx context.Context) (string, error)
	GetRawReceiptByDate(ctx context.Context, date time.Time) (string, error)
	GetUnconfirmedReceipt(ctx context.Context, retailer, date string) ([]model.PurchasedProductNew, error)
	GetUnconfirmedReceiptSummaries(ctx context.Context) ([]model.UnconfirmedReceiptSummary, error)
	GetProductNameAlias(ctx context.Context, parsedNames []string) (map[string]model.ProductAndVarietyName, error)
	GetLastReceiptDates(ctx context.Context) ([]model.LastReceiptDate, error)
	GetProductsWithMissingInfo(ctx context.Context, dateFrom string) ([]model.ProductAndVarietyName, error)
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

	aliasByParsedName, err := s.ReceiptRepo.GetProductNameAlias(ctx, products.GetNames())
	if err != nil {
		return model.ParseReceiptFromTextResponse{}, fmt.Errorf("get product name alias: %w", err)
	}

	products.UpdateProductNames(aliasByParsedName)

	if err := s.ReceiptRepo.InsertRawReceipt(ctx, date, receipt, receiptParser.GetRetailer(), products); err != nil {
		slog.ErrorContext(ctx, "insert raw receipt", "error", err)
	}

	return model.ParseReceiptFromTextResponse{
		Date:     date.Format(time.DateOnly),
		Retailer: receiptParser.GetRetailer(),
		Products: products,
	}, nil
}

func (s *Service) ProcessReceiptFromDB(ctx context.Context, receiptDate string) (model.ParseReceiptFromTextResponse, error) {
	receipt, err := s.getReceipt(ctx, receiptDate)
	if err != nil {
		return model.ParseReceiptFromTextResponse{}, fmt.Errorf("get receipt: %w", err)
	}
	return s.ProcessReceipt(ctx, receipt)
}

func (s *Service) GetUnconfirmedReceipt(ctx context.Context, retailer, date string) ([]model.PurchasedProductNew, error) {
	unconfirmedProducts, err := s.ReceiptRepo.GetUnconfirmedReceipt(ctx, retailer, date)
	if err != nil {
		return nil, fmt.Errorf("get unconfirmed receipt: %w", err)
	}

	products := model.ReceiptProducts(unconfirmedProducts)

	aliasByParsedName, err := s.ReceiptRepo.GetProductNameAlias(ctx, products.GetNames())
	if err != nil {
		return nil, fmt.Errorf("get product name alias: %w", err)
	}

	products.UpdateProductNames(aliasByParsedName)

	return products, nil
}

func (s *Service) GetUnconfirmedReceiptSummaries(ctx context.Context) ([]model.UnconfirmedReceiptSummary, error) {
	return s.ReceiptRepo.GetUnconfirmedReceiptSummaries(ctx)
}

func (s *Service) getReceipt(ctx context.Context, receiptDate string) (string, error) {
	if len(receiptDate) == 0 {
		receipt, err := s.ReceiptRepo.GetUnprocessedReceipt(ctx)
		if err != nil {
			return "", fmt.Errorf("get unprocessed receipt: %w", err)
		}
		return receipt, nil
	}

	parsedDate, err := time.Parse(time.DateOnly, receiptDate)
	if err != nil {
		return "", fmt.Errorf("parse date: %w", err)
	}
	receipt, err := s.ReceiptRepo.GetRawReceiptByDate(ctx, parsedDate)
	if err != nil {
		return "", fmt.Errorf("get raw receipt by date: %w", err)
	}

	return receipt, nil
}

func (s *Service) GetLastReceiptDates(ctx context.Context) ([]model.LastReceiptDate, error) {
	lastReceiptDates, err := s.ReceiptRepo.GetLastReceiptDates(ctx)
	if err != nil {
		return nil, fmt.Errorf("get last receipt dates: %w", err)
	}
	return lastReceiptDates, nil
}

func (s *Service) GetProductsWithMissingInfo(ctx context.Context, dateFrom string) ([]model.ProductAndVarietyName, error) {
	products, err := s.ReceiptRepo.GetProductsWithMissingInfo(ctx, dateFrom)
	if err != nil {
		return nil, fmt.Errorf("get products with missing info: %w", err)
	}
	return products, nil
}
