package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/SarunasBucius/nutri-price-server/internal/model"
	"github.com/SarunasBucius/nutri-price-server/internal/utils/uerror"
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

func (p *ProductRepo) GetProductGroups(ctx context.Context) ([]string, error) {
	query := `
	SELECT DISTINCT product_group 
	FROM purchased_products WHERE product_group != ''`

	rows, err := p.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var productGroups []string
	for rows.Next() {
		var productGroup string
		if err := rows.Scan(&productGroup); err != nil {
			return nil, err
		}
		productGroups = append(productGroups, productGroup)
	}
	return productGroups, nil
}

func (p *ProductRepo) GetProductsByGroup(ctx context.Context, productGroups []string) ([]model.PurchasedProduct, error) {
	query := `
	SELECT
		id, product_name, retailer, product_group, measurement_unit, quantity, full_price, paid_price, discount, notes, purchase_date
	FROM purchased_products
	WHERE product_group=ANY($1)
	ORDER BY purchase_date DESC`

	rows, err := p.DB.Query(ctx, query, productGroups)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []model.PurchasedProduct
	for rows.Next() {
		var prod model.PurchasedProduct
		if err := rows.Scan(
			&prod.ID, &prod.Name, &prod.Retailer, &prod.Group, &prod.Quantity.Unit, &prod.Quantity.Amount, &prod.Price.Full, &prod.Price.Paid, &prod.Price.Discount, &prod.Notes, &prod.Date,
		); err != nil {
			return nil, err
		}
		products = append(products, prod)
	}
	return products, nil
}

func (p *ProductRepo) GetProduct(ctx context.Context, productID int) (model.PurchasedProduct, error) {
	query := `
	SELECT 
		id, product_name, retailer, product_group, measurement_unit, quantity, full_price, paid_price, discount, notes, purchase_date
	FROM purchased_products
	WHERE id=$1`

	var prod model.PurchasedProduct
	err := p.DB.QueryRow(ctx, query, productID).Scan(
		&prod.ID, &prod.Name, &prod.Retailer, &prod.Group,
		&prod.Quantity.Unit, &prod.Quantity.Amount,
		&prod.Price.Full, &prod.Price.Paid, &prod.Price.Discount,
		&prod.Notes, &prod.Date)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.PurchasedProduct{}, uerror.NewNotFound("product not found", err)
	}
	if err != nil {
		return model.PurchasedProduct{}, err
	}

	return prod, nil
}

func (p *ProductRepo) UpdateProduct(ctx context.Context, prod model.PurchasedProduct) error {
	query := `
	UPDATE purchased_products
	SET
		retailer = $1,
		product_name = $2,
		product_group = $3,
		measurement_unit = $4,
		quantity = $5,
		full_price = $6,
		paid_price = $7,
		discount = $8,
		notes = $9,
		purchase_date = $10
	WHERE id = $11`

	status, err := p.DB.Exec(ctx, query, prod.Retailer, prod.Name, prod.Group,
		prod.Quantity.Unit, prod.Quantity.Amount,
		prod.Price.Full, prod.Price.Paid, prod.Price.Discount,
		prod.Notes, prod.Date, prod.ID)
	if err != nil {
		return err
	}

	if status.RowsAffected() != 1 {
		return uerror.NewNotFound(fmt.Sprintf("product with id %q does not exist", prod.ID), nil)
	}

	return nil
}

func (p *ProductRepo) DeleteProduct(ctx context.Context, productID int) error {
	query := `DELETE FROM purchased_products WHERE id = $1`
	if _, err := p.DB.Exec(ctx, query, productID); err != nil {
		return err
	}
	return nil
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
