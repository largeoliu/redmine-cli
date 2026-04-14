// Package issues provides types for managing Redmine issues.
package issues

import "time"

// Issue represents a Redmine issue.
type Issue struct {
	ID           int           `json:"id"`
	Subject      string        `json:"subject"`
	Description  string        `json:"description"`
	Project      *Reference    `json:"project,omitempty"`
	Tracker      *Reference    `json:"tracker,omitempty"`
	Status       *Reference    `json:"status,omitempty"`
	Priority     *Reference    `json:"priority,omitempty"`
	Author       *Reference    `json:"author,omitempty"`
	AssignedTo   *Reference    `json:"assigned_to,omitempty"`
	Category     *Reference    `json:"category,omitempty"`
	FixedVersion *Reference    `json:"fixed_version,omitempty"`
	Parent       *Reference    `json:"parent,omitempty"`
	StartDate    string        `json:"start_date,omitempty"`
	DueDate      string        `json:"due_date,omitempty"`
	DoneRatio    int           `json:"done_ratio"`
	CreatedOn    *time.Time    `json:"created_on,omitempty"`
	UpdatedOn    *time.Time    `json:"updated_on,omitempty"`
	ClosedOn     *time.Time    `json:"closed_on,omitempty"`
	Notes        string        `json:"notes,omitempty"`
	PrivateNotes bool          `json:"private_notes"`
	Watchers     []Reference   `json:"watchers,omitempty"`
	CustomFields []CustomField `json:"custom_fields,omitempty"`
}

// Reference represents a reference to another Redmine entity.
type Reference struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// CustomField represents a custom field value.
type CustomField struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Value any    `json:"value"`
}

// IssueList represents a list of issues with pagination info.
type IssueList struct {
	Issues     []Issue `json:"issues"`
	TotalCount int     `json:"total_count"`
	Limit      int     `json:"limit"`
	Offset     int     `json:"offset"`
}

// IssueCreateRequest represents a request to create an issue.
type IssueCreateRequest struct {
	ProjectID      int    `json:"project_id,omitempty"`
	Subject        string `json:"subject,omitempty"`
	Description    string `json:"description,omitempty"`
	TrackerID      int    `json:"tracker_id,omitempty"`
	StatusID       int    `json:"status_id,omitempty"`
	PriorityID     int    `json:"priority_id,omitempty"`
	AssignedToID   int    `json:"assigned_to_id,omitempty"`
	CategoryID     int    `json:"category_id,omitempty"`
	FixedVersionID int    `json:"fixed_version_id,omitempty"`
	ParentIssueID  int    `json:"parent_issue_id,omitempty"`
	StartDate      string `json:"start_date,omitempty"`
	DueDate        string `json:"due_date,omitempty"`
	DoneRatio      int    `json:"done_ratio,omitempty"`
	WatcherUserIDs []int  `json:"watcher_user_ids,omitempty"`
}

// IssueUpdateRequest represents a request to update an issue.
type IssueUpdateRequest struct {
	Subject        string `json:"subject,omitempty"`
	Description    string `json:"description,omitempty"`
	StatusID       int    `json:"status_id,omitempty"`
	PriorityID     int    `json:"priority_id,omitempty"`
	AssignedToID   int    `json:"assigned_to_id,omitempty"`
	CategoryID     int    `json:"category_id,omitempty"`
	FixedVersionID int    `json:"fixed_version_id,omitempty"`
	StartDate      string `json:"start_date,omitempty"`
	DueDate        string `json:"due_date,omitempty"`
	DoneRatio      int    `json:"done_ratio,omitempty"`
	Notes          string `json:"notes,omitempty"`
	PrivateNotes   bool   `json:"private_notes,omitempty"`
}
