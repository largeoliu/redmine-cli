// test/e2e/edge_cases_test.go
package e2e

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestIssueCreateWithLongSubject(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"issue": map[string]any{"id": 1},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	longSubject := strings.Repeat("a", 500)
	stdout, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"issues", "create",
		"--project-id", "1",
		"--subject", longSubject,
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "1") {
		t.Error("Expected output to contain issue ID")
	}
}

func TestIssueCreateWithSpecialCharacters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"issue": map[string]any{"id": 1},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	testCases := []struct {
		name    string
		subject string
	}{
		{"chinese", "测试中文 Subject"},
		{"emoji", "Subject with emoji 🎉"},
		{"quotes", `Subject with "quotes" and 'apostrophes'`},
		{"newlines", "Subject\nwith\nnewlines"},
		{"tabs", "Subject\twith\ttabs"},
		{"unicode", "Subject with üñíçödé and 日本語"},
		{"special_chars", "Subject with <>&| characters"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, exitCode := runCommand(
				"--url", server.URL,
				"--key", "test-api-key",
				"issues", "create",
				"--project-id", "1",
				"--subject", tc.subject,
			)

			if exitCode != 0 {
				t.Errorf("Failed for %s: exit code %d", tc.name, exitCode)
			}
		})
	}
}

func TestEmptyProjectList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"projects":    []map[string]any{},
			"total_count": 0,
		}
		json.NewEncoder(w).Encode(response)
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

	if !strings.Contains(stdout, "0") {
		t.Error("Expected output to show zero count")
	}
}

func TestEmptyIssueList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"issues":      []map[string]any{},
			"total_count": 0,
		}
		json.NewEncoder(w).Encode(response)
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

	if !strings.Contains(stdout, "0") {
		t.Error("Expected output to show zero count")
	}
}

func TestLimitZero(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"projects":    []map[string]any{},
			"total_count": 0,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	_, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"--limit", "0",
		"projects", "list",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestOffsetZero(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"projects":    []map[string]any{},
			"total_count": 0,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	_, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"--offset", "0",
		"projects", "list",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestVeryLargeLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limit := r.URL.Query().Get("limit")
		if limit != "99999" {
			t.Errorf("Expected limit=99999, got %s", limit)
		}

		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"projects":    []map[string]any{},
			"total_count": 0,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	_, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"--limit", "99999",
		"projects", "list",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestConcurrentRequestsEdge(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"projects": []map[string]any{{"id": 1}},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	for i := 0; i < 5; i++ {
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

func TestNonJSONResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("<html>Not JSON</html>"))
	}))
	defer server.Close()

	_, stderr, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"projects", "list",
	)

	if exitCode == 0 {
		t.Error("Expected non-zero exit code for non-JSON response")
	}

	if !strings.Contains(stderr, "decode") && !strings.Contains(stderr, "JSON") && !strings.Contains(stderr, "invalid") {
		t.Logf("stderr: %s", stderr)
	}
}

func TestEmptyResponseBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(""))
	}))
	defer server.Close()

	_, stderr, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"projects", "list",
	)

	if exitCode == 0 {
		t.Error("Expected non-zero exit code for empty response")
	}

	t.Logf("Empty body stderr: %s", stderr)
}

func TestServerRedirect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/projects.json", http.StatusMovedPermanently)
	}))
	defer server.Close()

	_, stderr, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"projects", "list",
	)

	if exitCode == 0 {
		t.Error("Expected non-zero exit code for redirect")
	}

	t.Logf("Redirect stderr: %s", stderr)
}

func TestServerBadGateway(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	_, stderr, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"projects", "list",
	)

	if exitCode == 0 {
		t.Error("Expected non-zero exit code for bad gateway")
	}

	if !strings.Contains(stderr, "502") && !strings.Contains(stderr, "Bad Gateway") {
		t.Logf("stderr: %s", stderr)
	}
}

func TestServerGatewayTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusGatewayTimeout)
	}))
	defer server.Close()

	_, stderr, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"projects", "list",
	)

	if exitCode == 0 {
		t.Error("Expected non-zero exit code for gateway timeout")
	}

	if !strings.Contains(stderr, "504") && !strings.Contains(stderr, "Gateway Timeout") {
		t.Logf("stderr: %s", stderr)
	}
}

