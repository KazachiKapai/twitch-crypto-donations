-- +goose Up
-- +goose StatementBegin
ALTER TABLE donations_history
    ADD COLUMN receiver TEXT,
    DROP COLUMN sender_address;

UPDATE donations_history
    SET receiver = 'HwPZZSsCdPiTcPtZ7fNZmx4eMXXKvNB1LFtdVvWg2xnS';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE donations_history
    DROP COLUMN receiver;
-- +goose StatementEnd