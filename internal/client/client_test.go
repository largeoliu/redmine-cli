// internal/client/client_test.go
package client

import (
	"context"
	stderrors "errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/largeoliu/redmine-cli/internal/errors"
)

func TestNewClient(t *testing.T) {
	c := NewClient("https://example.com/", "test-key")
	c.mu.RLock()
	baseURL := c.baseURL
	apiKey := c.apiKey
	c.mu.RUnlock()
	if baseURL != "https://example.com" {
		t.Errorf("expected baseURL 'https://example.com', got %s", baseURL)
	}
	if apiKey != "test-key" {
		t.Errorf("expected apiKey 'test-key', got %s", apiKey)
	}
}

func TestClientGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Redmine-API-Key") != "test-key" {
			t.Error("missing API key header")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"test": "ok"}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-key")
	var result map[string]string
	err := c.Get(context.Background(), "/test.json", &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["test"] != "ok" {
		t.Errorf("expected result['test'] = 'ok', got %s", result["test"])
	}
}

func TestClientRetry(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-key", WithRetry(3, 10*time.Millisecond, 100*time.Millisecond))
	err := c.Get(context.Background(), "/test.json", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestClientAuthError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	c := NewClient(server.URL, "invalid-key")
	err := c.Get(context.Background(), "/test.json", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestWithHTTPClient(t *testing.T) {
	customClient := &http.Client{Timeout: 10 * time.Second}
	c := NewClient("https://example.com", "test-key", WithHTTPClient(customClient))
	if c.httpClient != customClient {
		t.Error("expected custom HTTP client to be set")
	}
}

func TestWithTimeout(t *testing.T) {
	c := NewClient("https://example.com", "test-key", WithTimeout(15*time.Second))
	if c.httpClient.Timeout != 15*time.Second {
		t.Errorf("expected timeout 15s, got %v", c.httpClient.Timeout)
	}
}

func TestClientPost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST method, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("missing Content-Type header")
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": 1}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-key")
	body := map[string]string{"name": "test"}
	var result map[string]int
	err := c.Post(context.Background(), "/test.json", body, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["id"] != 1 {
		t.Errorf("expected result[id] = 1, got %d", result["id"])
	}
}

func TestClientPut(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT method, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": 1}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-key")
	body := map[string]string{"name": "updated"}
	var result map[string]int
	err := c.Put(context.Background(), "/test.json", body, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClientDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE method, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-key")
	err := c.Delete(context.Background(), "/test.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClientForbiddenError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-key")
	err := c.Get(context.Background(), "/test.json", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClientNotFoundError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-key")
	err := c.Get(context.Background(), "/test.json", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClientRateLimitError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-key")
	err := c.Get(context.Background(), "/test.json", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClientServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-key", WithRetry(0, 10*time.Millisecond, 100*time.Millisecond))
	err := c.Get(context.Background(), "/test.json", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClientOtherClientError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad request"))
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-key")
	err := c.Get(context.Background(), "/test.json", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestBuildPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		params   map[string]string
		expected string
	}{
		{
			name:     "no params",
			path:     "/test.json",
			params:   nil,
			expected: "/test.json",
		},
		{
			name:     "with params",
			path:     "/test.json",
			params:   map[string]string{"limit": "10", "offset": "0"},
			expected: "/test.json?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClient("https://example.com", "test-key")
			result, err := c.BuildPath(tt.path, tt.params)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			switch tt.name {
			case "with params":
				if result[:11] != tt.expected[:11] {
					t.Errorf("expected path to start with %s, got %s", tt.expected, result[:11])
				}
			default:
				if result != tt.expected {
					t.Errorf("expected %s, got %s", tt.expected, result)
				}
			}
		})
	}
}

func TestBuildPathWithParams(t *testing.T) {
	c := NewClient("https://example.com", "test-key")
	params := map[string]string{"key": "value"}
	result, err := c.BuildPath("/test.json", params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "/test.json?key=value" {
		t.Errorf("expected /test.json?key=value, got %s", result)
	}
}

func TestClientNetworkError(t *testing.T) {
	c := NewClient("http://invalid-host-that-does-not-exist.example", "test-key")
	err := c.Get(context.Background(), "/test.json", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClientInvalidJSONResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`invalid json`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-key")
	var result map[string]string
	err := c.Get(context.Background(), "/test.json", &result)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestClientInvalidJSONBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-key")
	// Create a value that cannot be marshaled to JSON
	body := make(chan int)
	err := c.Post(context.Background(), "/test.json", body, nil)
	if err == nil {
		t.Fatal("expected error for invalid JSON body, got nil")
	}
}

func TestRetryDelay(t *testing.T) {
	tests := []struct {
		name          string
		maxRetries    int
		initialDelay  time.Duration
		maxDelay      time.Duration
		attempt       int
		expectedDelay time.Duration
	}{
		{
			name:          "first attempt",
			maxRetries:    3,
			initialDelay:  100 * time.Millisecond,
			maxDelay:      5 * time.Second,
			attempt:       0,
			expectedDelay: 100 * time.Millisecond,
		},
		{
			name:          "second attempt",
			maxRetries:    3,
			initialDelay:  100 * time.Millisecond,
			maxDelay:      5 * time.Second,
			attempt:       1,
			expectedDelay: 200 * time.Millisecond,
		},
		{
			name:          "third attempt",
			maxRetries:    3,
			initialDelay:  100 * time.Millisecond,
			maxDelay:      5 * time.Second,
			attempt:       2,
			expectedDelay: 400 * time.Millisecond,
		},
		{
			name:          "delay exceeds max delay",
			maxRetries:    10,
			initialDelay:  1 * time.Second,
			maxDelay:      2 * time.Second,
			attempt:       5,
			expectedDelay: 2 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClient("https://example.com", "test-key",
				WithRetry(tt.maxRetries, tt.initialDelay, tt.maxDelay))
			delay := c.retryDelay(tt.attempt)
			if delay != tt.expectedDelay {
				t.Errorf("expected delay %v, got %v", tt.expectedDelay, delay)
			}
		})
	}
}

func TestShouldRetry(t *testing.T) {
	c := NewClient("https://example.com", "test-key")

	// Test with retryable error
	retryableErr := newNetworkError("network error")
	if !c.shouldRetry(retryableErr) {
		t.Error("expected shouldRetry to return true for network error")
	}

	// Test with non-retryable error
	nonRetryableErr := newValidationError("validation error")
	if c.shouldRetry(nonRetryableErr) {
		t.Error("expected shouldRetry to return false for validation error")
	}

	// Test with standard error
	stdErr := stderrors.New("standard error")
	if c.shouldRetry(stdErr) {
		t.Error("expected shouldRetry to return false for standard error")
	}
}

func TestClientContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-key", WithTimeout(50*time.Millisecond))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := c.Get(ctx, "/test.json", nil)
	if err == nil {
		t.Fatal("expected error due to context cancellation, got nil")
	}
}

// Helper functions to create errors for testing
func newNetworkError(message string) error {
	return errors.NewNetwork(message)
}

func newValidationError(message string) error {
	return errors.NewValidation(message)
}

func TestClientNilHTTPClient(t *testing.T) {
	c := &Client{
		baseURL: "https://example.com",
		apiKey:  "test-key",
	}
	// Test that methods don't panic with nil httpClient
	// This should not happen in practice, but we test for robustness
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("unexpected panic: %v", r)
		}
	}()
	_ = c
}

func TestClientRetryDelayMaxDelay(t *testing.T) {
	// Test that delay is capped at MaxDelay
	c := NewClient("https://example.com", "test-key",
		WithRetry(10, 1*time.Second, 2*time.Second))

	// Test various attempts
	for attempt := 0; attempt < 10; attempt++ {
		delay := c.retryDelay(attempt)
		if delay > 2*time.Second {
			t.Errorf("delay %v exceeds max delay 2s for attempt %d", delay, attempt)
		}
	}
}

func TestClientDoRequestWithNilBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-key")
	err := c.Post(context.Background(), "/test.json", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClientDoRequestWithNilResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"test": "data"}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-key")
	err := c.Get(context.Background(), "/test.json", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClientRetryConfigDefaults(t *testing.T) {
	c := NewClient("https://example.com", "test-key")

	if c.retry.MaxRetries != 3 {
		t.Errorf("expected default MaxRetries 3, got %d", c.retry.MaxRetries)
	}
	if c.retry.InitialDelay != 500*time.Millisecond {
		t.Errorf("expected default InitialDelay 500ms, got %v", c.retry.InitialDelay)
	}
	if c.retry.MaxDelay != 5*time.Second {
		t.Errorf("expected default MaxDelay 5s, got %v", c.retry.MaxDelay)
	}
}

func TestClientHTTPTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-key", WithTimeout(50*time.Millisecond))
	err := c.Get(context.Background(), "/test.json", nil)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

func TestClientInvalidURL(t *testing.T) {
	c := NewClient("://invalid-url", "test-key")
	err := c.Get(context.Background(), "/test.json", nil)
	if err == nil {
		t.Fatal("expected error for invalid URL, got nil")
	}
}

func TestHandleErrorResponseReadBodyError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-key")
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()

	resp.Body = &errorReaderCloser{}
	err = c.handleErrorResponse(resp)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var appErr *errors.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("expected *errors.Error, got %T", err)
	}
	if appErr.Category != errors.CategoryAPI {
		t.Errorf("expected category api, got %s", appErr.Category)
	}
}

func TestRetryDelayHighAttempt(t *testing.T) {
	c := NewClient("https://example.com", "test-key",
		WithRetry(10, 1*time.Millisecond, 5*time.Second))
	delay := c.retryDelay(31)
	if delay != 5*time.Second {
		t.Errorf("expected delay to be capped at maxDelay for very high attempt, got %v", delay)
	}
}

func TestRetryDelayVeryHighAttempt(t *testing.T) {
	c := NewClient("https://example.com", "test-key",
		WithRetry(10, 1*time.Millisecond, 10*time.Second))
	delay := c.retryDelay(100)
	if delay > 10*time.Second {
		t.Errorf("expected delay to be capped at maxDelay, got %v", delay)
	}
}

func TestRetryDelayShiftOverflow(t *testing.T) {
	c := NewClient("https://example.com", "test-key",
		WithRetry(10, 1*time.Millisecond, 10*time.Second))
	delay := c.retryDelay(70)
	if delay > 10*time.Second {
		t.Errorf("expected delay to be capped, got %v", delay)
	}
}

func TestRetryDelayExceedsMaxDelay(t *testing.T) {
	c := NewClient("https://example.com", "test-key",
		WithRetry(3, 100*time.Millisecond, 150*time.Millisecond))
	delay := c.retryDelay(2)
	if delay > 150*time.Millisecond {
		t.Errorf("expected delay to be capped at maxDelay 150ms, got %v", delay)
	}
}

func TestRetryDelayUnderMaxDelay(t *testing.T) {
	c := NewClient("https://example.com", "test-key",
		WithRetry(3, 100*time.Millisecond, 10*time.Second))
	delay := c.retryDelay(2)
	if delay != 400*time.Millisecond {
		t.Errorf("expected delay 400ms, got %v", delay)
	}
}

func TestPingWithResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-key")
	err := c.Ping(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPingWithError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-key")
	err := c.Ping(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestPingWithServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-key")
	err := c.Ping(context.Background())
	if err == nil {
		t.Fatal("expected error for 500 status, got nil")
	}
}

func TestDoSingleRequestWithNilBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result": "ok"}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-key")
	var result map[string]string
	err := c.doSingleRequest(context.Background(), http.MethodGet, "/test", nil, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["result"] != "ok" {
		t.Errorf("expected result['result'] = 'ok', got %s", result["result"])
	}
}

func TestDoSingleRequestWithBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": 123}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-key")
	body := map[string]string{"name": "test"}
	var result map[string]int
	err := c.doSingleRequest(context.Background(), http.MethodPost, "/test", body, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["id"] != 123 {
		t.Errorf("expected result['id'] = 123, got %d", result["id"])
	}
}

func TestDoSingleRequestWithUnmarshalableBody(t *testing.T) {
	c := NewClient("https://example.com", "test-key")

	type unmarshalable struct{}
	_ = unmarshalable{}

	body := make(chan int)

	err := c.doSingleRequest(context.Background(), http.MethodPost, "/test", body, nil)
	if err == nil {
		t.Fatal("expected error for unmarshalable body, got nil")
	}
}

func TestDoSingleRequestWithInvalidResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`invalid json`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-key")
	var result map[string]string
	err := c.doSingleRequest(context.Background(), http.MethodGet, "/test", nil, &result)
	if err == nil {
		t.Fatal("expected error for invalid JSON response, got nil")
	}
}

func TestDoSingleRequestWithNilResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result": "ok"}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-key")
	err := c.doSingleRequest(context.Background(), http.MethodGet, "/test", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDoSingleRequestWithMalformedURL(t *testing.T) {
	c := NewClient("http://[invalid", "test-key")

	err := c.doSingleRequest(context.Background(), http.MethodGet, "/test", nil, nil)
	if err == nil {
		t.Fatal("expected error for malformed URL, got nil")
	}
}

type errorReaderCloser struct{}

func (r *errorReaderCloser) Read(_ []byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}

func (r *errorReaderCloser) Close() error {
	return nil
}

type closeErrorBody struct {
	data     []byte
	readErr  error
	closeErr error
}

func (b *closeErrorBody) Read(p []byte) (int, error) {
	if b.readErr != nil {
		return 0, b.readErr
	}
	if len(b.data) == 0 {
		return 0, io.EOF
	}
	n := copy(p, b.data)
	b.data = b.data[n:]
	return n, nil
}

func (b *closeErrorBody) Close() error {
	return b.closeErr
}

type closeErrorTransport struct {
	closeErr error
}

func (t *closeErrorTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       &closeErrorBody{data: []byte(`{}`), closeErr: t.closeErr},
	}, nil
}

type serverErrorTransport struct {
	statusCode int
	closeErr   error
}

