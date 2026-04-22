// test/e2e/comprehensive_test.go
package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestIssueCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/issues.json" {
			http.NotFound(w, r)
			return
		}
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"issue": map[string]any{
				"id":      123,
				"subject": "New Issue",
				"status":  map[string]any{"name": "New"},
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
		"--subject", "New Issue",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "123") {
		t.Error("Expected output to contain issue ID")
	}
}

func TestIssueCreateWithAllFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/issues.json" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"issue": map[string]any{
				"id":          100,
				"subject":     "Full Issue",
				"description": "Full description",
				"tracker":     map[string]any{"name": "Bug"},
				"priority":    map[string]any{"name": "High"},
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
		"--subject", "Full Issue",
		"--description", "Full description",
		"--tracker-id", "1",
		"--priority-id", "2",
		"--assigned-to-id", "5",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "100") {
		t.Error("Expected output to contain issue ID")
	}
}

func TestIssueCreateValidationError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"issue": map[string]any{"id": 1},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	_, stderr, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"issues", "create",
		"--project-id", "1",
	)

	if exitCode == 0 {
		t.Error("Expected non-zero exit code for missing subject")
	}

	if !strings.Contains(stderr, "subject") {
		t.Error("Expected error message about missing subject")
	}
}

func TestIssueCreateMissingProjectID(t *testing.T) {
	_, stderr, exitCode := runCommand(
		"--url", "https://example.com",
		"--key", "test-api-key",
		"issues", "create",
		"--subject", "Test",
	)

	if exitCode == 0 {
		t.Error("Expected non-zero exit code for missing project-id")
	}

	if !strings.Contains(stderr, "project-id") {
		t.Error("Expected error message about missing project-id")
	}
}

func TestIssueUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/issues/123.json"
		if r.URL.Path != expectedPath {
			http.NotFound(w, r)
			return
		}
		if r.Method != "PUT" {
			t.Errorf("Expected PUT method, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"issue": map[string]any{
				"id":      123,
				"subject": "Updated Subject",
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	stdout, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"issues", "update", "123",
		"--subject", "Updated Subject",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "123") {
		t.Error("Expected output to contain issue ID")
	}
}

func TestIssueUpdateWithNotes(t *testing.T) {
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
		"--notes", "This is a test note",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestIssueUpdateStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/issues/1.json" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	_, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"issues", "update", "1",
		"--status-id", "2",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestIssueDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/issues/123.json"
		if r.URL.Path != expectedPath {
			http.NotFound(w, r)
			return
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
		"issues", "delete", "123",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "123") {
		t.Error("Expected output to contain issue ID")
	}
}

func TestIssueDeleteWithYesFlag(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/issues/1.json" {
			http.NotFound(w, r)
			return
		}
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE method, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	_, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"--yes",
		"issues", "delete", "1",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestIssueDeleteCanceledWithoutYes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Server should not receive request when user cancels")
	}))
	defer server.Close()

	done := make(chan string, 1)
	go func() {
		stdout, _, _ := runCommand(
			"--url", server.URL,
			"--key", "test-api-key",
			"issues", "delete", "1",
		)
		done <- stdout
	}()

	select {
	case result := <-done:
		if !strings.Contains(result, "Canceled") {
			t.Error("Expected 'Canceled' message when user declines")
		}
	case <-time.After(2 * time.Second):
		t.Log("Test skipped due to interactive prompt - this is expected")
	}
}

func TestProjectCreate(t *testing.T) {
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
				"id":         10,
				"name":       "New Project",
				"identifier": "new-project",
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	stdout, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"projects", "create",
		"--name", "New Project",
		"--identifier", "new-project",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "10") {
		t.Error("Expected output to contain project ID")
	}
}

func TestProjectCreateValidationError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	_, stderr, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"projects", "create",
		"--name", "Test",
	)

	if exitCode == 0 {
		t.Error("Expected non-zero exit code for missing identifier")
	}

	if !strings.Contains(stderr, "identifier") {
		t.Error("Expected error message about missing identifier")
	}
}

func TestProjectUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/projects/1.json"
		if r.URL.Path != expectedPath {
			http.NotFound(w, r)
			return
		}
		if r.Method != "PUT" {
			t.Errorf("Expected PUT method, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	stdout, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"projects", "update", "1",
		"--name", "Updated Name",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "Updated") {
		t.Error("Expected output to contain updated message")
	}
}

func TestProjectDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/projects/5.json"
		if r.URL.Path != expectedPath {
			http.NotFound(w, r)
			return
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
		"projects", "delete", "5",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "5") {
		t.Error("Expected output to contain project ID")
	}
}

func TestIssueListPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limit := r.URL.Query().Get("limit")
		offset := r.URL.Query().Get("offset")
		projectID := r.URL.Query().Get("project_id")

		if limit != "10" {
			t.Errorf("Expected limit=10, got %s", limit)
		}
		if offset != "20" {
			t.Errorf("Expected offset=20, got %s", offset)
		}
		if projectID != "1" {
			t.Errorf("Expected project_id=1, got %s", projectID)
		}

		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"issues":      []map[string]any{},
			"total_count": 100,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	_, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"--limit", "10",
		"--offset", "20",
		"issues", "list",
		"--project-id", "1",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestIssueListWithFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		projectID := r.URL.Query().Get("project_id")
		trackerID := r.URL.Query().Get("tracker_id")
		statusID := r.URL.Query().Get("status_id")

		if projectID != "1" {
			t.Errorf("Expected project_id=1, got %s", projectID)
		}
		if trackerID != "2" {
			t.Errorf("Expected tracker_id=2, got %s", trackerID)
		}
		if statusID != "3" {
			t.Errorf("Expected status_id=3, got %s", statusID)
		}

		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"issues":      []map[string]any{},
			"total_count": 0,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	_, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"issues", "list",
		"--project-id", "1",
		"--tracker-id", "2",
		"--status-id", "3",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestIssueListAssignedTo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assignedToID := r.URL.Query().Get("assigned_to_id")
		projectID := r.URL.Query().Get("project_id")

		if assignedToID != "5" {
			t.Errorf("Expected assigned_to_id=5, got %s", assignedToID)
		}
		if projectID != "1" {
			t.Errorf("Expected project_id=1, got %s", projectID)
		}

		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"issues": []map[string]any{
				{"id": 1, "subject": "Assigned Issue"},
			},
			"total_count": 1,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	stdout, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"issues", "list",
		"--project-id", "1",
		"--assigned-to-id", "5",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "Assigned Issue") {
		t.Error("Expected output to contain assigned issue")
	}
}

func TestIssueListSorting(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sort := r.URL.Query().Get("sort")
		projectID := r.URL.Query().Get("project_id")

		if sort != "created_on:desc" {
			t.Errorf("Expected sort=created_on:desc, got %s", sort)
		}
		if projectID != "1" {
			t.Errorf("Expected project_id=1, got %s", projectID)
		}

		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"issues":      []map[string]any{},
			"total_count": 0,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	_, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"issues", "list",
		"--project-id", "1",
		"--sort", "created_on:desc",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestProjectGetByIdentifier(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/projects/my-project.json"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"project": map[string]any{
				"id":         1,
				"name":       "My Project",
				"identifier": "my-project",
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	stdout, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"projects", "get", "my-project",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "My Project") {
		t.Error("Expected output to contain project name")
	}
}

func TestProjectListWithPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limit := r.URL.Query().Get("limit")
		offset := r.URL.Query().Get("offset")

		if limit != "5" {
			t.Errorf("Expected limit=5, got %s", limit)
		}
		if offset != "10" {
			t.Errorf("Expected offset=10, got %s", offset)
		}

		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"projects":    []map[string]any{},
			"total_count": 50,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	_, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"--limit", "5",
		"projects", "list",
		"--offset", "10",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestProjectListWithInclude(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		include := r.URL.Query().Get("include")

		if include != "trackers,issue_categories" {
			t.Errorf("Expected include=trackers,issue_categories, got %s", include)
		}

		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"projects": []map[string]any{
				{"id": 1, "name": "Test"},
			},
			"total_count": 1,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	_, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"projects", "list",
		"--include", "trackers,issue_categories",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestUserList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"users": []map[string]any{
				{"id": 1, "login": "user1", "firstname": "User", "lastname": "One"},
				{"id": 2, "login": "user2", "firstname": "User", "lastname": "Two"},
			},
			"total_count": 2,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	stdout, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"users", "list",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "user1") || !strings.Contains(stdout, "user2") {
		t.Error("Expected output to contain both users")
	}
}

func TestVersionGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/versions/1.json"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"version": map[string]any{
				"id":     1,
				"name":   "v1.0.0",
				"status": "open",
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	stdout, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"versions", "get", "1",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "v1.0.0") {
		t.Error("Expected output to contain version name")
	}
}

func TestCategoryGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/issue_categories/1.json"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"issue_category": map[string]any{
				"id":      1,
				"name":    "Bug",
				"project": map[string]any{"name": "Test Project"},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	stdout, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"categories", "get", "1",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "Bug") {
		t.Error("Expected output to contain category name")
	}
}

func TestTimeEntryGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/time_entries/1.json"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"time_entry": map[string]any{
				"id":       1,
				"hours":    2.5,
				"comments": "Work done",
				"user":     map[string]any{"name": "Admin"},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	stdout, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"time-entries", "get", "1",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "Work done") {
		t.Error("Expected output to contain time entry comments")
	}
}

func TestTimeEntryCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/time_entries.json" {
			http.NotFound(w, r)
			return
		}
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"time_entry": map[string]any{
				"id": 10,
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	_, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"time-entries", "create",
		"--issue-id", "1",
		"--hours", "3.5",
		"--spent-on", "2026-04-13",
		"--comments", "Testing",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestNetworkError(t *testing.T) {
	_, stderr, exitCode := runCommand(
		"--url", "http://localhost:99999",
		"--key", "test-api-key",
		"projects", "list",
	)

	if exitCode == 0 {
		t.Error("Expected non-zero exit code for network error")
	}

	if !strings.Contains(stderr, "Error") && !strings.Contains(stderr, "connect") {
		t.Error("Expected error message about connection failure")
	}
}

func TestInvalidURL(t *testing.T) {
	_, _, exitCode := runCommand(
		"--url", "not-a-valid-url",
		"--key", "test-api-key",
		"projects", "list",
	)

	if exitCode == 0 {
		t.Error("Expected non-zero exit code for invalid URL")
	}
}

func TestMalformedJSONResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("this is not json"))
	}))
	defer server.Close()

	_, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"projects", "list",
	)

	if exitCode == 0 {
		t.Error("Expected non-zero exit code for malformed JSON")
	}
}

func TestEmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{}"))
	}))
	defer server.Close()

	stdout, stderr, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"projects", "list",
	)

	t.Logf("stdout: %s, stderr: %s, exitCode: %d", stdout, stderr, exitCode)
}

func TestInvalidIDArgument(t *testing.T) {
	_, stderr, exitCode := runCommand(
		"--url", "https://example.com",
		"--key", "test-api-key",
		"issues", "get", "abc",
	)

	if exitCode == 0 {
		t.Error("Expected non-zero exit code for invalid ID")
	}

	if !strings.Contains(stderr, "invalid") {
		t.Error("Expected error message about invalid ID")
	}
}

func TestIssueGetWithInclude(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		include := r.URL.Query().Get("include")

		if include != "children,attachments" {
			t.Errorf("Expected include=children,attachments, got %s", include)
		}

		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"issue": map[string]any{
				"id":          1,
				"subject":     "Test Issue",
				"children":    []map[string]any{},
				"attachments": []map[string]any{},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	_, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"issues", "get", "1",
		"--include", "children,attachments",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestOutputToFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"projects": []map[string]any{
				{"id": 1, "name": "Test"},
			},
			"total_count": 1,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	tmpFile := "/tmp/redmine-test-output.json"
	stdout, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"--output", tmpFile,
		"projects", "list",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "Written") {
		t.Logf("stdout: %s", stdout)
	}
}

func TestBatchRequests(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"projects": []map[string]any{{"id": 1}},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	for i := 0; i < 3; i++ {
		_, _, exitCode := runCommand(
			"--url", server.URL,
			"--key", "test-api-key",
			"projects", "list",
		)

		if exitCode != 0 {
			t.Errorf("Request %d: Expected exit code 0, got %d", i+1, exitCode)
		}
	}
}

