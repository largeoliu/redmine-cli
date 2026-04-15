package client

import (
	"net/http"
	"testing"
	"time"
)

func TestDefaultConnectionPoolConfig(t *testing.T) {
	config := DefaultConnectionPoolConfig()

	if config == nil {
		t.Fatal("DefaultConnectionPoolConfig() returned nil")
	}

	// Verify default values
	if config.MaxIdleConns != 10 {
		t.Errorf("DefaultConnectionPoolConfig().MaxIdleConns = %v, want 10", config.MaxIdleConns)
	}

	if config.MaxIdleConnsPerHost != 5 {
		t.Errorf("DefaultConnectionPoolConfig().MaxIdleConnsPerHost = %v, want 5", config.MaxIdleConnsPerHost)
	}

	if config.MaxConnsPerHost != 10 {
		t.Errorf("DefaultConnectionPoolConfig().MaxConnsPerHost = %v, want 10", config.MaxConnsPerHost)
	}

	expectedTimeout := 30 * time.Second
	if config.IdleConnTimeout != expectedTimeout {
		t.Errorf("DefaultConnectionPoolConfig().IdleConnTimeout = %v, want %v", config.IdleConnTimeout, expectedTimeout)
	}
}

func TestWithConnectionPool(t *testing.T) {
	tests := []struct {
		name   string
		config *ConnectionPoolConfig
		want   *ConnectionPoolConfig
	}{
		{
			name:   "with nil config uses defaults",
			config: nil,
			want:   DefaultConnectionPoolConfig(),
		},
		{
			name: "with custom config",
			config: &ConnectionPoolConfig{
				MaxIdleConns:        20,
				MaxIdleConnsPerHost: 10,
				MaxConnsPerHost:     20,
				IdleConnTimeout:     60 * time.Second,
			},
			want: &ConnectionPoolConfig{
				MaxIdleConns:        20,
				MaxIdleConnsPerHost: 10,
				MaxConnsPerHost:     20,
				IdleConnTimeout:     60 * time.Second,
			},
		},
		{
			name: "with partial config",
			config: &ConnectionPoolConfig{
				MaxIdleConns: 15,
			},
			want: &ConnectionPoolConfig{
				MaxIdleConns:        15,
				MaxIdleConnsPerHost: 0,
				MaxConnsPerHost:     0,
				IdleConnTimeout:     0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				httpClient: &http.Client{},
			}

			opt := WithConnectionPool(tt.config)
			opt(client)

			transport, ok := client.httpClient.Transport.(*http.Transport)
			if !ok {
				t.Fatal("Expected Transport to be *http.Transport")
			}

			if transport.MaxIdleConns != tt.want.MaxIdleConns {
				t.Errorf("MaxIdleConns = %v, want %v", transport.MaxIdleConns, tt.want.MaxIdleConns)
			}

			if transport.MaxIdleConnsPerHost != tt.want.MaxIdleConnsPerHost {
				t.Errorf("MaxIdleConnsPerHost = %v, want %v", transport.MaxIdleConnsPerHost, tt.want.MaxIdleConnsPerHost)
			}

			if transport.MaxConnsPerHost != tt.want.MaxConnsPerHost {
				t.Errorf("MaxConnsPerHost = %v, want %v", transport.MaxConnsPerHost, tt.want.MaxConnsPerHost)
			}

			if transport.IdleConnTimeout != tt.want.IdleConnTimeout {
				t.Errorf("IdleConnTimeout = %v, want %v", transport.IdleConnTimeout, tt.want.IdleConnTimeout)
			}
		})
	}
}

func TestWithConnectionPool_WithExistingTransport(t *testing.T) {
	existingTransport := &http.Transport{
		MaxIdleConns:        5,
		MaxIdleConnsPerHost: 2,
	}

	client := &Client{
		httpClient: &http.Client{
			Transport: existingTransport,
		},
	}

	config := &ConnectionPoolConfig{
		MaxIdleConns:        25,
		MaxIdleConnsPerHost: 12,
		MaxConnsPerHost:     25,
		IdleConnTimeout:     45 * time.Second,
	}

	opt := WithConnectionPool(config)
	opt(client)

	transport, ok := client.httpClient.Transport.(*http.Transport)
	if !ok {
		t.Fatal("Expected Transport to be *http.Transport")
	}

	if transport.MaxIdleConns != 25 {
		t.Errorf("MaxIdleConns = %v, want 25", transport.MaxIdleConns)
	}

	if transport.MaxIdleConnsPerHost != 12 {
		t.Errorf("MaxIdleConnsPerHost = %v, want 12", transport.MaxIdleConnsPerHost)
	}
}

