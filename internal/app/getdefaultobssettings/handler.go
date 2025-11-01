package getdefaultobssettings

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"twitch-crypto-donations/internal/pkg/middleware"
	"twitch-crypto-donations/internal/pkg/obsservice"
)

type Database interface {
	QueryRow(query string, args ...any) *sql.Row
	Exec(query string, args ...any) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type ObsService interface {
	GetAlertSettings(channel string) (*obsservice.GetAlertSettingsResponse, error)
}

type (
	Request  = middleware.Request[struct{}]
	Response = middleware.Response[obsservice.GetAlertSettingsResponse]
)

type Handler struct {
	db         Database
	obsService ObsService
}

func New(db Database, obsService ObsService) *Handler {
	return &Handler{
		db:         db,
		obsService: obsService,
	}
}

func (h *Handler) Handle(_ context.Context, request Request) (*Response, error) {
	address, ok := request.PathParams["address"]
	if !ok {
		return nil, errors.New("address is required")
	}

	channel, err := h.getChannelByWallet(address)
	if err != nil {
		return nil, err
	}

	resp, err := h.obsService.GetAlertSettings(channel)
	if err != nil {
		return nil, err
	}

	return &Response{Body: *resp, StatusCode: http.StatusOK}, nil
}

func (h *Handler) getChannelByWallet(wallet string) (string, error) {
	const query = `SELECT channel FROM users WHERE wallet = $1;`

	var channel string
	err := h.db.QueryRow(query, wallet).Scan(&channel)
	if err != nil {
		return "", err
	}

	return channel, err
}
