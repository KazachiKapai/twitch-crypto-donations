package noncegeneration

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"
	"twitch-crypto-donations/internal/pkg/middleware"

	"github.com/google/uuid"
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
	db Database
}

func New(db Database) *Handler {
	return &Handler{db: db}
}

func (h *Handler) Handle(_ context.Context, request Request) (*Response, error) {
	nonce := uuid.New().String()
	timestamp := time.Now().UTC().Format(time.RFC3339)
	message := fmt.Sprintf("Sign in to MyDApp. Nonce: %s, Address: %s, Timestamp: %s", nonce, request.Body.Address, timestamp)

	if err := h.saveNonce(nonce, request.Body.Address, timestamp); err != nil {
		return nil, err
	}

	return &Response{Body: ResponseBody{Message: message}, StatusCode: http.StatusOK}, nil
}

func (h *Handler) saveNonce(nonce, address, timestamp string) error {
	const insertQuery = `
		INSERT INTO nonces (nonce, address, created_at)
		VALUES ($1, $2, $3);
	`

	_, err := h.db.Exec(insertQuery, nonce, address, timestamp)
	return err
}
