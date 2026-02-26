package nslsolver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewClient_Defaults(t *testing.T) {
	client := NewClient("test-key")

	if client.ApiKey != "test-key" {
		t.Errorf("expected ApiKey 'test-key', got '%s'", client.ApiKey)
	}
	if client.BaseURL != DefaultBaseURL {
		t.Errorf("expected BaseURL '%s', got '%s'", DefaultBaseURL, client.BaseURL)
	}
	if client.MaxRetries != DefaultMaxRetries {
		t.Errorf("expected MaxRetries %d, got %d", DefaultMaxRetries, client.MaxRetries)
	}
	if client.HTTPClient == nil {
		t.Fatal("expected HTTPClient to be non-nil")
	}
	if client.HTTPClient.Timeout != DefaultTimeout {
		t.Errorf("expected timeout %v, got %v", DefaultTimeout, client.HTTPClient.Timeout)
	}
}

func TestNewClient_WithOptions(t *testing.T) {
	customHTTP := &http.Client{Timeout: 30 * time.Second}
	client := NewClient("test-key",
		WithBaseURL("https://custom.api.com"),
		WithMaxRetries(5),
		WithHTTPClient(customHTTP),
	)

	if client.BaseURL != "https://custom.api.com" {
		t.Errorf("expected BaseURL 'https://custom.api.com', got '%s'", client.BaseURL)
	}
	if client.MaxRetries != 5 {
		t.Errorf("expected MaxRetries 5, got %d", client.MaxRetries)
	}
	if client.HTTPClient != customHTTP {
		t.Error("expected custom HTTP client")
	}
}

func TestNewClient_WithTimeout(t *testing.T) {
	client := NewClient("test-key", WithTimeout(30*time.Second))

	if client.HTTPClient.Timeout != 30*time.Second {
		t.Errorf("expected timeout 30s, got %v", client.HTTPClient.Timeout)
	}
}

func TestSolveTurnstile_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/solve" {
			t.Errorf("expected /solve, got %s", r.URL.Path)
		}
		if r.Header.Get("X-API-Key") != "test-key" {
			t.Errorf("expected X-API-Key 'test-key', got '%s'", r.Header.Get("X-API-Key"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type 'application/json', got '%s'", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("User-Agent") != "nslsolver-go/"+Version {
			t.Errorf("expected User-Agent 'nslsolver-go/%s', got '%s'", Version, r.Header.Get("User-Agent"))
		}

		var body solveRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if body.Type != "turnstile" {
			t.Errorf("expected type 'turnstile', got '%s'", body.Type)
		}
		if body.SiteKey != "0x4AAAA" {
			t.Errorf("expected site_key '0x4AAAA', got '%s'", body.SiteKey)
		}
		if body.URL != "https://example.com" {
			t.Errorf("expected url 'https://example.com', got '%s'", body.URL)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TurnstileResult{
			Success: true,
			Token:   "solved-token-123",
			Type:    "turnstile",
		})
	}))
	defer server.Close()

	client := NewClient("test-key", WithBaseURL(server.URL), WithMaxRetries(0))
	result, err := client.SolveTurnstile(context.Background(), TurnstileParams{
		SiteKey: "0x4AAAA",
		URL:     "https://example.com",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Error("expected Success to be true")
	}
	if result.Token != "solved-token-123" {
		t.Errorf("expected token 'solved-token-123', got '%s'", result.Token)
	}
	if result.Type != "turnstile" {
		t.Errorf("expected type 'turnstile', got '%s'", result.Type)
	}
}

