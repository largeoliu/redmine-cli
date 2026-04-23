// Package versions provides types for managing Redmine versions.
package versions

import (
	"time"

	"github.com/largeoliu/redmine-cli/internal/resources/projects"
)

// Version represents a Redmine version.
type Version struct {
	ID          int                 `json:"id"`
	Project     *projects.Reference `json:"project,omitempty"`
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	Status      string              `json:"status"`
	DueDate     string              `json:"due_date,omitempty"`
	CreatedOn   *time.Time          `json:"created_on,omitempty"`
	UpdatedOn   *time.Time          `json:"updated_on,omitempty"`
}

// VersionList represents a list of versions with pagination info.
type VersionList struct {
	Versions   []Version `json:"versions"`
	TotalCount int       `json:"total_count"`
	Limit      int       `json:"limit"`
	Offset     int       `json:"offset"`
}

// VersionCreateRequest represents a request to create a version.
type VersionCreateRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status,omitempty"`
	DueDate     string `json:"due_date,omitempty"`
	Sharing     string `json:"sharing,omitempty"`
}

// VersionUpdateRequest represents a request to update a version.
type VersionUpdateRequest = VersionCreateRequest
