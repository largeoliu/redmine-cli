package sprints

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/largeoliu/redmine-cli/internal/client"
	"github.com/largeoliu/redmine-cli/internal/output"
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
	if len(cmd.Commands()) != 1 || cmd.Commands()[0].Name() != "list" {
		t.Fatalf("expected list subcommand, got %v", cmd.Commands())
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
	cmd.SetArgs([]string{"42"})

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

func newSprintDetailsClient(t *testing.T) *client.Client {
	t.Helper()
	return client.NewClient("https://example.com", "test-key", client.WithHTTPClient(&http.Client{
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
						{"id": 7, "name": "Sprint 7", "status": "active", "description": "Release hardening", "start_date": "2026-04-01", "end_date": "2026-04-14"},
						{"id": 8, "name": "Sprint 8", "status": "open", "description": "Stabilization", "start_date": "2026-04-15", "end_date": "2026-04-28"},
					},
				}), nil
			default:
				t.Fatalf("unexpected path: %s", req.URL.Path)
				return nil, nil
			}
		}),
	}))
}

func TestListCommand_DetailsExpandsSprintPayload(t *testing.T) {
	flags := &types.GlobalFlags{Format: "json"}
	var payload any
	c := newSprintDetailsClient(t)
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) { return c, nil },
		writeOutputFunc: func(_ io.Writer, _ *types.GlobalFlags, p any) error {
			payload = p
			return nil
		},
	}

	cmd := newListCommand(flags, resolver)
	cmd.SetArgs([]string{"--details", "42"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sprints, ok := payload.([]agile.Sprint)
	if !ok {
		t.Fatalf("expected []agile.Sprint payload, got %T", payload)
	}
	if sprints[0].Description != "Release hardening" {
		t.Fatalf("expected expanded description, got %+v", sprints[0])
	}
}

func TestListCommand_DetailsRendersTablePayload(t *testing.T) {
	flags := &types.GlobalFlags{Format: "table"}
	var rendered bytes.Buffer
	c := newSprintDetailsClient(t)
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) { return c, nil },
		writeOutputFunc: func(_ io.Writer, _ *types.GlobalFlags, p any) error {
			return output.Write(&rendered, output.FormatTable, p)
		},
	}

	cmd := newListCommand(flags, resolver)
	cmd.SetArgs([]string{"--details", "42"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(rendered.String(), "Release hardening") {
		t.Fatalf("expected table rendering to include the description, got:\n%s", rendered.String())
	}
}
