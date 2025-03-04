-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS purchased_products_aliases (
    id SERIAL PRIMARY KEY,
    parsed_product_name TEXT UNIQUE NOT NULL,
    user_defined_product_name TEXT
);
-- +goose StatementEnd