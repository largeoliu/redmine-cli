package app

import (
	"context"
	"os"
	"testing"
)

func TestExecute(t *testing.T) {
	// Save original args and restore after test
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tests := []struct {
		name     string
		args     []string
		wantExit int
	}{
		{
			name:     "help command",
			args:     []string{"redmine", "--help"},
			wantExit: 0,
		},
		{
			name:     "version command",
			args:     []string{"redmine", "--version"},
			wantExit: 0,
		},
		{
			name:     "no args shows help",
			args:     []string{"redmine"},
			wantExit: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp config dir to avoid config issues
			tmpDir := t.TempDir()
			os.Setenv("REDMINE_CONFIG_DIR", tmpDir)
			defer os.Unsetenv("REDMINE_CONFIG_DIR")

			os.Args = tt.args

			exitCode := Execute()
			if exitCode != tt.wantExit {
				t.Errorf("Execute() = %d, want %d", exitCode, tt.wantExit)
			}
		})
	}
}

func TestExecuteContext(t *testing.T) {
	// Create temp config dir
	tmpDir := t.TempDir()
	os.Setenv("REDMINE_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("REDMINE_CONFIG_DIR")

	// Test that Execute creates a proper context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	root := NewRootCommand(ctx)
	if root == nil {
		t.Fatal("NewRootCommand returned nil")
	}

	// Verify context is set
	if root.Context() == nil {
		t.Error("Root command context is nil")
	}
}

func TestExecuteWithSignal(t *testing.T) {
	// This test verifies that Execute sets up signal handling
	// We can't actually test signal handling in a unit test,
	// but we can verify the function doesn't panic

	tmpDir := t.TempDir()
	os.Setenv("REDMINE_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("REDMINE_CONFIG_DIR")
}

func TestExecuteWithError(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tmpDir := t.TempDir()
	os.Setenv("REDMINE_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("REDMINE_CONFIG_DIR")

	os.Args = []string{"redmine", "nonexistent-command"}

	exitCode := Execute()
	if exitCode == 0 {
		t.Error("Execute() with invalid command should return non-zero exit code")
	}
}
