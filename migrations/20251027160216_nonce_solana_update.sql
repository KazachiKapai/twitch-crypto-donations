-- +goose Up
-- +goose StatementBegin

-- Add expires_at column
ALTER TABLE nonces
    ADD COLUMN expires_at TIMESTAMP WITHOUT TIME ZONE;

-- Backfill expires_at for existing rows (5 minutes after created_at)
UPDATE nonces
SET expires_at = created_at + INTERVAL '5 minutes'
WHERE expires_at IS NULL;

-- Make expires_at NOT NULL after backfilling
ALTER TABLE nonces
    ALTER COLUMN expires_at SET NOT NULL;

-- Add id column as primary key
ALTER TABLE nonces
    ADD COLUMN id SERIAL PRIMARY KEY;

-- Add unique constraint on nonce
ALTER TABLE nonces
    ADD CONSTRAINT nonces_nonce_unique UNIQUE (nonce);

-- Clean up duplicate addresses - keep only the most recent one for each address
DELETE FROM nonces a
    USING nonces b
WHERE a.ctid < b.ctid
  AND a.address = b.address;

-- Now add unique constraint on address for ON CONFLICT support
ALTER TABLE nonces
    ADD CONSTRAINT nonces_address_unique UNIQUE (address);

-- Add check constraint to ensure expiration is after creation
ALTER TABLE nonces
    ADD CONSTRAINT nonces_expiration_check CHECK (expires_at > created_at);

-- Change column types for better performance and validation
ALTER TABLE nonces
    ALTER COLUMN nonce TYPE VARCHAR(255),
    ALTER COLUMN address TYPE VARCHAR(44);

-- Create indexes for better query performance
CREATE INDEX idx_nonces_address ON nonces(address);
CREATE INDEX idx_nonces_expires_at ON nonces(expires_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop indexes
DROP INDEX IF EXISTS idx_nonces_address;
DROP INDEX IF EXISTS idx_nonces_expires_at;

-- Remove constraints
ALTER TABLE nonces DROP CONSTRAINT IF EXISTS nonces_nonce_unique;
ALTER TABLE nonces DROP CONSTRAINT IF EXISTS nonces_address_unique;
ALTER TABLE nonces DROP CONSTRAINT IF EXISTS nonces_expiration_check;

-- Remove id column
ALTER TABLE nonces DROP COLUMN IF EXISTS id;

-- Revert column types
ALTER TABLE nonces
    ALTER COLUMN nonce TYPE TEXT,
    ALTER COLUMN address TYPE TEXT;

-- Remove expires_at column
ALTER TABLE nonces DROP COLUMN IF EXISTS expires_at;

-- +goose StatementEnd