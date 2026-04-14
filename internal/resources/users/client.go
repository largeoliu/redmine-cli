// Package users provides a client for managing Redmine users.
package users

import (
	"context"
	"fmt"
	"strconv"

	"github.com/largeoliu/redmine-cli/internal/client"
)

// Client is a client for managing Redmine users.
type Client struct {
	client *client.Client
}

// NewClient creates a new users client.
func NewClient(c *client.Client) *Client {
	return &Client{client: c}
}

// List retrieves users based on the provided parameters.
func (c *Client) List(ctx context.Context, params map[string]string) (*UserList, error) {
	path := "/users.json"
	if len(params) > 0 {
		var err error
		if path, err = c.client.BuildPath(path, params); err != nil {
			return nil, err
		}
	}
	var result UserList
	if err := c.client.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Get retrieves a user by ID.
func (c *Client) Get(ctx context.Context, id int, params map[string]string) (*User, error) {
	path := fmt.Sprintf("/users/%d.json", id)
	if len(params) > 0 {
		var err error
		if path, err = c.client.BuildPath(path, params); err != nil {
			return nil, err
		}
	}
	var result struct {
		User User `json:"user"`
	}
	if err := c.client.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result.User, nil
}

// GetCurrent retrieves the current authenticated user.
func (c *Client) GetCurrent(ctx context.Context) (*User, error) {
	var result struct {
		User User `json:"user"`
	}
	if err := c.client.Get(ctx, "/users/current.json", &result); err != nil {
		return nil, err
	}
	return &result.User, nil
}

// Create creates a new user.
func (c *Client) Create(ctx context.Context, req *UserCreateRequest) (*User, error) {
	var result struct {
		User User `json:"user"`
	}
	body := map[string]any{"user": req}
	if err := c.client.Post(ctx, "/users.json", body, &result); err != nil {
		return nil, err
	}
	return &result.User, nil
}

// Update updates an existing user.
func (c *Client) Update(ctx context.Context, id int, req *UserUpdateRequest) error {
	path := fmt.Sprintf("/users/%d.json", id)
	body := map[string]any{"user": req}
	return c.client.Put(ctx, path, body, nil)
}

// Delete deletes a user.
func (c *Client) Delete(ctx context.Context, id int) error {
	path := fmt.Sprintf("/users/%d.json", id)
	return c.client.Delete(ctx, path)
}

// ListFlags represents flags for filtering users in a list operation.
type ListFlags struct {
	Status  int
	Name    string
	GroupID int
	Limit   int
	Offset  int
}

// BuildListParams converts ListFlags to query parameters.
func BuildListParams(flags ListFlags) map[string]string {
	params := make(map[string]string)
	if flags.Status > 0 {
		params["status"] = strconv.Itoa(flags.Status)
	}
	if flags.Name != "" {
		params["name"] = flags.Name
	}
	if flags.GroupID > 0 {
		params["group_id"] = strconv.Itoa(flags.GroupID)
	}
	if flags.Limit > 0 {
		params["limit"] = strconv.Itoa(flags.Limit)
	}
	if flags.Offset > 0 {
		params["offset"] = strconv.Itoa(flags.Offset)
	}
	return params
}
