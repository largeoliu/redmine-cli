package e2e

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

func TestLoginSavesConfigAndUsesInstance(t *testing.T) {
	tempDir := t.TempDir()
	var sawProjectsRequest atomic.Bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodHead && r.URL.Path == "/":
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/users/current.json":
			if got := r.Header.Get("X-Redmine-API-Key"); got != "secret-key" {
				t.Errorf("expected API key header %q, got %q", "secret-key", got)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"user": map[string]any{
					"id":    1,
					"login": "tester",
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/projects.json":
			sawProjectsRequest.Store(true)
			if got := r.Header.Get("X-Redmine-API-Key"); got != "secret-key" {
				t.Errorf("expected API key header %q, got %q", "secret-key", got)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"projects": []map[string]any{
					{"id": 1, "name": "Alpha"},
				},
				"total_count": 1,
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	env := map[string]string{"REDMINE_CONFIG_DIR": tempDir}
	input := strings.Join([]string{server.URL, "secret-key", "team"}, "\n") + "\n"

	stdout, stderr, exitCode := runCommandWithEnvInput(env, input, "login")
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d (stderr: %s)", exitCode, stderr)
	}
	if !strings.Contains(stdout, "登录成功") {
		t.Fatalf("expected login success output, got: %s", stdout)
	}

	configPath := filepath.Join(tempDir, "config.yaml")
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("expected config file to be written: %v", err)
	}

	stdout, stderr, exitCode = runCommandWithEnv(env,
		"--instance", "team",
		"projects", "list",
	)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d (stderr: %s)", exitCode, stderr)
	}
	if !sawProjectsRequest.Load() {
		t.Fatal("expected projects request to be routed through saved instance")
	}
	if !strings.Contains(stdout, "Alpha") {
		t.Fatalf("expected projects output to contain Alpha, got: %s", stdout)
	}
}