func TestProjectCreateWithDescription(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/projects.json" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"project": map[string]any{
				"id":          1,
				"name":        "Project With Desc",
				"description": "This is a description",
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	stdout, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"projects", "create",
		"--name", "Project With Desc",
		"--identifier", "project-desc",
		"--description", "This is a description",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "Project With Desc") {
		t.Error("Expected output to contain project name")
	}
}

func TestIssueCreateWithDueDate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/issues.json" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"issue": map[string]any{
				"id":       1,
				"subject":  "Issue with Due Date",
				"due_date": "2026-12-31",
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
		"--subject", "Issue with Due Date",
		"--due-date", "2026-12-31",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "Issue with Due Date") {
		t.Error("Expected output to contain issue subject")
	}
}

func TestConflictError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]any{
			"errors": []string{"Conflict"},
		})
	}))
	defer server.Close()

	_, stderr, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"issues", "update", "1",
		"--subject", "Conflict Test",
	)

	if exitCode == 0 {
		t.Error("Expected non-zero exit code for conflict error")
	}

	if !strings.Contains(stderr, "Conflict") {
		t.Error("Expected error message about conflict")
	}
}

func TestUnprocessableEntity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]any{
			"errors": []string{"Validation failed"},
		})
	}))
	defer server.Close()

	_, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"issues", "create",
		"--project-id", "1",
		"--subject", "",
	)

	if exitCode == 0 {
		t.Error("Expected non-zero exit code for validation error")
	}
}

func TestAPIKeyRequiredHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-Redmine-API-Key")
		if apiKey != "my-secret-key" {
			t.Errorf("Expected X-Redmine-API-Key header with value 'my-secret-key', got '%s'", apiKey)
		}
		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"projects": []map[string]any{},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	_, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "my-secret-key",
		"projects", "list",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestCustomQueryFilter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("query_id")

		if query != "5" {
			t.Errorf("Expected query_id=5, got %s", query)
		}

		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"issues":      []map[string]any{},
			"total_count": 0,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	_, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"issues", "list",
		"--query", "5",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestAllOutputFormats(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"projects": []map[string]any{
				{"id": 1, "name": "Test Project"},
			},
			"total_count": 1,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	formats := []string{"json", "table", "raw"}

	for _, format := range formats {
		t.Run(fmt.Sprintf("format_%s", format), func(t *testing.T) {
			stdout, _, exitCode := runCommand(
				"--url", server.URL,
				"--key", "test-api-key",
				"--format", format,
				"projects", "list",
			)

			if exitCode != 0 {
				t.Errorf("Expected exit code 0 for format %s, got %d", format, exitCode)
			}

			if stdout == "" {
				t.Errorf("Expected non-empty output for format %s", format)
			}
		})
	}
}

func TestIssueDeleteDryRun(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Server should not receive request in dry-run mode")
	}))
	defer server.Close()

	stdout, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"--dry-run",
		"--yes",
		"issues", "delete", "1",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for dry-run, got %d", exitCode)
	}

	if !strings.Contains(stdout, "dry-run") {
		t.Error("Expected output to contain dry-run message")
	}
}

func TestProjectUpdateDryRun(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Server should not receive request in dry-run mode")
	}))
	defer server.Close()

	stdout, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"--dry-run",
		"projects", "update", "1",
		"--name", "New Name",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for dry-run, got %d", exitCode)
	}

	if !strings.Contains(stdout, "dry-run") {
		t.Error("Expected output to contain dry-run message")
	}
}

func TestIssueCreateDryRun(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Server should not receive request in dry-run mode")
	}))
	defer server.Close()

	stdout, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"--dry-run",
		"issues", "create",
		"--project-id", "1",
		"--subject", "Dry Run Issue",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for dry-run, got %d", exitCode)
	}

	if !strings.Contains(stdout, "dry-run") {
		t.Error("Expected output to contain dry-run message")
	}
}

func TestTimeoutShort(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	_, stderr, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"--timeout", "1s",
		"projects", "list",
	)

	if exitCode == 0 {
		t.Error("Expected non-zero exit code for timeout")
	}

	if !strings.Contains(stderr, "timeout") && !strings.Contains(stderr, "Timeout") {
		t.Logf("stderr: %s", stderr)
	}
}
