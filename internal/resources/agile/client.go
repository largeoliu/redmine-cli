package agile

import (
	"context"
	"fmt"

	"github.com/largeoliu/redmine-cli/internal/client"
)

// Client is a client for Redmine agile endpoints.
type Client struct {
	client *client.Client
}

// NewClient creates a new agile client.
func NewClient(c *client.Client) *Client {
	return &Client{client: c}
}

// ListSprints retrieves agile sprints for a project.
func (c *Client) ListSprints(ctx context.Context, projectID int) (*SprintList, error) {
	var result SprintList
	if err := c.client.Get(ctx, fmt.Sprintf("/projects/%d/agile_sprints.json", projectID), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetSprint retrieves a sprint by project and sprint ID.
func (c *Client) GetSprint(ctx context.Context, projectID, sprintID int) (*Sprint, error) {
	var result struct {
		AgileSprint Sprint `json:"agile_sprint"`
	}
	if err := c.client.Get(ctx, fmt.Sprintf("/projects/%d/agile_sprints/%d.json", projectID, sprintID), &result); err != nil {
		return nil, err
	}
	return &result.AgileSprint, nil
}

// GetIssueAgileData retrieves agile metadata for an issue.
func (c *Client) GetIssueAgileData(ctx context.Context, issueID int) (*AgileData, error) {
	var result struct {
		AgileData AgileData `json:"agile_data"`
	}
	if err := c.client.Get(ctx, fmt.Sprintf("/issues/%d/agile_data.json", issueID), &result); err != nil {
		return nil, err
	}
	return &result.AgileData, nil
}