func TestSolveTurnstile_OptionalFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body solveRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if body.Action != "login" {
			t.Errorf("expected action 'login', got '%s'", body.Action)
		}
		if body.CData != "custom-data" {
			t.Errorf("expected cdata 'custom-data', got '%s'", body.CData)
		}
		if body.Proxy != "http://proxy:8080" {
			t.Errorf("expected proxy 'http://proxy:8080', got '%s'", body.Proxy)
		}
		if body.UserAgent != "Mozilla/5.0" {
			t.Errorf("expected user_agent 'Mozilla/5.0', got '%s'", body.UserAgent)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TurnstileResult{Success: true, Token: "t", Type: "turnstile"})
	}))
	defer server.Close()

	client := NewClient("test-key", WithBaseURL(server.URL), WithMaxRetries(0))
	_, err := client.SolveTurnstile(context.Background(), TurnstileParams{
		SiteKey:   "0x4AAAA",
		URL:       "https://example.com",
		Action:    "login",
		CData:     "custom-data",
		Proxy:     "http://proxy:8080",
		UserAgent: "Mozilla/5.0",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSolveTurnstile_MissingSiteKey(t *testing.T) {
	client := NewClient("test-key")
	_, err := client.SolveTurnstile(context.Background(), TurnstileParams{
		URL: "https://example.com",
	})

	if err == nil {
		t.Fatal("expected error for missing SiteKey")
	}
	if err.Error() != "nslsolver: SiteKey is required" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestSolveTurnstile_MissingURL(t *testing.T) {
	client := NewClient("test-key")
	_, err := client.SolveTurnstile(context.Background(), TurnstileParams{
		SiteKey: "0x4AAAA",
	})

	if err == nil {
		t.Fatal("expected error for missing URL")
	}
	if err.Error() != "nslsolver: URL is required" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestSolveChallenge_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body solveRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if body.Type != "challenge" {
			t.Errorf("expected type 'challenge', got '%s'", body.Type)
		}
		if body.Proxy != "http://user:pass@host:port" {
			t.Errorf("expected proxy, got '%s'", body.Proxy)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ChallengeResult{
			Success: true,
			Cookies: ChallengeCookies{
				CFClearance: "cf_clearance_value",
			},
			UserAgent: "Mozilla/5.0",
			Type:      "challenge",
		})
	}))
	defer server.Close()

	client := NewClient("test-key", WithBaseURL(server.URL), WithMaxRetries(0))
	result, err := client.SolveChallenge(context.Background(), ChallengeParams{
		URL:   "https://example.com/protected",
		Proxy: "http://user:pass@host:port",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Error("expected Success to be true")
	}
	if result.Cookies.CFClearance != "cf_clearance_value" {
		t.Errorf("expected cf_clearance 'cf_clearance_value', got '%s'", result.Cookies.CFClearance)
	}
	if result.UserAgent != "Mozilla/5.0" {
		t.Errorf("expected user_agent 'Mozilla/5.0', got '%s'", result.UserAgent)
	}
	if result.Type != "challenge" {
		t.Errorf("expected type 'challenge', got '%s'", result.Type)
	}
}

func TestSolveChallenge_MissingProxy(t *testing.T) {
	client := NewClient("test-key")
	_, err := client.SolveChallenge(context.Background(), ChallengeParams{
		URL: "https://example.com",
	})

	if err == nil {
		t.Fatal("expected error for missing Proxy")
	}
	if err.Error() != "nslsolver: Proxy is required for challenge type" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestSolveChallenge_MissingURL(t *testing.T) {
	client := NewClient("test-key")
	_, err := client.SolveChallenge(context.Background(), ChallengeParams{
		Proxy: "http://proxy:8080",
	})

	if err == nil {
		t.Fatal("expected error for missing URL")
	}
}

func TestGetBalance_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/balance" {
			t.Errorf("expected /balance, got %s", r.URL.Path)
		}
		if r.Header.Get("X-API-Key") != "test-key" {
			t.Errorf("expected X-API-Key 'test-key', got '%s'", r.Header.Get("X-API-Key"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(BalanceResult{
			Balance:      42.50,
			MaxThreads:   10,
			AllowedTypes: []string{"turnstile", "challenge"},
		})
	}))
	defer server.Close()

	client := NewClient("test-key", WithBaseURL(server.URL))
	result, err := client.GetBalance(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Balance != 42.50 {
		t.Errorf("expected balance 42.50, got %f", result.Balance)
	}
	if result.MaxThreads != 10 {
		t.Errorf("expected max_threads 10, got %d", result.MaxThreads)
	}
	if len(result.AllowedTypes) != 2 {
		t.Fatalf("expected 2 allowed types, got %d", len(result.AllowedTypes))
	}
}

func TestAPIError_401(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"message": "invalid API key"})
	}))
	defer server.Close()

	client := NewClient("bad-key", WithBaseURL(server.URL), WithMaxRetries(0))
	_, err := client.SolveTurnstile(context.Background(), TurnstileParams{
		SiteKey: "0x4AAAA",
		URL:     "https://example.com",
	})

	if err == nil {
		t.Fatal("expected error")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 401 {
		t.Errorf("expected status 401, got %d", apiErr.StatusCode)
	}
	if apiErr.Retryable {
		t.Error("expected non-retryable error")
	}
	if !IsAuthError(err) {
		t.Error("expected IsAuthError to return true")
	}
	if IsBalanceError(err) {
		t.Error("expected IsBalanceError to return false")
	}
}

