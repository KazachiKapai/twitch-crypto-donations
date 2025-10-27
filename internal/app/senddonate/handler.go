package senddonate

import (
	"context"
	"net/http"
	"twitch-crypto-donations/internal/pkg/middleware"
	"twitch-crypto-donations/internal/pkg/obsservice"
)

type ObsService interface {
	WebhookAlert(wallet string, request obsservice.AlertEvent) (any, error)
	WebhookMedia(wallet string, request obsservice.MediaEvent) (any, error)
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
}

func New(obsService ObsService) *Handler {
	return &Handler{obsService: obsService}
}

func (h *Handler) Handle(_ context.Context, request Request) (*Response, error) {
	response := ResponseBody{Errors: make([]Error, 0, 2)}

	if request.Body.MediaEvent != nil && request.Body.MediaEvent.Enable {
		_, err := h.obsService.WebhookMedia(request.Body.Wallet, obsservice.MediaEvent{
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
	}

	if request.Body.AlertEvent != nil && request.Body.AlertEvent.Enable {
		_, err := h.obsService.WebhookAlert(request.Body.Wallet, obsservice.AlertEvent{
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
	}

	statusCode := http.StatusOK
	if len(response.Errors) > 0 {
		statusCode = http.StatusInternalServerError
	}

	return &Response{Body: response, StatusCode: statusCode}, nil
}
