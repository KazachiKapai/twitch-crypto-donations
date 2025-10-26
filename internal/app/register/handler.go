package register

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"twitch-crypto-donations/internal/pkg/environment"
	"twitch-crypto-donations/internal/pkg/middleware"
)

type Database interface {
	QueryRow(query string, args ...any) *sql.Row
	Exec(query string, args ...any) (sql.Result, error)
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

type Handler struct {
	db                Database
	donationUrlPrefix environment.DonationUrlPrefix
}

func New(db Database, donationUrlPrefix environment.DonationUrlPrefix) *Handler {
	return &Handler{db: db, donationUrlPrefix: donationUrlPrefix}
}

func (h *Handler) Handle(_ context.Context, request Request) (*Response, error) {
	if donationUrl, exists := h.donationUrlFromWallet(request.Body.Wallet); exists {
		return &Response{
			Body:       ResponseBody{DonationUrl: donationUrl},
			StatusCode: http.StatusConflict,
		}, nil
	}

	donationUrl, err := h.createNewWallet(request.Body.Wallet)
	if err != nil {
		return nil, err
	}

	return &Response{
		Body:       ResponseBody{DonationUrl: donationUrl},
		StatusCode: http.StatusCreated,
	}, nil
}

func (h *Handler) donationUrlFromWallet(wallet string) (string, bool) {
	const query = `SELECT donation_url FROM users WHERE wallet = $1;`

	var donationURL string
	err := h.db.QueryRow(query, wallet).Scan(&donationURL)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false
	}
	if err != nil {
		return "", false
	}

	return donationURL, true
}

func (h *Handler) createNewWallet(wallet string) (string, error) {
	donationURL := fmt.Sprintf("%s%s", h.donationUrlPrefix, wallet)

	const insertQuery = `
		INSERT INTO users (wallet, donation_url)
		VALUES ($1, $2)
		RETURNING donation_url;
	`

	var returnedURL string
	err := h.db.QueryRow(insertQuery, wallet, donationURL).Scan(&returnedURL)
	if err != nil {
		return "", err
	}

	return returnedURL, nil
}
