package agile

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/largeoliu/redmine-cli/internal/client"
	"github.com/largeoliu/redmine-cli/internal/errors"
	issuespkg "github.com/largeoliu/redmine-cli/internal/resources/issues"
	projectspkg "github.com/largeoliu/redmine-cli/internal/resources/projects"
	statusespkg "github.com/largeoliu/redmine-cli/internal/resources/statuses"
	"github.com/largeoliu/redmine-cli/internal/resources/trackers"
	"github.com/largeoliu/redmine-cli/internal/testutil"
	"github.com/largeoliu/redmine-cli/internal/types"
)

type mockResolver struct {
	resolveClientFunc func(flags *types.GlobalFlags) (*client.Client, error)
	writeOutputFunc   func(w io.Writer, flags *types.GlobalFlags, payload any) error
}

func (m *mockResolver) ResolveClient(flags *types.GlobalFlags) (*client.Client, error) {
	if m.resolveClientFunc != nil {
		return m.resolveClientFunc(flags)
	}
	return nil, nil
}

func (m *mockResolver) WriteOutput(w io.Writer, flags *types.GlobalFlags, payload any) error {
	if m.writeOutputFunc != nil {
		return m.writeOutputFunc(w, flags, payload)
	}
	return nil
}

func TestNewCommand(t *testing.T) {
	cmd := NewCommand(&types.GlobalFlags{}, &mockResolver{})
	if cmd == nil {
		t.Fatal("expected command, got nil")
	}
	if cmd.Use != "agile" {
		t.Fatalf("expected Use agile, got %s", cmd.Use)
	}
	if len(cmd.Commands()) != 1 || cmd.Commands()[0].Name() != "board" {
		t.Fatalf("expected board subcommand, got %v", cmd.Commands())
	}
}

func TestBoardCommandRawOutput(t *testing.T) {
	mock := setupBoardMock(t)
	defer mock.Close()

	flags := &types.GlobalFlags{Format: "raw"}
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) {
			return client.NewClient(mock.URL, "test-key"), nil
		},
	}

	cmd := newBoardCommand(flags, resolver)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"42"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	for _, want := range []string{"Project: City (city)", "Sprint: Sprint 7", "Sprint 7", "#101", "#102"} {
		if !bytes.Contains(buf.Bytes(), []byte(want)) {
			t.Fatalf("expected output to contain %q, got %s", want, output)
		}
	}
	for _, want := range []string{"No Sprint", "#103", "#104"} {
		if bytes.Contains(buf.Bytes(), []byte(want)) {
			t.Fatalf("expected output to omit %q, got %s", want, output)
		}
	}
}

func TestBoardCommandJSONOutput(t *testing.T) {
	mock := setupBoardMock(t)
	defer mock.Close()

	flags := &types.GlobalFlags{Format: "json"}
	var payload any
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) {
			return client.NewClient(mock.URL, "test-key"), nil
		},
		writeOutputFunc: func(_ io.Writer, _ *types.GlobalFlags, p any) error {
			payload = p
			return nil
		},
	}

	cmd := newBoardCommand(flags, resolver)
	cmd.SetArgs([]string{"city"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	report, ok := payload.(*BoardReport)
	if !ok {
		t.Fatalf("expected BoardReport payload, got %T", payload)
	}
	if report.CurrentSprint == nil || report.CurrentSprint.ID != 7 {
		t.Fatalf("expected current sprint 7, got %+v", report.CurrentSprint)
	}
	if len(report.Groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(report.Groups))
	}
	if len(report.Groups[0].Cards) != 2 {
		t.Fatalf("expected 2 cards in current sprint group, got %d", len(report.Groups[0].Cards))
	}
}

