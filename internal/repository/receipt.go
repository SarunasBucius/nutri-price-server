package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/SarunasBucius/nutri-price-server/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReceiptRepo struct {
	DB *pgxpool.Pool
}

func NewReceiptRepo(db *pgxpool.Pool) *ReceiptRepo {
	return &ReceiptRepo{DB: db}
}

func (r *ReceiptRepo) InsertRawReceipt(ctx context.Context, receiptDate time.Time, receipt, retailer string, parsedProducts model.ReceiptProducts) error {
	query := `
	INSERT INTO raw_receipts (purchase_date, receipt, retailer, parsed_products) 
	VALUES ($1, $2, $3, $4) 
	ON CONFLICT (purchase_date, retailer) 
	DO UPDATE SET parsed_products = EXCLUDED.parsed_products`

	productsJSON, err := json.Marshal(parsedProducts)
	if err != nil {
		return err
	}

	if _, err := r.DB.Exec(ctx, query, receiptDate, receipt, retailer, productsJSON); err != nil {
		return err
	}

	return r.insertParsedProducts(ctx, parsedProducts)
}

func (r *ReceiptRepo) insertParsedProducts(ctx context.Context, parsedProducts model.ReceiptProducts) error {

	batch := &pgx.Batch{}
	for _, product := range parsedProducts {
		batch.Queue("INSERT INTO purchased_products_aliases(parsed_product_name) VALUES($1) ON CONFLICT DO NOTHING", product.Name)
	}
	br := r.DB.SendBatch(ctx, batch)

	return br.Close()
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

func (r *ReceiptRepo) GetUnprocessedReceipt(ctx context.Context) (string, error) {
	query := `
	SELECT receipt
	FROM raw_receipts
	WHERE submitted_products IS NULL
	LIMIT 1`

	var receipt string
	if err := r.DB.QueryRow(ctx, query).Scan(&receipt); err != nil {
		return "", err
	}

	return receipt, nil
}

func (r *ReceiptRepo) GetRawReceiptByDate(ctx context.Context, date time.Time) (string, error) {
	query := `
	SELECT receipt
	FROM raw_receipts
	WHERE purchase_date = $1`

	var receipt string
	if err := r.DB.QueryRow(ctx, query, date).Scan(&receipt); err != nil {
		return "", err
	}

	return receipt, nil
}

func (r *ReceiptRepo) UpdateProductNameAlias(ctx context.Context, editedNameByParsedName map[string]string) error {
	batch := &pgx.Batch{}
	for parsedName, editedName := range editedNameByParsedName {
		batch.Queue(`
		INSERT INTO purchased_products_aliases 
		(user_defined_product_name, parsed_product_name) VALUES ($1, $2) 
		ON CONFLICT (parsed_product_name) DO UPDATE
		SET user_defined_product_name = EXCLUDED.user_defined_product_name`, editedName, parsedName)
	}
	br := r.DB.SendBatch(ctx, batch)

	return br.Close()
}

func (r *ReceiptRepo) GetProductNameAlias(ctx context.Context, parsedNames []string) (map[string]string, error) {
	query := `
	SELECT parsed_product_name, user_defined_product_name
	FROM purchased_products_aliases
	WHERE parsed_product_name = ANY($1)`

	rows, err := r.DB.Query(ctx, query, parsedNames)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	aliases := make(map[string]string)
	for rows.Next() {
		var parsedName, alias string
		if err := rows.Scan(&parsedName, &alias); err != nil {
			return nil, err
		}
		aliases[parsedName] = alias
	}

	return aliases, nil
}

func (r *ReceiptRepo) GetUnconfirmedReceiptSummaries(ctx context.Context) ([]model.UnconfirmedReceiptSummary, error) {
	query := `
	SELECT purchase_date, retailer
	FROM raw_receipts
	WHERE NOT is_confirmed ORDER BY purchase_date DESC`

	rows, err := r.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []model.UnconfirmedReceiptSummary
	for rows.Next() {
		var summary model.UnconfirmedReceiptSummary
		var receiptDate *pgtype.Date
		if err := rows.Scan(&receiptDate, &summary.Retailer); err != nil {
			return nil, err
		}
		summary.Date = receiptDate.Time.Format("2006-01-02")
		summaries = append(summaries, summary)
	}

	return summaries, nil
}

func (r *ReceiptRepo) GetUnconfirmedReceipt(ctx context.Context, retailer, date string) ([]model.PurchasedProductNew, error) {
	query := `
	SELECT parsed_products
	FROM raw_receipts
	WHERE retailer = $1 AND purchase_date = $2`

	rows, err := r.DB.Query(ctx, query, retailer, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var daysParsedProducts []model.PurchasedProductNew
	for rows.Next() {
		var parsedProducts []model.PurchasedProductNew
		if err := rows.Scan(&parsedProducts); err != nil {
			return nil, err
		}
		daysParsedProducts = append(daysParsedProducts, parsedProducts...)
	}

	return daysParsedProducts, nil
}

func (r *ReceiptRepo) ConfirmReceipts(ctx context.Context, retailer, date string) error {
	query := `
	UPDATE raw_receipts
	SET is_confirmed = true
	WHERE purchase_date = $1 AND retailer = $2`

	if _, err := r.DB.Exec(ctx, query, date, retailer); err != nil {
		return err
	}

	return nil
}
