// Package projects provides types for managing Redmine projects.
package projects

import "time"

// Project represents a Redmine project.
type Project struct {
	ID              int         `json:"id"`
	Name            string      `json:"name"`
	Identifier      string      `json:"identifier"`
	Description     string      `json:"description"`
	Homepage        string      `json:"homepage"`
	Status          int         `json:"status"`
	Parent          *Reference  `json:"parent,omitempty"`
	CreatedOn       *time.Time  `json:"created_on,omitempty"`
	UpdatedOn       *time.Time  `json:"updated_on,omitempty"`
	Trackers        []Reference `json:"trackers,omitempty"`
	IssueCategories []Reference `json:"issue_categories,omitempty"`
	EnabledModules  []Reference `json:"enabled_modules,omitempty"`
}

// Reference represents a reference to another Redmine entity.
type Reference struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// ProjectList represents a list of projects with pagination info.
type ProjectList struct {
	Projects   []Project `json:"projects"`
	TotalCount int       `json:"total_count"`
	Limit      int       `json:"limit"`
	Offset     int       `json:"offset"`
}

// ProjectCreateRequest represents a request to create a project.
type ProjectCreateRequest struct {
	Name           string `json:"name,omitempty"`
	Identifier     string `json:"identifier,omitempty"`
	Description    string `json:"description,omitempty"`
	Homepage       string `json:"homepage,omitempty"`
	IsPublic       bool   `json:"is_public,omitempty"`
	ParentID       int    `json:"parent_id,omitempty"`
	InheritMembers bool   `json:"inherit_members,omitempty"`
}

// ProjectUpdateRequest represents a request to update a project.
type ProjectUpdateRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Homepage    string `json:"homepage,omitempty"`
	Status      int    `json:"status,omitempty"`
}