func TestBoardCommandTableOutput(t *testing.T) {
	mock := setupBoardMock(t)
	defer mock.Close()

	flags := &types.GlobalFlags{Format: "table"}
	var payload any
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) {
			return client.NewClient(mock.URL, "test-key"), nil
		},
		writeOutputFunc: func(_ io.Writer, _ *types.GlobalFlags, p any) error {
			payload = p
			return nil
		},
	}

	cmd := newBoardCommand(flags, resolver)
	cmd.SetArgs([]string{"city"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cards, ok := payload.([]BoardCard)
	if !ok {
		t.Fatalf("expected []BoardCard payload, got %T", payload)
	}
	if len(cards) != 2 {
		t.Fatalf("expected 2 cards, got %d", len(cards))
	}
	if cards[0].ID != 102 || cards[1].ID != 101 {
		t.Fatalf("unexpected card order: %+v", cards)
	}
}

func TestBoardCommandTrackerFilter(t *testing.T) {
	mock := setupBoardTrackerMock(t)
	defer mock.Close()

	flags := &types.GlobalFlags{Format: "json"}
	var payload any
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) {
			return client.NewClient(mock.URL, "test-key"), nil
		},
		writeOutputFunc: func(_ io.Writer, _ *types.GlobalFlags, p any) error {
			payload = p
			return nil
		},
	}

	cmd := newBoardCommand(flags, resolver)
	cmd.SetArgs([]string{"--tracker", "需求", "42"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	report, ok := payload.(*BoardReport)
	if !ok {
		t.Fatalf("expected BoardReport payload, got %T", payload)
	}
	if report.CurrentSprint == nil || report.CurrentSprint.ID != 7 {
		t.Fatalf("expected current sprint 7, got %+v", report.CurrentSprint)
	}
	if len(report.Groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(report.Groups))
	}
	if len(report.Groups[0].Cards) != 1 {
		t.Fatalf("expected 1 card in sprint group, got %d", len(report.Groups[0].Cards))
	}
	if report.Groups[0].Cards[0].ID != 101 {
		t.Fatalf("expected tracker-filtered card #101, got %+v", report.Groups[0].Cards[0])
	}
}

func TestBoardCommandTrackerAllSkipsLookup(t *testing.T) {
	mock := setupBoardMock(t)
	defer mock.Close()

	flags := &types.GlobalFlags{Format: "json"}
	var payload any
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) {
			return client.NewClient(mock.URL, "test-key"), nil
		},
		writeOutputFunc: func(_ io.Writer, _ *types.GlobalFlags, p any) error {
			payload = p
			return nil
		},
	}

	cmd := newBoardCommand(flags, resolver)
	cmd.SetArgs([]string{"--tracker", "全部", "42"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	report, ok := payload.(*BoardReport)
	if !ok {
		t.Fatalf("expected BoardReport payload, got %T", payload)
	}
	if len(report.Groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(report.Groups))
	}
	if len(report.Groups[0].Cards) != 2 {
		t.Fatalf("expected 2 cards with tracker=全部, got %d", len(report.Groups[0].Cards))
	}
}

func TestBoardCommandSprintSelection(t *testing.T) {
	mock := setupBoardMock(t)
	defer mock.Close()

	flags := &types.GlobalFlags{Format: "json"}
	var payload any
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) {
			return client.NewClient(mock.URL, "test-key"), nil
		},
		writeOutputFunc: func(_ io.Writer, _ *types.GlobalFlags, p any) error {
			payload = p
			return nil
		},
	}

	cmd := newBoardCommand(flags, resolver)
	cmd.SetArgs([]string{"--sprint", "8", "42"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	report, ok := payload.(*BoardReport)
	if !ok {
		t.Fatalf("expected BoardReport payload, got %T", payload)
	}
	if report.CurrentSprint == nil || report.CurrentSprint.ID != 8 {
		t.Fatalf("expected sprint 8, got %+v", report.CurrentSprint)
	}
	if len(report.Groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(report.Groups))
	}
	if len(report.Groups[0].Cards) != 1 {
		t.Fatalf("expected 1 sprint 8 card, got %d", len(report.Groups[0].Cards))
	}
	if report.Groups[0].Cards[0].ID != 104 {
		t.Fatalf("expected sprint 8 card #104, got %+v", report.Groups[0].Cards[0])
	}
}

