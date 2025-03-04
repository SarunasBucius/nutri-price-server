-- +goose Up
-- +goose StatementBegin
ALTER TABLE raw_receipts ADD COLUMN retailer TEXT DEFAULT '';
-- +goose StatementEnd
