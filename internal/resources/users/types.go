// Package users provides types for managing Redmine users.
package users

import "time"

// User represents a Redmine user.
type User struct {
	ID                 int           `json:"id"`
	Login              string        `json:"login"`
	Firstname          string        `json:"firstname"`
	Lastname           string        `json:"lastname"`
	Mail               string        `json:"mail"`
	CreatedOn          *time.Time    `json:"created_on,omitempty"`
	LastLoginOn        *time.Time    `json:"last_login_on,omitempty"`
	Admin              bool          `json:"admin"`
	Status             int           `json:"status"`
	AuthSourceID       int           `json:"auth_source_id,omitempty"`
	MustChangePassword bool          `json:"must_change_passwd,omitempty"`
	AvatarURL          string        `json:"avatar_url,omitempty"`
	CustomFields       []CustomField `json:"custom_fields,omitempty"`
}

// CustomField represents a custom field value.
type CustomField struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Value any    `json:"value"`
}

// UserList represents a list of users with pagination info.
type UserList struct {
	Users      []User `json:"users"`
	TotalCount int    `json:"total_count"`
	Limit      int    `json:"limit"`
	Offset     int    `json:"offset"`
}

// UserCreateRequest represents a request to create a user.
type UserCreateRequest struct {
	Login              string `json:"login,omitempty"`
	Firstname          string `json:"firstname,omitempty"`
	Lastname           string `json:"lastname,omitempty"`
	Mail               string `json:"mail,omitempty"`
	Password           string `json:"password,omitempty"`
	Admin              bool   `json:"admin,omitempty"`
	Status             int    `json:"status,omitempty"`
	AuthSourceID       int    `json:"auth_source_id,omitempty"`
	MustChangePassword bool   `json:"must_change_passwd,omitempty"`
}

// UserUpdateRequest represents a request to update a user.
type UserUpdateRequest = UserCreateRequest
