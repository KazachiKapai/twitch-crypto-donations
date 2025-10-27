package updatedefaultobssettings

import (
	"context"
	"fmt"
	"net/http"
	"twitch-crypto-donations/internal/pkg/middleware"
	"twitch-crypto-donations/internal/pkg/obsservice"
)

type ObsService interface {
	UpdateAlertSettings(wallet string, request obsservice.AlertSettings) (any, error)
}

type RequestBody struct {
	DefaultNotificationSound *string `json:"default_notification_sound"`
	DefaultAlertImage        *string `json:"default_alert_image"`
	DefaultAlertDuration     *int64  `json:"default_alert_duration"`
}

type ResponseBody struct{}

type (
	Request  = middleware.Request[RequestBody]
	Response = middleware.Response[ResponseBody]
)

type Handler struct {
	obsService ObsService
}

func New(obsService ObsService) *Handler {
	return &Handler{
		obsService: obsService,
	}
}

func (h *Handler) Handle(_ context.Context, request Request) (*Response, error) {
	address, exists := request.Context[middleware.AddressKey].(string)
	if !exists || address == "" {
		return &Response{
			StatusCode: http.StatusUnauthorized,
			Body:       ResponseBody{},
		}, fmt.Errorf("jwt is not found or api middleware is failed")
	}

	_, err := h.obsService.UpdateAlertSettings(address, obsservice.AlertSettings{
		DefaultAlertImage:        request.Body.DefaultAlertImage,
		DefaultNotificationSound: request.Body.DefaultNotificationSound,
		DefaultAlertDuration:     request.Body.DefaultAlertDuration,
	})
	return &Response{StatusCode: http.StatusNoContent}, err
}
