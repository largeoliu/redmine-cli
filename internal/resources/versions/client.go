// Package versions provides a client for managing Redmine versions.
package versions

import (
	"context"
	"fmt"

	"github.com/largeoliu/redmine-cli/internal/client"
)

// Client is a client for managing Redmine versions.
type Client struct {
	client *client.Client
}

// NewClient creates a new versions client.
func NewClient(c *client.Client) *Client {
	return &Client{client: c}
}

// List retrieves versions for a project.
func (c *Client) List(ctx context.Context, projectID int, params map[string]string) (*VersionList, error) {
	path := fmt.Sprintf("/projects/%d/versions.json", projectID)
	if len(params) > 0 {
		var err error
		if path, err = c.client.BuildPath(path, params); err != nil {
			return nil, err
		}
	}
	var result VersionList
	if err := c.client.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Get retrieves a version by ID.
func (c *Client) Get(ctx context.Context, id int) (*Version, error) {
	path := fmt.Sprintf("/versions/%d.json", id)
	var result struct {
		Version Version `json:"version"`
	}
	if err := c.client.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result.Version, nil
}

// Create creates a new version for a project.
func (c *Client) Create(ctx context.Context, projectID int, req *VersionCreateRequest) (*Version, error) {
	path := fmt.Sprintf("/projects/%d/versions.json", projectID)
	var result struct {
		Version Version `json:"version"`
	}
	body := map[string]any{"version": req}
	if err := c.client.Post(ctx, path, body, &result); err != nil {
		return nil, err
	}
	return &result.Version, nil
}

// Update updates an existing version.
func (c *Client) Update(ctx context.Context, id int, req *VersionUpdateRequest) error {
	path := fmt.Sprintf("/versions/%d.json", id)
	body := map[string]any{"version": req}
	return c.client.Put(ctx, path, body, nil)
}

// Delete deletes a version.
func (c *Client) Delete(ctx context.Context, id int) error {
	path := fmt.Sprintf("/versions/%d.json", id)
	return c.client.Delete(ctx, path)
}
