// Package nslsolver provides a Go client for the NSLSolver captcha solving API,
// supporting Cloudflare Turnstile and Challenge solves with automatic retries.
package nslsolver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"
)

const (
	DefaultBaseURL    = "https://api.nslsolver.com"
	DefaultTimeout    = 120 * time.Second
	DefaultMaxRetries = 3
	Version           = "1.0.0"
)

// Client is the NSLSolver API client. Safe for concurrent use.
type Client struct {
	ApiKey     string
	BaseURL    string
	HTTPClient *http.Client
	MaxRetries int
}

// Option configures the Client.
type Option func(*Client)

func WithBaseURL(url string) Option {
	return func(c *Client) { c.BaseURL = url }
}

func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) { c.HTTPClient.Timeout = timeout }
}

func WithMaxRetries(maxRetries int) Option {
	return func(c *Client) { c.MaxRetries = maxRetries }
}

// WithHTTPClient overrides the default HTTP client entirely.
// When used, WithTimeout has no effect.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) { c.HTTPClient = httpClient }
}

// NewClient creates a new NSLSolver client.
func NewClient(apiKey string, opts ...Option) *Client {
	c := &Client{
		ApiKey:  apiKey,
		BaseURL: DefaultBaseURL,
		HTTPClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		MaxRetries: DefaultMaxRetries,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// SolveTurnstile solves a Cloudflare Turnstile captcha. SiteKey and URL are required.
func (c *Client) SolveTurnstile(ctx context.Context, params TurnstileParams) (*TurnstileResult, error) {
	if params.SiteKey == "" {
		return nil, fmt.Errorf("nslsolver: SiteKey is required")
	}
	if params.URL == "" {
		return nil, fmt.Errorf("nslsolver: URL is required")
	}

	reqBody := solveRequest{
		Type:      "turnstile",
		SiteKey:   params.SiteKey,
		URL:       params.URL,
		Action:    params.Action,
		CData:     params.CData,
		Proxy:     params.Proxy,
		UserAgent: params.UserAgent,
	}

	var result TurnstileResult
	if err := c.doSolve(ctx, reqBody, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SolveChallenge solves a Cloudflare Challenge page. URL and Proxy are required.
func (c *Client) SolveChallenge(ctx context.Context, params ChallengeParams) (*ChallengeResult, error) {
	if params.URL == "" {
		return nil, fmt.Errorf("nslsolver: URL is required")
	}
	if params.Proxy == "" {
		return nil, fmt.Errorf("nslsolver: Proxy is required for challenge type")
	}

	reqBody := solveRequest{
		Type:      "challenge",
		URL:       params.URL,
		Proxy:     params.Proxy,
		UserAgent: params.UserAgent,
	}

	var result ChallengeResult
	if err := c.doSolve(ctx, reqBody, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetBalance returns the account balance and configuration for the API key.
func (c *Client) GetBalance(ctx context.Context) (*BalanceResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/balance", nil)
	if err != nil {
		return nil, fmt.Errorf("nslsolver: failed to create request: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("nslsolver: request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("nslsolver: failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, parseErrorResponse(resp.StatusCode, body)
	}

	var result BalanceResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("nslsolver: failed to parse response: %w", err)
	}
	return &result, nil
}

func (c *Client) doSolve(ctx context.Context, reqBody solveRequest, result interface{}) error {
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("nslsolver: failed to marshal request: %w", err)
	}

	var lastErr error
	attempts := 1 + c.MaxRetries

	for attempt := 0; attempt < attempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("nslsolver: %w", err)
		}

		if attempt > 0 {
			backoff := time.Duration(math.Pow(2, float64(attempt-1))) * time.Second
			select {
			case <-ctx.Done():
				return fmt.Errorf("nslsolver: %w", ctx.Err())
			case <-time.After(backoff):
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/solve", bytes.NewReader(jsonBody))
		if err != nil {
			return fmt.Errorf("nslsolver: failed to create request: %w", err)
		}

		c.setHeaders(req)
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("nslsolver: request failed: %w", err)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("nslsolver: failed to read response: %w", err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			apiErr := parseErrorResponse(resp.StatusCode, body)
			if apiErr.Retryable && attempt < attempts-1 {
				lastErr = apiErr
				continue
			}
			return apiErr
		}

		if err := json.Unmarshal(body, result); err != nil {
			return fmt.Errorf("nslsolver: failed to parse response: %w", err)
		}
		return nil
	}

	return lastErr
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("X-API-Key", c.ApiKey)
	req.Header.Set("User-Agent", "nslsolver-go/"+Version)
	req.Header.Set("Accept", "application/json")
}

func parseErrorResponse(statusCode int, body []byte) *APIError {
	var parsed struct {
		Message string `json:"message"`
		Error   string `json:"error"`
	}

	message := ""
	if err := json.Unmarshal(body, &parsed); err == nil {
		if parsed.Message != "" {
			message = parsed.Message
		} else if parsed.Error != "" {
			message = parsed.Error
		}
	}

	if message == "" {
		switch statusCode {
		case 400:
			message = "bad request: invalid or missing parameters"
		case 401:
			message = "unauthorized: invalid or missing API key"
		case 402:
			message = "insufficient balance"
		case 403:
			message = "captcha type not allowed for this API key"
		case 429:
			message = "rate limit exceeded"
		case 503:
			message = "backend service temporarily unavailable"
		default:
			message = fmt.Sprintf("unexpected error (HTTP %d)", statusCode)
		}
	}

	return newAPIError(statusCode, message)
}
