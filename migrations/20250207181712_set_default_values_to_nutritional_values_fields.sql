-- +goose Up
-- +goose StatementBegin
ALTER TABLE nutritional_values
  ALTER COLUMN measurement_unit SET DEFAULT '',
  ALTER COLUMN energy_value_kcal SET DEFAULT 0,
  ALTER COLUMN fat SET DEFAULT 0,
  ALTER COLUMN saturated_fat SET DEFAULT 0,
  ALTER COLUMN carbohydrate SET DEFAULT 0,
  ALTER COLUMN carbohydrate_sugars SET DEFAULT 0,
  ALTER COLUMN fibre SET DEFAULT 0,
  ALTER COLUMN soluble_fibre SET DEFAULT 0,
  ALTER COLUMN insoluble_fibre SET DEFAULT 0,
  ALTER COLUMN protein SET DEFAULT 0,
  ALTER COLUMN salt SET DEFAULT 0;
-- +goose StatementEnd