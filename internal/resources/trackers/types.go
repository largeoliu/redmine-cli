// Package trackers provides types for managing Redmine trackers.
package trackers

// ValueLabel represents a list field option.
type ValueLabel struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// TrackerCustomField represents a custom field definition on a tracker.
type TrackerCustomField struct {
	ID             int          `json:"id"`
	Name           string       `json:"name"`
	FieldFormat    string       `json:"field_format"`
	PossibleValues []ValueLabel `json:"possible_values,omitempty"`
}

// Tracker represents a Redmine tracker.
type Tracker struct {
	ID            int                  `json:"id"`
	Name          string               `json:"name"`
	DefaultStatus *int                 `json:"default_status,omitempty"`
	Description   string               `json:"description,omitempty"`
	CustomFields  []TrackerCustomField `json:"custom_fields,omitempty"`
}

// TrackerList represents a list of trackers.
type TrackerList struct {
	Trackers []Tracker `json:"trackers"`
}
