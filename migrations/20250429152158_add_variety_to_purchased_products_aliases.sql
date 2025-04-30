-- +goose Up
-- +goose StatementBegin
ALTER TABLE purchased_products_aliases ADD COLUMN user_defined_variety_name TEXT DEFAULT '';
-- +goose StatementEnd
