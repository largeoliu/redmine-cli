// Package statuses provides types for managing Redmine issue statuses.
package statuses

// IssueStatus represents a Redmine issue status.
type IssueStatus struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	IsClosed  bool   `json:"is_closed"`
	IsDefault bool   `json:"is_default"`
	Position  int    `json:"position"`
}

// IssueStatusList represents a list of issue statuses.
type IssueStatusList struct {
	IssueStatuses []IssueStatus `json:"issue_statuses"`
}
