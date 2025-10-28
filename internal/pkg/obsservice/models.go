package obsservice

type AlertEvent struct {
	Channel    string   `json:"channel"`
	Username   *string  `json:"username"`
	Amount     *float64 `json:"amount"`
	Currency   *string  `json:"currency"`
	Message    *string  `json:"message"`
	DurationMs *int64   `json:"duration_ms"`

	NotificationSound *string `json:"notification_sound"`
	VoiceUrl          *string `json:"voice_url"`
	ImageUrl          *string `json:"image_url"`
	GifUrl            *string `json:"gif_url"`
}

type AlertSettings struct {
	Channel                  string  `json:"channel"`
	DefaultNotificationSound *string `json:"default_notification_sound"`
	DefaultAlertImage        *string `json:"default_alert_image"`
	DefaultAlertDuration     *int64  `json:"default_alert_duration"`
}

type ChannelCreateRequest struct {
	StreamerId string `json:"streamer_id"`
}

type ChannelCreateResponse struct {
	Ok              bool   `json:"ok"`
	Channel         string `json:"channel"`
	WidgetToken     string `json:"widget_token"`
	WebhookSecret   string `json:"webhook_secret"`
	AlertsWidgetUrl string `json:"alerts_widget_url"`
	MediaWidgetUrl  string `json:"media_widget_url"`
	WebhookUrl      string `json:"webhook_url"`
}

type MediaEvent struct {
	Channel    string   `json:"channel"`
	Username   *string  `json:"username"`
	Amount     *float64 `json:"amount"`
	Currency   *string  `json:"currency"`
	Message    *string  `json:"message"`
	DurationMs *int64   `json:"duration_ms"`

	YoutubeUrl string `json:"youtube_url"`
	StartTime  *int64 `json:"start_time"`
	EndTime    *int64 `json:"end_time"`
	AutoPlay   *bool  `json:"auto_play"`
	Controls   *bool  `json:"controls"`
	Mute       *bool  `json:"mute"`
}

type SkipRequest struct {
	Channel    string `json:"channel"`
	WidgetType string `json:"widget_type"`
}

type GetAlertSettingsResponse struct {
	Ok                       bool    `json:"ok"`
	Channel                  string  `json:"channel"`
	DefaultNotificationSound *string `json:"default_notification_sound"`
	DefaultAlertImage        *string `json:"default_alert_image"`
	DefaultAlertDuration     *int64  `json:"default_alert_duration"`
}

type ErrorResponse struct {
	Detail []struct {
		Items struct {
			Loc []struct {
				Items *string `json:"items"`
			} `json:"loc"`
			Msg  string `json:"msg"`
			Type string `json:"type"`
		} `json:"items"`
	} `json:"detail"`
}