func TestConfigSetAndListUpdatesDefaultInstance(t *testing.T) {
	tempDir := t.TempDir()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodHead && r.URL.Path == "/":
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/users/current.json":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"user": map[string]any{"id": 1, "login": "tester"},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	env := map[string]string{"REDMINE_CONFIG_DIR": tempDir}

	alphaInput := strings.Join([]string{server.URL, "alpha-key", "alpha"}, "\n") + "\n"
	stdout, stderr, exitCode := runCommandWithEnvInput(env, alphaInput, "login")
	if exitCode != 0 {
		t.Fatalf("expected first login to succeed, got %d (stderr: %s)", exitCode, stderr)
	}
	if !strings.Contains(stdout, "登录成功") {
		t.Fatalf("expected first login success output, got: %s", stdout)
	}

	betaInput := strings.Join([]string{server.URL, "beta-key", "beta", "n"}, "\n") + "\n"
	stdout, stderr, exitCode = runCommandWithEnvInput(env, betaInput, "login")
	if exitCode != 0 {
		t.Fatalf("expected second login to succeed, got %d (stderr: %s)", exitCode, stderr)
	}
	if !strings.Contains(stdout, "登录成功") {
		t.Fatalf("expected second login success output, got: %s", stdout)
	}

	stdout, stderr, exitCode = runCommandWithEnv(env, "config", "set", "beta")
	if exitCode != 0 {
		t.Fatalf("expected config set to succeed, got %d (stderr: %s)", exitCode, stderr)
	}
	if !strings.Contains(stdout, "beta") {
		t.Fatalf("expected config set output to mention beta, got: %s", stdout)
	}

	stdout, stderr, exitCode = runCommandWithEnv(env, "config", "get")
	if exitCode != 0 {
		t.Fatalf("expected config get to succeed, got %d (stderr: %s)", exitCode, stderr)
	}
	if !strings.Contains(stdout, "Default instance: beta") {
		t.Fatalf("expected beta to be default, got: %s", stdout)
	}

	stdout, stderr, exitCode = runCommandWithEnv(env, "config", "list")
	if exitCode != 0 {
		t.Fatalf("expected config list to succeed, got %d (stderr: %s)", exitCode, stderr)
	}

	var payload struct {
		Instances []struct {
			Name    string `json:"name"`
			URL     string `json:"url"`
			Default bool   `json:"default"`
		} `json:"instances"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &payload); err != nil {
		t.Fatalf("failed to decode config list output: %v\noutput: %s", err, stdout)
	}
	if len(payload.Instances) != 2 {
		t.Fatalf("expected 2 instances, got %d", len(payload.Instances))
	}

	seen := map[string]bool{}
	var defaultName string
	for _, inst := range payload.Instances {
		seen[inst.Name] = true
		if inst.Default {
			defaultName = inst.Name
		}
	}
	if !seen["alpha"] || !seen["beta"] {
		t.Fatalf("expected both alpha and beta instances, got: %+v", payload.Instances)
	}
	if defaultName != "beta" {
		t.Fatalf("expected beta to be marked default, got %q", defaultName)
	}
}

func TestLogoutCommandWithYesFlagDeletesInstance(t *testing.T) {
	tempDir := t.TempDir()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodHead && r.URL.Path == "/":
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/users/current.json":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"user": map[string]any{"id": 1, "login": "tester"},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	env := map[string]string{"REDMINE_CONFIG_DIR": tempDir}
	loginInput := strings.Join([]string{server.URL, "secret-key", "team"}, "\n") + "\n"
	stdout, stderr, exitCode := runCommandWithEnvInput(env, loginInput, "login")
	if exitCode != 0 {
		t.Fatalf("expected login to succeed, got %d (stderr: %s)", exitCode, stderr)
	}
	if !strings.Contains(stdout, "登录成功") {
		t.Fatalf("expected login success output, got: %s", stdout)
	}

	stdout, stderr, exitCode = runCommandWithEnv(env, "--yes", "logout", "team")
	if exitCode != 0 {
		t.Fatalf("expected logout with --yes to succeed, got %d (stderr: %s)", exitCode, stderr)
	}
	if !strings.Contains(stdout, "已删除实例: team") {
		t.Fatalf("expected logout confirmation output, got: %s", stdout)
	}

	stdout, stderr, exitCode = runCommandWithEnv(env, "config", "get")
	if exitCode != 0 {
		t.Fatalf("expected config get to succeed after logout, got %d (stderr: %s)", exitCode, stderr)
	}
	if !strings.Contains(stdout, "No instances configured.") {
		t.Fatalf("expected config to be empty after logout, got: %s", stdout)
	}
}

func TestLogoutCommandConfirmationDeletesInstance(t *testing.T) {
	tempDir := t.TempDir()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodHead && r.URL.Path == "/":
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/users/current.json":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"user": map[string]any{"id": 1, "login": "tester"},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	env := map[string]string{"REDMINE_CONFIG_DIR": tempDir}
	loginInput := strings.Join([]string{server.URL, "secret-key", "team"}, "\n") + "\n"
	stdout, stderr, exitCode := runCommandWithEnvInput(env, loginInput, "login")
	if exitCode != 0 {
		t.Fatalf("expected login to succeed, got %d (stderr: %s)", exitCode, stderr)
	}
	if !strings.Contains(stdout, "登录成功") {
		t.Fatalf("expected login success output, got: %s", stdout)
	}

	stdout, stderr, exitCode = runCommandWithEnvInput(env, "yes\n", "logout", "team")
	if exitCode != 0 {
		t.Fatalf("expected logout confirmation to succeed, got %d (stderr: %s)", exitCode, stderr)
	}
	if !strings.Contains(stdout, "已删除实例: team") {
		t.Fatalf("expected logout confirmation output, got: %s", stdout)
	}

	stdout, stderr, exitCode = runCommandWithEnv(env, "config", "get")
	if exitCode != 0 {
		t.Fatalf("expected config get to succeed after confirmed logout, got %d (stderr: %s)", exitCode, stderr)
	}
	if !strings.Contains(stdout, "No instances configured.") {
		t.Fatalf("expected config to be empty after confirmed logout, got: %s", stdout)
	}
}
