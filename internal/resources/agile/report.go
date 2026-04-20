package agile

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/largeoliu/redmine-cli/internal/client"
	issuespkg "github.com/largeoliu/redmine-cli/internal/resources/issues"
	projectspkg "github.com/largeoliu/redmine-cli/internal/resources/projects"
	statusespkg "github.com/largeoliu/redmine-cli/internal/resources/statuses"
)

// BoardReport is the structured representation of an agile content view.
type BoardReport struct {
	Project       *projectspkg.Project `json:"project"`
	CurrentSprint *Sprint              `json:"current_sprint,omitempty"`
	Groups        []BoardGroup         `json:"groups"`
	Cards         []BoardCard          `json:"cards"`
}

// BoardGroup represents a sprint or no-sprint grouping.
type BoardGroup struct {
	Name   string      `json:"name"`
	Sprint *Sprint     `json:"sprint,omitempty"`
	Cards  []BoardCard `json:"cards"`
}

// BoardCard represents one issue entry in the agile content view.
type BoardCard struct {
	ID             int     `json:"id"`
	Subject        string  `json:"subject"`
	Status         string  `json:"status,omitempty"`
	StatusPosition int     `json:"status_position,omitempty"`
	Assignee       string  `json:"assignee,omitempty"`
	Priority       string  `json:"priority,omitempty"`
	DueDate        string  `json:"due_date,omitempty"`
	StoryPoints    float64 `json:"story_points,omitempty"`
	Position       int     `json:"position,omitempty"`
	SprintID       int     `json:"sprint_id,omitempty"`
	SprintName     string  `json:"sprint_name,omitempty"`
}

func buildBoardReport(ctx context.Context, c *client.Client, project *projectspkg.Project) (*BoardReport, []BoardCard, error) {
	agileClient := NewClient(c)
	issuesClient := issuespkg.NewClient(c)
	statusClient := statusespkg.NewClient(c)

	sprintList, err := agileClient.ListSprints(ctx, project.ID)
	if err != nil {
		return nil, nil, err
	}

	currentSprint := selectCurrentSprint(sprintList.AgileSprints)
	if currentSprint != nil {
		detailed, err := agileClient.GetSprint(ctx, project.ID, currentSprint.ID)
		if err != nil {
			return nil, nil, err
		}
		currentSprint = detailed
	}

	statusPositions, err := loadStatusPositions(ctx, statusClient)
	if err != nil {
		return nil, nil, err
	}

	allIssues, err := collectIssues(ctx, issuesClient, project.ID)
	if err != nil {
		return nil, nil, err
	}

	agileDataByIssueID, err := loadIssueAgileData(ctx, agileClient, allIssues)
	if err != nil {
		return nil, nil, err
	}

	report := &BoardReport{
		Project:       project,
		CurrentSprint: currentSprint,
	}

	var currentGroup *BoardGroup
	if currentSprint != nil {
		currentGroup = &BoardGroup{Name: currentSprint.Name, Sprint: currentSprint}
	}
	noSprintGroup := BoardGroup{Name: "No Sprint"}

	for _, issue := range allIssues {
		data := agileDataByIssueID[issue.ID]
		card := buildBoardCard(issue, data, statusPositions, currentSprint)

		switch {
		case currentSprint != nil && card.SprintID == currentSprint.ID:
			if currentGroup != nil {
				currentGroup.Cards = append(currentGroup.Cards, card)
			}
		case card.SprintID == 0:
			noSprintGroup.Cards = append(noSprintGroup.Cards, card)
		}
	}

	if currentGroup != nil {
		sortBoardCards(currentGroup.Cards)
		report.Groups = append(report.Groups, *currentGroup)
	}
	sortBoardCards(noSprintGroup.Cards)
	report.Groups = append(report.Groups, noSprintGroup)

	report.Cards = flattenCards(report.Groups)
	return report, report.Cards, nil
}

func loadStatusPositions(ctx context.Context, c *statusespkg.Client) (map[int]int, error) {
	result, err := c.List(ctx)
	if err != nil {
		return nil, err
	}
	positions := make(map[int]int, len(result.IssueStatuses))
	for _, status := range result.IssueStatuses {
		positions[status.ID] = status.Position
	}
	return positions, nil
}

func collectIssues(ctx context.Context, c *issuespkg.Client, projectID int) ([]issuespkg.Issue, error) {
	fetcher := func(innerCtx context.Context, offset, limit int) ([]issuespkg.Issue, int, error) {
		params := map[string]string{
			"project_id": strconv.Itoa(projectID),
			"status_id":  "*",
			"limit":      strconv.Itoa(limit),
			"offset":     strconv.Itoa(offset),
		}
		result, err := c.List(innerCtx, params)
		if err != nil {
			return nil, 0, err
		}
		return result.Issues, result.TotalCount, nil
	}
	return client.CollectAll(client.Paginate(ctx, fetcher, 100))
}

