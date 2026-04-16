// Package projects provides a client for managing Redmine projects.
package projects

import (
	"context"
	"fmt"

	"github.com/largeoliu/redmine-cli/internal/client"
)

// Client is a client for managing Redmine projects.
type Client struct {
	client *client.Client
}

// NewClient creates a new projects client.
func NewClient(c *client.Client) *Client {
	return &Client{client: c}
}

// List retrieves projects based on the provided parameters.
func (c *Client) List(ctx context.Context, params map[string]string) (*ProjectList, error) {
	path := c.client.BuildPath("/projects.json", params)
	var result ProjectList
	if err := c.client.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Get retrieves a project by ID.
func (c *Client) Get(ctx context.Context, id int, params map[string]string) (*Project, error) {
	return c.get(ctx, fmt.Sprintf("/projects/%d.json", id), params)
}

// GetByIdentifier retrieves a project by identifier.
func (c *Client) GetByIdentifier(ctx context.Context, identifier string, params map[string]string) (*Project, error) {
	return c.get(ctx, fmt.Sprintf("/projects/%s.json", identifier), params)
}

func (c *Client) get(ctx context.Context, path string, params map[string]string) (*Project, error) {
	path = c.client.BuildPath(path, params)
	var result struct {
		Project Project `json:"project"`
	}
	if err := c.client.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result.Project, nil
}

// Create creates a new project.
func (c *Client) Create(ctx context.Context, req *ProjectCreateRequest) (*Project, error) {
	var result struct {
		Project Project `json:"project"`
	}
	body := map[string]any{"project": req}
	if err := c.client.Post(ctx, "/projects.json", body, &result); err != nil {
		return nil, err
	}
	return &result.Project, nil
}

// Update updates an existing project.
func (c *Client) Update(ctx context.Context, id int, req *ProjectUpdateRequest) error {
	path := fmt.Sprintf("/projects/%d.json", id)
	body := map[string]any{"project": req}
	return c.client.Put(ctx, path, body, nil)
}

// Delete deletes a project.
func (c *Client) Delete(ctx context.Context, id int) error {
	path := fmt.Sprintf("/projects/%d.json", id)
	return c.client.Delete(ctx, path)
}