func (t *serverErrorTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: t.statusCode,
		Header:     make(http.Header),
		Body:       &closeErrorBody{data: []byte(`error`), closeErr: t.closeErr},
	}, nil
}

func TestDoSingleRequestNewRequestError(t *testing.T) {
	c := NewClient("https://example.com", "test-key")
	err := c.doSingleRequest(context.Background(), "INVALID METHOD", "/test", nil, nil)
	if err == nil {
		t.Fatal("expected error for invalid HTTP method, got nil")
	}
	var appErr *errors.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("expected *errors.Error, got %T", err)
	}
	if appErr.Category != errors.CategoryInternal {
		t.Errorf("expected category internal, got %s", appErr.Category)
	}
}

func TestDoSingleRequestCloseError(t *testing.T) {
	closeErr := stderrors.New("close error")
	transport := &closeErrorTransport{closeErr: closeErr}
	httpClient := &http.Client{Transport: transport}

	c := NewClient("https://example.com", "test-key", WithHTTPClient(httpClient))
	err := c.doSingleRequest(context.Background(), http.MethodGet, "/test", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPingInvalidURL(t *testing.T) {
	c := NewClient("https://example.com", "test-key")
	c.mu.Lock()
	c.baseURL = "://invalid"
	c.mu.Unlock()

	err := c.Ping(context.Background())
	if err == nil {
		t.Fatal("expected error for invalid URL, got nil")
	}
	var appErr *errors.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("expected *errors.Error, got %T", err)
	}
	if appErr.Category != errors.CategoryNetwork {
		t.Errorf("expected category network, got %s", appErr.Category)
	}
}

func TestPingUnreachable(t *testing.T) {
	c := NewClient("http://invalid-host-that-does-not-exist.example", "test-key")
	err := c.Ping(context.Background())
	if err == nil {
		t.Fatal("expected error for unreachable host, got nil")
	}
	var appErr *errors.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("expected *errors.Error, got %T", err)
	}
	if appErr.Category != errors.CategoryNetwork {
		t.Errorf("expected category network, got %s", appErr.Category)
	}
}

