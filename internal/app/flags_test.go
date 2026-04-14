// internal/app/flags_test.go
package app

import (
	"context"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

func TestGlobalFlagsAlias(_ *testing.T) {
	// 测试 GlobalFlags �?types.GlobalFlags 的别�?
	var flags GlobalFlags
	var _ = flags
}

func TestGlobalFlagsAllFields(t *testing.T) {
	flags := &GlobalFlags{
		URL:      "https://example.com",
		Key:      "test-key",
		Format:   "json",
		JQ:       ".id",
		Fields:   "id,name",
		DryRun:   true,
		Yes:      true,
		Output:   "output.json",
		Limit:    10,
		Offset:   5,
		Timeout:  60 * time.Second,
		Retries:  5,
		Verbose:  true,
		Debug:    true,
		Instance: "production",
	}

	if flags.URL != "https://example.com" {
		t.Errorf("expected URL 'https://example.com', got %s", flags.URL)
	}
	if flags.Key != "test-key" {
		t.Errorf("expected Key 'test-key', got %s", flags.Key)
	}
	if flags.Format != "json" {
		t.Errorf("expected Format 'json', got %s", flags.Format)
	}
	if flags.JQ != ".id" {
		t.Errorf("expected JQ '.id', got %s", flags.JQ)
	}
	if flags.Fields != "id,name" {
		t.Errorf("expected Fields 'id,name', got %s", flags.Fields)
	}
	if !flags.DryRun {
		t.Error("expected DryRun to be true")
	}
	if !flags.Yes {
		t.Error("expected Yes to be true")
	}
	if flags.Output != "output.json" {
		t.Errorf("expected Output 'output.json', got %s", flags.Output)
	}
	if flags.Limit != 10 {
		t.Errorf("expected Limit 10, got %d", flags.Limit)
	}
	if flags.Offset != 5 {
		t.Errorf("expected Offset 5, got %d", flags.Offset)
	}
	if flags.Timeout != 60*time.Second {
		t.Errorf("expected Timeout 60s, got %v", flags.Timeout)
	}
	if flags.Retries != 5 {
		t.Errorf("expected Retries 5, got %d", flags.Retries)
	}
	if !flags.Verbose {
		t.Error("expected Verbose to be true")
	}
	if !flags.Debug {
		t.Error("expected Debug to be true")
	}
	if flags.Instance != "production" {
		t.Errorf("expected Instance 'production', got %s", flags.Instance)
	}
}

func TestGlobalFlagsParsing(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected GlobalFlags
	}{
		{
			name: "url and key flags",
			args: []string{"--url", "https://test.com", "--key", "secret-key"},
			expected: GlobalFlags{
				URL:     "https://test.com",
				Key:     "secret-key",
				Format:  "json",           // default
				Timeout: 30 * time.Second, // default
				Retries: 3,                // default
			},
		},
		{
			name: "format flag",
			args: []string{"--format", "table"},
			expected: GlobalFlags{
				Format:  "table",
				Timeout: 30 * time.Second, // default
				Retries: 3,                // default
			},
		},
		{
			name: "boolean flags",
			args: []string{"--dry-run", "--yes", "--verbose", "--debug"},
			expected: GlobalFlags{
				DryRun:  true,
				Yes:     true,
				Verbose: true,
				Debug:   true,
				Format:  "json",           // default
				Timeout: 30 * time.Second, // default
				Retries: 3,                // default
			},
		},
		{
			name: "numeric flags",
			args: []string{"--limit", "100", "--offset", "50", "--timeout", "120s", "--retries", "10"},
			expected: GlobalFlags{
				Limit:   100,
				Offset:  50,
				Timeout: 120 * time.Second,
				Retries: 10,
				Format:  "json", // default
			},
		},
		{
			name: "output flag",
			args: []string{"--output", "/tmp/result.json"},
			expected: GlobalFlags{
				Output:  "/tmp/result.json",
				Format:  "json",           // default
				Timeout: 30 * time.Second, // default
				Retries: 3,                // default
			},
		},
		{
			name: "jq and fields flags",
			args: []string{"--jq", ".issues[]", "--fields", "id,subject,status"},
			expected: GlobalFlags{
				JQ:      ".issues[]",
				Fields:  "id,subject,status",
				Format:  "json",           // default
				Timeout: 30 * time.Second, // default
				Retries: 3,                // default
			},
		},
		{
			name: "instance flag",
			args: []string{"--instance", "staging"},
			expected: GlobalFlags{
				Instance: "staging",
				Format:   "json",           // default
				Timeout:  30 * time.Second, // default
				Retries:  3,                // default
			},
		},
		{
			name: "shorthand flags",
			args: []string{"-u", "https://short.com", "-k", "short-key", "-f", "raw", "-y", "-v", "-o", "out.json", "-l", "25"},
			expected: GlobalFlags{
				URL:     "https://short.com",
				Key:     "short-key",
				Format:  "raw",
				Yes:     true,
				Verbose: true,
				Output:  "out.json",
				Limit:   25,
				Timeout: 30 * time.Second, // default
				Retries: 3,                // default
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := &GlobalFlags{}
			cmd := &cobra.Command{}
			bindGlobalFlags(cmd, flags)

			// 设置一个空运行函数
			cmd.Run = func(_ *cobra.Command, _ []string) {}

			cmd.SetArgs(tt.args)
			err := cmd.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// 验证解析的标志�?
			if flags.URL != tt.expected.URL {
				t.Errorf("URL: expected %s, got %s", tt.expected.URL, flags.URL)
			}
			if flags.Key != tt.expected.Key {
				t.Errorf("Key: expected %s, got %s", tt.expected.Key, flags.Key)
			}
			if flags.Format != tt.expected.Format {
				t.Errorf("Format: expected %s, got %s", tt.expected.Format, flags.Format)
			}
			if flags.JQ != tt.expected.JQ {
				t.Errorf("JQ: expected %s, got %s", tt.expected.JQ, flags.JQ)
			}
			if flags.Fields != tt.expected.Fields {
				t.Errorf("Fields: expected %s, got %s", tt.expected.Fields, flags.Fields)
			}
			if flags.DryRun != tt.expected.DryRun {
				t.Errorf("DryRun: expected %v, got %v", tt.expected.DryRun, flags.DryRun)
			}
			if flags.Yes != tt.expected.Yes {
				t.Errorf("Yes: expected %v, got %v", tt.expected.Yes, flags.Yes)
			}
			if flags.Output != tt.expected.Output {
				t.Errorf("Output: expected %s, got %s", tt.expected.Output, flags.Output)
			}
			if flags.Limit != tt.expected.Limit {
				t.Errorf("Limit: expected %d, got %d", tt.expected.Limit, flags.Limit)
			}
			if flags.Offset != tt.expected.Offset {
				t.Errorf("Offset: expected %d, got %d", tt.expected.Offset, flags.Offset)
			}
			if flags.Timeout != tt.expected.Timeout {
				t.Errorf("Timeout: expected %v, got %v", tt.expected.Timeout, flags.Timeout)
			}
			if flags.Retries != tt.expected.Retries {
				t.Errorf("Retries: expected %d, got %d", tt.expected.Retries, flags.Retries)
			}
			if flags.Verbose != tt.expected.Verbose {
				t.Errorf("Verbose: expected %v, got %v", tt.expected.Verbose, flags.Verbose)
			}
			if flags.Debug != tt.expected.Debug {
				t.Errorf("Debug: expected %v, got %v", tt.expected.Debug, flags.Debug)
			}
			if flags.Instance != tt.expected.Instance {
				t.Errorf("Instance: expected %s, got %s", tt.expected.Instance, flags.Instance)
			}
		})
	}
}

