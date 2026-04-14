// internal/app/logout_test.go
package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/largeoliu/redmine-cli/internal/config"
	"github.com/largeoliu/redmine-cli/internal/errors"
)

func TestNewLogoutCommand(t *testing.T) {
	flags := &GlobalFlags{}
	cmd := newLogoutCommand(flags)

	if cmd == nil {
		t.Fatal("expected logout command, got nil")
	}

	if cmd.Use != "logout" {
		t.Errorf("expected Use 'logout', got %s", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description, got empty")
	}

	if cmd.RunE == nil {
		t.Error("expected RunE function, got nil")
	}
}

func TestLogoutCommandExecuteWithInstanceArg(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("REDMINE_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("REDMINE_CONFIG_DIR")

	store := config.NewStore(tmpDir)
	if err := store.SaveInstance("instance1", config.Instance{
		URL:    "https://example.com",
		APIKey: "test-key",
	}); err != nil {
		t.Fatalf("SaveInstance() error = %v", err)
	}

	cmd := newLogoutCommand(&GlobalFlags{Yes: true})
	cmd.SetArgs([]string{"instance1"})
	cmd.SetContext(context.Background())

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if _, ok := cfg.Instances["instance1"]; ok {
		t.Error("instance1 should be deleted")
	}
}

func TestRunLogoutNoDefaultInstance(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("REDMINE_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("REDMINE_CONFIG_DIR")

	ctx := context.Background()
	flags := &GlobalFlags{}
	err := runLogout(ctx, flags, "")

	if err == nil {
		t.Error("expected error for no default instance")
	}

	appErr, ok := err.(*errors.Error)
	if !ok {
		t.Fatalf("expected *errors.Error, got %T", err)
	}
	if appErr.Category != errors.CategoryValidation {
		t.Errorf("expected validation error, got %s", appErr.Category)
	}
}

func TestRunLogoutInstanceNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("REDMINE_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("REDMINE_CONFIG_DIR")

	store := config.NewStore(tmpDir)
	if err := store.SaveInstance("default", config.Instance{
		URL:    "https://example.com",
		APIKey: "test-key",
	}); err != nil {
		t.Fatalf("SaveInstance() error = %v", err)
	}

	ctx := context.Background()
	flags := &GlobalFlags{}
	err := runLogout(ctx, flags, "nonexistent")

	if err == nil {
		t.Error("expected error for nonexistent instance")
	}

	appErr, ok := err.(*errors.Error)
	if !ok {
		t.Fatalf("expected *errors.Error, got %T", err)
	}
	if appErr.Category != errors.CategoryValidation {
		t.Errorf("expected validation error, got %s", appErr.Category)
	}
}

func TestRunLogoutLoadErrorWithoutInstanceName(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("REDMINE_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("REDMINE_CONFIG_DIR")

	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("invalid: yaml: ["), 0600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	err := runLogout(context.Background(), &GlobalFlags{}, "")
	if err == nil {
		t.Fatal("expected error for invalid config")
	}

	appErr, ok := err.(*errors.Error)
	if !ok {
		t.Fatalf("expected *errors.Error, got %T", err)
	}
	if appErr.Category != errors.CategoryInternal {
		t.Errorf("expected internal error, got %s", appErr.Category)
	}
}

func TestRunLogoutLoadErrorWithExplicitInstance(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("REDMINE_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("REDMINE_CONFIG_DIR")

	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("invalid: yaml: ["), 0600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	err := runLogout(context.Background(), &GlobalFlags{}, "default")
	if err == nil {
		t.Fatal("expected error for invalid config")
	}

	appErr, ok := err.(*errors.Error)
	if !ok {
		t.Fatalf("expected *errors.Error, got %T", err)
	}
	if appErr.Category != errors.CategoryInternal {
		t.Errorf("expected internal error, got %s", appErr.Category)
	}
}

func TestRunLogoutConfirmationCanceled(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("REDMINE_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("REDMINE_CONFIG_DIR")

	store := config.NewStore(tmpDir)
	if err := store.SaveInstance("default", config.Instance{
		URL:    "https://example.com",
		APIKey: "test-key",
	}); err != nil {
		t.Fatalf("SaveInstance() error = %v", err)
	}

	input := "no\n"
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	go func() {
		defer w.Close()
		_, _ = w.Write([]byte(input))
	}()

	ctx := context.Background()
	flags := &GlobalFlags{}
	err := runLogout(ctx, flags, "")

	os.Stdin = oldStdin

	if err != nil {
		t.Errorf("expected no error when canceled, got %v", err)
	}

	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if _, ok := cfg.Instances["default"]; !ok {
		t.Error("instance should not be deleted when confirmation is canceled")
	}
}

func TestRunLogoutConfirmationReadError(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("REDMINE_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("REDMINE_CONFIG_DIR")

	store := config.NewStore(tmpDir)
	if err := store.SaveInstance("default", config.Instance{
		URL:    "https://example.com",
		APIKey: "test-key",
	}); err != nil {
		t.Fatalf("SaveInstance() error = %v", err)
	}

	oldStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()
	if err := w.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	err = runLogout(context.Background(), &GlobalFlags{}, "")
	if err != nil {
		t.Errorf("expected no error when stdin read fails, got %v", err)
	}

	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if _, ok := cfg.Instances["default"]; !ok {
		t.Error("instance should not be deleted when confirmation read fails")
	}
}

func TestRunLogoutWithYesFlag(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("REDMINE_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("REDMINE_CONFIG_DIR")

	store := config.NewStore(tmpDir)
	if err := store.SaveInstance("default", config.Instance{
		URL:    "https://example.com",
		APIKey: "test-key",
	}); err != nil {
		t.Fatalf("SaveInstance() error = %v", err)
	}

	ctx := context.Background()
	flags := &GlobalFlags{Yes: true}
	err := runLogout(ctx, flags, "")

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if _, ok := cfg.Instances["default"]; ok {
		t.Error("instance should be deleted")
	}
}

func TestRunLogoutWithInstanceName(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("REDMINE_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("REDMINE_CONFIG_DIR")

	store := config.NewStore(tmpDir)
	if err := store.SaveInstance("instance1", config.Instance{
		URL:    "https://example1.com",
		APIKey: "key1",
	}); err != nil {
		t.Fatalf("SaveInstance(instance1) error = %v", err)
	}
	if err := store.SaveInstance("instance2", config.Instance{
		URL:    "https://example2.com",
		APIKey: "key2",
	}); err != nil {
		t.Fatalf("SaveInstance(instance2) error = %v", err)
	}
	if err := store.SetDefault("instance1"); err != nil {
		t.Fatalf("SetDefault() error = %v", err)
	}

	ctx := context.Background()
	flags := &GlobalFlags{Yes: true}
	err := runLogout(ctx, flags, "instance2")

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if _, ok := cfg.Instances["instance2"]; ok {
		t.Error("instance2 should be deleted")
	}
	if _, ok := cfg.Instances["instance1"]; !ok {
		t.Error("instance1 should still exist")
	}
}

func TestRunLogoutDefaultInstanceDeleted(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("REDMINE_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("REDMINE_CONFIG_DIR")

	store := config.NewStore(tmpDir)
	if err := store.SaveInstance("instance1", config.Instance{
		URL:    "https://example1.com",
		APIKey: "key1",
	}); err != nil {
		t.Fatalf("SaveInstance(instance1) error = %v", err)
	}
	if err := store.SaveInstance("instance2", config.Instance{
		URL:    "https://example2.com",
		APIKey: "key2",
	}); err != nil {
		t.Fatalf("SaveInstance(instance2) error = %v", err)
	}
	if err := store.SetDefault("instance1"); err != nil {
		t.Fatalf("SetDefault() error = %v", err)
	}

	ctx := context.Background()
	flags := &GlobalFlags{Yes: true}
	err := runLogout(ctx, flags, "instance1")

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if _, ok := cfg.Instances["instance1"]; ok {
		t.Error("instance1 should be deleted")
	}
	if cfg.Default != "instance2" {
		t.Errorf("expected default to be instance2, got %s", cfg.Default)
	}
}

func TestRunLogoutConfirmationYes(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("REDMINE_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("REDMINE_CONFIG_DIR")

	store := config.NewStore(tmpDir)
	if err := store.SaveInstance("default", config.Instance{
		URL:    "https://example.com",
		APIKey: "test-key",
	}); err != nil {
		t.Fatalf("SaveInstance() error = %v", err)
	}

	input := "yes\n"
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	go func() {
		defer w.Close()
		_, _ = w.Write([]byte(input))
	}()

	ctx := context.Background()
	flags := &GlobalFlags{}
	err := runLogout(ctx, flags, "")

	os.Stdin = oldStdin

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if _, ok := cfg.Instances["default"]; ok {
		t.Error("instance should be deleted after confirmation")
	}
}

func TestRunLogoutDeleteInstanceError(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("REDMINE_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("REDMINE_CONFIG_DIR")

	store := config.NewStore(tmpDir)
	if err := store.SaveInstance("default", config.Instance{
		URL:    "https://example.com",
		APIKey: "test-key",
	}); err != nil {
		t.Fatalf("SaveInstance() error = %v", err)
	}

	configPath := filepath.Join(tmpDir, "config.yaml")
	oldStdin := os.Stdin
	stdinR, stdinW, _ := os.Pipe()
	os.Stdin = stdinR
	defer func() { os.Stdin = oldStdin }()
	oldStdout := os.Stdout
	stdoutR, stdoutW, _ := os.Pipe()
	os.Stdout = stdoutW
	defer func() { os.Stdout = oldStdout }()

	syncErr := make(chan error, 1)

	go func() {
		defer stdinW.Close()

		var output strings.Builder
		buf := make([]byte, 64)
		prompt := "确定要删除此实例吗？请输入 \"yes\" 确认: "

		for {
			n, err := stdoutR.Read(buf)
			if n > 0 {
				output.Write(buf[:n])
				if strings.Contains(output.String(), prompt) {
					if err := os.Remove(configPath); err != nil {
						syncErr <- fmt.Errorf("remove config: %w", err)
						return
					}
					if err := os.Mkdir(configPath, 0755); err != nil {
						syncErr <- fmt.Errorf("replace config with directory: %w", err)
						return
					}
					if _, err := stdinW.Write([]byte("yes\n")); err != nil {
						syncErr <- fmt.Errorf("write confirmation: %w", err)
						return
					}
					syncErr <- nil
					return
				}
			}

			if err != nil {
				syncErr <- fmt.Errorf("prompt not observed: %w", err)
				return
			}
		}
	}()

	err := runLogout(context.Background(), &GlobalFlags{}, "default")
	_ = stdoutW.Close()
	if err := <-syncErr; err != nil {
		t.Fatalf("failed to synchronize delete failure: %v", err)
	}
	if err == nil {
		t.Fatal("expected error when delete instance load fails")
	}

	appErr, ok := err.(*errors.Error)
	if !ok {
		t.Fatalf("expected *errors.Error, got %T", err)
	}
	if appErr.Category != errors.CategoryInternal {
		t.Errorf("expected internal error, got %s", appErr.Category)
	}
}
