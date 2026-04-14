// Package statuses provides a client for managing Redmine issue statuses.
package statuses

import (
	"context"

	"github.com/largeoliu/redmine-cli/internal/client"
)

// Client is a client for managing Redmine issue statuses.
type Client struct {
	client *client.Client
}

// NewClient creates a new statuses client.
func NewClient(c *client.Client) *Client {
	return &Client{client: c}
}

// List retrieves all issue statuses.
func (c *Client) List(ctx context.Context) (*IssueStatusList, error) {
	var result IssueStatusList
	if err := c.client.Get(ctx, "/issue_statuses.json", &result); err != nil {
		return nil, err
	}
	return &result, nil
}
