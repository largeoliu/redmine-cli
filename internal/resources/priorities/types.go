// Package priorities provides types for managing Redmine issue priorities.
package priorities

// Priority represents a Redmine issue priority.
type Priority struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	IsDefault bool   `json:"is_default"`
	Position  int    `json:"position"`
}

// PriorityList represents a list of issue priorities.
type PriorityList struct {
	Priorities []Priority `json:"issue_priorities"`
}
