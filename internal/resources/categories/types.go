// Package categories provides types for managing Redmine issue categories.
package categories

import "github.com/largeoliu/redmine-cli/internal/resources/projects"

// Category represents a Redmine issue category.
type Category struct {
	ID         int                 `json:"id"`
	Project    *projects.Reference `json:"project,omitempty"`
	Name       string              `json:"name"`
	AssignedTo *projects.Reference `json:"assigned_to,omitempty"`
}

// CategoryList represents a list of issue categories.
type CategoryList struct {
	IssueCategories []Category `json:"issue_categories"`
}

// CategoryCreateRequest represents a request to create an issue category.
type CategoryCreateRequest struct {
	Name         string `json:"name,omitempty"`
	AssignedToID int    `json:"assigned_to_id,omitempty"`
}

// CategoryUpdateRequest represents a request to update an issue category.
type CategoryUpdateRequest = CategoryCreateRequest
