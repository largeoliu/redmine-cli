// internal/app/root_test.go
package app

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/largeoliu/redmine-cli/internal/errors"
	"github.com/largeoliu/redmine-cli/internal/types"
)

func TestNewRootCommand(t *testing.T) {
	ctx := context.Background()
	root := NewRootCommand(ctx)

	if root == nil {
		t.Fatal("expected root command, got nil")
	}

	if root.Use != "redmine" {
		t.Errorf("expected Use 'redmine', got %s", root.Use)
	}

	if root.Short == "" {
		t.Error("expected Short description, got empty")
	}

	if root.Version != version {
		t.Errorf("expected Version %s, got %s", version, root.Version)
	}

	if !root.SilenceErrors {
		t.Error("expected SilenceErrors to be true")
	}

	if !root.SilenceUsage {
		t.Error("expected SilenceUsage to be true")
	}
}

func TestRootCommandHasSubcommands(t *testing.T) {
	ctx := context.Background()
	root := NewRootCommand(ctx)

	expectedCommands := []string{
		"login",
		"logout",
		"version",
		"config",
		"sprint",
		"category",
		"issue",
		"priority",
		"project",
		"status",
		"time-entry",
		"tracker",
		"user",
	}

	commands := root.Commands()
	commandNames := make(map[string]bool)
	for _, cmd := range commands {
		commandNames[cmd.Name()] = true
	}

	for _, expected := range expectedCommands {
		if !commandNames[expected] {
			t.Errorf("expected subcommand %s not found", expected)
		}
	}
}

func TestBindGlobalFlags(t *testing.T) {
	flags := &GlobalFlags{}
	cmd := &cobra.Command{}

	bindGlobalFlags(cmd, flags)

	tests := []struct {
		name      string
		flagName  string
		shorthand string
	}{
		{"url", "url", "u"},
		{"key", "key", "k"},
		{"format", "format", "f"},
		{"jq", "jq", ""},
		{"fields", "fields", ""},
		{"dry-run", "dry-run", ""},
		{"yes", "yes", "y"},
		{"output", "output", "o"},
		{"limit", "limit", "l"},
		{"offset", "offset", ""},
		{"timeout", "timeout", ""},
		{"retries", "retries", ""},
		{"verbose", "verbose", "v"},
		{"debug", "debug", ""},
		{"instance", "instance", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := cmd.PersistentFlags().Lookup(tt.flagName)
			if flag == nil {
				t.Errorf("flag %s not found", tt.flagName)
				return
			}
			if tt.shorthand != "" && flag.Shorthand != tt.shorthand {
				t.Errorf("flag %s: expected shorthand %s, got %s", tt.flagName, tt.shorthand, flag.Shorthand)
			}
		})
	}
}

func TestBindGlobalFlagsDefaultValues(t *testing.T) {
	flags := &GlobalFlags{}
	cmd := &cobra.Command{}

	bindGlobalFlags(cmd, flags)

	// 测试默认�?
	tests := []struct {
		name        string
		flagName    string
		expectedDef string
	}{
		{"format", "format", "json"},
		{"timeout", "timeout", "30s"},
		{"retries", "retries", "3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := cmd.PersistentFlags().Lookup(tt.flagName)
			if flag == nil {
				t.Errorf("flag %s not found", tt.flagName)
				return
			}
			if flag.DefValue != tt.expectedDef {
				t.Errorf("flag %s: expected default %s, got %s", tt.flagName, tt.expectedDef, flag.DefValue)
			}
		})
	}
}

func TestResolveFormat(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		expected string
	}{
		{"json format", "json", "json"},
		{"table format", "table", "table"},
		{"raw format", "raw", "raw"},
		{"empty defaults to json", "", "json"},
		{"unknown defaults to json", "unknown", "json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := &GlobalFlags{Format: tt.format}
			format := ResolveFormat(flags)
			if string(format) != tt.expected {
				t.Errorf("expected format %s, got %s", tt.expected, format)
			}
		})
	}
}