func loadIssueAgileData(ctx context.Context, c *Client, allIssues []issuespkg.Issue) (map[int]AgileData, error) {
	issueIDs := make([]int, 0, len(allIssues))
	for _, issue := range allIssues {
		issueIDs = append(issueIDs, issue.ID)
	}
	if len(issueIDs) == 0 {
		return map[int]AgileData{}, nil
	}

	results := client.BatchGetFunc(issueIDs, ctx, func(innerCtx context.Context, _ int, issueID int) (AgileData, error) {
		data, err := c.GetIssueAgileData(innerCtx, issueID)
		if err != nil {
			return AgileData{}, err
		}
		return *data, nil
	}, 5)

	agileDataByIssueID := make(map[int]AgileData, len(results))
	for _, result := range results {
		if result.Error != nil {
			return nil, result.Error
		}
		issueID := issueIDs[result.Index]
		agileDataByIssueID[issueID] = result.Result
	}
	return agileDataByIssueID, nil
}

func buildBoardCard(issue issuespkg.Issue, data AgileData, statusPositions map[int]int, currentSprint *Sprint) BoardCard {
	card := BoardCard{
		ID:          issue.ID,
		Subject:     issue.Subject,
		StoryPoints: data.StoryPoints,
		Position:    data.Position,
	}

	if issue.Status != nil {
		card.Status = issue.Status.Name
		card.StatusPosition = statusPositionOrDefault(statusPositions, issue.Status.ID)
	}
	if issue.AssignedTo != nil {
		card.Assignee = issue.AssignedTo.Name
	}
	if issue.Priority != nil {
		card.Priority = issue.Priority.Name
	}
	card.DueDate = issue.DueDate

	if data.AgileSprintID != nil {
		card.SprintID = *data.AgileSprintID
	}
	if currentSprint != nil && card.SprintID == currentSprint.ID {
		card.SprintName = currentSprint.Name
	}

	return card
}

func flattenCards(groups []BoardGroup) []BoardCard {
	cards := make([]BoardCard, 0)
	for _, group := range groups {
		cards = append(cards, group.Cards...)
	}
	return cards
}

func sortBoardCards(cards []BoardCard) {
	sort.SliceStable(cards, func(i, j int) bool {
		if cards[i].StatusPosition != cards[j].StatusPosition {
			return cards[i].StatusPosition < cards[j].StatusPosition
		}
		if cards[i].Position != cards[j].Position {
			return cards[i].Position < cards[j].Position
		}
		return cards[i].ID < cards[j].ID
	})
}

func selectCurrentSprint(sprints []Sprint) *Sprint {
	now := time.Now().UTC()

	for i := range sprints {
		if sprintMatchesCurrent(sprints[i], now) {
			return &sprints[i]
		}
	}

	for i := range sprints {
		if !sprints[i].IsClosed && !sprints[i].IsArchived {
			return &sprints[i]
		}
	}

	return nil
}

func sprintMatchesCurrent(sprint Sprint, now time.Time) bool {
	status := strings.ToLower(strings.TrimSpace(sprint.Status))
	if status == "active" || status == "current" {
		return true
	}

	if sprint.IsClosed || sprint.IsArchived {
		return false
	}

	start, err := time.Parse("2006-01-02", sprint.StartDate)
	if err == nil && !start.After(now) {
		end, endErr := time.Parse("2006-01-02", sprint.EndDate)
		if endErr == nil && !end.Before(now) {
			return true
		}
	}

	return sprint.IsDefault
}

func renderBoardReport(report *BoardReport) string {
	var b strings.Builder
	if report.Project != nil {
		fmt.Fprintf(&b, "Project: %s (%s)\n", report.Project.Name, report.Project.Identifier)
	}
	if report.CurrentSprint != nil {
		fmt.Fprintf(&b, "Current sprint: %s", report.CurrentSprint.Name)
		if report.CurrentSprint.StartDate != "" || report.CurrentSprint.EndDate != "" {
			fmt.Fprintf(&b, " [%s - %s]", renderSprintDate(report.CurrentSprint.StartDate), renderSprintDate(report.CurrentSprint.EndDate))
		}
		b.WriteString("\n")
	}

	for _, group := range report.Groups {
		b.WriteString("\n")
		b.WriteString(group.Name)
		b.WriteString("\n")
		if len(group.Cards) == 0 {
			b.WriteString("  (empty)\n")
			continue
		}
		for _, card := range group.Cards {
			fmt.Fprintf(&b, "- #%d %s | %s | %s | sp=%s | pri=%s | due=%s | pos=%d\n",
				card.ID,
				card.Subject,
				renderTextValue(card.Status),
				renderTextValue(card.Assignee),
				renderFloat(card.StoryPoints),
				renderTextValue(card.Priority),
				renderTextValue(card.DueDate),
				card.Position,
			)
		}
	}
	return b.String()
}

func renderSprintDate(value string) string {
	if value == "" {
		return "-"
	}
	return value
}

func renderTextValue(value string) string {
	if value == "" {
		return "-"
	}
	return value
}

func renderFloat(value float64) string {
	if value == 0 {
		return "-"
	}
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func statusPositionOrDefault(positions map[int]int, statusID int) int {
	if pos, ok := positions[statusID]; ok {
		return pos
	}
	return math.MaxInt
}
