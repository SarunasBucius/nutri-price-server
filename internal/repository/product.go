package repository

import (
	"context"
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

func (p *ProductRepo) InsertProducts(ctx context.Context, retailer string, purchaseDate time.Time, products []model.PurchasedProductNew) error {
	rows := make([][]interface{}, 0, len(products))
	for _, p := range products {
		row := []interface{}{p.Name, retailer, p.Group, p.Quantity.Unit, p.Quantity.Amount, p.Price.Full, p.Price.Paid, p.Price.Discount, p.Notes, purchaseDate}
		rows = append(rows, row)
	}

	_, err := p.DB.CopyFrom(ctx,
		pgx.Identifier{"purchased_products"},
		[]string{"product_name", "retailer", "product_group", "measurement_unit", "quantity", "full_price", "paid_price", "discount", "notes", "purchase_date"},
		pgx.CopyFromRows(rows),
	)

	return err
}

func (p *ProductRepo) GetDistinctProductsByNames(ctx context.Context, productNames []string) (map[string]model.PurchasedProduct, error) {
	query := `
	SELECT DISTINCT ON (product_name) 
		id, product_name, retailer, product_group, measurement_unit, quantity, full_price, paid_price, discount, notes, purchase_date
	FROM purchased_products
	WHERE product_name=ANY($1)
	ORDER BY product_name, purchase_date DESC`

	rows, err := p.DB.Query(ctx, query, productNames)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make(map[string]model.PurchasedProduct)
	for rows.Next() {
		var p model.PurchasedProduct
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Retailer, &p.Group, &p.Quantity.Unit, &p.Quantity.Amount, &p.Price.Full, &p.Price.Paid, &p.Price.Discount, &p.Notes, &p.Date,
		); err != nil {
			return nil, err
		}
		products[p.Name] = p
	}
	return products, nil
}

func (p *ProductRepo) GetLastBoughtProductsByNamesOrGroups(ctx context.Context, productNamesOrGroups []string) ([]model.PurchasedProduct, error) {
	query := `
	SELECT DISTINCT ON (product_name) 
		id, product_name, retailer, product_group, measurement_unit, quantity, full_price, paid_price, discount, notes, purchase_date
	FROM purchased_products
	WHERE product_name=ANY($1) OR product_group=ANY($1)
	ORDER BY product_name, purchase_date DESC`

	rows, err := p.DB.Query(ctx, query, productNamesOrGroups)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []model.PurchasedProduct
	for rows.Next() {
		var p model.PurchasedProduct
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Retailer, &p.Group, &p.Quantity.Unit, &p.Quantity.Amount, &p.Price.Full, &p.Price.Paid, &p.Price.Discount, &p.Notes, &p.Date,
		); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}
