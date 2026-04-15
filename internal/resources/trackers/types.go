// Package trackers provides types for managing Redmine trackers.
package trackers

type ValueLabel struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type TrackerCustomField struct {
	ID             int          `json:"id"`
	Name           string       `json:"name"`
	FieldFormat    string       `json:"field_format"`
	PossibleValues []ValueLabel `json:"possible_values,omitempty"`
}

type Tracker struct {
	ID            int                  `json:"id"`
	Name          string               `json:"name"`
	DefaultStatus *int                 `json:"default_status,omitempty"`
	Description   string               `json:"description,omitempty"`
	CustomFields  []TrackerCustomField `json:"custom_fields,omitempty"`
}

type TrackerList struct {
	Trackers []Tracker `json:"trackers"`
}
