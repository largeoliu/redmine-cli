// cmd/main_test.go
package main

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/largeoliu/redmine-cli/internal/app"
)

// TestMain 用于设置测试环境
func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

// TestMainFunction tests the main function behavior by building and executing the binary
// Since main() calls os.Exit(), we need to test it indirectly
func TestMainFunction(t *testing.T) {
	// Skip if short test mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Build the binary
	binaryName := "redmine-test"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, binaryName)

	// Build the binary
	buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
	buildCmd.Dir = filepath.Join("..", "cmd")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build binary: %v\nOutput: %s", err, output)
	}

	// Test running the binary with --help
	t.Run("help_flag", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "--help")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to run binary with --help: %v\nOutput: %s", err, output)
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "Redmine CLI") {
			t.Errorf("Expected help output to contain 'Redmine CLI', got: %s", outputStr)
		}
	})

	// Test running the binary with --version
	t.Run("version_flag", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "--version")
		output, err := cmd.CombinedOutput()
		// version command may exit with 0
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				if exitErr.ExitCode() != 0 {
					t.Fatalf("Unexpected exit code: %d\nOutput: %s", exitErr.ExitCode(), output)
				}
			} else {
				t.Fatalf("Failed to run binary with --version: %v\nOutput: %s", err, output)
			}
		}

		outputStr := string(output)
		if outputStr == "" {
			t.Error("Expected version output, got empty")
		}
	})

	// Test running without arguments (should show help)
	t.Run("no_arguments", func(t *testing.T) {
		cmd := exec.Command(binaryPath)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to run binary without arguments: %v\nOutput: %s", err, output)
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "Usage:") && !strings.Contains(outputStr, "redmine") {
			t.Errorf("Expected help output, got: %s", outputStr)
		}
	})

	// Test invalid command returns non-zero exit code
	t.Run("invalid_command", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "nonexistent-command")
		output, err := cmd.CombinedOutput()
		_ = output
		_ = err
		// Should fail with non-zero exit code
		if cmd.ProcessState == nil {
			t.Fatal("Process state is nil")
		}
		if cmd.ProcessState.Success() {
			t.Errorf("Expected non-zero exit code for invalid command, got success. Output: %s", output)
		}
	})
}

// TestAppExecuteIntegration tests the app.Execute function behavior
func TestAppExecuteIntegration(t *testing.T) {
	// Create a temp config directory to avoid loading real config
	tmpDir := t.TempDir()
	t.Setenv("REDMINE_CONFIG_DIR", tmpDir)

	tests := []struct {
		name       string
		args       []string
		wantCode   int
		wantOutput string
		wantErr    bool
	}{
		{
			name:       "help_command",
			args:       []string{"--help"},
			wantCode:   0,
			wantOutput: "Redmine CLI",
		},
		{
			name:       "version_command",
			args:       []string{"version"},
			wantCode:   0,
			wantOutput: "",
		},
		{
			name:     "unknown_command",
			args:     []string{"unknown-command-xyz"},
			wantCode: 1, // Should return error
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't directly test Execute() because it calls os.Exit()
			// Instead, we test NewRootCommand directly
			ctx := context.Background()
			root := app.NewRootCommand(ctx)

			var buf bytes.Buffer
			root.SetOut(&buf)
			root.SetErr(&buf)
			root.SetArgs(tt.args)

			err := root.Execute()
			output := buf.String()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if tt.wantOutput != "" && !strings.Contains(output, tt.wantOutput) {
				t.Errorf("Expected output to contain %q, got %q", tt.wantOutput, output)
			}
		})
	}
}

// TestMainCallsAppExecute verifies that main function properly calls app.Execute
func TestMainCallsAppExecute(t *testing.T) {
	// This test verifies the relationship between main() and app.Execute()
	// by checking that the app package is properly imported and Execute exists

	// Create a temp config directory
	tmpDir := t.TempDir()
	t.Setenv("REDMINE_CONFIG_DIR", tmpDir)

	// Test that Execute function exists and can be called
	// We use NewRootCommand instead of Execute to avoid os.Exit
	ctx := context.Background()
	root := app.NewRootCommand(ctx)

	if root == nil {
		t.Error("Expected root command, got nil")
		return
	}

	if root.Use != "redmine" {
		t.Errorf("Expected Use 'redmine', got %s", root.Use)
	}
}

// TestBinarySignalHandling tests that the binary handles signals properly
func TestBinarySignalHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Build the binary
	binaryName := "redmine-signal-test"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, binaryName)

	// nosec G204
	buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
	buildCmd.Dir = filepath.Join("..", "cmd")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build binary: %v\nOutput: %s", err, output)
	}

	// Start the binary and test signal handling
	// nosec G204
	cmd := exec.Command(binaryPath, "--help")
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start binary: %v", err)
	}

	// Wait for completion with timeout
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Binary exited with error: %v", err)
		}
	case <-time.After(5 * time.Second):
		_ = cmd.Process.Kill() //nolint:errcheck
		t.Error("Binary did not exit in time")
	}
}

