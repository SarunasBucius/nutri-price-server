package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/SarunasBucius/nutri-price-server/internal/model"
	"github.com/SarunasBucius/nutri-price-server/internal/utils/uerror"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NutritionalValueRepo struct {
	DB *pgxpool.Pool
}

func NewNutritionalValueRepo(db *pgxpool.Pool) *NutritionalValueRepo {
	return &NutritionalValueRepo{DB: db}
}

func (n *NutritionalValueRepo) InsertProductNutritionalValue(ctx context.Context, product, measurementUnit string, nv model.NutritionalValue) error {
	query := `
INSERT INTO nutritional_values (
	product, measurement_unit, energy_value_kcal, fat, saturated_fat, carbohydrate, 
	carbohydrate_sugars, fibre, soluble_fibre, insoluble_fibre, protein, salt
) VALUES (
	$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
)`
	if _, err := n.DB.Exec(ctx, query,
		product, measurementUnit, nv.EnergyValueKCAL, nv.Fat, nv.SaturatedFat, nv.Carbohydrate, nv.CarbohydrateSugars, nv.Fibre, nv.SolubleFibre, nv.InsolubleFibre, nv.Protein, nv.Salt); err != nil {
		return err
	}

	return nil
}

func (n *NutritionalValueRepo) GetProductsNutritionalValue(ctx context.Context) ([]model.ProductNutritionalValue, error) {
	query := `
	SELECT 
		id, product, measurement_unit, energy_value_kcal,
		fat, saturated_fat, carbohydrate, carbohydrate_sugars, 
		fibre, soluble_fibre, insoluble_fibre, protein, salt
	FROM nutritional_values`

	rows, err := n.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var productNVs []model.ProductNutritionalValue
	for rows.Next() {
		var n model.ProductNutritionalValue
		if err := rows.Scan(
			&n.ID, &n.Product, &n.Unit, &n.NutritionalValue.EnergyValueKCAL,
			&n.NutritionalValue.Fat, &n.NutritionalValue.SaturatedFat, &n.NutritionalValue.Carbohydrate, &n.NutritionalValue.CarbohydrateSugars,
			&n.NutritionalValue.Fibre, &n.NutritionalValue.SolubleFibre, &n.NutritionalValue.InsolubleFibre, &n.NutritionalValue.Protein, &n.NutritionalValue.Salt,
		); err != nil {
			return nil, err
		}
		productNVs = append(productNVs, n)
	}
	return productNVs, nil
}

func (n *NutritionalValueRepo) GetProductsNutritionalValueByProductNames(ctx context.Context, productNames []string) ([]model.ProductNutritionalValue, error) {
	query := `
	SELECT 
		id, product, measurement_unit, energy_value_kcal,
		fat, saturated_fat, carbohydrate, carbohydrate_sugars, 
		fibre, soluble_fibre, insoluble_fibre, protein, salt
	FROM nutritional_values
	WHERE product=ANY($1)`

	rows, err := n.DB.Query(ctx, query, productNames)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var productNVs []model.ProductNutritionalValue
	for rows.Next() {
		var pnv model.ProductNutritionalValue
		if err := rows.Scan(
			&pnv.ID, &pnv.Product, &pnv.Unit, &pnv.NutritionalValue.EnergyValueKCAL,
			&pnv.NutritionalValue.Fat, &pnv.NutritionalValue.SaturatedFat, &pnv.NutritionalValue.Carbohydrate, &pnv.NutritionalValue.CarbohydrateSugars,
			&pnv.NutritionalValue.Fibre, &pnv.NutritionalValue.SolubleFibre, &pnv.NutritionalValue.InsolubleFibre, &pnv.NutritionalValue.Protein, &pnv.NutritionalValue.Salt,
		); err != nil {
			return nil, err
		}
		productNVs = append(productNVs, pnv)
	}
	return productNVs, nil
}

func (n *NutritionalValueRepo) GetProductNutritionalValue(ctx context.Context, nutritionalValueID int) (model.ProductNutritionalValue, error) {
	query := `
	SELECT 
		id, product, measurement_unit, energy_value_kcal,
		fat, saturated_fat, carbohydrate, carbohydrate_sugars, 
		fibre, soluble_fibre, insoluble_fibre, protein, salt
	FROM nutritional_values
	WHERE id=$1`

	var pnv model.ProductNutritionalValue
	err := n.DB.QueryRow(ctx, query, nutritionalValueID).Scan(
		&pnv.ID, &pnv.Product, &pnv.Unit, &pnv.NutritionalValue.EnergyValueKCAL,
		&pnv.NutritionalValue.Fat, &pnv.NutritionalValue.SaturatedFat, &pnv.NutritionalValue.Carbohydrate, &pnv.NutritionalValue.CarbohydrateSugars,
		&pnv.NutritionalValue.Fibre, &pnv.NutritionalValue.SolubleFibre, &pnv.NutritionalValue.InsolubleFibre, &pnv.NutritionalValue.Protein, &pnv.NutritionalValue.Salt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.ProductNutritionalValue{}, uerror.NewNotFound("nutritional value not found", err)
	}
	if err != nil {
		return model.ProductNutritionalValue{}, err
	}

	return pnv, nil
}

func (n *NutritionalValueRepo) UpdateProductNutritionalValue(ctx context.Context, pnv model.ProductNutritionalValue) error {
	query := `
	UPDATE nutritional_values 
	SET 
		product = $1,
		measurement_unit = $2,
		energy_value_kcal = $3,
		fat = $4,
		saturated_fat = $5,
		carbohydrate = $6,
		carbohydrate_sugars = $7,
		fibre = $8,
		soluble_fibre = $9,
		insoluble_fibre = $10,
		protein = $11,
		salt = $12
	WHERE id = $13`
	status, err := n.DB.Exec(ctx, query, pnv.Product, pnv.Unit, pnv.NutritionalValue.EnergyValueKCAL,
		pnv.NutritionalValue.Fat, pnv.NutritionalValue.SaturatedFat, pnv.NutritionalValue.Carbohydrate, pnv.NutritionalValue.CarbohydrateSugars,
		pnv.NutritionalValue.Fibre, pnv.NutritionalValue.SolubleFibre, &pnv.NutritionalValue.InsolubleFibre,
		pnv.NutritionalValue.Protein, pnv.NutritionalValue.Salt, pnv.ID)
	if err != nil {
		return err
	}

	if status.RowsAffected() != 1 {
		return uerror.NewNotFound(fmt.Sprintf("nutrition value with id %q does not exist", pnv.ID), nil)
	}

	return nil
}

func (n *NutritionalValueRepo) DeleteProductNutritionalValue(ctx context.Context, id int) error {
	query := `DELETE FROM nutritional_values WHERE id = $1`
	if _, err := n.DB.Exec(ctx, query, id); err != nil {
		return err
	}
	return nil
}