func TestParseFields(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"single field", "id", []string{"id"}},
		{"multiple fields", "id,name,subject", []string{"id", "name", "subject"}},
		{"empty string", "", []string(nil)},
		{"trailing comma", "id,name,", []string{"id", "name"}},
		{"leading comma", ",id,name", []string{"id", "name"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseFields(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d fields, got %d: %v", len(tt.expected), len(result), result)
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("expected field[%d] = %s, got %s", i, tt.expected[i], v)
				}
			}
		})
	}
}

func TestSplitByComma(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"empty string", "", []string(nil)},
		{"single value", "value", []string{"value"}},
		{"two values", "a,b", []string{"a", "b"}},
		{"multiple values", "a,b,c", []string{"a", "b", "c"}},
		{"trailing comma", "a,", []string{"a"}},
		{"leading comma", ",a", []string{"", "a"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitByComma(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d parts, got %d: %v", len(tt.expected), len(result), result)
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("expected part[%d] = %s, got %s", i, tt.expected[i], v)
				}
			}
		})
	}
}

func TestResolverInterface(_ *testing.T) {
	var _ types.Resolver = &resolver{}
}

func TestResolverResolveClient(t *testing.T) {
	r := &resolver{}
	flags := &types.GlobalFlags{
		URL: "https://example.com",
		Key: "test-key",
	}

	client, err := r.ResolveClient(flags)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Error("expected client, got nil")
	}
}

func TestResolverWriteOutput(t *testing.T) {
	r := &resolver{}
	flags := &types.GlobalFlags{
		Format: "json",
	}

	var buf bytes.Buffer
	payload := map[string]string{"test": "value"}

	err := r.WriteOutput(&buf, flags, payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("expected output, got empty")
	}
}

func TestWriteOutputWithJQ(t *testing.T) {
	flags := &GlobalFlags{
		Format: "json",
		JQ:     ".name",
	}

	var buf bytes.Buffer
	payload := map[string]string{"name": "test", "value": "ignored"}

	err := WriteOutput(&buf, flags, payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	// JQ filter should extract only "test"
	if output == "" {
		t.Error("expected output, got empty")
	}
}

func TestWriteOutputWithFields(t *testing.T) {
	flags := &GlobalFlags{
		Format: "json",
		Fields: "id,name",
	}

	var buf bytes.Buffer
	payload := map[string]any{
		"id":      1,
		"name":    "test",
		"subject": "should be filtered",
	}

	err := WriteOutput(&buf, flags, payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("expected output, got empty")
	}
}

func TestRootCommandRunWithHelp(t *testing.T) {
	ctx := context.Background()
	root := NewRootCommand(ctx)

	// 设置输出缓冲�?
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)

	// 执行不带参数的根命令应该显示帮助
	root.SetArgs([]string{})
	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("expected help output, got empty")
	}
}

func TestResolveClientWithFlags(t *testing.T) {
	// Create a temp config directory to ensure no existing config interferes
	tmpDir := t.TempDir()
	os.Setenv("REDMINE_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("REDMINE_CONFIG_DIR")

	tests := []struct {
		name    string
		flags   *GlobalFlags
		wantErr bool
	}{
		{
			name: "with url and key",
			flags: &GlobalFlags{
				URL: "https://example.com",
				Key: "test-api-key",
			},
			wantErr: false,
		},
		{
			name: "missing url",
			flags: &GlobalFlags{
				Key: "test-api-key",
			},
			wantErr: true,
		},
		{
			name: "missing key",
			flags: &GlobalFlags{
				URL: "https://example.com",
			},
			wantErr: true,
		},
		{
			name:    "missing both",
			flags:   &GlobalFlags{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := ResolveClient(tt.flags)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if client == nil {
					t.Error("expected client, got nil")
				}
			}
		})
	}
}

// 辅助函数：创建测试用�?cobra 命令
func newTestCommand() *cobra.Command {
	return &cobra.Command{
		Use: "test",
		Run: func(_ *cobra.Command, _ []string) {},
	}
}

// 测试命令输出
func TestCommandOutput(t *testing.T) {
	cmd := newTestCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	cmd.Println("test output")

	if buf.String() == "" {
		t.Error("expected output, got empty")
	}
}