func TestGlobalFlagsDefaults(t *testing.T) {
	flags := &GlobalFlags{}
	cmd := &cobra.Command{}
	bindGlobalFlags(cmd, flags)

	// 不设置任何参数，检查默认�?
	cmd.Run = func(_ *cobra.Command, _ []string) {}
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 检查默认�?
	if flags.Format != "json" {
		t.Errorf("expected default Format 'json', got %s", flags.Format)
	}
	if flags.Timeout != 30*time.Second {
		t.Errorf("expected default Timeout 30s, got %v", flags.Timeout)
	}
	if flags.Retries != 3 {
		t.Errorf("expected default Retries 3, got %d", flags.Retries)
	}
	if flags.DryRun {
		t.Error("expected default DryRun to be false")
	}
	if flags.Yes {
		t.Error("expected default Yes to be false")
	}
	if flags.Verbose {
		t.Error("expected default Verbose to be false")
	}
	if flags.Debug {
		t.Error("expected default Debug to be false")
	}
}

func TestGlobalFlagsWithRootCommand(t *testing.T) {
	// 简化测试，只验证命令可以正确解析标�?
	tests := []struct {
		name string
		args []string
	}{
		{"with url shorthand", []string{"-u", "https://example.com", "--help"}},
		{"with key shorthand", []string{"-k", "test-key", "--help"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newTestContext()
			root := NewRootCommand(ctx)
			root.SetArgs(tt.args)

			// 执行命令（help 子命令不会报错）
			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestGlobalFlagsPersistence(t *testing.T) {
	// 测试持久化标志是否可以被子命令继�?
	flags := &GlobalFlags{}
	root := &cobra.Command{Use: "root"}
	bindGlobalFlags(root, flags)

	// 创建子命�?
	child := &cobra.Command{
		Use: "child",
		Run: func(_ *cobra.Command, _ []string) {},
	}
	root.AddCommand(child)

	// 在子命令上设置标�?
	root.SetArgs([]string{"child", "--url", "https://child.com", "--key", "child-key"})
	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 验证标志值被正确设置
	if flags.URL != "https://child.com" {
		t.Errorf("expected URL 'https://child.com', got %s", flags.URL)
	}
	if flags.Key != "child-key" {
		t.Errorf("expected Key 'child-key', got %s", flags.Key)
	}
}

func TestGlobalFlagsInvalidValues(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "invalid limit (negative)",
			args:    []string{"--limit", "-1"},
			wantErr: false, // cobra 不会验证负数
		},
		{
			name:    "invalid timeout (negative)",
			args:    []string{"--timeout", "-10s"},
			wantErr: false, // DurationVar accepts negative durations
		},
		{
			name:    "invalid format",
			args:    []string{"--format", "invalid"},
			wantErr: false, // 格式验证�?ResolveFormat 中进�?
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := &GlobalFlags{}
			cmd := &cobra.Command{}
			bindGlobalFlags(cmd, flags)
			cmd.Run = func(_ *cobra.Command, _ []string) {}

			cmd.SetArgs(tt.args)
			err := cmd.Execute()

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// 辅助函数
func newTestContext() context.Context {
	// 返回一个简单的 context 用于测试
	return context.Background()
}
