// test/e2e/e2e_test.go
package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

var binaryPath string

func TestMain(m *testing.M) {
	// 获取项目根目录
	projectRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		fmt.Printf("Failed to get project root: %v\n", err)
		os.Exit(1)
	}

	// 设置二进制文件路径
	binaryName := "redmine"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	binaryPath = filepath.Join(projectRoot, "bin", binaryName)

	// 确保每次都重新构建二进制文件
	fmt.Printf("Building binary for E2E tests at: %s\n", binaryPath)

	// 删除旧的二进制文件（如果存在）
	os.Remove(binaryPath)

	// 构建新的二进制文件
	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd")
	cmd.Dir = projectRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Printf("Failed to build binary: %v\n%s\n", err, output)
		os.Exit(1)
	}

	// 验证二进制文件是否创建成功
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		fmt.Printf("Binary not found at %s after build\n", binaryPath)
		os.Exit(1)
	}

	fmt.Println("Binary built successfully")
	os.Exit(m.Run())
}

func runCommand(args ...string) (stdout, stderr string, exitCode int) {
	cmd := exec.Command(binaryPath, args...)
	cmd.Env = append(os.Environ(), "NO_COLOR=1")

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	stdout = stdoutBuf.String()
	stderr = stderrBuf.String()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	}
	return
}

func runCommandWithEnv(env map[string]string, args ...string) (stdout, stderr string, exitCode int) {
	cmd := exec.Command(binaryPath, args...)
	cmdEnv := append(os.Environ(), "NO_COLOR=1")
	for k, v := range env {
		cmdEnv = append(cmdEnv, fmt.Sprintf("%s=%s", k, v))
	}
	cmd.Env = cmdEnv

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	stdout = stdoutBuf.String()
	stderr = stderrBuf.String()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	}
	return
}

func TestHelpCommand(t *testing.T) {
	stdout, _, exitCode := runCommand("--help")

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "Redmine CLI") {
		t.Error("Expected help output to contain 'Redmine CLI'")
	}

	if !strings.Contains(stdout, "Usage:") {
		t.Error("Expected help output to contain 'Usage:'")
	}
}

func TestVersionCommand(t *testing.T) {
	stdout, _, exitCode := runCommand("version")

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "redmine version") {
		t.Error("Expected version output to contain 'redmine version'")
	}
}

func TestVersionFlag(t *testing.T) {
	stdout, _, exitCode := runCommand("--version")

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "redmine version") {
		t.Error("Expected version output to contain 'redmine version'")
	}
}

func TestUnknownCommand(t *testing.T) {
	_, stderr, exitCode := runCommand("unknown-command")

	if exitCode == 0 {
		t.Error("Expected non-zero exit code for unknown command")
	}

	if !strings.Contains(stderr, "unknown") && !strings.Contains(stderr, "Error") {
		t.Error("Expected error message for unknown command")
	}
}

func TestConfigHelp(t *testing.T) {
	stdout, _, exitCode := runCommand("config", "--help")

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "config") {
		t.Error("Expected config help to contain 'config'")
	}
}

func TestLoginHelp(t *testing.T) {
	stdout, _, exitCode := runCommand("login", "--help")

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "login") {
		t.Error("Expected login help to contain 'login'")
	}
}

func TestProjectsHelp(t *testing.T) {
	stdout, _, exitCode := runCommand("projects", "--help")

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "projects") {
		t.Error("Expected projects help to contain 'projects'")
	}
}

func TestIssuesHelp(t *testing.T) {
	stdout, _, exitCode := runCommand("issues", "--help")

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "issues") {
		t.Error("Expected issues help to contain 'issues'")
	}
}

func TestUsersHelp(t *testing.T) {
	stdout, _, exitCode := runCommand("users", "--help")

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "users") {
		t.Error("Expected users help to contain 'users'")
	}
}