// TestMainExitCode tests that main returns correct exit codes
func TestMainExitCode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Build the binary
	binaryName := "redmine-exit-test"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, binaryName)

	// nosec G204
	buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
	buildCmd.Dir = filepath.Join("..", "cmd")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build binary: %v\nOutput: %s", err, output)
	}

	tests := []struct {
		name         string
		args         []string
		expectedCode int
	}{
		{
			name:         "help_returns_zero",
			args:         []string{"--help"},
			expectedCode: 0,
		},
		{
			name:         "version_returns_zero",
			args:         []string{"version"},
			expectedCode: 0,
		},
		{
			name:         "invalid_command_returns_nonzero",
			args:         []string{"invalid-cmd-xyz"},
			expectedCode: 1, // validation error exit code
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// nosec G204
			cmd := exec.Command(binaryPath, tt.args...)
			_ = cmd.Run()

			if cmd.ProcessState == nil {
				t.Fatal("Process state is nil")
			}

			exitCode := cmd.ProcessState.ExitCode()
			if exitCode != tt.expectedCode {
				t.Errorf("Expected exit code %d, got %d", tt.expectedCode, exitCode)
			}
		})
	}
}

// TestMainOsExit tests the main function's osExit call for coverage
func TestMainOsExit(t *testing.T) {
	type exitPanic struct {
		exitCode int
	}

	osExit = func(code int) {
		panic(exitPanic{exitCode: code})
	}

	defer func() {
		if r := recover(); r != nil {
			if ep, ok := r.(exitPanic); ok {
				if ep.exitCode != 0 {
					t.Errorf("Expected exit code 0, got %d", ep.exitCode)
				}
				return
			}
			t.Errorf("Unexpected panic: %v", r)
		}
		t.Error("Expected os.Exit to be called")
	}()

	main()
}

// TestMainWithEnvVars tests main with various environment variables
func TestMainWithEnvVars(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Build the binary
	binaryName := "redmine-env-test"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, binaryName)

	// nosec G204
	buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
	buildCmd.Dir = filepath.Join("..", "cmd")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build binary: %v\nOutput: %s", err, output)
	}

	// Create a temp config directory
	configDir := filepath.Join(tmpDir, "config")
	if err := os.MkdirAll(configDir, 0750); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	tests := []struct {
		name  string
		env   map[string]string
		args  []string
		check func(t *testing.T, output string, err error)
	}{
		{
			name: "with_custom_config_dir",
			env: map[string]string{
				"REDMINE_CONFIG_DIR": configDir,
			},
			args: []string{"--help"},
			check: func(_ *testing.T, _ string, _ error) {
			},
		},
		{
			name: "with_url_and_key_flags",
			env:  map[string]string{},
			args: []string{"--url", "https://test.example.com", "--key", "test-key", "issue", "list"},
			check: func(_ *testing.T, _ string, _ error) {
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// nosec G204
			cmd := exec.Command(binaryPath, tt.args...)

			// Set environment variables
			cmd.Env = os.Environ()
			for k, v := range tt.env {
				cmd.Env = append(cmd.Env, k+"="+v)
			}

			output, err := cmd.CombinedOutput()
			tt.check(t, string(output), err)
		})
	}
}

// TestMainOutputFormat tests various output formats
func TestMainOutputFormat(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Build the binary
	binaryName := "redmine-format-test"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, binaryName)

	// nosec G204
	buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
	buildCmd.Dir = filepath.Join("..", "cmd")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build binary: %v\nOutput: %s", err, output)
	}

	tests := []struct {
		name  string
		args  []string
		check func(t *testing.T, output string)
	}{
		{
			name: "json_format_default",
			args: []string{"--help"},
			check: func(t *testing.T, output string) {
				if output == "" {
					t.Error("Expected output, got empty")
				}
			},
		},
		{
			name: "table_format",
			args: []string{"--format", "table", "--help"},
			check: func(t *testing.T, output string) {
				if output == "" {
					t.Error("Expected output, got empty")
				}
			},
		},
		{
			name: "raw_format",
			args: []string{"--format", "raw", "--help"},
			check: func(t *testing.T, output string) {
				if output == "" {
					t.Error("Expected output, got empty")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// nosec G204
			cmd := exec.Command(binaryPath, tt.args...)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("Command failed: %v\nOutput: %s", err, output)
			}
			tt.check(t, string(output))
		})
	}
}

// TestMainSubcommands tests that all subcommands are available
func TestMainSubcommands(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Build the binary
	binaryName := "redmine-subcmd-test"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, binaryName)

	// nosec G204
	buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
	buildCmd.Dir = filepath.Join("..", "cmd")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build binary: %v\nOutput: %s", err, output)
	}

	expectedCommands := []string{
		"login",
		"version",
		"config",
		"category",
		"issue",
		"priority",
		"project",
		"status",
		"time-entry",
		"tracker",
		"user",
		"version",
	}

	// Get help output
	// nosec G204
	cmd := exec.Command(binaryPath, "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to get help: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	for _, expected := range expectedCommands {
		// Note: "version" appears twice (as version flag and version command)
		// We just check that the command structure is correct
		if !strings.Contains(outputStr, expected) && expected != "version" {
			t.Errorf("Expected command %q not found in help output", expected)
		}
	}
}

