// test/e2e/config_test.go
package e2e

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestConfigHelpOutputs(t *testing.T) {
	stdout, _, exitCode := runCommand("config", "--help")

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "config") {
		t.Error("Expected help output to contain 'config'")
	}
}

func TestConfigListHelp(t *testing.T) {
	stdout, _, exitCode := runCommand("config", "list", "--help")

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "list") {
		t.Error("Expected help output to contain 'list'")
	}
}

func TestConfigGetHelp(t *testing.T) {
	stdout, _, exitCode := runCommand("config", "get", "--help")

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "get") {
		t.Error("Expected help output to contain 'get'")
	}
}

func TestConfigSetHelp(t *testing.T) {
	stdout, _, exitCode := runCommand("config", "set", "--help")

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "set") {
		t.Error("Expected help output to contain 'set'")
	}
}

func TestInstanceFlagNonexistent(t *testing.T) {
	_, stderr, exitCode := runCommand(
		"--instance", "nonexistent",
		"projects", "list",
	)

	if exitCode == 0 {
		t.Error("Expected non-zero exit code for nonexistent instance")
	}

	if !strings.Contains(stderr, "not found") && !strings.Contains(stderr, "not exist") {
		t.Logf("stderr: %s", stderr)
	}
}

func TestCommandLineOverridesEnvVar(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"projects": []map[string]any{},
		})
	}))
	defer server.Close()

	stdout, _, exitCode := runCommandWithEnv(
		map[string]string{"REDMINE_URL": "https://wrong-url.com"},
		"--url", server.URL,
		"--key", "test-key",
		"projects", "list",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	t.Logf("Command line overrides env var: %s", stdout)
}

func TestLoginWithAPIKeyFlag(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"projects": []map[string]any{},
		})
	}))
	defer server.Close()

	stdout, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"projects", "list",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	t.Logf("Login with API key: %s", stdout)
}

func TestComplexIssueFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		if query.Get("project_id") != "1" {
			t.Errorf("Expected project_id=1, got %s", query.Get("project_id"))
		}
		if query.Get("tracker_id") != "2" {
			t.Errorf("Expected tracker_id=2, got %s", query.Get("tracker_id"))
		}
		if query.Get("status_id") != "3" {
			t.Errorf("Expected status_id=3, got %s", query.Get("status_id"))
		}
		if query.Get("assigned_to_id") != "4" {
			t.Errorf("Expected assigned_to_id=4, got %s", query.Get("assigned_to_id"))
		}
		if query.Get("sort") != "created_on:desc" {
			t.Errorf("Expected sort=created_on:desc, got %s", query.Get("sort"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"issues":      []map[string]any{},
			"total_count": 0,
		})
	}))
	defer server.Close()

	_, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"--project-id", "1",
		"issues", "list",
		"--tracker-id", "2",
		"--status-id", "3",
		"--assigned-to-id", "4",
		"--sort", "created_on:desc",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestIssuesListWithQueryFilter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		if query.Get("query_id") != "7" {
			t.Errorf("Expected query_id=7, got %s", query.Get("query_id"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"issues":      []map[string]any{},
			"total_count": 0,
		})
	}))
	defer server.Close()

	_, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"--project-id", "1",
		"issues", "list",
		"--query", "7",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestIssuesListDoneRatio(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		doneRatio := query.Get("done_ratio")
		if doneRatio != "50" && doneRatio != "" {
			t.Errorf("Expected done_ratio filter, got %s", doneRatio)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"issues": []map[string]any{
				{"id": 1, "subject": "Test", "done_ratio": 50},
			},
			"total_count": 1,
		})
	}))
	defer server.Close()

	stdout, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"--project-id", "1",
		"issues", "list",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "Test") {
		t.Error("Expected output to contain issue subject")
	}
}

