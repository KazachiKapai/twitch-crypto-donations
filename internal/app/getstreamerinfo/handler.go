package getstreamerinfo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"
	"twitch-crypto-donations/internal/pkg/middleware"
)

type Database interface {
	QueryRow(query string, args ...any) *sql.Row
	Exec(query string, args ...any) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type UserInfo struct {
	Wallet      string     `json:"wallet"`
	Username    *string    `json:"username"`
	Email       *string    `json:"email"`
	DisplayName *string    `json:"display_name"`
	Bio         *string    `json:"bio"`
	AvatarUrl   *string    `json:"avatar_url"`
	CreatedAt   *time.Time `json:"created_at"`

	AlertsWidgetUrl *string `json:"alerts_widget_url"`
	MediaWidgetUrl  *string `json:"media_widget_url"`
}

type (
	Request  = middleware.Request[struct{}]
	Response = middleware.Response[*UserInfo]
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
	address, authorized := request.Context[middleware.AddressKey].(string)

	if !authorized {
		streamerUsername := request.PathParams["username"]

		streamerInfo, err := h.getStreamerInfo(streamerUsername)
		if err != nil {
			return nil, err
		}

		return &Response{Body: streamerInfo, StatusCode: http.StatusOK}, nil
	}

	if address == "" {
		return &Response{
			StatusCode: http.StatusUnauthorized,
		}, fmt.Errorf("jwt is not found or api middleware is failed")
	}

	userInfo, err := h.getUserInfo(address)
	if err != nil {
		return nil, err
	}

	return &Response{
		Body:       userInfo,
		StatusCode: http.StatusOK,
	}, nil
}

func (h *Handler) getUserInfo(wallet string) (*UserInfo, error) {
	query := `
        SELECT wallet, username, email,
            display_name, bio,
            avatar_url, created_at,
            alerts_widget_url, media_widget_url
        FROM users
        WHERE wallet = $1
    `

	var userInfo UserInfo
	err := h.db.QueryRow(query, wallet).Scan(
		&userInfo.Wallet, &userInfo.Username,
		&userInfo.Email, &userInfo.DisplayName,
		&userInfo.Bio, &userInfo.AvatarUrl,
		&userInfo.CreatedAt, &userInfo.AlertsWidgetUrl,
		&userInfo.MediaWidgetUrl,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}

		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	return &userInfo, nil
}

func (h *Handler) getStreamerInfo(username string) (*UserInfo, error) {
	query := `
        SELECT wallet, username, email,
            display_name, bio,
            avatar_url, created_at
        FROM users
        WHERE username = $1
    `

	var userInfo UserInfo
	err := h.db.QueryRow(query, username).Scan(
		&userInfo.Wallet,
		&userInfo.Username,
		&userInfo.Email,
		&userInfo.DisplayName,
		&userInfo.Bio,
		&userInfo.AvatarUrl,
		&userInfo.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}

		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	return &userInfo, nil
}
