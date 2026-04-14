// Package categories provides a client for managing Redmine issue categories.
package categories

import (
	"context"
	"fmt"

	"github.com/largeoliu/redmine-cli/internal/client"
)

// Client is a client for managing Redmine issue categories.
type Client struct {
	client *client.Client
}

// NewClient creates a new categories client.
func NewClient(c *client.Client) *Client {
	return &Client{client: c}
}

// List retrieves all issue categories for a project.
func (c *Client) List(ctx context.Context, projectID int) (*CategoryList, error) {
	path := fmt.Sprintf("/projects/%d/issue_categories.json", projectID)
	var result CategoryList
	if err := c.client.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Get retrieves an issue category by ID.
func (c *Client) Get(ctx context.Context, id int) (*Category, error) {
	path := fmt.Sprintf("/issue_categories/%d.json", id)
	var result struct {
		IssueCategory Category `json:"issue_category"`
	}
	if err := c.client.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result.IssueCategory, nil
}

// Create creates a new issue category for a project.
func (c *Client) Create(ctx context.Context, projectID int, req *CategoryCreateRequest) (*Category, error) {
	path := fmt.Sprintf("/projects/%d/issue_categories.json", projectID)
	var result struct {
		IssueCategory Category `json:"issue_category"`
	}
	body := map[string]any{"issue_category": req}
	if err := c.client.Post(ctx, path, body, &result); err != nil {
		return nil, err
	}
	return &result.IssueCategory, nil
}

// Update updates an existing issue category.
func (c *Client) Update(ctx context.Context, id int, req *CategoryUpdateRequest) error {
	path := fmt.Sprintf("/issue_categories/%d.json", id)
	body := map[string]any{"issue_category": req}
	return c.client.Put(ctx, path, body, nil)
}

// Delete deletes an issue category.
func (c *Client) Delete(ctx context.Context, id int) error {
	path := fmt.Sprintf("/issue_categories/%d.json", id)
	return c.client.Delete(ctx, path)
}