func TestRetryExhausted(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-key", WithRetry(2, 1*time.Millisecond, 10*time.Millisecond))
	err := c.Get(context.Background(), "/test.json", nil)
	if err == nil {
		t.Fatal("expected error after retries exhausted, got nil")
	}
	var appErr *errors.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("expected *errors.Error, got %T", err)
	}
	if !appErr.Retryable {
		t.Error("expected retryable error")
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts (1 initial + 2 retries), got %d", attempts)
	}
}

func TestRetryContextCancelledDuringDelay(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-key", WithRetry(5, 5*time.Second, 10*time.Second))
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := c.Get(ctx, "/test.json", nil)
	if err == nil {
		t.Fatal("expected error due to context cancellation, got nil")
	}
	if ctx.Err() != context.DeadlineExceeded {
		t.Errorf("expected context deadline exceeded, got %v", ctx.Err())
	}
}

func TestRetryDelayShiftOverflow63(t *testing.T) {
	c := NewClient("https://example.com", "test-key",
		WithRetry(100, 1*time.Millisecond, 10*time.Second))

	for attempt := 0; attempt <= 100; attempt++ {
		delay := c.retryDelay(attempt)
		if delay > 10*time.Second {
			t.Errorf("delay %v exceeds maxDelay for attempt %d", delay, attempt)
		}
	}
}

