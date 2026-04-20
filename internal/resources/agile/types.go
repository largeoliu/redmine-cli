package agile

// Sprint represents a Redmine agile sprint.
type Sprint struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Status     string `json:"status,omitempty"`
	StartDate  string `json:"start_date,omitempty"`
	EndDate    string `json:"end_date,omitempty"`
	Goal       string `json:"goal,omitempty"`
	IsDefault  bool   `json:"is_default,omitempty"`
	IsClosed   bool   `json:"is_closed,omitempty"`
	IsArchived bool   `json:"is_archived,omitempty"`
}

// SprintList represents a list of agile sprints.
type SprintList struct {
	AgileSprints []Sprint `json:"agile_sprints"`
	TotalCount   int      `json:"total_count,omitempty"`
	Limit        int      `json:"limit,omitempty"`
	Offset       int      `json:"offset,omitempty"`
}

// AgileData represents agile metadata for an issue.
type AgileData struct {
	AgileSprintID *int    `json:"agile_sprint_id,omitempty"`
	StoryPoints   float64 `json:"story_points,omitempty"`
	Position      int     `json:"position,omitempty"`
}