func TestWithMaxIdleConns(t *testing.T) {
	client := &Client{
		httpClient: &http.Client{},
	}

	opt := WithMaxIdleConns(50)
	opt(client)

	transport, ok := client.httpClient.Transport.(*http.Transport)
	if !ok {
		t.Fatal("Expected Transport to be *http.Transport")
	}

	if transport.MaxIdleConns != 50 {
		t.Errorf("MaxIdleConns = %v, want 50", transport.MaxIdleConns)
	}
}

func TestWithMaxIdleConns_WithExistingTransport(t *testing.T) {
	existingTransport := &http.Transport{
		MaxIdleConns: 10,
	}

	client := &Client{
		httpClient: &http.Client{
			Transport: existingTransport,
		},
	}

	opt := WithMaxIdleConns(30)
	opt(client)

	transport, ok := client.httpClient.Transport.(*http.Transport)
	if !ok {
		t.Fatal("Expected Transport to be *http.Transport")
	}

	if transport.MaxIdleConns != 30 {
		t.Errorf("MaxIdleConns = %v, want 30", transport.MaxIdleConns)
	}
}

func TestWithMaxIdleConnsPerHost(t *testing.T) {
	client := &Client{
		httpClient: &http.Client{},
	}

	opt := WithMaxIdleConnsPerHost(20)
	opt(client)

	transport, ok := client.httpClient.Transport.(*http.Transport)
	if !ok {
		t.Fatal("Expected Transport to be *http.Transport")
	}

	if transport.MaxIdleConnsPerHost != 20 {
		t.Errorf("MaxIdleConnsPerHost = %v, want 20", transport.MaxIdleConnsPerHost)
	}
}

func TestWithIdleConnTimeout(t *testing.T) {
	client := &Client{
		httpClient: &http.Client{},
	}

	timeout := 120 * time.Second
	opt := WithIdleConnTimeout(timeout)
	opt(client)

	transport, ok := client.httpClient.Transport.(*http.Transport)
	if !ok {
		t.Fatal("Expected Transport to be *http.Transport")
	}

	if transport.IdleConnTimeout != timeout {
		t.Errorf("IdleConnTimeout = %v, want %v", transport.IdleConnTimeout, timeout)
	}
}

func TestWithMaxConnsPerHost(t *testing.T) {
	client := &Client{
		httpClient: &http.Client{},
	}

	opt := WithMaxConnsPerHost(100)
	opt(client)

	transport, ok := client.httpClient.Transport.(*http.Transport)
	if !ok {
		t.Fatal("Expected Transport to be *http.Transport")
	}

	if transport.MaxConnsPerHost != 100 {
		t.Errorf("MaxConnsPerHost = %v, want 100", transport.MaxConnsPerHost)
	}
}

func TestGetConnectionPoolConfig(t *testing.T) {
	tests := []struct {
		name       string
		client     *Client
		wantConfig *ConnectionPoolConfig
		wantNil    bool
	}{
		{
			name:       "nil httpClient",
			client:     &Client{httpClient: nil},
			wantConfig: nil,
			wantNil:    true,
		},
		{
			name: "nil transport",
			client: &Client{
				httpClient: &http.Client{},
			},
			wantConfig: nil,
			wantNil:    true,
		},
		{
			name: "non-http.Transport",
			client: &Client{
				httpClient: &http.Client{
					Transport: &customTransport{},
				},
			},
			wantConfig: nil,
			wantNil:    true,
		},
		{
			name: "valid transport",
			client: &Client{
				httpClient: &http.Client{
					Transport: &http.Transport{
						MaxIdleConns:        15,
						MaxIdleConnsPerHost: 8,
						MaxConnsPerHost:     15,
						IdleConnTimeout:     90 * time.Second,
					},
				},
			},
			wantConfig: &ConnectionPoolConfig{
				MaxIdleConns:        15,
				MaxIdleConnsPerHost: 8,
				MaxConnsPerHost:     15,
				IdleConnTimeout:     90 * time.Second,
			},
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.client.GetConnectionPoolConfig()

			if tt.wantNil {
				if got != nil {
					t.Errorf("GetConnectionPoolConfig() = %v, want nil", got)
				}
				return
			}

			if got == nil {
				t.Fatal("GetConnectionPoolConfig() returned nil, want non-nil")
			}

			if got.MaxIdleConns != tt.wantConfig.MaxIdleConns {
				t.Errorf("MaxIdleConns = %v, want %v", got.MaxIdleConns, tt.wantConfig.MaxIdleConns)
			}

			if got.MaxIdleConnsPerHost != tt.wantConfig.MaxIdleConnsPerHost {
				t.Errorf("MaxIdleConnsPerHost = %v, want %v", got.MaxIdleConnsPerHost, tt.wantConfig.MaxIdleConnsPerHost)
			}

			if got.MaxConnsPerHost != tt.wantConfig.MaxConnsPerHost {
				t.Errorf("MaxConnsPerHost = %v, want %v", got.MaxConnsPerHost, tt.wantConfig.MaxConnsPerHost)
			}

			if got.IdleConnTimeout != tt.wantConfig.IdleConnTimeout {
				t.Errorf("IdleConnTimeout = %v, want %v", got.IdleConnTimeout, tt.wantConfig.IdleConnTimeout)
			}
		})
	}
}

