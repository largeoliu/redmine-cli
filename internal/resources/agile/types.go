package agile

// Sprint represents a Redmine agile sprint.
type Sprint struct {
	ID             int     `json:"id"`
	Name           string  `json:"name"`
	Description    string  `json:"description,omitempty"`
	Status         string  `json:"status,omitempty"`
	StartDate      string  `json:"start_date,omitempty"`
	EndDate        string  `json:"end_date,omitempty"`
	Goal           string  `json:"goal,omitempty"`
	IsDefault      bool    `json:"is_default,omitempty"`
	IsClosed       bool    `json:"is_closed,omitempty"`
	IsArchived     bool    `json:"is_archived,omitempty"`
	ProjectID      int     `json:"project_id,omitempty"`
	StoryPoints    float64 `json:"story_points,omitempty"`
	DoneRatio      float64 `json:"done_ratio,omitempty"`
	EstimatedHours float64 `json:"estimated_hours,omitempty"`
	SpentHours     float64 `json:"spent_hours,omitempty"`
}

// SprintList represents a list of agile sprints.
type SprintList struct {
	AgileSprints []Sprint `json:"agile_sprints"`
	TotalCount   int      `json:"total_count,omitempty"`
	Limit        int      `json:"limit,omitempty"`
	Offset       int      `json:"offset,omitempty"`
}

// Data represents agile metadata for an issue.
type Data struct {
	AgileSprintID *int    `json:"agile_sprint_id,omitempty"`
	StoryPoints   float64 `json:"story_points,omitempty"`
	Position      int     `json:"position,omitempty"`
}
