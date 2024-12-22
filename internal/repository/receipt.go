package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/SarunasBucius/nutri-price-server/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReceiptRepo struct {
	DB *pgxpool.Pool
}

func NewReceiptRepo(db *pgxpool.Pool) *ReceiptRepo {
	return &ReceiptRepo{DB: db}
}

func (r *ReceiptRepo) InsertRawReceipt(ctx context.Context, receiptDate time.Time, receipt string, parsedProducts model.ReceiptProducts) error {
	query := `
	INSERT INTO raw_receipts (purchase_date, receipt, parsed_products) 
	VALUES ($1, $2, $3) 
	ON CONFLICT (purchase_date) DO NOTHING`

	productsJSON, err := json.Marshal(parsedProducts)
	if err != nil {
		return err
	}

	if _, err := r.DB.Exec(ctx, query, receiptDate, receipt, productsJSON); err != nil {
		return err
	}

	return nil
}

func (r *ReceiptRepo) SetRawReceiptSubmittedProducts(ctx context.Context, receiptDate time.Time, submittedProducts []model.PurchasedProductNew) error {
	query := `
	UPDATE raw_receipts 
	SET submitted_products = $1
	WHERE purchase_date = $2`

	productsJSON, err := json.Marshal(submittedProducts)
	if err != nil {
		return err
	}

	if _, err := r.DB.Exec(ctx, query, productsJSON, receiptDate); err != nil {
		return err
	}

	return nil
}
