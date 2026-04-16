// Package trackers provides a client for managing Redmine trackers.
package trackers

import (
	"context"
	"strconv"

	"github.com/largeoliu/redmine-cli/internal/client"
)

// Client is a client for managing Redmine trackers.
type Client struct {
	client *client.Client
}

// NewClient creates a new trackers client.
func NewClient(c *client.Client) *Client {
	return &Client{client: c}
}

// List retrieves all trackers.
func (c *Client) List(ctx context.Context) (*TrackerList, error) {
	var result TrackerList
	if err := c.client.Get(ctx, "/trackers.json", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Get retrieves a tracker by ID.
func (c *Client) Get(ctx context.Context, id int) (*Tracker, error) {
	var result struct {
		Tracker *Tracker `json:"tracker"`
	}
	if err := c.client.Get(ctx, "/trackers/"+strconv.Itoa(id)+".json", &result); err != nil {
		return nil, err
	}
	return result.Tracker, nil
}
