package client

import (
	"net/http"
	"testing"
	"time"
)

type customRoundTripper struct{}

func (c *customRoundTripper) RoundTrip(*http.Request) (*http.Response, error) { return nil, nil }

func TestDefaultConnectionPoolConfig(t *testing.T) {
	cfg := DefaultConnectionPoolConfig()
	if cfg.MaxIdleConns != 10 {
		t.Errorf("expected MaxIdleConns 10, got %d", cfg.MaxIdleConns)
	}
	if cfg.MaxIdleConnsPerHost != 5 {
		t.Errorf("expected MaxIdleConnsPerHost 5, got %d", cfg.MaxIdleConnsPerHost)
	}
	if cfg.MaxConnsPerHost != 10 {
		t.Errorf("expected MaxConnsPerHost 10, got %d", cfg.MaxConnsPerHost)
	}
	if cfg.IdleConnTimeout != 30*time.Second {
		t.Errorf("expected IdleConnTimeout 30s, got %v", cfg.IdleConnTimeout)
	}
}

func TestWithConnectionPool(t *testing.T) {
	tests := []struct {
		name   string
		config *ConnectionPoolConfig
		setup  func(c *Client)
		want   *ConnectionPoolConfig
	}{
		{
			name:   "nil config uses defaults",
			config: nil,
			setup:  nil,
			want:   DefaultConnectionPoolConfig(),
		},
		{
			name: "custom config",
			config: &ConnectionPoolConfig{
				MaxIdleConns:        20,
				MaxIdleConnsPerHost: 10,
				MaxConnsPerHost:     30,
				IdleConnTimeout:     60 * time.Second,
			},
			setup: nil,
			want: &ConnectionPoolConfig{
				MaxIdleConns:        20,
				MaxIdleConnsPerHost: 10,
				MaxConnsPerHost:     30,
				IdleConnTimeout:     60 * time.Second,
			},
		},
		{
			name: "non-Transport RoundTripper already set",
			config: &ConnectionPoolConfig{
				MaxIdleConns:        20,
				MaxIdleConnsPerHost: 10,
				MaxConnsPerHost:     30,
				IdleConnTimeout:     60 * time.Second,
			},
			setup: func(c *Client) {
				c.httpClient.Transport = &customRoundTripper{}
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClient("https://example.com", "test-key")
			if tt.setup != nil {
				tt.setup(c)
			}
			WithConnectionPool(tt.config)(c)
			got := c.GetConnectionPoolConfig()
			if tt.want == nil {
				if got != nil {
					t.Fatalf("expected nil, got %+v", got)
				}
				return
			}
			if got.MaxIdleConns != tt.want.MaxIdleConns {
				t.Errorf("MaxIdleConns: expected %d, got %d", tt.want.MaxIdleConns, got.MaxIdleConns)
			}
			if got.MaxIdleConnsPerHost != tt.want.MaxIdleConnsPerHost {
				t.Errorf("MaxIdleConnsPerHost: expected %d, got %d", tt.want.MaxIdleConnsPerHost, got.MaxIdleConnsPerHost)
			}
			if got.MaxConnsPerHost != tt.want.MaxConnsPerHost {
				t.Errorf("MaxConnsPerHost: expected %d, got %d", tt.want.MaxConnsPerHost, got.MaxConnsPerHost)
			}
			if got.IdleConnTimeout != tt.want.IdleConnTimeout {
				t.Errorf("IdleConnTimeout: expected %v, got %v", tt.want.IdleConnTimeout, got.IdleConnTimeout)
			}
		})
	}
}

func TestWithMaxIdleConns(t *testing.T) {
	tests := []struct {
		name         string
		maxIdleConns int
		setup        func(c *Client)
		expected     int
		transportOk  bool
	}{
		{
			name:         "creates transport if nil",
			maxIdleConns: 50,
			setup:        nil,
			expected:     50,
			transportOk:  true,
		},
		{
			name:         "non-Transport RoundTripper already set",
			maxIdleConns: 50,
			setup: func(c *Client) {
				c.httpClient.Transport = &customRoundTripper{}
			},
			expected:    0,
			transportOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClient("https://example.com", "test-key")
			if tt.setup != nil {
				tt.setup(c)
			}
			WithMaxIdleConns(tt.maxIdleConns)(c)
			transport, ok := c.httpClient.Transport.(*http.Transport)
			if !tt.transportOk {
				if ok {
					t.Fatal("expected transport to not be *http.Transport")
				}
				return
			}
			if !ok {
				t.Fatal("expected transport to be *http.Transport")
			}
			if transport.MaxIdleConns != tt.expected {
				t.Errorf("expected MaxIdleConns %d, got %d", tt.expected, transport.MaxIdleConns)
			}
		})
	}
}

func TestWithMaxIdleConnsPerHost(t *testing.T) {
	tests := []struct {
		name                string
		maxIdleConnsPerHost int
		setup               func(c *Client)
		expected            int
		transportOk         bool
	}{
		{
			name:                "creates transport if nil",
			maxIdleConnsPerHost: 25,
			setup:               nil,
			expected:            25,
			transportOk:         true,
		},
		{
			name:                "non-Transport RoundTripper already set",
			maxIdleConnsPerHost: 25,
			setup: func(c *Client) {
				c.httpClient.Transport = &customRoundTripper{}
			},
			expected:    0,
			transportOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClient("https://example.com", "test-key")
			if tt.setup != nil {
				tt.setup(c)
			}
			WithMaxIdleConnsPerHost(tt.maxIdleConnsPerHost)(c)
			transport, ok := c.httpClient.Transport.(*http.Transport)
			if !tt.transportOk {
				if ok {
					t.Fatal("expected transport to not be *http.Transport")
				}
				return
			}
			if !ok {
				t.Fatal("expected transport to be *http.Transport")
			}
			if transport.MaxIdleConnsPerHost != tt.expected {
				t.Errorf("expected MaxIdleConnsPerHost %d, got %d", tt.expected, transport.MaxIdleConnsPerHost)
			}
		})
	}
}

