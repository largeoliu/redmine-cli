// Package client provides the HTTP client for interacting with Redmine API.
package client

import (
	"net/http"
	"time"
)

// ConnectionPoolConfig 定义 HTTP 连接池配置
type ConnectionPoolConfig struct {
	// MaxIdleConns 控制所有主机的最大空闲连接数
	// 默认值: 10
	MaxIdleConns int

	// MaxIdleConnsPerHost 控制每个主机的最大空闲连接数
	// 默认值: 5
	MaxIdleConnsPerHost int

	// MaxConnsPerHost 控制每个主机的最大连接数（包括空闲和活跃）
	// 默认值: 10
	MaxConnsPerHost int

	// IdleConnTimeout 是空闲连接在关闭前保持打开的最长时间
	// 默认值: 30 * time.Second
	IdleConnTimeout time.Duration
}

// DefaultConnectionPoolConfig 返回默认的连接池配置
func DefaultConnectionPoolConfig() *ConnectionPoolConfig {
	return &ConnectionPoolConfig{
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 5,
		MaxConnsPerHost:     10,
		IdleConnTimeout:     30 * time.Second,
	}
}

// WithConnectionPool 返回一个配置连接池的选项函数
// 如果传入 nil，则使用默认配置
func WithConnectionPool(config *ConnectionPoolConfig) Option {
	return func(c *Client) {
		if config == nil {
			config = DefaultConnectionPoolConfig()
		}

		// 如果 httpClient 还没有 Transport，创建一个新的
		if c.httpClient.Transport == nil {
			c.httpClient.Transport = &http.Transport{}
		}

		// 配置 Transport 的连接池参数
		if transport, ok := c.httpClient.Transport.(*http.Transport); ok {
			transport.MaxIdleConns = config.MaxIdleConns
			transport.MaxIdleConnsPerHost = config.MaxIdleConnsPerHost
			transport.MaxConnsPerHost = config.MaxConnsPerHost
			transport.IdleConnTimeout = config.IdleConnTimeout
		}
	}
}

// WithMaxIdleConns 返回一个配置最大空闲连接数的选项函数
func WithMaxIdleConns(maxIdleConns int) Option {
	return func(c *Client) {
		if c.httpClient.Transport == nil {
			c.httpClient.Transport = &http.Transport{}
		}

		if transport, ok := c.httpClient.Transport.(*http.Transport); ok {
			transport.MaxIdleConns = maxIdleConns
		}
	}
}

// WithMaxIdleConnsPerHost 返回一个配置每个主机最大空闲连接数的选项函数
func WithMaxIdleConnsPerHost(maxIdleConnsPerHost int) Option {
	return func(c *Client) {
		if c.httpClient.Transport == nil {
			c.httpClient.Transport = &http.Transport{}
		}

		if transport, ok := c.httpClient.Transport.(*http.Transport); ok {
			transport.MaxIdleConnsPerHost = maxIdleConnsPerHost
		}
	}
}

// WithIdleConnTimeout 返回一个配置空闲连接超时的选项函数
func WithIdleConnTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		if c.httpClient.Transport == nil {
			c.httpClient.Transport = &http.Transport{}
		}

		if transport, ok := c.httpClient.Transport.(*http.Transport); ok {
			transport.IdleConnTimeout = timeout
		}
	}
}

// WithMaxConnsPerHost 返回一个配置每个主机最大连接数的选项函数
func WithMaxConnsPerHost(maxConnsPerHost int) Option {
	return func(c *Client) {
		if c.httpClient.Transport == nil {
			c.httpClient.Transport = &http.Transport{}
		}

		if transport, ok := c.httpClient.Transport.(*http.Transport); ok {
			transport.MaxConnsPerHost = maxConnsPerHost
		}
	}
}

// GetConnectionPoolConfig 从 Client 获取当前的连接池配置
// 如果 Transport 不是 *http.Transport 类型，返回 nil
func (c *Client) GetConnectionPoolConfig() *ConnectionPoolConfig {
	if c.httpClient == nil || c.httpClient.Transport == nil {
		return nil
	}

	transport, ok := c.httpClient.Transport.(*http.Transport)
	if !ok {
		return nil
	}

	return &ConnectionPoolConfig{
		MaxIdleConns:        transport.MaxIdleConns,
		MaxIdleConnsPerHost: transport.MaxIdleConnsPerHost,
		MaxConnsPerHost:     transport.MaxConnsPerHost,
		IdleConnTimeout:     transport.IdleConnTimeout,
	}
}