func setupBoardMock(t *testing.T) *testutil.MockServer {
	t.Helper()

	mock := testutil.NewMockServer(t)

	mock.HandleJSON("/projects/42.json", map[string]any{
		"project": projectspkg.Project{ID: 42, Name: "City", Identifier: "city"},
	})
	mock.HandleJSON("/projects/city.json", map[string]any{
		"project": projectspkg.Project{ID: 42, Name: "City", Identifier: "city"},
	})
	mock.HandleJSON("/projects/42/agile_sprints.json", SprintList{
		AgileSprints: []Sprint{
			{ID: 7, Name: "Sprint 7", Status: "active", StartDate: "2026-04-01", EndDate: "2026-04-14"},
			{ID: 8, Name: "Sprint 8", Status: "open", StartDate: "2026-04-15", EndDate: "2026-04-28"},
		},
	})
	mock.HandleJSON("/projects/42/agile_sprints/7.json", map[string]any{
		"agile_sprint": Sprint{ID: 7, Name: "Sprint 7", Status: "active", StartDate: "2026-04-01", EndDate: "2026-04-14"},
	})
	mock.HandleJSON("/projects/42/agile_sprints/8.json", map[string]any{
		"agile_sprint": Sprint{ID: 8, Name: "Sprint 8", Status: "open", StartDate: "2026-04-15", EndDate: "2026-04-28"},
	})
	mock.HandleJSON("/issue_statuses.json", statusespkg.IssueStatusList{
		IssueStatuses: []statusespkg.IssueStatus{
			{ID: 1, Name: "New", Position: 1},
			{ID: 2, Name: "In Progress", Position: 2},
		},
	})
	mock.HandleJSON("/issues.json", issuespkg.IssueList{
		Issues: []issuespkg.Issue{
			{
				ID:         101,
				Subject:    "Current sprint issue A",
				Tracker:    &issuespkg.Reference{ID: 1, Name: "需求"},
				Status:     &issuespkg.Reference{ID: 1, Name: "New"},
				Priority:   &issuespkg.Reference{ID: 1, Name: "Normal"},
				AssignedTo: &issuespkg.Reference{ID: 5, Name: "Alice"},
				DueDate:    "2026-04-10",
			},
			{
				ID:         102,
				Subject:    "Current sprint issue B",
				Tracker:    &issuespkg.Reference{ID: 2, Name: "缺陷"},
				Status:     &issuespkg.Reference{ID: 1, Name: "New"},
				Priority:   &issuespkg.Reference{ID: 1, Name: "Normal"},
				AssignedTo: &issuespkg.Reference{ID: 6, Name: "Bob"},
				DueDate:    "2026-04-11",
			},
			{
				ID:         103,
				Subject:    "No sprint issue",
				Tracker:    &issuespkg.Reference{ID: 1, Name: "需求"},
				Status:     &issuespkg.Reference{ID: 1, Name: "New"},
				Priority:   &issuespkg.Reference{ID: 1, Name: "Normal"},
				AssignedTo: &issuespkg.Reference{ID: 7, Name: "Carol"},
				DueDate:    "2026-04-12",
			},
			{
				ID:         104,
				Subject:    "Other sprint issue",
				Tracker:    &issuespkg.Reference{ID: 1, Name: "需求"},
				Status:     &issuespkg.Reference{ID: 2, Name: "In Progress"},
				Priority:   &issuespkg.Reference{ID: 2, Name: "High"},
				AssignedTo: &issuespkg.Reference{ID: 8, Name: "Dora"},
				DueDate:    "2026-04-13",
			},
		},
		TotalCount: 4,
		Limit:      100,
		Offset:     0,
	})
	mock.HandleJSON("/issues/101/agile_data.json", map[string]any{
		"agile_data": AgileData{AgileSprintID: intPtr(7), StoryPoints: 5, Position: 2},
	})
	mock.HandleJSON("/issues/102/agile_data.json", map[string]any{
		"agile_data": AgileData{AgileSprintID: intPtr(7), StoryPoints: 1, Position: 1},
	})
	mock.HandleJSON("/issues/103/agile_data.json", map[string]any{
		"agile_data": AgileData{StoryPoints: 1, Position: 3},
	})
	mock.HandleJSON("/issues/104/agile_data.json", map[string]any{
		"agile_data": AgileData{AgileSprintID: intPtr(8), StoryPoints: 8, Position: 1},
	})

	return mock
}

