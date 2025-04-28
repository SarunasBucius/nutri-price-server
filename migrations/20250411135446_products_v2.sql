-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS purchases (
    id SERIAL PRIMARY KEY,
    product_id INT NOT NULL,
    variety_name TEXT NOT NULL DEFAULT '',
    retailer TEXT NOT NULL DEFAULT '',
    purchase_date DATE,
    unit TEXT NOT NULL DEFAULT '',
    quantity NUMERIC(9, 3) NOT NULL DEFAULT 0,
    price NUMERIC(6, 2) NOT NULL DEFAULT 0,
    notes TEXT NOT NULL DEFAULT '',
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS nutritional_values_v2 (
    id SERIAL PRIMARY KEY,
    product_id INT NOT NULL,
    variety_name TEXT NOT NULL DEFAULT '',
    unit TEXT NOT NULL DEFAULT '',
    energy_value_kcal SMALLINT NOT NULL DEFAULT 0,
    fat NUMERIC(6, 3) NOT NULL DEFAULT 0,
    saturated_fat NUMERIC(6, 3) NOT NULL DEFAULT 0,
    carbohydrate NUMERIC(6, 3) NOT NULL DEFAULT 0,
    carbohydrate_sugars NUMERIC(6, 3) NOT NULL DEFAULT 0,
    fibre NUMERIC(6, 3) NOT NULL DEFAULT 0,
    protein NUMERIC(6, 3) NOT NULL DEFAULT 0,
    salt NUMERIC(6, 3) NOT NULL DEFAULT 0,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE,
    UNIQUE (variety_name, product_id)
);

-- +goose StatementEnd