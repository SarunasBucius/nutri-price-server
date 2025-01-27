-- +goose Up
-- +goose StatementBegin
ALTER TABLE
    recipe_ingredients DROP cut_style,
    DROP measurement_unit,
    DROP quantity;

ALTER TABLE
    recipe_ingredients RENAME metric_measurement_unit TO unit;

ALTER TABLE
    recipe_ingredients RENAME metric_quantity TO amount;

-- +goose StatementEnd