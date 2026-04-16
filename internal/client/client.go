// Package client provides the HTTP client for interacting with Redmine API.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/largeoliu/redmine-cli/internal/errors"
)

// Client is the HTTP client for Redmine API.
type Client struct {
	mu         sync.RWMutex
	baseURL    string
	apiKey     string
	httpClient *http.Client
	retry      *RetryConfig
}

// RetryConfig configures retry behavior.
type RetryConfig struct {
	MaxRetries   int
	InitialDelay time.Duration
	MaxDelay     time.Duration
}

// Pagination represents pagination parameters.
type Pagination struct {
	Limit  int `url:"limit,omitempty"`
	Offset int `url:"offset,omitempty"`
}

// PaginatedResponse is the base response for paginated results.
type PaginatedResponse struct {
	TotalCount int `json:"total_count"`
	Limit      int `json:"limit"`
	Offset     int `json:"offset"`
}

// NewClient creates a new Redmine API client.
func NewClient(baseURL, apiKey string, opts ...Option) *Client {
	baseURL = strings.TrimRight(baseURL, "/")
	c := &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		retry: &RetryConfig{
			MaxRetries:   3,
			InitialDelay: 500 * time.Millisecond,
			MaxDelay:     5 * time.Second,
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Option is a functional option for configuring the Client.
type Option func(*Client)

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) { c.httpClient = httpClient }
}

// WithTimeout sets the request timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) { c.httpClient.Timeout = timeout }
}

// WithRetry configures retry behavior.
func WithRetry(maxRetries int, initialDelay, maxDelay time.Duration) Option {
	return func(c *Client) {
		c.retry = &RetryConfig{
			MaxRetries:   maxRetries,
			InitialDelay: initialDelay,
			MaxDelay:     maxDelay,
		}
	}
}

// Get makes a GET request to the specified path.
func (c *Client) Get(ctx context.Context, path string, result any) error {
	return c.doRequest(ctx, http.MethodGet, path, nil, result)
}

// Post makes a POST request to the specified path.
func (c *Client) Post(ctx context.Context, path string, body, result any) error {
	return c.doRequest(ctx, http.MethodPost, path, body, result)
}

// Put makes a PUT request to the specified path.
func (c *Client) Put(ctx context.Context, path string, body, result any) error {
	return c.doRequest(ctx, http.MethodPut, path, body, result)
}

// Delete makes a DELETE request to the specified path.
func (c *Client) Delete(ctx context.Context, path string) error {
	return c.doRequest(ctx, http.MethodDelete, path, nil, nil)
}

func (c *Client) doRequest(ctx context.Context, method, path string, body, result any) error {
	return c.WithRetry(ctx, func() error {
		return c.doSingleRequest(ctx, method, path, body, result)
	})
}

func (c *Client) doSingleRequest(ctx context.Context, method, path string, body, result any) error {
	c.mu.RLock()
	baseURL := c.baseURL
	apiKey := c.apiKey
	c.mu.RUnlock()

	reqURL, err := url.Parse(baseURL + path)
	if err != nil {
		return errors.NewInternal("invalid URL", errors.WithCause(err))
	}

	var reqBody io.Reader
	if body != nil {
		data, marshalErr := json.Marshal(body)
		if marshalErr != nil {
			return errors.NewInternal("failed to marshal request body", errors.WithCause(marshalErr))
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL.String(), reqBody)
	if err != nil {
		return errors.NewInternal("failed to create request", errors.WithCause(err))
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Redmine-API-Key", apiKey)

	resp, doErr := c.httpClient.Do(req)
	if doErr != nil {
		return errors.NewNetwork(fmt.Sprintf("request failed: %v", doErr), errors.WithCause(doErr))
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			_ = closeErr
		}
	}()

	if resp.StatusCode >= 400 {
		return c.handleErrorResponse(resp)
	}

	if result != nil {
		if decodeErr := json.NewDecoder(resp.Body).Decode(result); decodeErr != nil {
			return errors.NewInternal("failed to decode response", errors.WithCause(decodeErr))
		}
	}

	return nil
}

func (c *Client) handleErrorResponse(resp *http.Response) error {
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024))
	if err != nil {
		body = []byte("unable to read response body")
	}

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return errors.NewAuth("authentication failed",
			errors.WithHint("check your API key"),
			errors.WithActions("redmine login"))
	case http.StatusForbidden:
		return errors.NewAuth("permission denied",
			errors.WithHint("you don't have permission to access this resource"))
	case http.StatusNotFound:
		return errors.NewAPI("resource not found")
	case http.StatusTooManyRequests:
		return errors.NewRateLimit("rate limit exceeded")
	default:
		if resp.StatusCode >= 500 {
			return errors.NewAPI(fmt.Sprintf("server error: HTTP %d", resp.StatusCode),
				errors.WithRetryable(true))
		}
		return errors.NewAPI(fmt.Sprintf("request failed: HTTP %d - %s", resp.StatusCode, string(body)))
	}
}

func (c *Client) shouldRetry(err error) bool {
	var appErr *errors.Error
	if errors.As(err, &appErr) {
		return appErr.Retryable
	}
	return false
}

func (c *Client) retryDelay(attempt int) time.Duration {
	const maxShift = 30
	if attempt > maxShift {
		attempt = maxShift
	}
	//nolint:gosec // G115: attempt is bounded to [0, 30], safe to convert #nosec G115
	shift := uint(attempt)
	if shift > 63 {
		shift = 63
	}
	delay := c.retry.InitialDelay * time.Duration(1<<shift)
	if delay > c.retry.MaxDelay {
		delay = c.retry.MaxDelay
	}
	return delay
}

// BuildPath builds a URL path with query parameters.
func (c *Client) BuildPath(path string, params map[string]string) string {
	if len(params) == 0 {
		return path
	}
	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}
	return path + "?" + values.Encode()
}

// Ping tests the connectivity to the given URL using HTTP HEAD request.
func (c *Client) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, c.baseURL, nil)
	if err != nil {
		return errors.NewNetwork("URL 格式无效", errors.WithCause(err))
	}

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return errors.NewNetwork("URL 无法访问", errors.WithActions(
			"1) URL 正确",
			"2) 网络畅通",
			"3) 服务正常运行",
		), errors.WithCause(err))
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			_ = closeErr
		}
	}()

	if resp.StatusCode >= 400 {
		return errors.NewAPI("服务器返回错误状态码", errors.WithCause(fmt.Errorf("status: %d", resp.StatusCode)))
	}

	return nil
}
