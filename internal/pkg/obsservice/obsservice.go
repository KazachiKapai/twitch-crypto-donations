package obsservice

import (
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
	"twitch-crypto-donations/internal/pkg/environment"
	"twitch-crypto-donations/internal/pkg/http"
	"twitch-crypto-donations/internal/pkg/logger"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type HttpClient interface {
	NewRequest(method, url string) *http.RequestBuilder
	Post(url string) *http.RequestBuilder
	Get(url string) *http.RequestBuilder
	WithLogger(logger http.Logger) *http.Client
}

type Database interface {
	QueryRow(query string, args ...any) *sql.Row
	Exec(query string, args ...any) (sql.Result, error)
}

type ObsService struct {
	obsDomain  environment.OBSServiceDomain
	httpClient HttpClient
	db         Database
}

func New(db Database, httpClient HttpClient, obsDomain environment.OBSServiceDomain) *ObsService {
	return &ObsService{
		db:         db,
		httpClient: httpClient,
		obsDomain:  obsDomain,
	}
}

func (s *ObsService) CreateChannel(request ChannelCreateRequest) (*ChannelCreateResponse, error) {
	url := fmt.Sprintf("%s/channels/create", s.obsDomain)

	var response ChannelCreateResponse
	err := s.httpClient.
		WithLogger(logger.New(logrus.StandardLogger())).
		Post(url).
		WithJSON(request).
		DecodeResponseJSON().
		Parse(&response)
	return &response, err
}

func (s *ObsService) UpdateAlertSettings(wallet string, request AlertSettings) (any, error) {
	url := fmt.Sprintf("%s/channels/settings", s.obsDomain)

	channel, webhookSecret, ok := s.getChannelInfo(wallet)
	if ok {
		request.Channel = channel
	}

	timestamp, nonce, signature, err := s.generateSignature(webhookSecret, request)
	if err != nil {
		return nil, err
	}

	var response any
	err = s.httpClient.
		WithLogger(logger.New(logrus.StandardLogger())).
		Post(url).
		WithJSON(request).
		WithHeaders(map[string]string{
			"x-signature": signature,
			"x-nonce":     nonce,
			"x-timestamp": timestamp,
		}).
		DecodeResponseJSON().
		Parse(&response)
	return response, err
}

func (s *ObsService) WebhookAlert(wallet string, request AlertEvent) (any, error) {
	url := fmt.Sprintf("%s/webhooks/alert", s.obsDomain)

	channel, webhookSecret, ok := s.getChannelInfo(wallet)
	if ok {
		request.Channel = channel
	}

	timestamp, nonce, signature, err := s.generateSignature(webhookSecret, request)
	if err != nil {
		return "", err
	}

	var response any
	err = s.httpClient.
		WithLogger(logger.New(logrus.StandardLogger())).
		Post(url).
		WithJSON(request).
		WithHeaders(map[string]string{
			"x-signature": signature,
			"x-nonce":     nonce,
			"x-timestamp": timestamp,
		}).
		DecodeResponseJSON().
		Parse(&response)
	return response, err
}

func (s *ObsService) WebhookMedia(wallet string, request MediaEvent) (any, error) {
	url := fmt.Sprintf("%s/webhooks/media", s.obsDomain)

	channel, webhookSecret, ok := s.getChannelInfo(wallet)
	if ok {
		request.Channel = channel
	}

	timestamp, nonce, signature, err := s.generateSignature(webhookSecret, request)
	if err != nil {
		return "", err
	}

	var response any
	err = s.httpClient.
		WithLogger(logger.New(logrus.StandardLogger())).
		Post(url).
		WithJSON(request).
		WithHeaders(map[string]string{
			"x-signature": signature,
			"x-nonce":     nonce,
			"x-timestamp": timestamp,
		}).
		DecodeResponseJSON().
		Parse(&response)
	return response, err
}

func (s *ObsService) WebhookSkip(wallet string, request MediaEvent) (any, error) {
	url := fmt.Sprintf("%s/webhooks/skip", s.obsDomain)

	channel, webhookSecret, ok := s.getChannelInfo(wallet)
	if ok {
		request.Channel = channel
	}

	timestamp, nonce, signature, err := s.generateSignature(webhookSecret, request)
	if err != nil {
		return "", err
	}

	var response any
	err = s.httpClient.
		WithLogger(logger.New(logrus.StandardLogger())).
		Post(url).
		WithJSON(request).
		WithHeaders(map[string]string{
			"x-signature": signature,
			"x-nonce":     nonce,
			"x-timestamp": timestamp,
		}).
		DecodeResponseJSON().
		Parse(&response)
	return response, err
}

func (s *ObsService) getChannelInfo(wallet string) (string, string, bool) {
	const query = `SELECT channel, webhook_secret FROM users WHERE wallet = $1;`

	var channel, webhookSecret string
	err := s.db.QueryRow(query, wallet).Scan(&channel, &webhookSecret)
	if err != nil {
		return "", "", false
	}

	return channel, webhookSecret, true
}

func (s *ObsService) generateSignature(webhookSecret string, requestBody any) (string, string, string, error) {
	requestBodyData, err := json.Marshal(requestBody)
	if err != nil {
		return "", "", "", err
	}

	timestamp := strconv.FormatInt(time.Now().UTC().Unix(), 10)

	nonce := strings.ReplaceAll(uuid.New().String(), "-", "")

	mac := hmac.New(sha256.New, []byte(webhookSecret))
	mac.Write(requestBodyData)
	signatureHex := hex.EncodeToString(mac.Sum(nil))

	signature := fmt.Sprintf("sha256=%s", signatureHex)

	return timestamp, nonce, signature, nil
}
