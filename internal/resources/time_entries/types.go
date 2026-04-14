// Package time_entries provides types for managing Redmine time entries.
//
//nolint:revive // package name has underscore due to Redmine API naming convention
package time_entries

import "time"

// TimeEntry represents a Redmine time entry.
type TimeEntry struct {
	ID        int        `json:"id"`
	Project   *Reference `json:"project,omitempty"`
	Issue     *Reference `json:"issue,omitempty"`
	User      *Reference `json:"user,omitempty"`
	Activity  *Reference `json:"activity,omitempty"`
	Hours     float64    `json:"hours"`
	Comments  string     `json:"comments"`
	SpentOn   string     `json:"spent_on"`
	CreatedOn *time.Time `json:"created_on,omitempty"`
	UpdatedOn *time.Time `json:"updated_on,omitempty"`
}

// Reference represents a reference to another Redmine entity.
type Reference struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// TimeEntryList represents a list of time entries with pagination info.
type TimeEntryList struct {
	TimeEntries []TimeEntry `json:"time_entries"`
	TotalCount  int         `json:"total_count"`
	Limit       int         `json:"limit"`
	Offset      int         `json:"offset"`
}

// TimeEntryCreateRequest represents a request to create a time entry.
type TimeEntryCreateRequest struct {
	IssueID    int     `json:"issue_id,omitempty"`
	ProjectID  int     `json:"project_id,omitempty"`
	SpentOn    string  `json:"spent_on,omitempty"`
	Hours      float64 `json:"hours,omitempty"`
	ActivityID int     `json:"activity_id,omitempty"`
	Comments   string  `json:"comments,omitempty"`
}

// TimeEntryUpdateRequest represents a request to update a time entry.
type TimeEntryUpdateRequest = TimeEntryCreateRequest

// TimeEntryActivity represents a Redmine time entry activity.
type TimeEntryActivity struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	IsDefault bool   `json:"is_default"`
	Active    bool   `json:"active"`
}