func TestGlobalFlags(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		check func(stdout, stderr string) bool
	}{
		{
			name: "format flag",
			args: []string{"--format", "json", "--help"},
			check: func(stdout, stderr string) bool {
				return strings.Contains(stdout, "format")
			},
		},
		{
			name: "verbose flag",
			args: []string{"--verbose", "--help"},
			check: func(stdout, stderr string) bool {
				return strings.Contains(stdout, "verbose")
			},
		},
		{
			name: "output flag",
			args: []string{"--output", "test.json", "--help"},
			check: func(stdout, stderr string) bool {
				return strings.Contains(stdout, "output")
			},
		},
		{
			name: "timeout flag",
			args: []string{"--timeout", "60s", "--help"},
			check: func(stdout, stderr string) bool {
				return strings.Contains(stdout, "timeout")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, exitCode := runCommand(tt.args...)
			if exitCode != 0 {
				t.Errorf("Expected exit code 0, got %d", exitCode)
			}
			if !tt.check(stdout, stderr) {
				t.Errorf("Check failed for test %s", tt.name)
			}
		})
	}
}

func TestProjectsListWithMockServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/projects.json" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"projects": []map[string]any{
				{"id": 1, "name": "Project Alpha", "identifier": "alpha"},
				{"id": 2, "name": "Project Beta", "identifier": "beta"},
			},
			"total_count": 2,
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

	if !strings.Contains(stdout, "Project Alpha") {
		t.Error("Expected output to contain 'Project Alpha'")
	}
}

func TestIssuesListWithMockServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/issues.json" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"issues": []map[string]any{
				{"id": 1, "subject": "Test Issue 1", "status": map[string]any{"name": "New"}},
				{"id": 2, "subject": "Test Issue 2", "status": map[string]any{"name": "In Progress"}},
			},
			"total_count": 2,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	stdout, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"issues", "list",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "Test Issue 1") {
		t.Error("Expected output to contain 'Test Issue 1'")
	}
}

func TestUsersListWithMockServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users.json" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"users": []map[string]any{
				{"id": 1, "login": "admin", "firstname": "Admin", "lastname": "User"},
				{"id": 2, "login": "john", "firstname": "John", "lastname": "Doe"},
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

	if !strings.Contains(stdout, "john") {
		t.Error("Expected output to contain 'john'")
	}
}

func TestOutputFormatJSON(t *testing.T) {
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

	stdout, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"--format", "json",
		"projects", "list",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Errorf("Expected valid JSON output, got error: %v", err)
	}
}

func TestOutputFormatTable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"projects": []map[string]any{
				{"id": 1, "name": "Test Project", "identifier": "test"},
			},
			"total_count": 1,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	stdout, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"--format", "table",
		"projects", "list",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "id") || !strings.Contains(stdout, "name") {
		t.Error("Expected table output to contain column headers")
	}
}

func TestJQFilter(t *testing.T) {
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

	stdout, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"--jq", ".total_count",
		"projects", "list",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "1") {
		t.Errorf("Expected JQ filtered output to contain '1', got: %s", stdout)
	}
}

func TestFieldsFilter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{
			"projects": []map[string]any{
				{"id": 1, "name": "Test Project", "identifier": "test", "description": "A test project"},
			},
			"total_count": 1,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	stdout, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"--fields", "id,name",
		"projects", "list",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "id") || !strings.Contains(stdout, "name") {
		t.Error("Expected output to contain 'id' and 'name' fields")
	}
}

func TestAuthenticationError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]any{
			"errors": []string{"Unauthorized"},
		})
	}))
	defer server.Close()

	_, stderr, exitCode := runCommand(
		"--url", server.URL,
		"--key", "invalid-key",
		"projects", "list",
	)

	if exitCode == 0 {
		t.Error("Expected non-zero exit code for authentication error")
	}

	if !strings.Contains(stderr, "Error") {
		t.Error("Expected error message in stderr")
	}
}

func TestNotFoundError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]any{
			"errors": []string{"Not found"},
		})
	}))
	defer server.Close()

	_, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"issues", "get", "99999",
	)

	if exitCode == 0 {
		t.Error("Expected non-zero exit code for not found error")
	}
}

func TestServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"errors": []string{"Internal server error"},
		})
	}))
	defer server.Close()

	_, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"projects", "list",
	)

	if exitCode == 0 {
		t.Error("Expected non-zero exit code for server error")
	}
}

func TestDryRunFlag(t *testing.T) {
	requestReceived := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestReceived = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	_, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"--dry-run",
		"issues", "create", "--project-id", "1", "--subject", "Test",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for dry-run, got %d", exitCode)
	}

	if requestReceived {
		t.Error("Expected no request to be sent in dry-run mode")
	}
}

