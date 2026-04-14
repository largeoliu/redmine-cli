// Package trackers provides types for managing Redmine trackers.
package trackers

// Tracker represents a Redmine tracker.
type Tracker struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	DefaultStatus *int   `json:"default_status,omitempty"`
	Description   string `json:"description,omitempty"`
}

// TrackerList represents a list of trackers.
type TrackerList struct {
	Trackers []Tracker `json:"trackers"`
}