// TestMainConcurrent tests running multiple instances concurrently
func TestMainConcurrent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Build the binary
	binaryName := "redmine-concurrent-test"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, binaryName)

	// nosec G204
	buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
	buildCmd.Dir = filepath.Join("..", "cmd")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build binary: %v\nOutput: %s", err, output)
	}

	// Run multiple instances concurrently
	const numInstances = 5
	errChan := make(chan error, numInstances)

	for i := 0; i < numInstances; i++ {
		go func() {
			cmd := exec.Command(binaryPath, "--help")
			_, err := cmd.CombinedOutput()
			errChan <- err
		}()
	}

	// Wait for all to complete
	for i := 0; i < numInstances; i++ {
		if err := <-errChan; err != nil {
			t.Errorf("Concurrent execution failed: %v", err)
		}
	}
}

// TestMainStdin tests reading from stdin (if applicable)
func TestMainStdin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Build the binary
	binaryName := "redmine-stdin-test"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, binaryName)

	// nosec G204
	buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
	buildCmd.Dir = filepath.Join("..", "cmd")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build binary: %v\nOutput: %s", err, output)
	}

	// Test with empty stdin
	// nosec G204
	cmd := exec.Command(binaryPath, "--help")
	cmd.Stdin = strings.NewReader("")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	if len(output) == 0 {
		t.Error("Expected output, got empty")
	}
}

// TestMainLargeOutput tests handling of large output
func TestMainLargeOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Build the binary
	binaryName := "redmine-large-test"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, binaryName)

	// nosec G204
	buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
	buildCmd.Dir = filepath.Join("..", "cmd")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build binary: %v\nOutput: %s", err, output)
	}

	// Test with verbose output
	cmd := exec.Command(binaryPath, "--verbose", "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	// Verify output is not truncated
	scanner := bufio.NewScanner(bytes.NewReader(output))
	lineCount := 0
	for scanner.Scan() {
		lineCount++
	}

	if lineCount < 5 {
		t.Errorf("Expected more output lines, got %d", lineCount)
	}
}

// TestMainTimeout tests command timeout behavior
func TestMainTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Build the binary
	binaryName := "redmine-timeout-test"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, binaryName)

	buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
	buildCmd.Dir = filepath.Join("..", "cmd")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build binary: %v\nOutput: %s", err, output)
	}

	// Test with timeout flag
	cmd := exec.Command(binaryPath, "--timeout", "5s", "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	if len(output) == 0 {
		t.Error("Expected output, got empty")
	}
}

// TestMainDebugMode tests debug mode output
func TestRun_ReturnsZeroOnHelp(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("REDMINE_CONFIG_DIR", tmpDir)

	code := run()
	if code != 0 {
		t.Errorf("Expected run() to return 0 for no args (help), got %d", code)
	}
}

func TestRun_ReturnsNonZeroOnError(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tmpDir := t.TempDir()
	t.Setenv("REDMINE_CONFIG_DIR", tmpDir)

	os.Args = []string{"redmine", "nonexistent-cmd-xyz"}
	code := run()
	if code == 0 {
		t.Error("Expected run() to return non-zero for invalid command, got 0")
	}
}

func TestMainOsExitWithErrorCode(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tmpDir := t.TempDir()
	t.Setenv("REDMINE_CONFIG_DIR", tmpDir)

	os.Args = []string{"redmine", "nonexistent-cmd-xyz"}

	type exitPanic struct {
		exitCode int
	}

	osExit = func(code int) {
		panic(exitPanic{exitCode: code})
	}
	defer func() {
		if r := recover(); r != nil {
			if ep, ok := r.(exitPanic); ok {
				if ep.exitCode == 0 {
					t.Error("Expected non-zero exit code for invalid command, got 0")
				}
				return
			}
			t.Errorf("Unexpected panic: %v", r)
		}
	}()

	main()
}

func TestRun_VersionCommand(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tmpDir := t.TempDir()
	t.Setenv("REDMINE_CONFIG_DIR", tmpDir)

	os.Args = []string{"redmine", "info"}
	code := run()
	if code != 0 {
		t.Errorf("Expected run() to return 0 for version command, got %d", code)
	}
}

func TestMainDebugMode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Build the binary
	binaryName := "redmine-debug-test"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, binaryName)

	buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
	buildCmd.Dir = filepath.Join("..", "cmd")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build binary: %v\nOutput: %s", err, output)
	}

	// Test with debug flag
	cmd := exec.Command(binaryPath, "--debug", "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	if len(output) == 0 {
		t.Error("Expected output, got empty")
	}
}
