// Package trackers provides a client for managing Redmine trackers.
package trackers

import (
	"context"
	"strconv"
	"strings"

	"github.com/largeoliu/redmine-cli/internal/client"
	"github.com/largeoliu/redmine-cli/internal/errors"
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

// FindByName retrieves a tracker by its exact name.
func (c *Client) FindByName(ctx context.Context, name string) (*Tracker, error) {
	result, err := c.List(ctx)
	if err != nil {
		return nil, err
	}

	needle := strings.TrimSpace(name)
	for i := range result.Trackers {
		if result.Trackers[i].Name == needle {
			return &result.Trackers[i], nil
		}
	}

	return nil, errors.NewValidation(
		"tracker not found: "+name,
		errors.WithHint("Use 'redmine tracker list' to find the correct tracker name."),
		errors.WithActions("redmine tracker list"),
	)
}
