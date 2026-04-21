package agile

import (
	"context"
	"testing"

	"github.com/largeoliu/redmine-cli/internal/client"
	"github.com/largeoliu/redmine-cli/internal/resources/issues"
	"github.com/largeoliu/redmine-cli/internal/resources/projects"
	"github.com/largeoliu/redmine-cli/internal/resources/statuses"
	"github.com/largeoliu/redmine-cli/internal/testutil"
)

func TestBuildBoardReport(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.HandleJSON("/projects/42/agile_sprints.json", SprintList{
		AgileSprints: []Sprint{
			{ID: 7, Name: "Sprint 7", Status: "active", StartDate: "2026-04-01", EndDate: "2026-04-14"},
			{ID: 8, Name: "Sprint 8", Status: "open", StartDate: "2026-04-15", EndDate: "2026-04-28"},
		},
	})
	mock.HandleJSON("/projects/42/agile_sprints/7.json", struct {
		AgileSprint Sprint `json:"agile_sprint"`
	}{
		AgileSprint: Sprint{ID: 7, Name: "Sprint 7", Status: "active", StartDate: "2026-04-01", EndDate: "2026-04-14"},
	})
	mock.HandleJSON("/issue_statuses.json", statuses.IssueStatusList{
		IssueStatuses: []statuses.IssueStatus{
			{ID: 1, Name: "New", Position: 1},
			{ID: 2, Name: "In Progress", Position: 2},
		},
	})
	mock.HandleJSON("/issues.json", issues.IssueList{
		Issues: []issues.Issue{
			{
				ID:      101,
				Subject: "Current sprint low status",
				Status:  &issues.Reference{ID: 1, Name: "New"},
				Priority: &issues.Reference{
					ID:   2,
					Name: "Normal",
				},
				AssignedTo: &issues.Reference{ID: 5, Name: "Alice"},
				DueDate:    "2026-04-10",
			},
			{
				ID:      102,
				Subject: "Current sprint high status",
				Status:  &issues.Reference{ID: 1, Name: "New"},
				Priority: &issues.Reference{
					ID:   3,
					Name: "High",
				},
				AssignedTo: &issues.Reference{ID: 6, Name: "Bob"},
				DueDate:    "2026-04-12",
			},
			{
				ID:      103,
				Subject: "Current sprint later status",
				Status:  &issues.Reference{ID: 2, Name: "In Progress"},
				Priority: &issues.Reference{
					ID:   2,
					Name: "Normal",
				},
				AssignedTo: &issues.Reference{ID: 7, Name: "Carol"},
				DueDate:    "2026-04-13",
			},
			{
				ID:      104,
				Subject: "No sprint issue",
				Status:  &issues.Reference{ID: 1, Name: "New"},
				Priority: &issues.Reference{
					ID:   1,
					Name: "Low",
				},
				AssignedTo: &issues.Reference{ID: 8, Name: "Dora"},
				DueDate:    "2026-04-14",
			},
			{
				ID:      105,
				Subject: "Other sprint issue",
				Status:  &issues.Reference{ID: 1, Name: "New"},
				Priority: &issues.Reference{
					ID:   1,
					Name: "Low",
				},
				AssignedTo: &issues.Reference{ID: 9, Name: "Eve"},
				DueDate:    "2026-04-15",
			},
		},
		TotalCount: 5,
		Limit:      100,
		Offset:     0,
	})
	mock.HandleJSON("/issues/101/agile_data.json", struct {
		AgileData AgileData `json:"agile_data"`
	}{AgileData: AgileData{AgileSprintID: intPtr(7), StoryPoints: 3, Position: 4}})
	mock.HandleJSON("/issues/102/agile_data.json", struct {
		AgileData AgileData `json:"agile_data"`
	}{AgileData: AgileData{AgileSprintID: intPtr(7), StoryPoints: 5, Position: 1}})
	mock.HandleJSON("/issues/103/agile_data.json", struct {
		AgileData AgileData `json:"agile_data"`
	}{AgileData: AgileData{AgileSprintID: intPtr(7), StoryPoints: 8, Position: 9}})
	mock.HandleJSON("/issues/104/agile_data.json", struct {
		AgileData AgileData `json:"agile_data"`
	}{AgileData: AgileData{StoryPoints: 2, Position: 2}})
	mock.HandleJSON("/issues/105/agile_data.json", struct {
		AgileData AgileData `json:"agile_data"`
	}{AgileData: AgileData{AgileSprintID: intPtr(8), StoryPoints: 1, Position: 1}})

	c := client.NewClient(mock.URL, "test-key")
	project := &projects.Project{ID: 42, Name: "City", Identifier: "city"}

	report, cards, err := buildBoardReport(context.Background(), c, project)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.CurrentSprint == nil {
		t.Fatal("expected current sprint, got nil")
	}
	if report.CurrentSprint.ID != 7 {
		t.Fatalf("expected current sprint 7, got %d", report.CurrentSprint.ID)
	}
	if len(report.Groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(report.Groups))
	}
	if report.Groups[0].Name != "Sprint 7" {
		t.Fatalf("expected first group to be Sprint 7, got %s", report.Groups[0].Name)
	}
	if len(report.Groups[0].Cards) != 3 {
		t.Fatalf("expected 3 current sprint cards, got %d", len(report.Groups[0].Cards))
	}
	if len(cards) != 3 {
		t.Fatalf("expected 3 flattened cards, got %d", len(cards))
	}

	wantIDs := []int{102, 101, 103}
	for i, card := range cards {
		if card.ID != wantIDs[i] {
			t.Fatalf("expected card[%d] to be %d, got %d", i, wantIDs[i], card.ID)
		}
	}
}

