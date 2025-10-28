package senddonate

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"twitch-crypto-donations/internal/pkg/middleware"
	"twitch-crypto-donations/internal/pkg/obsservice"
)

type Database interface {
	QueryRow(query string, args ...any) *sql.Row
	Exec(query string, args ...any) (sql.Result, error)
}

type ObsService interface {
	WebhookAlert(wallet string, request obsservice.AlertEvent) (any, string, error)
	WebhookMedia(wallet string, request obsservice.MediaEvent) (any, string, error)
}

type RequestBody struct {
	Wallet     string   `json:"wallet"`
	Username   *string  `json:"username"`
	Amount     *float64 `json:"amount"`
	Currency   *string  `json:"currency"`
	Message    *string  `json:"message"`
	DurationMs *int64   `json:"duration_ms"`

	AlertEvent *AlertRequest `json:"alert_event"`
	MediaEvent *MediaRequest `json:"media_event"`
}

type AlertRequest struct {
	Enable            bool    `json:"enable"`
	NotificationSound *string `json:"notification_sound"`
	VoiceUrl          *string `json:"voice_url"`
	ImageUrl          *string `json:"image_url"`
	GifUrl            *string `json:"gif_url"`
}

type MediaRequest struct {
	Enable     bool   `json:"enable"`
	YoutubeUrl string `json:"youtube_url"`
	StartTime  *int64 `json:"start_time"`
	EndTime    *int64 `json:"end_time"`
	AutoPlay   *bool  `json:"auto_play"`
	Controls   *bool  `json:"controls"`
	Mute       *bool  `json:"mute"`
}

type ResponseBody struct {
	Errors []Error `json:"errors"`
}

type Error struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

type (
	Request  = middleware.Request[RequestBody]
	Response = middleware.Response[ResponseBody]
)

type Handler struct {
	obsService ObsService
	db         Database
}

func New(obsService ObsService, db Database) *Handler {
	return &Handler{obsService: obsService, db: db}
}

func (h *Handler) Handle(_ context.Context, request Request) (*Response, error) {
	response := ResponseBody{Errors: make([]Error, 0, 2)}
	channels := make(map[string]struct{})

	if request.Body.MediaEvent != nil && request.Body.MediaEvent.Enable {
		_, channel, err := h.obsService.WebhookMedia(request.Body.Wallet, obsservice.MediaEvent{
			Username:   request.Body.Username,
			Amount:     request.Body.Amount,
			Currency:   request.Body.Currency,
			Message:    request.Body.Message,
			DurationMs: request.Body.DurationMs,
			YoutubeUrl: request.Body.MediaEvent.YoutubeUrl,
			StartTime:  request.Body.MediaEvent.StartTime,
			EndTime:    request.Body.MediaEvent.EndTime,
			AutoPlay:   request.Body.MediaEvent.AutoPlay,
			Controls:   request.Body.MediaEvent.Controls,
			Mute:       request.Body.MediaEvent.Mute,
		})

		if err != nil {
			response.Errors = append(response.Errors, Error{Message: err.Error()})
		}

		channels[channel] = struct{}{}
	}

	if request.Body.AlertEvent != nil && request.Body.AlertEvent.Enable {
		_, channel, err := h.obsService.WebhookAlert(request.Body.Wallet, obsservice.AlertEvent{
			Username:          request.Body.Username,
			Amount:            request.Body.Amount,
			Currency:          request.Body.Currency,
			Message:           request.Body.Message,
			DurationMs:        request.Body.DurationMs,
			NotificationSound: request.Body.AlertEvent.NotificationSound,
			VoiceUrl:          request.Body.AlertEvent.VoiceUrl,
			ImageUrl:          request.Body.AlertEvent.ImageUrl,
			GifUrl:            request.Body.AlertEvent.GifUrl,
		})

		if err != nil {
			response.Errors = append(response.Errors, Error{Message: err.Error()})
		}

		channels[channel] = struct{}{}
	}

	if len(response.Errors) > 0 {
		return &Response{Body: response, StatusCode: http.StatusInternalServerError}, nil
	}

	if errors := h.saveDonation(request, channels); len(errors) > 0 {
		return &Response{Body: ResponseBody{Errors: errors}, StatusCode: http.StatusInternalServerError}, nil
	}

	return &Response{Body: response}, nil
}

func (h *Handler) saveDonation(request Request, channels map[string]struct{}) []Error {
	errors := make([]Error, 0, len(channels))

	for channel := range channels {
		var layout string
		if request.Body.MediaEvent != nil && request.Body.MediaEvent.Enable {
			layout = "media"
		} else if request.Body.AlertEvent != nil && request.Body.AlertEvent.Enable {
			layout = "alert"
		}

		amount := ""
		if request.Body.Amount != nil {
			amount = fmt.Sprintf("%f", *request.Body.Amount)
		}

		username := ""
		if request.Body.Username != nil {
			username = *request.Body.Username
		}

		currency := ""
		if request.Body.Currency != nil {
			currency = *request.Body.Currency
		}

		var durationMs *float64
		if request.Body.DurationMs != nil {
			duration := float64(*request.Body.DurationMs)
			durationMs = &duration
		}

		var audioURL, imageURL *string
		if request.Body.AlertEvent != nil {
			audioURL = request.Body.AlertEvent.VoiceUrl
			if request.Body.AlertEvent.ImageUrl != nil {
				imageURL = request.Body.AlertEvent.ImageUrl
			} else {
				imageURL = request.Body.AlertEvent.GifUrl
			}
		}

		_, err := h.db.Exec(
			`INSERT INTO donations_history 
			(donation_amount, sender_address, sender_username, currency, text, audio_url, image_url, duration_ms, layout, channel) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
			amount, request.Body.Wallet,
			username, currency,
			request.Body.Message, audioURL,
			imageURL, durationMs,
			layout, channel,
		)

		if err != nil {
			errors = append(errors, Error{Message: "Failed to save donation history: " + err.Error()})
		}
	}

	return errors
}
