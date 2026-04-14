// Package client provides the HTTP client for interacting with Redmine API.
package client

import (
	"context"
)

// TestAuth tests the authentication by making a request to the current user endpoint.
func (c *Client) TestAuth(ctx context.Context) error {
	var result struct {
		User struct {
			ID    int    `json:"id"`
			Login string `json:"login"`
		} `json:"user"`
	}
	return c.Get(ctx, "/users/current.json", &result)
}

// SetAPIKey sets the API key for the client.
func (c *Client) SetAPIKey(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.apiKey = key
}

// SetBaseURL sets the base URL for the client.
func (c *Client) SetBaseURL(newURL string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.baseURL = newURL
}
