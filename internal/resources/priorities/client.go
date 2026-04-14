// Package priorities provides a client for managing Redmine issue priorities.
package priorities

import (
	"context"

	"github.com/largeoliu/redmine-cli/internal/client"
)

// Client is a client for managing Redmine issue priorities.
type Client struct {
	client *client.Client
}

// NewClient creates a new priorities client.
func NewClient(c *client.Client) *Client {
	return &Client{client: c}
}

// List retrieves all issue priorities.
func (c *Client) List(ctx context.Context) (*PriorityList, error) {
	var result PriorityList
	if err := c.client.Get(ctx, "/enumerations/issue_priorities.json", &result); err != nil {
		return nil, err
	}
	return &result, nil
}
