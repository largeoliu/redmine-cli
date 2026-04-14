// Package time_entries provides a client for managing Redmine time entries.
//
//nolint:revive // package name has underscore due to Redmine API naming convention
package time_entries

import (
	"context"
	"fmt"
	"strconv"

	"github.com/largeoliu/redmine-cli/internal/client"
)

// Client is a client for managing Redmine time entries.
type Client struct {
	client *client.Client
}

// NewClient creates a new time_entries client.
func NewClient(c *client.Client) *Client {
	return &Client{client: c}
}

// List retrieves time entries based on the provided parameters.
func (c *Client) List(ctx context.Context, params map[string]string) (*TimeEntryList, error) {
	path := "/time_entries.json"
	if len(params) > 0 {
		var err error
		if path, err = c.client.BuildPath(path, params); err != nil {
			return nil, err
		}
	}
	var result TimeEntryList
	if err := c.client.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Get retrieves a time entry by ID.
func (c *Client) Get(ctx context.Context, id int) (*TimeEntry, error) {
	path := fmt.Sprintf("/time_entries/%d.json", id)
	var result struct {
		TimeEntry TimeEntry `json:"time_entry"`
	}
	if err := c.client.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result.TimeEntry, nil
}

// Create creates a new time entry.
func (c *Client) Create(ctx context.Context, req *TimeEntryCreateRequest) (*TimeEntry, error) {
	var result struct {
		TimeEntry TimeEntry `json:"time_entry"`
	}
	body := map[string]any{"time_entry": req}
	if err := c.client.Post(ctx, "/time_entries.json", body, &result); err != nil {
		return nil, err
	}
	return &result.TimeEntry, nil
}

// Update updates an existing time entry.
func (c *Client) Update(ctx context.Context, id int, req *TimeEntryUpdateRequest) error {
	path := fmt.Sprintf("/time_entries/%d.json", id)
	body := map[string]any{"time_entry": req}
	return c.client.Put(ctx, path, body, nil)
}

// Delete deletes a time entry.
func (c *Client) Delete(ctx context.Context, id int) error {
	path := fmt.Sprintf("/time_entries/%d.json", id)
	return c.client.Delete(ctx, path)
}

// ListActivities retrieves all time entry activities.
func (c *Client) ListActivities(ctx context.Context) ([]TimeEntryActivity, error) {
	path := "/enumerations/time_entry_activities.json"
	var result struct {
		TimeEntryActivities []TimeEntryActivity `json:"time_entry_activities"`
	}
	if err := c.client.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return result.TimeEntryActivities, nil
}

// ListFlags represents flags for filtering time entries in a list operation.
type ListFlags struct {
	ProjectID  int
	IssueID    int
	UserID     int
	ActivityID int
	From       string
	To         string
	Limit      int
	Offset     int
}

// BuildListParams converts ListFlags to query parameters.
func BuildListParams(flags ListFlags) map[string]string {
	params := make(map[string]string)
	if flags.ProjectID > 0 {
		params["project_id"] = strconv.Itoa(flags.ProjectID)
	}
	if flags.IssueID > 0 {
		params["issue_id"] = strconv.Itoa(flags.IssueID)
	}
	if flags.UserID > 0 {
		params["user_id"] = strconv.Itoa(flags.UserID)
	}
	if flags.ActivityID > 0 {
		params["activity_id"] = strconv.Itoa(flags.ActivityID)
	}
	if flags.From != "" {
		params["from"] = flags.From
	}
	if flags.To != "" {
		params["to"] = flags.To
	}
	if flags.Limit > 0 {
		params["limit"] = strconv.Itoa(flags.Limit)
	}
	if flags.Offset > 0 {
		params["offset"] = strconv.Itoa(flags.Offset)
	}
	return params
}