func TestBuildBoardReportSelectedSprint(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.HandleJSON("/projects/42/agile_sprints.json", SprintList{
		AgileSprints: []Sprint{
			{ID: 7, Name: "Sprint 7", Status: "active", StartDate: "2026-04-01", EndDate: "2026-04-14"},
			{ID: 8, Name: "Sprint 8", Status: "open", StartDate: "2026-04-15", EndDate: "2026-04-28"},
		},
	})
	mock.HandleJSON("/projects/42/agile_sprints/8.json", struct {
		AgileSprint Sprint `json:"agile_sprint"`
	}{
		AgileSprint: Sprint{ID: 8, Name: "Sprint 8", Status: "open", StartDate: "2026-04-15", EndDate: "2026-04-28"},
	})
	mock.HandleJSON("/issue_statuses.json", statuses.IssueStatusList{
		IssueStatuses: []statuses.IssueStatus{
			{ID: 1, Name: "New", Position: 1},
			{ID: 2, Name: "In Progress", Position: 2},
		},
	})
	mock.HandleJSON("/issues.json", issues.IssueList{
		Issues: []issues.Issue{
			{
				ID:      101,
				Subject: "Current sprint issue",
				Status:  &issues.Reference{ID: 1, Name: "New"},
				Priority: &issues.Reference{
					ID:   2,
					Name: "Normal",
				},
			},
			{
				ID:      105,
				Subject: "Other sprint issue",
				Status:  &issues.Reference{ID: 1, Name: "New"},
				Priority: &issues.Reference{
					ID:   1,
					Name: "Low",
				},
			},
		},
		TotalCount: 2,
		Limit:      100,
		Offset:     0,
	})
	mock.HandleJSON("/issues/101/agile_data.json", struct {
		AgileData AgileData `json:"agile_data"`
	}{AgileData: AgileData{AgileSprintID: intPtr(7), StoryPoints: 3, Position: 4}})
	mock.HandleJSON("/issues/105/agile_data.json", struct {
		AgileData AgileData `json:"agile_data"`
	}{AgileData: AgileData{AgileSprintID: intPtr(8), StoryPoints: 1, Position: 1}})

	c := client.NewClient(mock.URL, "test-key")
	project := &projects.Project{ID: 42, Name: "City", Identifier: "city"}

	report, cards, err := buildBoardReportWithOptions(context.Background(), c, project, boardOptions{
		Sprint:  "8",
		Tracker: "全部",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.CurrentSprint == nil || report.CurrentSprint.ID != 8 {
		t.Fatalf("expected sprint 8, got %+v", report.CurrentSprint)
	}
	if len(report.Groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(report.Groups))
	}
	if report.Groups[0].Name != "Sprint 8" {
		t.Fatalf("expected group Sprint 8, got %s", report.Groups[0].Name)
	}
	if len(report.Groups[0].Cards) != 1 {
		t.Fatalf("expected 1 sprint 8 card, got %d", len(report.Groups[0].Cards))
	}
	if len(cards) != 1 || cards[0].ID != 105 {
		t.Fatalf("expected sprint 8 card 105, got %+v", cards)
	}
}
