-- +goose Up
-- +goose StatementBegin
ALTER TABLE recipes DROP CONSTRAINT recipes_recipe_name_key;

ALTER TABLE recipes ADD COLUMN dish_made_date DATE;
-- +goose StatementEnd