func TestWithIdleConnTimeout(t *testing.T) {
	tests := []struct {
		name        string
		timeout     time.Duration
		setup       func(c *Client)
		expected    time.Duration
		transportOk bool
	}{
		{
			name:        "creates transport if nil",
			timeout:     45 * time.Second,
			setup:       nil,
			expected:    45 * time.Second,
			transportOk: true,
		},
		{
			name:    "non-Transport RoundTripper already set",
			timeout: 45 * time.Second,
			setup: func(c *Client) {
				c.httpClient.Transport = &customRoundTripper{}
			},
			expected:    0,
			transportOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClient("https://example.com", "test-key")
			if tt.setup != nil {
				tt.setup(c)
			}
			WithIdleConnTimeout(tt.timeout)(c)
			transport, ok := c.httpClient.Transport.(*http.Transport)
			if !tt.transportOk {
				if ok {
					t.Fatal("expected transport to not be *http.Transport")
				}
				return
			}
			if !ok {
				t.Fatal("expected transport to be *http.Transport")
			}
			if transport.IdleConnTimeout != tt.expected {
				t.Errorf("expected IdleConnTimeout %v, got %v", tt.expected, transport.IdleConnTimeout)
			}
		})
	}
}

func TestWithMaxConnsPerHost(t *testing.T) {
	tests := []struct {
		name            string
		maxConnsPerHost int
		setup           func(c *Client)
		expected        int
		transportOk     bool
	}{
		{
			name:            "creates transport if nil",
			maxConnsPerHost: 100,
			setup:           nil,
			expected:        100,
			transportOk:     true,
		},
		{
			name:            "non-Transport RoundTripper already set",
			maxConnsPerHost: 100,
			setup: func(c *Client) {
				c.httpClient.Transport = &customRoundTripper{}
			},
			expected:    0,
			transportOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClient("https://example.com", "test-key")
			if tt.setup != nil {
				tt.setup(c)
			}
			WithMaxConnsPerHost(tt.maxConnsPerHost)(c)
			transport, ok := c.httpClient.Transport.(*http.Transport)
			if !tt.transportOk {
				if ok {
					t.Fatal("expected transport to not be *http.Transport")
				}
				return
			}
			if !ok {
				t.Fatal("expected transport to be *http.Transport")
			}
			if transport.MaxConnsPerHost != tt.expected {
				t.Errorf("expected MaxConnsPerHost %d, got %d", tt.expected, transport.MaxConnsPerHost)
			}
		})
	}
}

func TestGetConnectionPoolConfig(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(c *Client)
		expect *ConnectionPoolConfig
	}{
		{
			name: "returns config from configured transport",
			setup: func(c *Client) {
				WithConnectionPool(&ConnectionPoolConfig{
					MaxIdleConns:        15,
					MaxIdleConnsPerHost: 7,
					MaxConnsPerHost:     25,
					IdleConnTimeout:     45 * time.Second,
				})(c)
			},
			expect: &ConnectionPoolConfig{
				MaxIdleConns:        15,
				MaxIdleConnsPerHost: 7,
				MaxConnsPerHost:     25,
				IdleConnTimeout:     45 * time.Second,
			},
		},
		{
			name: "nil Transport returns nil",
			setup: func(c *Client) {
				c.httpClient.Transport = nil
			},
			expect: nil,
		},
		{
			name: "nil httpClient returns nil",
			setup: func(c *Client) {
				c.httpClient = nil
			},
			expect: nil,
		},
		{
			name: "non-http.Transport returns nil",
			setup: func(c *Client) {
				c.httpClient.Transport = &customRoundTripper{}
			},
			expect: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClient("https://example.com", "test-key")
			if tt.setup != nil {
				tt.setup(c)
			}
			got := c.GetConnectionPoolConfig()
			if tt.expect == nil {
				if got != nil {
					t.Fatalf("expected nil, got %+v", got)
				}
				return
			}
			if got == nil {
				t.Fatal("expected non-nil config, got nil")
			}
			if got.MaxIdleConns != tt.expect.MaxIdleConns {
				t.Errorf("MaxIdleConns: expected %d, got %d", tt.expect.MaxIdleConns, got.MaxIdleConns)
			}
			if got.MaxIdleConnsPerHost != tt.expect.MaxIdleConnsPerHost {
				t.Errorf("MaxIdleConnsPerHost: expected %d, got %d", tt.expect.MaxIdleConnsPerHost, got.MaxIdleConnsPerHost)
			}
			if got.MaxConnsPerHost != tt.expect.MaxConnsPerHost {
				t.Errorf("MaxConnsPerHost: expected %d, got %d", tt.expect.MaxConnsPerHost, got.MaxConnsPerHost)
			}
			if got.IdleConnTimeout != tt.expect.IdleConnTimeout {
				t.Errorf("IdleConnTimeout: expected %v, got %v", tt.expect.IdleConnTimeout, got.IdleConnTimeout)
			}
		})
	}
}
