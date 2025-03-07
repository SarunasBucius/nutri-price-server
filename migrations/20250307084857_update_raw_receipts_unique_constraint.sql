-- +goose Up
-- +goose StatementBegin
ALTER TABLE raw_receipts DROP CONSTRAINT raw_receipts_purchase_date_key;
ALTER TABLE raw_receipts ADD CONSTRAINT raw_receipts_purchase_date_retailer_key UNIQUE (purchase_date, retailer);
-- +goose StatementEnd