func setupBoardTrackerMock(t *testing.T) *testutil.MockServer {
	t.Helper()

	mock := testutil.NewMockServer(t)

	mock.HandleJSON("/trackers.json", trackers.TrackerList{
		Trackers: []trackers.Tracker{
			{ID: 1, Name: "需求"},
			{ID: 2, Name: "缺陷"},
		},
	})
	mock.HandleJSON("/projects/42.json", map[string]any{
		"project": projectspkg.Project{ID: 42, Name: "City", Identifier: "city"},
	})
	mock.HandleJSON("/projects/42/agile_sprints.json", SprintList{
		AgileSprints: []Sprint{
			{ID: 7, Name: "Sprint 7", Status: "active", StartDate: "2026-04-01", EndDate: "2026-04-14"},
		},
	})
	mock.HandleJSON("/projects/42/agile_sprints/7.json", map[string]any{
		"agile_sprint": Sprint{ID: 7, Name: "Sprint 7", Status: "active", StartDate: "2026-04-01", EndDate: "2026-04-14"},
	})
	mock.HandleJSON("/issue_statuses.json", statusespkg.IssueStatusList{
		IssueStatuses: []statusespkg.IssueStatus{{ID: 1, Name: "New", Position: 1}},
	})
	mock.Handle("/issues.json", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("tracker_id"); got != "1" {
			t.Fatalf("expected tracker_id=1, got %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(issuespkg.IssueList{
			Issues: []issuespkg.Issue{
				{
					ID:         101,
					Subject:    "Current sprint issue A",
					Tracker:    &issuespkg.Reference{ID: 1, Name: "需求"},
					Status:     &issuespkg.Reference{ID: 1, Name: "New"},
					Priority:   &issuespkg.Reference{ID: 1, Name: "Normal"},
					AssignedTo: &issuespkg.Reference{ID: 5, Name: "Alice"},
					DueDate:    "2026-04-10",
				},
				{
					ID:         103,
					Subject:    "No sprint issue",
					Tracker:    &issuespkg.Reference{ID: 1, Name: "需求"},
					Status:     &issuespkg.Reference{ID: 1, Name: "New"},
					Priority:   &issuespkg.Reference{ID: 1, Name: "Normal"},
					AssignedTo: &issuespkg.Reference{ID: 7, Name: "Carol"},
					DueDate:    "2026-04-12",
				},
				{
					ID:         104,
					Subject:    "Other sprint issue",
					Tracker:    &issuespkg.Reference{ID: 1, Name: "需求"},
					Status:     &issuespkg.Reference{ID: 1, Name: "New"},
					Priority:   &issuespkg.Reference{ID: 1, Name: "Normal"},
					AssignedTo: &issuespkg.Reference{ID: 8, Name: "Dora"},
					DueDate:    "2026-04-13",
				},
			},
			TotalCount: 3,
			Limit:      100,
			Offset:     0,
		})
	})
	mock.HandleJSON("/issues/101/agile_data.json", map[string]any{
		"agile_data": AgileData{AgileSprintID: intPtr(7), StoryPoints: 5, Position: 2},
	})
	mock.HandleJSON("/issues/103/agile_data.json", map[string]any{
		"agile_data": AgileData{StoryPoints: 1, Position: 3},
	})
	mock.HandleJSON("/issues/104/agile_data.json", map[string]any{
		"agile_data": AgileData{AgileSprintID: intPtr(8), StoryPoints: 8, Position: 1},
	})

	return mock
}

func TestResolveProjectByIdentifier(t *testing.T) {
	mock := setupBoardMock(t)
	defer mock.Close()

	c := client.NewClient(mock.URL, "test-key")
	project, err := resolveProject(context.Background(), c, "city")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if project.ID != 42 {
		t.Fatalf("expected project 42, got %d", project.ID)
	}
}

func TestResolveProjectNotFound(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.HandleError("/projects/city.json", http.StatusNotFound, "Not found")

	c := client.NewClient(mock.URL, "test-key")
	_, err := resolveProject(context.Background(), c, "city")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var appErr *errors.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("expected *errors.Error, got %T", err)
	}
	if appErr.Category != errors.CategoryValidation {
		t.Fatalf("expected validation error, got %s", appErr.Category)
	}
	if appErr.Message != "project not found: city" {
		t.Fatalf("expected message %q, got %q", "project not found: city", appErr.Message)
	}
	if appErr.Hint == "" {
		t.Fatal("expected hint, got empty")
	}
	if len(appErr.Actions) == 0 {
		t.Fatal("expected actions, got none")
	}
}