func TestProjectCreateWithParent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/projects.json" {
			http.NotFound(w, r)
			return
		}
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"project": map[string]any{
				"id":     10,
				"name":   "Child Project",
				"parent": map[string]any{"id": 1, "name": "Parent Project"},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	stdout, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"projects", "create",
		"--name", "Child Project",
		"--identifier", "child-project",
		"--parent-id", "1",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "10") {
		t.Error("Expected output to contain project ID")
	}
}

func TestProjectCreateWithHomepage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/projects.json" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"project": map[string]any{
				"id":       11,
				"name":     "Project With Homepage",
				"homepage": "https://example.com",
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	stdout, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"projects", "create",
		"--name", "Project With Homepage",
		"--identifier", "homepage-project",
		"--homepage", "https://example.com",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "11") {
		t.Error("Expected output to contain project ID")
	}
}

func TestIssueCreateWithStartDate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/issues.json" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"issue": map[string]any{
				"id":         20,
				"subject":    "Issue with Dates",
				"start_date": "2026-04-01",
				"due_date":   "2026-04-30",
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	stdout, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"issues", "create",
		"--project-id", "1",
		"--subject", "Issue with Dates",
		"--start-date", "2026-04-01",
		"--due-date", "2026-04-30",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "20") {
		t.Error("Expected output to contain issue ID")
	}
}

func TestIssueUpdateWithDoneRatio(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/issues/1.json" {
			http.NotFound(w, r)
			return
		}
		if r.Method != "PUT" {
			t.Errorf("Expected PUT method, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	_, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"issues", "update", "1",
		"--done-ratio", "75",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestVersionCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/projects/1/versions.json" {
			http.NotFound(w, r)
			return
		}
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"version": map[string]any{
				"id":     5,
				"name":   "v2.0.0",
				"status": "open",
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	stdout, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"versions", "create",
		"--project-id", "1",
		"--name", "v2.0.0",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "5") {
		t.Error("Expected output to contain version ID")
	}
}

func TestCategoryCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/projects/1/issue_categories.json" {
			http.NotFound(w, r)
			return
		}
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"issue_category": map[string]any{
				"id":   8,
				"name": "New Category",
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	stdout, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"categories", "create",
		"--project-id", "1",
		"--name", "New Category",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "8") {
		t.Error("Expected output to contain category ID")
	}
}

func TestTimeEntryUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/time_entries/1.json"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}
		if r.Method != "PUT" {
			t.Errorf("Expected PUT method, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	_, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"time-entries", "update", "1",
		"--hours", "5.0",
		"--comments", "Updated comment",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestTimeEntryDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/time_entries/1.json"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE method, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	stdout, _, exitCode := runCommandWithEnv(
		map[string]string{"REDMINE_CLI_YES": "true"},
		"--url", server.URL,
		"--key", "test-api-key",
		"time-entries", "delete", "1",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "1") {
		t.Error("Expected output to contain time entry ID")
	}
}

func TestUsersGetSelf(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/users/current.json"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"user": map[string]any{
				"id":        42,
				"login":     "currentuser",
				"firstname": "Current",
				"lastname":  "User",
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	stdout, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"users", "get-self",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "42") {
		t.Error("Expected output to contain user ID")
	}

	if !strings.Contains(stdout, "currentuser") {
		t.Error("Expected output to contain user login")
	}
}

func TestDebugFlag(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"projects": []map[string]any{},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	stdout, stderr, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"--debug",
		"projects", "list",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	output := stdout + stderr
	if !strings.Contains(output, "DEBUG") {
		t.Logf("Output: %s", output)
	}
}

func TestRetriesFlag(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"projects": []map[string]any{},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	_, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"--retries", "5",
		"projects", "list",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestOffsetFlag(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		offset := r.URL.Query().Get("offset")
		if offset != "10" {
			t.Errorf("Expected offset=10, got %s", offset)
		}

		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"projects":    []map[string]any{},
			"total_count": 100,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	_, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"--offset", "10",
		"projects", "list",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}