// customTransport is a non-http.Transport type for testing
type customTransport struct {
	http.RoundTripper
}

func (c *customTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, nil
}

func TestConnectionPoolConfigIntegration(t *testing.T) {
	// Test that all options can be chained together
	client := &Client{
		httpClient: &http.Client{},
	}

	config := &ConnectionPoolConfig{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 50,
		MaxConnsPerHost:     100,
		IdleConnTimeout:     5 * time.Minute,
	}

	// Apply WithConnectionPool
	WithConnectionPool(config)(client)

	// Verify the config can be retrieved
	retrievedConfig := client.GetConnectionPoolConfig()
	if retrievedConfig == nil {
		t.Fatal("GetConnectionPoolConfig() returned nil")
	}

	if retrievedConfig.MaxIdleConns != 100 {
		t.Errorf("MaxIdleConns = %v, want 100", retrievedConfig.MaxIdleConns)
	}

	if retrievedConfig.MaxIdleConnsPerHost != 50 {
		t.Errorf("MaxIdleConnsPerHost = %v, want 50", retrievedConfig.MaxIdleConnsPerHost)
	}

	if retrievedConfig.MaxConnsPerHost != 100 {
		t.Errorf("MaxConnsPerHost = %v, want 100", retrievedConfig.MaxConnsPerHost)
	}

	if retrievedConfig.IdleConnTimeout != 5*time.Minute {
		t.Errorf("IdleConnTimeout = %v, want 5m", retrievedConfig.IdleConnTimeout)
	}
}

func TestIndividualOptionsOverride(t *testing.T) {
	// Test that individual options can override WithConnectionPool settings
	client := &Client{
		httpClient: &http.Client{},
	}

	// First apply a full config
	config := &ConnectionPoolConfig{
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 5,
		MaxConnsPerHost:     10,
		IdleConnTimeout:     30 * time.Second,
	}
	WithConnectionPool(config)(client)

	// Then override individual values
	WithMaxIdleConns(200)(client)
	WithMaxIdleConnsPerHost(100)(client)
	WithMaxConnsPerHost(200)(client)
	WithIdleConnTimeout(10 * time.Minute)(client)

	retrievedConfig := client.GetConnectionPoolConfig()
	if retrievedConfig == nil {
		t.Fatal("GetConnectionPoolConfig() returned nil")
	}

	if retrievedConfig.MaxIdleConns != 200 {
		t.Errorf("MaxIdleConns = %v, want 200", retrievedConfig.MaxIdleConns)
	}

	if retrievedConfig.MaxIdleConnsPerHost != 100 {
		t.Errorf("MaxIdleConnsPerHost = %v, want 100", retrievedConfig.MaxIdleConnsPerHost)
	}

	if retrievedConfig.MaxConnsPerHost != 200 {
		t.Errorf("MaxConnsPerHost = %v, want 200", retrievedConfig.MaxConnsPerHost)
	}

	if retrievedConfig.IdleConnTimeout != 10*time.Minute {
		t.Errorf("IdleConnTimeout = %v, want 10m", retrievedConfig.IdleConnTimeout)
	}
}
