package setuserinfo

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"twitch-crypto-donations/internal/pkg/middleware"
)

type Database interface {
	QueryRow(query string, args ...any) *sql.Row
	Exec(query string, args ...any) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type RequestBody struct {
	Username    *string `json:"username"`
	Email       *string `json:"email"`
	DisplayName *string `json:"display_name"`
	Bio         *string `json:"bio"`
	AvatarUrl   *string `json:"avatar_url"`
}

type (
	Request  = middleware.Request[RequestBody]
	Response = middleware.Response[struct{}]
)

type Handler struct {
	db Database
}

func New(db Database) *Handler {
	return &Handler{
		db: db,
	}
}

func (h *Handler) Handle(_ context.Context, request Request) (*Response, error) {
	address, exists := request.Context[middleware.AddressKey].(string)
	if !exists || address == "" {
		return &Response{
			StatusCode: http.StatusUnauthorized,
		}, fmt.Errorf("jwt is not found or api middleware is failed")
	}

	updates, args, argCount := h.buildQuery(request.Body)
	if len(updates) == 0 {
		return &Response{
			StatusCode: http.StatusBadRequest,
		}, fmt.Errorf("no fields to update")
	}

	args = append(args, address)
	if err := h.setUserInfo(updates, args, argCount); err != nil {
		return nil, err
	}

	return &Response{
		StatusCode: http.StatusOK,
	}, nil
}

func (h *Handler) setUserInfo(updates []string, args []interface{}, argCount int) error {
	query := fmt.Sprintf(
		"UPDATE users SET %s WHERE wallet = $%d",
		strings.Join(updates, ", "),
		argCount,
	)

	_, err := h.db.Exec(query, args...)
	return err
}

func (h *Handler) buildQuery(body RequestBody) ([]string, []interface{}, int) {
	var updates []string
	var args []interface{}
	argCount := 1

	if body.Username != nil {
		updates = append(updates, fmt.Sprintf("username = $%d", argCount))
		args = append(args, *body.Username)
		argCount++
	}

	if body.Email != nil {
		updates = append(updates, fmt.Sprintf("email = $%d", argCount))
		args = append(args, *body.Email)
		argCount++
	}

	if body.DisplayName != nil {
		updates = append(updates, fmt.Sprintf("display_name = $%d", argCount))
		args = append(args, *body.DisplayName)
		argCount++
	}

	if body.Bio != nil {
		updates = append(updates, fmt.Sprintf("bio = $%d", argCount))
		args = append(args, *body.Bio)
		argCount++
	}

	if body.AvatarUrl != nil {
		updates = append(updates, fmt.Sprintf("avatar_url = $%d", argCount))
		args = append(args, *body.AvatarUrl)
		argCount++
	}

	return updates, args, argCount
}
