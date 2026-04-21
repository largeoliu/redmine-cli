package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/largeoliu/redmine-cli/internal/client"
	"github.com/largeoliu/redmine-cli/internal/resources/categories"
	"github.com/largeoliu/redmine-cli/internal/resources/priorities"
	"github.com/largeoliu/redmine-cli/internal/resources/trackers"
	"github.com/largeoliu/redmine-cli/internal/resources/users"
	"github.com/largeoliu/redmine-cli/internal/resources/versions"
)

func newMockClient(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *client.Client) {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	return server, client.NewClient(server.URL, "test-api-key")
}

func TestCategoriesListAndGet(t *testing.T) {
	projectID := 42
	_, cli := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Redmine-API-Key"); got != "test-api-key" {
			t.Errorf("expected API key header, got %q", got)
		}
		switch r.URL.Path {
		case "/projects/42/issue_categories.json":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"issue_categories": []map[string]any{
					{"id": 1, "name": "Bug"},
					{"id": 2, "name": "Feature"},
				},
			})
		case "/issue_categories/1.json":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"issue_category": map[string]any{"id": 1, "name": "Bug"},
			})
		default:
			http.NotFound(w, r)
		}
	})

	catClient := categories.NewClient(cli)
	list, err := catClient.List(context.Background(), projectID)
	if err != nil {
		t.Fatalf("list categories: %v", err)
	}
	if len(list.IssueCategories) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(list.IssueCategories))
	}

	got, err := catClient.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("get category: %v", err)
	}
	if got.Name != "Bug" {
		t.Fatalf("expected category name Bug, got %q", got.Name)
	}
}

func TestVersionsListAndGet(t *testing.T) {
	_, cli := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/projects/42/versions.json":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"versions": []map[string]any{
					{"id": 1, "name": "v1.0", "status": "open"},
					{"id": 2, "name": "v2.0", "status": "locked"},
				},
				"total_count": 2,
			})
		case "/versions/1.json":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"version": map[string]any{"id": 1, "name": "v1.0", "status": "open"},
			})
		default:
			http.NotFound(w, r)
		}
	})

	versionClient := versions.NewClient(cli)
	list, err := versionClient.List(context.Background(), 42, nil)
	if err != nil {
		t.Fatalf("list versions: %v", err)
	}
	if len(list.Versions) != 2 {
		t.Fatalf("expected 2 versions, got %d", len(list.Versions))
	}

	got, err := versionClient.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("get version: %v", err)
	}
	if got.Name != "v1.0" {
		t.Fatalf("expected version name v1.0, got %q", got.Name)
	}
}

func TestTrackersListAndGet(t *testing.T) {
	_, cli := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/trackers.json":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"trackers": []map[string]any{
					{"id": 1, "name": "Bug"},
					{"id": 2, "name": "Feature"},
				},
			})
		case "/trackers/1.json":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"tracker": map[string]any{
					"id":   1,
					"name": "Bug",
					"custom_fields": []map[string]any{
						{"id": 5, "name": "Affected Version", "field_format": "list"},
					},
				},
			})
		default:
			http.NotFound(w, r)
		}
	})

	trackerClient := trackers.NewClient(cli)
	list, err := trackerClient.List(context.Background())
	if err != nil {
		t.Fatalf("list trackers: %v", err)
	}
	if len(list.Trackers) != 2 {
		t.Fatalf("expected 2 trackers, got %d", len(list.Trackers))
	}

	got, err := trackerClient.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("get tracker: %v", err)
	}
	if got.Name != "Bug" {
		t.Fatalf("expected tracker name Bug, got %q", got.Name)
	}
	if len(got.CustomFields) != 1 || got.CustomFields[0].Name != "Affected Version" {
		t.Fatalf("expected custom field coverage, got %+v", got.CustomFields)
	}
}

func TestPrioritiesList(t *testing.T) {
	_, cli := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/enumerations/issue_priorities.json" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"issue_priorities": []map[string]any{
				{"id": 1, "name": "Normal", "is_default": true, "position": 2},
				{"id": 2, "name": "High", "is_default": false, "position": 3},
			},
		})
	})

	priorityClient := priorities.NewClient(cli)
	list, err := priorityClient.List(context.Background())
	if err != nil {
		t.Fatalf("list priorities: %v", err)
	}
	if len(list.Priorities) != 2 {
		t.Fatalf("expected 2 priorities, got %d", len(list.Priorities))
	}
	if !list.Priorities[0].IsDefault {
		t.Fatalf("expected first priority to be default, got %+v", list.Priorities[0])
	}
}

func TestUsersListAndGet(t *testing.T) {
	_, cli := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/users.json":
			if r.URL.Query().Get("limit") != "1" {
				t.Errorf("expected limit=1, got %q", r.URL.Query().Get("limit"))
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"users": []map[string]any{
					{"id": 1, "login": "admin", "firstname": "Admin", "lastname": "User", "admin": true},
				},
				"total_count": 1,
			})
		case "/users/1.json":
			if r.URL.Query().Get("include") != "memberships,groups" {
				t.Errorf("expected include query, got %q", r.URL.Query().Get("include"))
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"user": map[string]any{
					"id":        1,
					"login":     "admin",
					"firstname": "Admin",
					"lastname":  "User",
					"admin":     true,
				},
			})
		default:
			http.NotFound(w, r)
		}
	})

	userClient := users.NewClient(cli)
	list, err := userClient.List(context.Background(), users.BuildListParams(users.ListFlags{Limit: 1}))
	if err != nil {
		t.Fatalf("list users: %v", err)
	}
	if len(list.Users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(list.Users))
	}

	got, err := userClient.Get(context.Background(), list.Users[0].ID, map[string]string{"include": "memberships,groups"})
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	if got.Login != "admin" {
		t.Fatalf("expected user login admin, got %q", got.Login)
	}
}
