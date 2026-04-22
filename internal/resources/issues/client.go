// Package issues provides a client for managing Redmine issues.
package issues

import (
	"context"
	"fmt"
	"strconv"

	"github.com/largeoliu/redmine-cli/internal/client"
)

// Client is a client for managing Redmine issues.
type Client struct {
	client *client.Client
}

// NewClient creates a new issues client.
func NewClient(c *client.Client) *Client {
	return &Client{client: c}
}

// List retrieves issues based on the provided parameters.
func (c *Client) List(ctx context.Context, params map[string]string) (*IssueList, error) {
	path := c.client.BuildPath("/issues.json", params)
	var result IssueList
	if err := c.client.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Get retrieves an issue by ID.
func (c *Client) Get(ctx context.Context, id int, params map[string]string) (*Issue, error) {
	path := c.client.BuildPath(fmt.Sprintf("/issues/%d.json", id), params)
	var result struct {
		Issue Issue `json:"issue"`
	}
	if err := c.client.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result.Issue, nil
}

// Create creates a new issue.
func (c *Client) Create(ctx context.Context, req *IssueCreateRequest) (*Issue, error) {
	var result struct {
		Issue Issue `json:"issue"`
	}
	body := map[string]any{"issue": req}
	if err := c.client.Post(ctx, "/issues.json", body, &result); err != nil {
		return nil, err
	}
	return &result.Issue, nil
}

// Update updates an existing issue.
func (c *Client) Update(ctx context.Context, id int, req *IssueUpdateRequest) error {
	path := fmt.Sprintf("/issues/%d.json", id)
	body := map[string]any{"issue": req}
	return c.client.Put(ctx, path, body, nil)
}

// Delete deletes an issue.
func (c *Client) Delete(ctx context.Context, id int) error {
	path := fmt.Sprintf("/issues/%d.json", id)
	return c.client.Delete(ctx, path)
}

// ListFlags represents flags for filtering issues in a list operation.
type ListFlags struct {
	ProjectID    int
	TrackerID    int
	VersionID    int
	StatusID     int
	AssignedToID int
	Limit        int
	Offset       int
	Query        string
	Sort         string
}

// BuildListParams converts ListFlags to query parameters.
func BuildListParams(flags ListFlags) map[string]string {
	params := make(map[string]string)
	if flags.ProjectID > 0 {
		params["project_id"] = strconv.Itoa(flags.ProjectID)
	}
	if flags.TrackerID > 0 {
		params["tracker_id"] = strconv.Itoa(flags.TrackerID)
	}
	if flags.VersionID > 0 {
		params["fixed_version_id"] = strconv.Itoa(flags.VersionID)
	}
	if flags.StatusID > 0 {
		params["status_id"] = strconv.Itoa(flags.StatusID)
	}
	if flags.AssignedToID > 0 {
		params["assigned_to_id"] = strconv.Itoa(flags.AssignedToID)
	}
	if flags.Limit > 0 {
		params["limit"] = strconv.Itoa(flags.Limit)
	}
	if flags.Offset > 0 {
		params["offset"] = strconv.Itoa(flags.Offset)
	}
	if flags.Query != "" {
		params["query_id"] = flags.Query
	}
	if flags.Sort != "" {
		params["sort"] = flags.Sort
	}
	return params
}
