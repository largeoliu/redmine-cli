// Package trackers provides a client for managing Redmine trackers.
package trackers

import (
	"context"
	"strconv"

	"github.com/largeoliu/redmine-cli/internal/client"
)

type Client struct {
	client *client.Client
}

func NewClient(c *client.Client) *Client {
	return &Client{client: c}
}

func (c *Client) List(ctx context.Context) (*TrackerList, error) {
	var result TrackerList
	if err := c.client.Get(ctx, "/trackers.json", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) Get(ctx context.Context, id int) (*Tracker, error) {
	var result struct {
		Tracker *Tracker `json:"tracker"`
	}
	if err := c.client.Get(ctx, "/trackers/"+strconv.Itoa(id)+".json", &result); err != nil {
		return nil, err
	}
	return result.Tracker, nil
}
