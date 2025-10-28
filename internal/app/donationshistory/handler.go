package donationshistory

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"
	"twitch-crypto-donations/internal/pkg/middleware"
)

type Database interface {
	QueryRow(query string, args ...any) *sql.Row
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type ResponseBody struct {
	Donations   []Donation `json:"donations"`
	TotalAmount int64      `json:"amount"`
}

type Donation struct {
	DonationAmount string    `json:"donation_amount"`
	SenderAddress  string    `json:"sender_address"`
	SenderUsername string    `json:"sender_username"`
	Currency       string    `json:"currency"`
	Text           *string   `json:"text"`
	AudioUrl       *string   `json:"audio_url"`
	ImageUrl       *string   `json:"image_url"`
	DurationMs     *float64  `json:"duration_ms"`
	Layout         *string   `json:"layout"`
	Channel        *string   `json:"channel"`
	CreatedAt      time.Time `json:"created_at"`
}

type (
	Request  = middleware.Request[struct{}]
	Response = middleware.Response[ResponseBody]
)

type Handler struct {
	db Database
}

func New(db Database) *Handler {
	return &Handler{db: db}
}

func (h *Handler) Handle(_ context.Context, request Request) (*Response, error) {
	address, exists := request.Context[middleware.AddressKey].(string)
	if !exists || address == "" {
		return &Response{
			StatusCode: http.StatusUnauthorized,
			Body:       ResponseBody{Donations: []Donation{}, TotalAmount: 0},
		}, fmt.Errorf("jwt is not found or api middleware is failed")
	}

	donationsHistory, totalAmount, err := h.getDonationsHistory(address)
	if err != nil {
		return nil, err
	}

	return &Response{
		Body:       ResponseBody{Donations: donationsHistory, TotalAmount: totalAmount},
		StatusCode: http.StatusOK,
	}, nil
}

func (h *Handler) getDonationsHistory(address string) ([]Donation, int64, error) {
	query := `
        SELECT 
            donation_amount, sender_address, sender_username, currency, 
            text, audio_url, image_url, duration_ms, layout, channel, created_at
        FROM donations_history
        WHERE sender_address = $1
        ORDER BY created_at DESC`

	rows, err := h.db.Query(query, address)
	if err != nil {
		return nil, 0, fmt.Errorf("query failed: %w", err)
	}

	defer rows.Close()

	donations := make([]Donation, 0, 10)
	var totalAmount int64 = 0

	for rows.Next() {
		var d Donation

		err = rows.Scan(
			&d.DonationAmount, &d.SenderAddress,
			&d.SenderUsername, &d.Currency,
			&d.Text, &d.AudioUrl, &d.ImageUrl,
			&d.DurationMs, &d.Layout,
			&d.Channel, &d.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan row: %w", err)
		}

		donations = append(donations, d)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("row iteration error: %w", err)
	}

	return donations, totalAmount, nil
}