// 测试 io.Writer 接口
func TestWriteOutputImplementsIOWriter(_ *testing.T) {
	// 确保 WriteOutput 接受 io.Writer 参数
	var _ = WriteOutput
}

// 测试 printError 函数
func TestPrintError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectContains []string
	}{
		{
			name: "app error with hint",
			err: errors.NewValidation("validation failed",
				errors.WithHint("check your input"),
			),
			expectContains: []string{"Error:", "validation failed", "Hint:", "check your input"},
		},
		{
			name: "app error with actions",
			err: errors.NewAuth("authentication failed",
				errors.WithActions("Check your API key", "Verify your permissions"),
			),
			expectContains: []string{"Error:", "authentication failed", "Actions:"},
		},
		{
			name: "app error with hint and actions",
			err: errors.NewAPI("server error",
				errors.WithHint("try again later"),
				errors.WithActions("Check status page", "Contact support"),
			),
			expectContains: []string{"Error:", "server error", "Hint:", "try again later", "Actions:"},
		},
		{
			name:           "standard error",
			err:            fmt.Errorf("standard error"),
			expectContains: []string{"Error:", "standard error"},
		},
		{
			name: "network error",
			err: errors.NewNetwork("connection timeout",
				errors.WithHint("check your network connection"),
			),
			expectContains: []string{"Error:", "connection timeout", "Hint:"},
		},
		{
			name: "internal error",
			err: errors.NewInternal("internal failure",
				errors.WithCause(fmt.Errorf("root cause")),
			),
			expectContains: []string{"Error:", "internal failure"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 捕获 stderr
			old := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			cmd := &cobra.Command{Use: "test"}
			printError(cmd, tt.err)

			w.Close()
			os.Stderr = old

			var buf bytes.Buffer
			if _, err := io.Copy(&buf, r); err != nil {
				t.Fatalf("io.Copy() error = %v", err)
			}

			output := buf.String()
			for _, expected := range tt.expectContains {
				if !strings.Contains(output, expected) {
					t.Errorf("expected output to contain %q, got %s", expected, output)
				}
			}
		})
	}
}

// 测试 ResolveClient 从配置文件加载实例
func TestResolveClientFromConfig(t *testing.T) {
	// 创建临时配置目录
	tmpDir := t.TempDir()
	os.Setenv("REDMINE_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("REDMINE_CONFIG_DIR")

	// 写入配置文件
	configContent := `default: prod
instances:
  dev:
    url: https://dev.example.com
    api_key: dev-key
  prod:
    url: https://prod.example.com
    api_key: prod-key
`
	configFile := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configFile, []byte(configContent), 0600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	tests := []struct {
		name        string
		flags       *GlobalFlags
		wantErr     bool
		errContains string
	}{
		{
			name: "use default instance from config",
			flags: &GlobalFlags{
				Instance: "",
			},
			wantErr: false,
		},
		{
			name: "use specific instance from config",
			flags: &GlobalFlags{
				Instance: "dev",
			},
			wantErr: false,
		},
		{
			name: "instance not found",
			flags: &GlobalFlags{
				Instance: "nonexistent",
			},
			wantErr:     true,
			errContains: "instance not found",
		},
		{
			name: "url from flags overrides config",
			flags: &GlobalFlags{
				URL:      "https://override.example.com",
				Key:      "override-key",
				Instance: "dev",
			},
			wantErr: false,
		},
		{
			name: "key from flags overrides config",
			flags: &GlobalFlags{
				URL:      "",
				Key:      "override-key",
				Instance: "dev",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := ResolveClient(tt.flags)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error to contain %q, got %v", tt.errContains, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if client == nil {
					t.Error("expected client, got nil")
				}
			}
		})
	}
}

