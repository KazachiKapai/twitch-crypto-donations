package noncegeneration

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"time"
	"twitch-crypto-donations/internal/pkg/middleware"
)

type Database interface {
	QueryRow(query string, args ...any) *sql.Row
	Exec(query string, args ...any) (sql.Result, error)
}

type RequestBody struct {
	Address string `json:"address"`
}

type ResponseBody struct {
	Message string `json:"message"`
}

type (
	Request  = middleware.Request[RequestBody]
	Response = middleware.Response[ResponseBody]
)

type Handler struct {
	db              Database
	nonceExpiration time.Duration
	appName         string
}

func New(db Database) *Handler {
	return &Handler{
		db:              db,
		nonceExpiration: 5 * time.Minute,
		appName:         "KapachiPay",
	}
}

func (h *Handler) Handle(_ context.Context, request Request) (*Response, error) {
	if len(request.Body.Address) < 32 || len(request.Body.Address) > 44 {
		return nil, errors.New("invalid Solana address format")
	}

	nonce, err := h.generateSecureNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	message := h.createSignMessage(nonce)

	if err = h.saveNonce(nonce, request.Body.Address); err != nil {
		return nil, fmt.Errorf("failed to save nonce: %w", err)
	}

	return &Response{
		Body: ResponseBody{
			Message: message,
		},
		StatusCode: http.StatusOK,
	}, nil
}

func (h *Handler) generateSecureNonce() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	nonce := base64.URLEncoding.EncodeToString(bytes)
	return nonce, nil
}

func (h *Handler) createSignMessage(nonce string) string {
	return fmt.Sprintf("Sign in to %s Nonce: %s", h.appName, nonce)
}

func (h *Handler) saveNonce(nonce, address string) error {
	now := time.Now().UTC()
	expiresAt := now.Add(h.nonceExpiration)

	const insertQuery = `
		INSERT INTO nonces (nonce, address, created_at, expires_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (address) 
		DO UPDATE SET 
			nonce = EXCLUDED.nonce,
			created_at = EXCLUDED.created_at,
			expires_at = EXCLUDED.expires_at;
	`

	_, err := h.db.Exec(insertQuery, nonce, address, now, expiresAt)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	return nil
}

func (h *Handler) CleanupExpiredNonces() error {
	const deleteQuery = `DELETE FROM nonces WHERE expires_at < $1;`

	_, err := h.db.Exec(deleteQuery, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("cleanup failed: %w", err)
	}

	return nil
}