func TestLimitFlag(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limit := r.URL.Query().Get("limit")
		if limit != "10" {
			t.Errorf("Expected limit=10, got %s", limit)
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
		"--limit", "10",
		"issues", "list",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestTimeoutFlag(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{"issues": []map[string]any{}}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	stdout, stderr, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"--timeout", "1s",
		"issues", "list",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d (stdout: %s, stderr: %s)", exitCode, stdout, stderr)
	}
}

func TestVerboseFlag(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{"projects": []map[string]any{}}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	_, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"--verbose",
		"projects", "list",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestMissingURL(t *testing.T) {
	_, stderr, exitCode := runCommand(
		"--key", "test-api-key",
		"projects", "list",
	)

	if exitCode == 0 {
		t.Error("Expected non-zero exit code when URL is missing")
	}

	if !strings.Contains(stderr, "URL") {
		t.Error("Expected error message about missing URL")
	}
}

func TestMissingAPIKey(t *testing.T) {
	_, stderr, exitCode := runCommand(
		"--url", "https://example.com",
		"projects", "list",
	)

	if exitCode == 0 {
		t.Error("Expected non-zero exit code when API key is missing")
	}

	if !strings.Contains(stderr, "key") {
		t.Error("Expected error message about missing API key")
	}
}

func TestProjectGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/projects/1.json" {
			w.Header().Set("Content-Type", "application/json")
			response := map[string]any{
				"project": map[string]any{
					"id":          1,
					"name":        "Test Project",
					"identifier":  "test",
					"description": "A test project",
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	stdout, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"projects", "get", "1",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "Test Project") {
		t.Error("Expected output to contain 'Test Project'")
	}
}

func TestIssueGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/issues/1.json" {
			w.Header().Set("Content-Type", "application/json")
			response := map[string]any{
				"issue": map[string]any{
					"id":      1,
					"subject": "Test Issue",
					"status":  map[string]any{"name": "New"},
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	stdout, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"issues", "get", "1",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "Test Issue") {
		t.Error("Expected output to contain 'Test Issue'")
	}
}

func TestUserGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/users/1.json" {
			w.Header().Set("Content-Type", "application/json")
			response := map[string]any{
				"user": map[string]any{
					"id":        1,
					"login":     "admin",
					"firstname": "Admin",
					"lastname":  "User",
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	stdout, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"users", "get", "1",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "admin") {
		t.Error("Expected output to contain 'admin'")
	}
}

func TestCategoriesList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/projects/1/issue_categories.json" {
			w.Header().Set("Content-Type", "application/json")
			response := map[string]any{
				"issue_categories": []map[string]any{
					{"id": 1, "name": "Bug"},
					{"id": 2, "name": "Feature"},
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	stdout, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"categories", "list", "--project-id", "1",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "Bug") {
		t.Error("Expected output to contain 'Bug'")
	}
}

func TestVersionsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/projects/1/versions.json" {
			w.Header().Set("Content-Type", "application/json")
			response := map[string]any{
				"versions": []map[string]any{
					{"id": 1, "name": "v1.0"},
					{"id": 2, "name": "v2.0"},
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	stdout, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"versions", "list", "--project-id", "1",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "v1.0") {
		t.Error("Expected output to contain 'v1.0'")
	}
}

func TestTimeEntriesList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/time_entries.json" {
			w.Header().Set("Content-Type", "application/json")
			response := map[string]any{
				"time_entries": []map[string]any{
					{"id": 1, "hours": 2.5, "comments": "Work on feature"},
				},
				"total_count": 1,
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	stdout, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"time-entries", "list",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "Work on feature") {
		t.Error("Expected output to contain 'Work on feature'")
	}
}

func TestStatusesList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/issue_statuses.json" {
			w.Header().Set("Content-Type", "application/json")
			response := map[string]any{
				"issue_statuses": []map[string]any{
					{"id": 1, "name": "New"},
					{"id": 2, "name": "In Progress"},
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	stdout, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"statuses", "list",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "New") {
		t.Error("Expected output to contain 'New'")
	}
}

func TestTrackersList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/trackers.json" {
			w.Header().Set("Content-Type", "application/json")
			response := map[string]any{
				"trackers": []map[string]any{
					{"id": 1, "name": "Bug"},
					{"id": 2, "name": "Feature"},
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	stdout, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"trackers", "list",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "Bug") {
		t.Error("Expected output to contain 'Bug'")
	}
}

func TestTrackerGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/trackers/1.json" {
			w.Header().Set("Content-Type", "application/json")
			response := map[string]any{
				"tracker": map[string]any{
					"id":          1,
					"name":        "Bug",
					"description": "Bug tracker",
					"custom_fields": []map[string]any{
						{
							"id":           5,
							"name":         "Affected Version",
							"field_format": "list",
							"possible_values": []map[string]any{
								{"value": "v1.0", "label": "Version 1.0"},
								{"value": "v2.0", "label": "Version 2.0"},
							},
						},
					},
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	stdout, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"trackers", "get", "1",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "Bug") {
		t.Error("Expected output to contain 'Bug'")
	}
	if !strings.Contains(stdout, "Affected Version") {
		t.Error("Expected output to contain custom field name 'Affected Version'")
	}
	if !strings.Contains(stdout, "list") {
		t.Error("Expected output to contain field_format 'list'")
	}
}

func TestIssueCreateWithCustomField(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/issues.json" {
			http.NotFound(w, r)
			return
		}
		if r.Method != "POST" {
			t.Errorf("expected POST request, got %s", r.Method)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		issue, ok := body["issue"].(map[string]any)
		if !ok {
			t.Fatal("expected issue in request body")
		}
		cfs, ok := issue["custom_fields"].([]any)
		if !ok {
			t.Fatal("expected custom_fields in issue")
		}
		if len(cfs) != 1 {
			t.Errorf("expected 1 custom field, got %d", len(cfs))
		}
		cf := cfs[0].(map[string]any)
		if cf["id"].(float64) != 5 {
			t.Errorf("expected custom field id 5, got %v", cf["id"])
		}
		if cf["value"] != "v1.0" {
			t.Errorf("expected custom field value 'v1.0', got %v", cf["value"])
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"issue": map[string]any{"id": 100, "subject": "Test Issue"},
		})
	}))
	defer server.Close()

	stdout, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"issues", "create",
		"--project-id", "1",
		"--subject", "Test Issue",
		"--custom-field", "id:5:v1.0",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "100") {
		t.Error("Expected output to contain issue ID")
	}
}