func TestIssueIDZero(t *testing.T) {
	_, stderr, exitCode := runCommand(
		"--url", "https://example.com",
		"--key", "test-api-key",
		"issues", "get", "0",
	)

	if exitCode == 0 {
		t.Error("Expected non-zero exit code for issue ID 0")
	}

	if !strings.Contains(stderr, "invalid") && !strings.Contains(stderr, "positive") {
		t.Logf("stderr: %s", stderr)
	}
}

func TestIssueIDNegative(t *testing.T) {
	_, stderr, exitCode := runCommand(
		"--url", "https://example.com",
		"--key", "test-api-key",
		"issues", "get", "-1",
	)

	if exitCode == 0 {
		t.Error("Expected non-zero exit code for negative issue ID")
	}

	if !strings.Contains(stderr, "invalid") && !strings.Contains(stderr, "positive") {
		t.Logf("stderr: %s", stderr)
	}
}

func TestProjectIdentifierWithHyphen(t *testing.T) {
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

func TestProjectIdentifierWithUnderscore(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/projects/my_project.json"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"project": map[string]any{
				"id":         1,
				"name":       "My Project",
				"identifier": "my_project",
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	stdout, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"projects", "get", "my_project",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "My Project") {
		t.Error("Expected output to contain project name")
	}
}

func TestAllOutputFormatsForIssues(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"issues": []map[string]any{
				{"id": 1, "subject": "Test Issue"},
			},
			"total_count": 1,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	formats := []string{"json", "table", "raw"}

	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			stdout, _, exitCode := runCommand(
				"--url", server.URL,
				"--key", "test-api-key",
				"--format", format,
				"--project-id", "1",
				"issues", "list",
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

func TestAllOutputFormatsForUsers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"users": []map[string]any{
				{"id": 1, "login": "testuser"},
			},
			"total_count": 1,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	formats := []string{"json", "table", "raw"}

	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			stdout, _, exitCode := runCommand(
				"--url", server.URL,
				"--key", "test-api-key",
				"--format", format,
				"users", "list",
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

func TestTimeEntryCreateWithActivity(t *testing.T) {
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
		"--hours", "2.0",
		"--activity-id", "5",
		"--spent-on", "2026-04-13",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestProjectUpdateStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/projects/1.json" {
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
		"projects", "update", "1",
		"--status", "5",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestIssueCreateWithPriority(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/issues.json" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"issue": map[string]any{
				"id":      100,
				"subject": "Issue with Priority",
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
		"--subject", "Issue with Priority",
		"--priority-id", "3",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "100") {
		t.Error("Expected output to contain issue ID")
	}
}

func TestVersionGetByProject(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/projects/1/versions.json"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"versions": []map[string]any{
				{"id": 1, "name": "v1.0.0"},
				{"id": 2, "name": "v2.0.0"},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	stdout, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"versions", "list",
		"--project-id", "1",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "v1.0.0") || !strings.Contains(stdout, "v2.0.0") {
		t.Error("Expected output to contain versions")
	}
}

func TestCategoryGetByProject(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/projects/1/issue_categories.json"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"issue_categories": []map[string]any{
				{"id": 1, "name": "Bug"},
				{"id": 2, "name": "Feature"},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	stdout, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"categories", "list",
		"--project-id", "1",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "Bug") || !strings.Contains(stdout, "Feature") {
		t.Error("Expected output to contain categories")
	}
}

func TestIssueWatcherFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/issues/1/agile_data.json" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"agile_data": map[string]any{}})
			return
		}

		query := r.URL.Query()
		include := query.Get("include")

		if include != "watchers" {
			t.Errorf("Expected include=watchers, got %s", include)
		}

		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"issue": map[string]any{
				"id":       1,
				"subject":  "Issue with Watchers",
				"watchers": []map[string]any{},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	_, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"issues", "get", "1",
		"--include", "watchers",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestMultipleFlagsCombined(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		if query.Get("limit") != "5" {
			t.Errorf("Expected limit=5, got %s", query.Get("limit"))
		}
		if query.Get("offset") != "10" {
			t.Errorf("Expected offset=10, got %s", query.Get("offset"))
		}
		if query.Get("sort") != "updated_on:desc" {
			t.Errorf("Expected sort=updated_on:desc, got %s", query.Get("sort"))
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
		"--limit", "5",
		"--offset", "10",
		"--format", "json",
		"--project-id", "1",
		"issues", "list",
		"--sort", "updated_on:desc",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}
