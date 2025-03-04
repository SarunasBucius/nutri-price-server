-- +goose Up
-- +goose StatementBegin
ALTER TABLE raw_receipts ADD COLUMN is_confirmed BOOLEAN DEFAULT FALSE;
-- +goose StatementEnd