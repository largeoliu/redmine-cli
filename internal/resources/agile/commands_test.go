package agile

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/largeoliu/redmine-cli/internal/client"
	issuespkg "github.com/largeoliu/redmine-cli/internal/resources/issues"
	projectspkg "github.com/largeoliu/redmine-cli/internal/resources/projects"
	statusespkg "github.com/largeoliu/redmine-cli/internal/resources/statuses"
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
	for _, want := range []string{"Project: City (city)", "Current sprint: Sprint 7", "Sprint 7", "No Sprint", "#101", "#102"} {
		if !bytes.Contains(buf.Bytes(), []byte(want)) {
			t.Fatalf("expected output to contain %q, got %s", want, output)
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
	if len(report.Groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(report.Groups))
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
	if cards[0].ID != 101 || cards[1].ID != 102 {
		t.Fatalf("unexpected card order: %+v", cards)
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
		},
	})
	mock.HandleJSON("/projects/42/agile_sprints/7.json", map[string]any{
		"agile_sprint": Sprint{ID: 7, Name: "Sprint 7", Status: "active", StartDate: "2026-04-01", EndDate: "2026-04-14"},
	})
	mock.HandleJSON("/issue_statuses.json", statusespkg.IssueStatusList{
		IssueStatuses: []statusespkg.IssueStatus{{ID: 1, Name: "New", Position: 1}},
	})
	mock.HandleJSON("/issues.json", issuespkg.IssueList{
		Issues: []issuespkg.Issue{
			{
				ID:         101,
				Subject:    "Current sprint issue",
				Status:     &issuespkg.Reference{ID: 1, Name: "New"},
				Priority:   &issuespkg.Reference{ID: 1, Name: "Normal"},
				AssignedTo: &issuespkg.Reference{ID: 5, Name: "Alice"},
				DueDate:    "2026-04-10",
			},
			{
				ID:         102,
				Subject:    "No sprint issue",
				Status:     &issuespkg.Reference{ID: 1, Name: "New"},
				Priority:   &issuespkg.Reference{ID: 1, Name: "Normal"},
				AssignedTo: &issuespkg.Reference{ID: 6, Name: "Bob"},
				DueDate:    "2026-04-11",
			},
		},
		TotalCount: 2,
		Limit:      100,
		Offset:     0,
	})
	mock.HandleJSON("/issues/101/agile_data.json", map[string]any{
		"agile_data": AgileData{AgileSprintID: intPtr(7), StoryPoints: 5, Position: 2},
	})
	mock.HandleJSON("/issues/102/agile_data.json", map[string]any{
		"agile_data": AgileData{StoryPoints: 1, Position: 1},
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