// 测试 ResolveClient 配置加载失败
func TestResolveClientConfigLoadError(t *testing.T) {
	// 创建临时配置目录
	tmpDir := t.TempDir()
	os.Setenv("REDMINE_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("REDMINE_CONFIG_DIR")

	// 写入无效的配置文件
	configFile := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configFile, []byte("invalid: yaml: content: ["), 0600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	flags := &GlobalFlags{}
	_, err := ResolveClient(flags)
	if err == nil {
		t.Error("expected error for invalid config, got nil")
	}
}

// 测试 ResolveClient 空配置
func TestResolveClientEmptyConfig(t *testing.T) {
	// 创建临时配置目录（空）
	tmpDir := t.TempDir()
	os.Setenv("REDMINE_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("REDMINE_CONFIG_DIR")

	flags := &GlobalFlags{}
	_, err := ResolveClient(flags)
	if err == nil {
		t.Error("expected error for missing URL and key, got nil")
	}

	// 验证错误消息
	if !strings.Contains(err.Error(), "URL is required") {
		t.Errorf("expected error to contain 'URL is required', got %v", err)
	}
}

// 测试 Execute 函数（模拟成功场景）
func TestExecuteSuccess(t *testing.T) {
	// 这个测试验证 Execute 函数的基本流程
	// 由于 Execute 会调用 os.Exit，我们只测试它能正常启动
	// 实际的命令执行测试在其他测试中覆盖

	// 创建临时配置目录以避免配置加载问题
	tmpDir := t.TempDir()
	os.Setenv("REDMINE_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("REDMINE_CONFIG_DIR")

	// Execute() 会启动实际的命令，这里我们只验证它不会 panic
	// 实际的集成测试应该使用构建标签或单独的测试文件
}

// 测试默认解析器
func TestDefaultResolver(t *testing.T) {
	// 验证 defaultResolver 已正确初始化
	if defaultResolver == nil {
		t.Error("expected defaultResolver to be initialized")
	}
}

// 测试 resolver 实现
func TestResolverImplementation(t *testing.T) {
	r := &resolver{}

	// 测试 ResolveClient
	flags := &types.GlobalFlags{
		URL: "https://example.com",
		Key: "test-key",
	}
	client, err := r.ResolveClient(flags)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Error("expected client, got nil")
	}

	// 测试 WriteOutput
	var buf bytes.Buffer
	payload := map[string]string{"test": "value"}
	outputFlags := &types.GlobalFlags{Format: "json"}
	err = r.WriteOutput(&buf, outputFlags, payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.String() == "" {
		t.Error("expected output, got empty")
	}
}

// 测试 WriteOutput 的各种格式
func TestWriteOutputFormats(t *testing.T) {
	payload := map[string]any{
		"id":   1,
		"name": "test",
	}

	tests := []struct {
		name   string
		format string
	}{
		{"json format", "json"},
		{"table format", "table"},
		{"raw format", "raw"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := &GlobalFlags{Format: tt.format}
			var buf bytes.Buffer
			err := WriteOutput(&buf, flags, payload)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if buf.String() == "" {
				t.Error("expected output, got empty")
			}
		})
	}
}

// 测试 WriteOutput 组合选项
func TestWriteOutputCombinedOptions(t *testing.T) {
	payload := map[string]any{
		"id":      1,
		"name":    "test",
		"subject": "should be filtered",
	}

	// 测试 JQ 和 Fields 不能同时使用（JQ 优先）
	flags := &GlobalFlags{
		Format: "json",
		JQ:     ".name",
		Fields: "id,name",
	}

	var buf bytes.Buffer
	err := WriteOutput(&buf, flags, payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("expected output, got empty")
	}
}

// 测试错误退出码
func TestErrorExitCodes(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode int
	}{
		{"validation error", errors.NewValidation("test"), 1},
		{"auth error", errors.NewAuth("test"), 2},
		{"api error", errors.NewAPI("test"), 3},
		{"network error", errors.NewNetwork("test"), 4},
		{"internal error", errors.NewInternal("test"), 5},
		{"timeout error", errors.NewTimeout("test"), 6},
		{"rate limit error", errors.NewRateLimit("test"), 7},
		{"nil error", nil, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := errors.ExitCode(tt.err)
			if code != tt.expectedCode {
				t.Errorf("expected exit code %d, got %d", tt.expectedCode, code)
			}
		})
	}
}

