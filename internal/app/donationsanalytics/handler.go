package donationsanalytics

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

type Donation struct {
	ReceiverAddress string    `json:"receiver_address"`
	DonationAmount  string    `json:"donation_amount"`
	SenderUsername  string    `json:"sender_username"`
	Currency        string    `json:"currency"`
	Text            *string   `json:"text"`
	AudioUrl        *string   `json:"audio_url"`
	ImageUrl        *string   `json:"image_url"`
	DurationMs      *float64  `json:"duration_ms"`
	Layout          *string   `json:"layout"`
	Channel         *string   `json:"channel"`
	CreatedAt       time.Time `json:"created_at"`
}

type ResponseBody struct {
	TopSingleDonations   []Donation `json:"top_single_donations"`
	TopVolumeDonations   []Donation `json:"top_volume_donations"`
	TopFrequentDonations []Donation `json:"top_frequent_donations"`
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
			Body: ResponseBody{
				TopSingleDonations:   []Donation{},
				TopVolumeDonations:   []Donation{},
				TopFrequentDonations: []Donation{},
			},
		}, fmt.Errorf("jwt is not found or api middleware is failed")
	}

	topSingleDonations, err := h.getTopSingleDonations(address)
	if err != nil {
		return nil, err
	}

	topVolumeDonations, err := h.getTopVolumeDonations(address)
	if err != nil {
		return nil, err
	}

	topFrequentDonations, err := h.getTopFrequentDonations(address)
	if err != nil {
		return nil, err
	}

	return &Response{
		Body: ResponseBody{
			TopSingleDonations:   topSingleDonations,
			TopVolumeDonations:   topVolumeDonations,
			TopFrequentDonations: topFrequentDonations,
		},
		StatusCode: http.StatusOK,
	}, nil
}

func (h *Handler) getTopSingleDonations(receiver string) ([]Donation, error) {
	query := `
        SELECT receiver, donation_amount, sender_username, 
               currency, text, audio_url, image_url, duration_ms, 
               layout, channel, created_at
        FROM donations_history
        WHERE receiver = $1
        ORDER BY CAST(donation_amount AS DECIMAL) DESC
        LIMIT 10
    `

	rows, err := h.db.Query(query, receiver)
	if err != nil {
		return nil, fmt.Errorf("failed to query top single donations: %w", err)
	}
	defer rows.Close()

	return scanDonations(rows)
}

func (h *Handler) getTopVolumeDonations(receiver string) ([]Donation, error) {
	query := `
        SELECT receiver, 
               SUM(CAST(donation_amount AS DECIMAL))::TEXT as donation_amount,
               sender_username,
               currency,
               MAX(text) as text,
               MAX(audio_url) as audio_url,
               MAX(image_url) as image_url,
               MAX(duration_ms) as duration_ms,
               MAX(layout) as layout,
               MAX(channel) as channel,
               MAX(created_at) as created_at
        FROM donations_history
        WHERE receiver = $1
        GROUP BY receiver, sender_username, currency
        ORDER BY SUM(CAST(donation_amount AS DECIMAL)) DESC
        LIMIT 10
    `

	rows, err := h.db.Query(query, receiver)
	if err != nil {
		return nil, fmt.Errorf("failed to query top volume donations: %w", err)
	}
	defer rows.Close()

	return scanDonations(rows)
}

func (h *Handler) getTopFrequentDonations(receiver string) ([]Donation, error) {
	query := `
        SELECT receiver,
               SUM(CAST(donation_amount AS DECIMAL))::TEXT as donation_amount,
               sender_username,
               currency,
               MAX(text) as text,
               MAX(audio_url) as audio_url,
               MAX(image_url) as image_url,
               MAX(duration_ms) as duration_ms,
               MAX(layout) as layout,
               MAX(channel) as channel,
               MAX(created_at) as created_at
        FROM donations_history
        WHERE receiver = $1
        GROUP BY receiver, sender_username, currency
        ORDER BY COUNT(*) DESC
        LIMIT 10
    `

	rows, err := h.db.Query(query, receiver)
	if err != nil {
		return nil, fmt.Errorf("failed to query top frequent donations: %w", err)
	}
	defer rows.Close()

	return scanDonations(rows)
}

func scanDonations(rows *sql.Rows) ([]Donation, error) {
	var donations []Donation

	for rows.Next() {
		var d Donation
		err := rows.Scan(
			&d.ReceiverAddress, &d.DonationAmount,
			&d.SenderUsername,
			&d.Currency, &d.Text,
			&d.AudioUrl, &d.ImageUrl,
			&d.DurationMs, &d.Layout,
			&d.Channel, &d.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan donation: %w", err)
		}
		donations = append(donations, d)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return donations, nil
}
