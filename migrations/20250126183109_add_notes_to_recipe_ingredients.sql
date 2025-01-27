-- +goose Up
-- +goose StatementBegin
ALTER TABLE recipe_ingredients ADD COLUMN notes TEXT;
-- +goose StatementEnd