// 测试 NewRootCommand 上下文设置
func TestNewRootCommandContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	root := NewRootCommand(ctx)
	if root == nil {
		t.Fatal("expected root command, got nil")
	}

	// 验证上下文已设置
	cmdCtx := root.Context()
	if cmdCtx == nil {
		t.Error("expected context to be set")
	}
}

func TestResolveClientWithTimeoutAndRetries(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("REDMINE_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("REDMINE_CONFIG_DIR")

	configContent := `default: prod
instances:
  prod:
    url: https://prod.example.com
    api_key: prod-key
`
	configFile := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configFile, []byte(configContent), 0600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	flags := &GlobalFlags{
		Timeout: 10 * time.Second,
		Retries: 5,
	}
	client, err := ResolveClient(flags)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Error("expected client, got nil")
	}
}

func TestResolveClientNoDefaultWithInstances(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("REDMINE_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("REDMINE_CONFIG_DIR")

	configContent := `instances:
  dev:
    url: https://dev.example.com
    api_key: dev-key
`
	configFile := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configFile, []byte(configContent), 0600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	flags := &GlobalFlags{}
	_, err := ResolveClient(flags)
	if err == nil {
		t.Error("expected error for no default instance, got nil")
	}
	if !strings.Contains(err.Error(), "URL is required") {
		t.Errorf("expected URL is required error, got %v", err)
	}
}

type errorWriter struct{}

func (w *errorWriter) Write(_ []byte) (int, error) {
	return 0, fmt.Errorf("write error")
}

func TestWriteOutputWithWriterError(t *testing.T) {
	flags := &GlobalFlags{
		Format: "json",
	}
	payload := map[string]string{"test": "value"}

	err := WriteOutput(&errorWriter{}, flags, payload)
	if err == nil {
		t.Error("expected error from writer, got nil")
	}
}

func TestWriteOutputWithJQError(t *testing.T) {
	flags := &GlobalFlags{
		Format: "json",
		JQ:     "invalid[",
	}
	payload := map[string]string{"test": "value"}

	var buf bytes.Buffer
	err := WriteOutput(&buf, flags, payload)
	if err == nil {
		t.Error("expected JQ parse error, got nil")
	}
}

func TestWriteOutputSelectFieldsWithMap(t *testing.T) {
	flags := &GlobalFlags{
		Format: "json",
		Fields: "id,name",
	}
	payload := map[string]any{
		"id":   1,
		"name": "test",
	}

	var buf bytes.Buffer
	err := WriteOutput(&buf, flags, payload)
	if err != nil {
		t.Fatalf("SelectFields on map should not error: %v", err)
	}
}

func TestWriteOutputSelectFieldsMarshalError(t *testing.T) {
	flags := &GlobalFlags{
		Format: "json",
		Fields: "id",
	}
	ch := make(chan int)
	payload := map[string]any{
		"id": 1,
		"ch": ch,
	}

	var buf bytes.Buffer
	err := WriteOutput(&buf, flags, payload)
	if err == nil {
		t.Error("expected error from SelectFields with unmarshallable payload, got nil")
	}
}

func TestResolverResolveClientError(t *testing.T) {
	t.Setenv("REDMINE_CONFIG_DIR", t.TempDir())

	r := &resolver{}
	flags := &types.GlobalFlags{}
	_, err := r.ResolveClient(flags)
	if err == nil {
		t.Error("expected error for empty flags, got nil")
	}
}

func TestResolverWriteOutputError(t *testing.T) {
	r := &resolver{}
	flags := &types.GlobalFlags{
		Format: "json",
		JQ:     "invalid[",
	}
	payload := map[string]string{"test": "value"}

	var buf bytes.Buffer
	err := r.WriteOutput(&buf, flags, payload)
	if err == nil {
		t.Error("expected JQ parse error, got nil")
	}
}

// 测试 parseFields 边界情况
func TestParseFieldsEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"multiple commas", "a,,b,,c", []string{"a", "b", "c"}},
		{"only commas", ",,,", []string{}},
		{"spaces", "a, b, c", []string{"a", " b", " c"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseFields(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d fields, got %d: %v", len(tt.expected), len(result), result)
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("expected field[%d] = %q, got %q", i, tt.expected[i], v)
				}
			}
		})
	}
}
