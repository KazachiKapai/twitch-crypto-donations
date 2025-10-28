-- +goose Up
-- +goose StatementBegin
ALTER TABLE users
    ADD COLUMN username TEXT,
    ADD COLUMN email TEXT,
    ADD COLUMN display_name TEXT,
    ADD COLUMN bio TEXT,
    ADD COLUMN avatar_url TEXT,
    ADD COLUMN created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_username;

ALTER TABLE users
    DROP COLUMN IF EXISTS updated_at,
    DROP COLUMN IF EXISTS created_at,
    DROP COLUMN IF EXISTS avatar_url,
    DROP COLUMN IF EXISTS bio,
    DROP COLUMN IF EXISTS display_name,
    DROP COLUMN IF EXISTS email,
    DROP COLUMN IF EXISTS username;
-- +goose StatementEnd