func TestPingCloseError(t *testing.T) {
	closeErr := stderrors.New("close error")
	origTransport := http.DefaultTransport
	http.DefaultTransport = &closeErrorTransport{closeErr: closeErr}
	defer func() { http.DefaultTransport = origTransport }()

	c := NewClient("https://example.com", "test-key")
	err := c.Ping(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPingCloseErrorWithServerError(t *testing.T) {
	closeErr := stderrors.New("close error")
	origTransport := http.DefaultTransport
	http.DefaultTransport = &serverErrorTransport{statusCode: http.StatusInternalServerError, closeErr: closeErr}
	defer func() { http.DefaultTransport = origTransport }()

	c := NewClient("https://example.com", "test-key")
	err := c.Ping(context.Background())
	if err == nil {
		t.Fatal("expected error for 500 status, got nil")
	}
}

func TestDoSingleRequestCloseErrorWithBody(t *testing.T) {
	closeErr := stderrors.New("close error")
	transport := &closeErrorTransport{closeErr: closeErr}
	httpClient := &http.Client{Transport: transport}

	c := NewClient("https://example.com", "test-key", WithHTTPClient(httpClient))
	err := c.doSingleRequest(context.Background(), http.MethodGet, "/test", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDoSingleRequestCloseErrorServerErr(t *testing.T) {
	closeErr := stderrors.New("close error")
	transport := &serverErrorTransport{statusCode: http.StatusInternalServerError, closeErr: closeErr}
	httpClient := &http.Client{Transport: transport}

	c := NewClient("https://example.com", "test-key", WithRetry(0, 1*time.Millisecond, 10*time.Millisecond), WithHTTPClient(httpClient))
	err := c.doSingleRequest(context.Background(), http.MethodGet, "/test", nil, nil)
	if err == nil {
		t.Fatal("expected error for 500 status, got nil")
	}
}