func TestAPIError_402(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusPaymentRequired)
		json.NewEncoder(w).Encode(map[string]string{"message": "insufficient balance"})
	}))
	defer server.Close()

	client := NewClient("test-key", WithBaseURL(server.URL), WithMaxRetries(0))
	_, err := client.GetBalance(context.Background())

	if err == nil {
		t.Fatal("expected error")
	}
	if !IsBalanceError(err) {
		t.Error("expected IsBalanceError to return true")
	}

	apiErr := err.(*APIError)
	if apiErr.Retryable {
		t.Error("expected non-retryable error for 402")
	}
}

func TestAPIError_403(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"message": "type not allowed"})
	}))
	defer server.Close()

	client := NewClient("test-key", WithBaseURL(server.URL), WithMaxRetries(0))
	_, err := client.SolveTurnstile(context.Background(), TurnstileParams{
		SiteKey: "0x4AAAA",
		URL:     "https://example.com",
	})

	if err == nil {
		t.Fatal("expected error")
	}
	if !IsNotAllowedError(err) {
		t.Error("expected IsNotAllowedError to return true")
	}

	apiErr := err.(*APIError)
	if apiErr.Retryable {
		t.Error("expected non-retryable error for 403")
	}
}

func TestAPIError_400(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "missing site_key"})
	}))
	defer server.Close()

	client := NewClient("test-key", WithBaseURL(server.URL), WithMaxRetries(0))
	_, err := client.SolveTurnstile(context.Background(), TurnstileParams{
		SiteKey: "0x4AAAA",
		URL:     "https://example.com",
	})

	if err == nil {
		t.Fatal("expected error")
	}
	if !IsBadRequestError(err) {
		t.Error("expected IsBadRequestError to return true")
	}
}

func TestRetry_429(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&attempts, 1)
		if count <= 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{"message": "rate limited"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TurnstileResult{
			Success: true,
			Token:   "retry-success",
			Type:    "turnstile",
		})
	}))
	defer server.Close()

	client := NewClient("test-key", WithBaseURL(server.URL), WithMaxRetries(3))
	result, err := client.SolveTurnstile(context.Background(), TurnstileParams{
		SiteKey: "0x4AAAA",
		URL:     "https://example.com",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Token != "retry-success" {
		t.Errorf("expected token 'retry-success', got '%s'", result.Token)
	}

	finalAttempts := atomic.LoadInt32(&attempts)
	if finalAttempts != 3 {
		t.Errorf("expected 3 attempts, got %d", finalAttempts)
	}
}

func TestRetry_503(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&attempts, 1)
		if count <= 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{"message": "backend unavailable"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TurnstileResult{
			Success: true,
			Token:   "recovered",
			Type:    "turnstile",
		})
	}))
	defer server.Close()

	client := NewClient("test-key", WithBaseURL(server.URL), WithMaxRetries(2))
	result, err := client.SolveTurnstile(context.Background(), TurnstileParams{
		SiteKey: "0x4AAAA",
		URL:     "https://example.com",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Token != "recovered" {
		t.Errorf("expected token 'recovered', got '%s'", result.Token)
	}
}

func TestRetry_Exhausted(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(map[string]string{"message": "rate limited"})
	}))
	defer server.Close()

	client := NewClient("test-key", WithBaseURL(server.URL), WithMaxRetries(2))
	_, err := client.SolveTurnstile(context.Background(), TurnstileParams{
		SiteKey: "0x4AAAA",
		URL:     "https://example.com",
	})

	if err == nil {
		t.Fatal("expected error after retries exhausted")
	}
	if !IsRateLimitError(err) {
		t.Error("expected rate limit error")
	}

	finalAttempts := atomic.LoadInt32(&attempts)
	if finalAttempts != 3 {
		t.Errorf("expected 3 attempts, got %d", finalAttempts)
	}
}

