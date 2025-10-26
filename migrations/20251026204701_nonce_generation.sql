-- +goose Up
-- +goose StatementBegin
CREATE TABLE nonces (
    nonce TEXT NOT NULL,
    address TEXT NOT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS nonces;
-- +goose StatementEnd