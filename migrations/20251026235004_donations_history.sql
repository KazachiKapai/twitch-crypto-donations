-- +goose Up
-- +goose StatementBegin
CREATE TABLE donations_history (
    id SERIAL PRIMARY KEY,
    donation_amount TEXT NOT NULL,
    sender_address TEXT NOT NULL,
    sender_username TEXT NOT NULL,
    currency TEXT NOT NULL,
    text TEXT,
    audio_url TEXT,
    image_url TEXT,
    duration_ms FLOAT8,
    layout TEXT,
    channel TEXT,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS donations_history;
-- +goose StatementEnd