func TestNoRetry_FatalError(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"message": "invalid key"})
	}))
	defer server.Close()

	client := NewClient("bad-key", WithBaseURL(server.URL), WithMaxRetries(3))
	_, err := client.SolveTurnstile(context.Background(), TurnstileParams{
		SiteKey: "0x4AAAA",
		URL:     "https://example.com",
	})

	if err == nil {
		t.Fatal("expected error")
	}

	finalAttempts := atomic.LoadInt32(&attempts)
	if finalAttempts != 1 {
		t.Errorf("expected 1 attempt (no retry for 401), got %d", finalAttempts)
	}
}

func TestContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TurnstileResult{Success: true, Token: "t", Type: "turnstile"})
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	client := NewClient("test-key", WithBaseURL(server.URL), WithMaxRetries(0))
	_, err := client.SolveTurnstile(ctx, TurnstileParams{
		SiteKey: "0x4AAAA",
		URL:     "https://example.com",
	})

	if err == nil {
		t.Fatal("expected context deadline error")
	}
}

func TestAPIError_ErrorInterface(t *testing.T) {
	err := newAPIError(429, "rate limited")
	expected := "nslsolver: API error 429: rate limited"
	if err.Error() != expected {
		t.Errorf("expected '%s', got '%s'", expected, err.Error())
	}
}

func TestErrorHelpers_NonAPIError(t *testing.T) {
	err := context.DeadlineExceeded

	if IsAuthError(err) {
		t.Error("IsAuthError should return false for non-APIError")
	}
	if IsBalanceError(err) {
		t.Error("IsBalanceError should return false for non-APIError")
	}
	if IsRateLimitError(err) {
		t.Error("IsRateLimitError should return false for non-APIError")
	}
	if IsNotAllowedError(err) {
		t.Error("IsNotAllowedError should return false for non-APIError")
	}
	if IsBadRequestError(err) {
		t.Error("IsBadRequestError should return false for non-APIError")
	}
	if IsBackendError(err) {
		t.Error("IsBackendError should return false for non-APIError")
	}
	if IsRetryableError(err) {
		t.Error("IsRetryableError should return false for non-APIError")
	}
}

func TestParseErrorResponse_FallbackMessage(t *testing.T) {
	err := parseErrorResponse(401, []byte("not json"))
	if err.Message != "unauthorized: invalid or missing API key" {
		t.Errorf("expected fallback message, got '%s'", err.Message)
	}

	err = parseErrorResponse(503, []byte("{}"))
	if err.Message != "backend service temporarily unavailable" {
		t.Errorf("expected fallback message, got '%s'", err.Message)
	}

	err = parseErrorResponse(400, []byte(`{"error": "custom error"}`))
	if err.Message != "custom error" {
		t.Errorf("expected 'custom error', got '%s'", err.Message)
	}
}

func TestSolveRequest_OmitEmpty(t *testing.T) {
	req := solveRequest{
		Type: "turnstile",
		URL:  "https://example.com",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	for _, field := range []string{"site_key", "action", "cdata", "proxy", "user_agent"} {
		if _, ok := parsed[field]; ok {
			t.Errorf("expected field '%s' to be omitted when empty", field)
		}
	}

	for _, field := range []string{"type", "url"} {
		if _, ok := parsed[field]; !ok {
			t.Errorf("expected field '%s' to be present", field)
		}
	}
}
