package senddonate

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"twitch-crypto-donations/internal/pkg/environment"
	"twitch-crypto-donations/internal/pkg/middleware"

	"github.com/google/uuid"
)

type Database interface {
	QueryRow(query string, args ...any) *sql.Row
	Exec(query string, args ...any) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type RequestBody struct {
	Wallet     string  `json:"wallet"`
	Amount     float64 `json:"amount"`
	AudioUrl   string  `json:"audio_url"`
	Currency   string  `json:"currency"`
	DurationMs float64 `json:"duration_ms"`
	ImageUrl   string  `json:"image_url"`
	Layout     string  `json:"layout"`
	Text       string  `json:"text"`
	Username   string  `json:"username"`
}

type ResponseBody struct{}

type (
	Request  = middleware.Request[RequestBody]
	Response = middleware.Response[ResponseBody]
)

type ObsEventsDonateRequest struct {
	Amount     float64 `json:"amount"`
	AudioUrl   string  `json:"audio_url"`
	Channel    string  `json:"channel"`
	Currency   string  `json:"currency"`
	DurationMs float64 `json:"duration_ms"`
	ImageUrl   string  `json:"image_url"`
	Layout     string  `json:"layout"`
	Text       string  `json:"text"`
	Username   string  `json:"username"`
}

type ObsEventsDonateResponse string

type Handler struct {
	db         Database
	httpClient HttpClient
	obsDomain  environment.OBSServiceDomain
}

func New(db Database, httpClient HttpClient, obsDomain environment.OBSServiceDomain) *Handler {
	return &Handler{db: db, httpClient: httpClient, obsDomain: obsDomain}
}

func (h *Handler) Handle(ctx context.Context, request Request) (*Response, error) {
	channel, webhookSecret, exists := h.getChannelInfo(request.Body.Wallet)
	if !exists {
		return &Response{
			StatusCode: http.StatusNotFound,
		}, nil
	}

	if err := h.sendDonate(ctx, channel, webhookSecret, request); err != nil {
		return nil, fmt.Errorf("failed to send donate: %w", err)
	}

	if err := h.saveDonateHistory(ctx, request); err != nil {
		log.Printf("failed to save donation history: %+v", err)
	}

	return &Response{StatusCode: http.StatusNoContent}, nil
}

func (h *Handler) getChannelInfo(wallet string) (string, string, bool) {
	const query = `SELECT channel, webhook_secret FROM users WHERE wallet = $1;`

	var channel, webhookSecret string
	err := h.db.QueryRow(query, wallet).Scan(&channel, &webhookSecret)
	if err != nil {
		return "", "", false
	}

	return channel, webhookSecret, true
}

func (h *Handler) sendDonate(ctx context.Context, channel, webhookSecret string, request Request) error {
	url := fmt.Sprintf("%s/events/donate", h.obsDomain)

	requestBody := ObsEventsDonateRequest{
		Channel:    channel,
		Username:   request.Body.Username,
		Amount:     request.Body.Amount,
		Currency:   request.Body.Currency,
		Text:       request.Body.Text,
		AudioUrl:   request.Body.AudioUrl,
		ImageUrl:   request.Body.ImageUrl,
		DurationMs: request.Body.DurationMs,
		Layout:     request.Body.Layout,
	}

	requestBodyJson, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	timestamp, nonce, signature := h.generateSignature(webhookSecret, requestBodyJson)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(requestBodyJson))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-signature", signature)
	req.Header.Set("x-nonce", nonce)
	req.Header.Set("x-timestamp", timestamp)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("OBS service error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

func (h *Handler) generateSignature(webhookSecret string, requestBody []byte) (string, string, string) {
	timestamp := strconv.FormatInt(time.Now().UTC().Unix(), 10)

	nonce := strings.ReplaceAll(uuid.New().String(), "-", "")

	mac := hmac.New(sha256.New, []byte(webhookSecret))
	mac.Write(requestBody)
	signatureHex := hex.EncodeToString(mac.Sum(nil))

	signature := fmt.Sprintf("sha256=%s", signatureHex)

	return timestamp, nonce, signature
}

func (h *Handler) saveDonateHistory(ctx context.Context, request Request) error {
	const query = `
       INSERT INTO donations_history (
          sender_address, sender_username,
          donation_amount, currency,
          text, audio_url,
          image_url, duration_ms,
          layout, created_at
       )
       VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);
    `

	now := time.Now().UTC()
	body := request.Body

	_, err := h.db.ExecContext(
		ctx, query, body.Wallet,
		body.Username, body.Amount,
		body.Currency, body.Text,
		body.AudioUrl, body.ImageUrl,
		body.DurationMs, body.Layout, now,
	)

	if err != nil {
		return fmt.Errorf("error executing donation insert: %w", err)
	}

	return nil
}