func TestIssueUpdateWithCustomField(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/issues/123.json" {
			http.NotFound(w, r)
			return
		}
		if r.Method != "PUT" {
			t.Errorf("expected PUT request, got %s", r.Method)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		issue, ok := body["issue"].(map[string]any)
		if !ok {
			t.Fatal("expected issue in request body")
		}
		cfs, ok := issue["custom_fields"].([]any)
		if !ok {
			t.Fatal("expected custom_fields in issue")
		}
		if len(cfs) != 1 {
			t.Errorf("expected 1 custom field, got %d", len(cfs))
		}
		cf := cfs[0].(map[string]any)
		if cf["id"].(float64) != 7 {
			t.Errorf("expected custom field id 7, got %v", cf["id"])
		}
		if cf["value"] != "test-value" {
			t.Errorf("expected custom field value 'test-value', got %v", cf["value"])
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	stdout, stderr, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"issues", "update", "123",
		"--custom-field", "id:7:test-value",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. stderr: %s", exitCode, stderr)
	}

	if !strings.Contains(stdout, "123") {
		t.Error("Expected output to contain issue ID")
	}
}

func TestPrioritiesList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/enumerations/issue_priorities.json" {
			w.Header().Set("Content-Type", "application/json")
			response := map[string]any{
				"issue_priorities": []map[string]any{
					{"id": 1, "name": "Low"},
					{"id": 2, "name": "Normal"},
					{"id": 3, "name": "High"},
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	stdout, _, exitCode := runCommand(
		"--url", server.URL,
		"--key", "test-api-key",
		"priorities", "list",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "Normal") {
		t.Error("Expected output to contain 'Normal'")
	}
}

func TestConcurrentRequests(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]any{"projects": []map[string]any{}}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	done := make(chan bool, 5)

	for i := 0; i < 5; i++ {
		go func() {
			_, _, exitCode := runCommand(
				"--url", server.URL,
				"--key", "test-api-key",
				"projects", "list",
			)
			done <- exitCode == 0
		}()
	}

	for i := 0; i < 5; i++ {
		if !<-done {
			t.Error("Concurrent request failed")
		}
	}
}
