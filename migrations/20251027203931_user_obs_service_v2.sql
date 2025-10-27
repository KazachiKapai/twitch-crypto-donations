-- +goose Up
-- +goose StatementBegin
DELETE FROM users;

ALTER TABLE users
    ADD COLUMN widget_token TEXT,
    ADD COLUMN alerts_widget_url TEXT,
    ADD COLUMN media_widget_url TEXT,
    ADD COLUMN webhook_url TEXT,
    DROP COLUMN IF EXISTS donation_url;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users
    DROP COLUMN IF EXISTS widget_token,
    DROP COLUMN IF EXISTS alerts_widget_url,
    DROP COLUMN IF EXISTS media_widget_url,
    DROP COLUMN IF EXISTS webhook_url;
-- +goose StatementEnd