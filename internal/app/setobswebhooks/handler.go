package setobswebhooks

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"twitch-crypto-donations/internal/pkg/environment"
	"twitch-crypto-donations/internal/pkg/middleware"
)

type Database interface {
	QueryRow(query string, args ...any) *sql.Row
	Exec(query string, args ...any) (sql.Result, error)
}

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type RequestBody struct {
	Wallet string `json:"wallet"`
}

type ResponseBody struct {
	DonationUrl string `json:"donation_url"`
}

type (
	Request  = middleware.Request[RequestBody]
	Response = middleware.Response[ResponseBody]
)

type ObsChannelUpsertResponse struct {
	Ok            bool   `json:"ok"`
	Channel       string `json:"channel"`
	AlertUrl      string `json:"alert_url"`
	ViewerToken   string `json:"viewer_token"`
	WebhookSecret string `json:"webhook_secret"`
}

type ObsChannelUpsertRequest struct {
	StreamerId string `json:"streamer_id"`
}

type Handler struct {
	db         Database
	httpClient HttpClient
	obsDomain  environment.OBSServiceDomain
}

func New(db Database, httpClient HttpClient, obsDomain environment.OBSServiceDomain) *Handler {
	return &Handler{db: db, obsDomain: obsDomain, httpClient: httpClient}
}

func (h *Handler) Handle(_ context.Context, request Request) (*Response, error) {
	if donationUrl, exists := h.donationUrlFromWallet(request.Body.Wallet); exists {
		return &Response{
			Body:       ResponseBody{DonationUrl: donationUrl},
			StatusCode: http.StatusConflict,
		}, nil
	}

	obsChannelResponse, err := h.receiveOBSData(request.Body.Wallet)
	if err != nil {
		return nil, err
	}

	donationUrl, err := h.createNewWallet(
		request.Body.Wallet,
		obsChannelResponse.AlertUrl,
		obsChannelResponse.ViewerToken,
		obsChannelResponse.WebhookSecret,
		obsChannelResponse.Channel,
	)
	if err != nil {
		return nil, err
	}

	return &Response{
		Body:       ResponseBody{DonationUrl: donationUrl},
		StatusCode: http.StatusCreated,
	}, nil
}

func (h *Handler) receiveOBSData(wallet string) (*ObsChannelUpsertResponse, error) {
	url := fmt.Sprintf("%s/channels/upsert", h.obsDomain)

	requestBody := ObsChannelUpsertRequest{StreamerId: wallet}
	requestBodyJson, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(requestBodyJson))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var response ObsChannelUpsertResponse
	if err = json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

func (h *Handler) donationUrlFromWallet(wallet string) (string, bool) {
	const query = `SELECT donation_url FROM users WHERE wallet = $1;`

	var donationURL string
	err := h.db.QueryRow(query, wallet).Scan(&donationURL)
	if err != nil {
		return "", false
	}

	return donationURL, true
}

func (h *Handler) createNewWallet(wallet, donationURL, viewerToken, webhookSecret, channel string) (string, error) {
	const insertQuery = `
		INSERT INTO users (wallet, donation_url, viewer_token, webhook_secret, channel)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING donation_url;
	`

	var returnedURL string
	err := h.db.QueryRow(insertQuery, wallet, donationURL, viewerToken, webhookSecret, channel).Scan(&returnedURL)
	if err != nil {
		return "", err
	}

	return returnedURL, nil
}
