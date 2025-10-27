package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Logger interface {
	Info(msg string, ctx ...interface{})
}

type Client struct {
	httpClient HttpClient
	logger     Logger
}

func New(client HttpClient) *Client {
	return &Client{
		httpClient: client,
	}
}

func (c *Client) WithLogger(logger Logger) *Client {
	c.logger = logger
	return c
}

type RetryPolicy struct {
	MaxRetries  int
	InitialWait time.Duration
	MaxWait     time.Duration
	Multiplier  float64
	RetryIf     func(resp *http.Response, err error) bool
}

func DefaultRetryPolicy() *RetryPolicy {
	return &RetryPolicy{
		MaxRetries:  3,
		InitialWait: 100 * time.Millisecond,
		MaxWait:     5 * time.Second,
		Multiplier:  2.0,
		RetryIf:     DefaultRetryCondition,
	}
}

func DefaultRetryCondition(resp *http.Response, err error) bool {
	if err != nil {
		return true
	}
	return resp.StatusCode >= 500 && resp.StatusCode < 600
}

type RequestBuilder struct {
	client       *Client
	ctx          context.Context
	method       string
	url          string
	body         interface{}
	headers      map[string]string
	err          error
	responseBody []byte
	retryPolicy  *RetryPolicy
}

func (c *Client) NewRequest(method, url string) *RequestBuilder {
	return &RequestBuilder{
		client:  c,
		ctx:     context.Background(),
		method:  method,
		url:     url,
		headers: make(map[string]string),
	}
}

func (c *Client) Get(url string) *RequestBuilder {
	return c.NewRequest(http.MethodGet, url)
}

func (c *Client) Post(url string) *RequestBuilder {
	return c.NewRequest(http.MethodPost, url)
}

func (c *Client) Put(url string) *RequestBuilder {
	return c.NewRequest(http.MethodPut, url)
}

func (c *Client) Delete(url string) *RequestBuilder {
	return c.NewRequest(http.MethodDelete, url)
}

func (rb *RequestBuilder) WithContext(ctx context.Context) *RequestBuilder {
	rb.ctx = ctx
	return rb
}

func (rb *RequestBuilder) WithJSON(body interface{}) *RequestBuilder {
	if rb.err != nil {
		return rb
	}
	rb.body = body
	rb.headers["Content-Type"] = "application/json"
	return rb
}

func (rb *RequestBuilder) WithHeader(key, value string) *RequestBuilder {
	rb.headers[key] = value
	return rb
}

func (rb *RequestBuilder) WithHeaders(headers map[string]string) *RequestBuilder {
	for k, v := range headers {
		rb.headers[k] = v
	}
	return rb
}

func (rb *RequestBuilder) WithRetry(policy *RetryPolicy) *RequestBuilder {
	rb.retryPolicy = policy
	return rb
}

func (rb *RequestBuilder) WithDefaultRetry() *RequestBuilder {
	rb.retryPolicy = DefaultRetryPolicy()
	return rb
}

func (rb *RequestBuilder) WithRetries(maxRetries int) *RequestBuilder {
	policy := DefaultRetryPolicy()
	policy.MaxRetries = maxRetries
	rb.retryPolicy = policy
	return rb
}

func (rb *RequestBuilder) Do() (*http.Response, error) {
	if rb.err != nil {
		return nil, rb.err
	}

	var bodyBytes []byte
	if rb.body != nil {
		var err error
		bodyBytes, err = json.Marshal(rb.body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	if rb.retryPolicy == nil {
		return rb.executeRequest(bodyBytes)
	}

	return rb.executeWithRetry(bodyBytes)
}

func (rb *RequestBuilder) executeRequest(bodyBytes []byte) (*http.Response, error) {
	var bodyReader io.Reader
	if bodyBytes != nil {
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(rb.ctx, rb.method, rb.url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for key, value := range rb.headers {
		req.Header.Set(key, value)
	}

	if rb.client.logger != nil {
		rb.client.logger.Info("http_request",
			"method", rb.method,
			"url", rb.url,
			"headers", rb.headers,
		)
	}

	start := time.Now()
	resp, err := rb.client.httpClient.Do(req)
	duration := time.Since(start)

	if rb.client.logger != nil {
		if err != nil {
			rb.client.logger.Info("http_request_error",
				"method", rb.method,
				"url", rb.url,
				"duration_ms", duration.Milliseconds(),
				"error", err.Error(),
			)
		} else {
			rb.client.logger.Info("http_response",
				"method", rb.method,
				"url", rb.url,
				"status_code", resp.StatusCode,
				"duration_ms", duration.Milliseconds(),
			)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	return resp, nil
}

func (rb *RequestBuilder) executeWithRetry(bodyBytes []byte) (*http.Response, error) {
	var resp *http.Response
	var err error

	wait := rb.retryPolicy.InitialWait

	for attempt := 0; attempt <= rb.retryPolicy.MaxRetries; attempt++ {
		if rb.client.logger != nil && attempt > 0 {
			rb.client.logger.Info("http_retry",
				"method", rb.method,
				"url", rb.url,
				"attempt", attempt,
				"wait_ms", wait.Milliseconds(),
			)
		}

		resp, err = rb.executeRequest(bodyBytes)

		shouldRetry := rb.retryPolicy.RetryIf(resp, err)

		if !shouldRetry || attempt == rb.retryPolicy.MaxRetries {
			return resp, err
		}

		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}

		select {
		case <-rb.ctx.Done():
			return nil, rb.ctx.Err()
		case <-time.After(wait):
			wait = time.Duration(float64(wait) * rb.retryPolicy.Multiplier)
			if wait > rb.retryPolicy.MaxWait {
				wait = rb.retryPolicy.MaxWait
			}
		}
	}

	return resp, err
}

func (rb *RequestBuilder) DecodeResponseJSON() *RequestBuilder {
	if rb.err != nil {
		return rb
	}

	resp, err := rb.Do()
	if err != nil {
		rb.err = err
		return rb
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		rb.err = fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
		return rb
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		rb.err = fmt.Errorf("failed to read response body: %w", err)
		return rb
	}

	rb.responseBody = body
	return rb
}

func (rb *RequestBuilder) Parse(target interface{}) error {
	if rb.err != nil {
		return rb.err
	}

	if rb.responseBody == nil {
		return fmt.Errorf("no response body to parse, call DecodeResponseJSON first")
	}

	if err := json.Unmarshal(rb.responseBody, target); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}
