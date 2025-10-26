-- +goose Up
-- +goose StatementBegin
ALTER TABLE users
    ADD COLUMN viewer_token TEXT,
    ADD COLUMN webhook_secret TEXT,
    ADD COLUMN channel TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users
    DROP COLUMN IF EXISTS viewer_token,
    DROP COLUMN IF EXISTS webhook_secret,
    DROP COLUMN IF EXISTS channel;
-- +goose StatementEnd
