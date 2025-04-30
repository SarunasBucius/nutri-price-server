package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/SarunasBucius/nutri-price-server/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductRepo struct {
	DB *pgxpool.Pool
}

func NewProductRepo(db *pgxpool.Pool) *ProductRepo {
	return &ProductRepo{DB: db}
}

func (p *ProductRepo) InsertProducts(ctx context.Context, productNames []string) error {
	if len(productNames) == 0 {
		return nil
	}

	queryPrefix := `INSERT INTO products (name) VALUES `
	queryConflict := ` ON CONFLICT (name) DO NOTHING`

	placeholders := make([]string, 0, len(productNames))
	values := make([]any, 0, len(productNames))
	for i := range productNames {
		placeholders = append(placeholders, fmt.Sprintf("($%d)", i+1))
		values = append(values, productNames[i])
	}

	query := queryPrefix + strings.Join(placeholders, ",") + queryConflict
	_, err := p.DB.Exec(ctx, query, values...)
	return err
}

func (p *ProductRepo) InsertPurchases(ctx context.Context, retailer string, purchaseDate time.Time, products []model.PurchasedProductNew) error {
	rows := make([][]interface{}, 0, len(products))
	for _, p := range products {
		row := []interface{}{p.ProductID, retailer, purchaseDate, p.Quantity.Unit, p.Quantity.Amount, p.Price, p.Notes, p.VarietyName}
		rows = append(rows, row)
	}

	_, err := p.DB.CopyFrom(ctx,
		pgx.Identifier{"purchases"},
		[]string{"product_id", "retailer", "purchase_date", "unit", "quantity", "price", "notes", "variety_name"},
		pgx.CopyFromRows(rows),
	)

	return err
}

func (p *ProductRepo) GetProductIDsByName(ctx context.Context, productNames []string) (map[string]string, error) {
	query := `SELECT id, name FROM products WHERE name=ANY($1)`
	rows, err := p.DB.Query(ctx, query, productNames)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	productIDs := make(map[string]string)
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		productIDs[name] = id
	}
	return productIDs, nil
}

func (p *ProductRepo) GetLastBoughtProductsByNamesOrGroups(ctx context.Context, productNames []string) ([]model.PurchasedProduct, error) {
	query := `
	SELECT DISTINCT ON (product_name) 
		id, variety_name, retailer, unit, quantity, price, notes, purchase_date
	FROM purchases
	WHERE variety_name=ANY($1)
	ORDER BY variety_name, purchase_date DESC`

	rows, err := p.DB.Query(ctx, query, productNames)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []model.PurchasedProduct
	for rows.Next() {
		var p model.PurchasedProduct
		if err := rows.Scan(
			&p.ID, &p.VarietyName, &p.Retailer, &p.Quantity.Unit, &p.Quantity.Amount, &p.Price, &p.Notes, &p.Date,
		); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}
