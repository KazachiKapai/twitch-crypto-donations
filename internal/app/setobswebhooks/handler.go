package setobswebhooks

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"
	"twitch-crypto-donations/internal/pkg/environment"
	"twitch-crypto-donations/internal/pkg/middleware"
	"twitch-crypto-donations/internal/pkg/obsservice"
)

type Database interface {
	QueryRow(query string, args ...any) *sql.Row
	Exec(query string, args ...any) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type ObsService interface {
	CreateChannel(request obsservice.ChannelCreateRequest) (*obsservice.ChannelCreateResponse, error)
}

type RequestBody struct {
	Wallet string `json:"wallet"`
}

type ResponseBody struct {
	AlertWidgetUrl string `json:"alert_widget_url"`
	MediaWidgetUrl string `json:"media_widget_url"`
}

type (
	Request  = middleware.Request[RequestBody]
	Response = middleware.Response[ResponseBody]
)

type Handler struct {
	db         Database
	obsService ObsService
	obsDomain  environment.OBSServiceDomain
}

func New(db Database, obsService ObsService, obsDomain environment.OBSServiceDomain) *Handler {
	return &Handler{db: db, obsDomain: obsDomain, obsService: obsService}
}

func (h *Handler) Handle(ctx context.Context, request Request) (*Response, error) {
	if alertsWidgetUrl, mediaWidgetUrl, exists := h.urlsFromWallet(request.Body.Wallet); exists {
		return &Response{
			Body: ResponseBody{
				AlertWidgetUrl: alertsWidgetUrl,
				MediaWidgetUrl: mediaWidgetUrl,
			},
			StatusCode: http.StatusConflict,
		}, nil
	}

	response, err := h.obsService.CreateChannel(obsservice.ChannelCreateRequest{
		StreamerId: request.Body.Wallet,
	})
	if err != nil {
		return nil, err
	}

	err = h.createNewWallet(
		ctx,
		request.Body.Wallet,
		response.Channel,
		response.WidgetToken,
		response.AlertsWidgetUrl,
		response.MediaWidgetUrl,
		response.WebhookUrl,
		response.WebhookSecret,
	)

	return &Response{
		Body: ResponseBody{
			MediaWidgetUrl: response.MediaWidgetUrl,
			AlertWidgetUrl: response.AlertsWidgetUrl,
		},
		StatusCode: http.StatusCreated,
	}, err
}

func (h *Handler) urlsFromWallet(wallet string) (string, string, bool) {
	const query = `SELECT alerts_widget_url, media_widget_url FROM users WHERE wallet = $1;`

	var (
		alertsWidgetURL string
		mediaWidgetURL  string
	)
	err := h.db.QueryRow(query, wallet).Scan(&alertsWidgetURL, &mediaWidgetURL)
	if err != nil {
		return "", "", false
	}

	return alertsWidgetURL, mediaWidgetURL, true
}

func (h *Handler) createNewWallet(ctx context.Context, wallet, channel, widgetToken, alertsWidgetUrl, mediaWidgetUrl, webhookUrl, webhookSecret string) error {
	const insertQuery = `
		INSERT INTO users (wallet, channel, widget_token, alerts_widget_url, media_widget_url, webhook_url, webhook_secret)
		VALUES ($1, $2, $3, $4, $5, $6, $7);
	`

	c, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := h.db.ExecContext(c, insertQuery, wallet, channel, widgetToken, alertsWidgetUrl, mediaWidgetUrl, webhookUrl, webhookSecret)
	if err != nil {
		return fmt.Errorf("failed to insert wallet: %w", err)
	}

	return nil
}
