package sprints

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/largeoliu/redmine-cli/internal/client"
	"github.com/largeoliu/redmine-cli/internal/resources/agile"
	"github.com/largeoliu/redmine-cli/internal/resources/projects"
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
	if cmd.Use != "sprint" {
		t.Fatalf("expected Use sprint, got %s", cmd.Use)
	}
	if len(cmd.Aliases) != 1 || cmd.Aliases[0] != "sprints" {
		t.Fatalf("expected alias sprints, got %v", cmd.Aliases)
	}
	if len(cmd.Commands()) != 2 {
		t.Fatalf("expected 2 subcommands, got %d", len(cmd.Commands()))
	}
	subcommandNames := make([]string, len(cmd.Commands()))
	for i, c := range cmd.Commands() {
		subcommandNames[i] = c.Name()
	}
	if subcommandNames[0] != "get" || subcommandNames[1] != "list" {
		t.Fatalf("expected [get, list], got %v", subcommandNames)
	}
}

func TestListCommand_Success(t *testing.T) {
	flags := &types.GlobalFlags{Format: "json"}
	var payload any
	c := client.NewClient("https://example.com", "test-key", client.WithHTTPClient(&http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/projects/42.json":
				return jsonHTTPResponse(t, http.StatusOK, map[string]any{
					"project": projects.Project{ID: 42, Name: "City", Identifier: "city"},
				}), nil
			case "/projects/42/agile_sprints.json":
				return jsonHTTPResponse(t, http.StatusOK, map[string]any{
					"project_id":   42,
					"project_name": "City",
					"sprints": []map[string]any{
						{
							"id":          7,
							"name":        "Sprint 7",
							"status":      "active",
							"start_date":  "2026-04-01",
							"end_date":    "2026-04-14",
							"is_default":  true,
							"is_closed":   false,
							"is_archived": false,
						},
						{
							"id":          8,
							"name":        "Sprint 8",
							"status":      "open",
							"start_date":  "2026-04-15",
							"end_date":    "2026-04-28",
							"is_default":  false,
							"is_closed":   false,
							"is_archived": false,
						},
					},
				}), nil
			default:
				t.Fatalf("unexpected path: %s", req.URL.Path)
				return nil, nil
			}
		}),
	}))
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) { return c, nil },
		writeOutputFunc: func(_ io.Writer, _ *types.GlobalFlags, p any) error {
			payload = p
			return nil
		},
	}

	cmd := newListCommand(flags, resolver)
	cmd.SetArgs([]string{"--project", "42"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sprints, ok := payload.([]agile.Sprint)
	if !ok {
		t.Fatalf("expected []agile.Sprint payload, got %T", payload)
	}
	if len(sprints) != 2 {
		t.Fatalf("expected 2 sprints, got %d", len(sprints))
	}
	if sprints[0].ID != 7 || sprints[1].ID != 8 {
		t.Fatalf("unexpected sprint ids: %+v", sprints)
	}
}

func TestGetCommand_Success(t *testing.T) {
	flags := &types.GlobalFlags{Format: "json"}
	var payload any
	c := client.NewClient("https://example.com", "test-key", client.WithHTTPClient(&http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/projects/42/agile_sprints/7.json":
				return jsonHTTPResponse(t, http.StatusOK, map[string]any{
					"agile_sprint": map[string]any{
						"id":              7,
						"project_id":      42,
						"name":            "Sprint 7",
						"status":          "active",
						"start_date":      "2026-04-01",
						"end_date":        "2026-04-14",
						"story_points":    68.0,
						"done_ratio":      100.0,
						"estimated_hours": 68.0,
						"spent_hours":     186.5,
					},
				}), nil
			default:
				t.Fatalf("unexpected path: %s", req.URL.Path)
				return nil, nil
			}
		}),
	}))
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) { return c, nil },
		writeOutputFunc: func(_ io.Writer, _ *types.GlobalFlags, p any) error {
			payload = p
			return nil
		},
	}

	cmd := newGetCommand(flags, resolver)
	cmd.SetArgs([]string{"--project-id", "42", "7"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sprint, ok := payload.(*agile.Sprint)
	if !ok {
		t.Fatalf("expected *agile.Sprint payload, got %T", payload)
	}
	if sprint.ID != 7 {
		t.Fatalf("expected sprint ID 7, got %d", sprint.ID)
	}
	if sprint.ProjectID != 42 {
		t.Fatalf("expected project_id 42, got %d", sprint.ProjectID)
	}
	if sprint.StoryPoints != 68.0 {
		t.Fatalf("expected story_points 68.0, got %f", sprint.StoryPoints)
	}
	if sprint.DoneRatio != 100.0 {
		t.Fatalf("expected done_ratio 100.0, got %f", sprint.DoneRatio)
	}
	if sprint.SpentHours != 186.5 {
		t.Fatalf("expected spent_hours 186.5, got %f", sprint.SpentHours)
	}
}

func TestGetCommand_MissingProjectID(t *testing.T) {
	flags := &types.GlobalFlags{Format: "json"}
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) { return nil, nil },
	}

	cmd := newGetCommand(flags, resolver)
	cmd.SetArgs([]string{"7"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when --project-id is missing")
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func jsonHTTPResponse(t *testing.T, status int, payload any) *http.Response {
	t.Helper()

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	return &http.Response{
		StatusCode: status,
		Status:     http.StatusText(status),
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(data)),
	}
}
