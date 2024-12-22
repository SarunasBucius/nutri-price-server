-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS purchased_products (
    id SERIAL PRIMARY KEY,
    retailer TEXT NOT NULL,
    product_name TEXT NOT NULL,
    product_group TEXT NOT NULL,
    measurement_unit TEXT NOT NULL,
    quantity NUMERIC(9, 3) NOT NULL,
    full_price NUMERIC(6, 2) NOT NULL,
    paid_price NUMERIC(6, 2) NOT NULL,
    discount NUMERIC(6, 2) NOT NULL,
    notes TEXT,
    purchase_date DATE
);

CREATE TABLE IF NOT EXISTS recipes (
    id SERIAL PRIMARY KEY,
    recipe_name TEXT UNIQUE NOT NULL,
    steps TEXT [],
    notes TEXT
);

CREATE TABLE IF NOT EXISTS recipe_ingredients (
    id SERIAL PRIMARY KEY,
    recipe_id INT NOT NULL,
    product_name TEXT NOT NULL,
    measurement_unit TEXT NOT NULL,
    quantity NUMERIC(9, 3) NOT NULL,
    metric_measurement_unit TEXT NOT NULL,
    metric_quantity NUMERIC(9, 3) NOT NULL,
    cut_style TEXT,
    FOREIGN KEY (recipe_id) REFERENCES recipes(id)
);

CREATE TABLE IF NOT EXISTS nutritional_values (
    id SERIAL PRIMARY KEY,
    product TEXT NOT NULL,
    measurement_unit TEXT NOT NULL,
    energy_value_kcal SMALLINT NOT NULL,
    fat NUMERIC(6, 3) NOT NULL,
    saturated_fat NUMERIC(6, 3) NOT NULL,
    carbohydrate NUMERIC(6, 3) NOT NULL,
    carbohydrate_sugars NUMERIC(6, 3) NOT NULL,
    fibre NUMERIC(6, 3) NOT NULL,
    soluble_fibre NUMERIC(6, 3),
    insoluble_fibre NUMERIC(6, 3),
    protein NUMERIC(6, 3) NOT NULL,
    salt NUMERIC(6, 3) NOT NULL
);

CREATE TABLE IF NOT EXISTS raw_receipts (
    purchase_date DATE UNIQUE,
    receipt TEXT NOT NULL,
    parsed_products JSON NOT NULL,
    submitted_products JSON
);

-- +goose StatementEnd