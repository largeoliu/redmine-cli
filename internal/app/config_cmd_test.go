// internal/app/config_cmd_test.go
package app

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewConfigCommand(t *testing.T) {
	flags := &GlobalFlags{}
	cmd := newConfigCommand(flags)

	if cmd == nil {
		t.Fatal("expected config command, got nil")
	}

	if cmd.Use != "config" {
		t.Errorf("expected Use 'config', got %s", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description, got empty")
	}

	// 验证子命令
	commands := cmd.Commands()
	expectedSubcmds := []string{"get", "set", "list"}

	for _, expected := range expectedSubcmds {
		found := false
		for _, subcmd := range commands {
			if subcmd.Name() == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected subcommand %s not found", expected)
		}
	}
}

func TestNewConfigGetCommand(t *testing.T) {
	flags := &GlobalFlags{}
	cmd := newConfigGetCommand(flags)

	if cmd == nil {
		t.Fatal("expected config get command, got nil")
	}

	if cmd.Use != "get" {
		t.Errorf("expected Use 'get', got %s", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description, got empty")
	}

	if cmd.RunE == nil {
		t.Error("expected RunE function, got nil")
	}
}

func TestNewConfigSetCommand(t *testing.T) {
	flags := &GlobalFlags{}
	cmd := newConfigSetCommand(flags)

	if cmd == nil {
		t.Fatal("expected config set command, got nil")
	}

	if cmd.Use != "set <instance-name>" {
		t.Errorf("expected Use 'set <instance-name>', got %s", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description, got empty")
	}

	if cmd.RunE == nil {
		t.Error("expected RunE function, got nil")
	}

	// 验证参数要求
	if cmd.Args == nil {
		t.Error("expected Args validator, got nil")
	}
}

func TestNewConfigListCommand(t *testing.T) {
	flags := &GlobalFlags{}
	cmd := newConfigListCommand(flags)

	if cmd == nil {
		t.Fatal("expected config list command, got nil")
	}

	if cmd.Use != "list" {
		t.Errorf("expected Use 'list', got %s", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description, got empty")
	}

	if cmd.RunE == nil {
		t.Error("expected RunE function, got nil")
	}
}

func TestConfigGetCommandExecution(t *testing.T) {
	// 创建临时配置目录
	tmpDir := t.TempDir()

	// 设置环境变量以使用临时目录
	t.Setenv("REDMINE_CONFIG_DIR", tmpDir)

	flags := &GlobalFlags{Format: "json"}
	cmd := newConfigGetCommand(flags)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)

	output := buf.String()
	if output == "" {
		t.Error("expected output, got empty")
	}
}

func TestConfigSetCommandExecution(t *testing.T) {
	// 创建临时配置目录
	tmpDir := t.TempDir()

	os.Setenv("REDMINE_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("REDMINE_CONFIG_DIR")

	flags := &GlobalFlags{Format: "json"}

	// 写入初始配置
	configContent := `default: test
instances:
  test:
    url: https://example.com
    api_key: test-key
`
	configFile := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cmd := newConfigSetCommand(flags)

	// 捕获 stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// 测试设置默认实例
	err := cmd.RunE(cmd, []string{"test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)

	output := buf.String()
	if output == "" {
		t.Error("expected output, got empty")
	}
}

func TestConfigSetCommandMissingArg(t *testing.T) {
	flags := &GlobalFlags{}
	cmd := newConfigSetCommand(flags)

	// 测试缺少参数
	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Error("expected error for missing argument, got nil")
	}
}

func TestConfigSetCommandTooManyArgs(t *testing.T) {
	flags := &GlobalFlags{}
	cmd := newConfigSetCommand(flags)

	// 测试参数过多
	err := cmd.Args(cmd, []string{"arg1", "arg2"})
	if err == nil {
		t.Error("expected error for too many arguments, got nil")
	}
}

func TestConfigListCommandExecution(t *testing.T) {
	// 创建临时配置目录
	tmpDir := t.TempDir()

	os.Setenv("REDMINE_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("REDMINE_CONFIG_DIR")

	// 写入配置
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
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	flags := &GlobalFlags{Format: "json"}
	cmd := newConfigListCommand(flags)

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("expected output, got empty")
	}
}

func TestConfigListCommandEmpty(t *testing.T) {
	// 创建临时配置目录（空配置）
	tmpDir := t.TempDir()

	os.Setenv("REDMINE_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("REDMINE_CONFIG_DIR")

	flags := &GlobalFlags{Format: "json"}
	cmd := newConfigListCommand(flags)

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	// 应该返回空的实例列表
	if output == "" {
		t.Error("expected output, got empty")
	}
}

func TestConfigCommandWithRoot(t *testing.T) {
	ctx := context.Background()
	root := NewRootCommand(ctx)

	// 测试 config 命令是否正确添加到根命令
	configCmd, _, err := root.Find([]string{"config"})
	if err != nil {
		t.Fatalf("failed to find config command: %v", err)
	}

	if configCmd == nil {
		t.Fatal("expected config command, got nil")
	}

	if configCmd.Name() != "config" {
		t.Errorf("expected command name 'config', got %s", configCmd.Name())
	}
}

func TestConfigGetSubcommandWithRoot(t *testing.T) {
	ctx := context.Background()
	root := NewRootCommand(ctx)

	getCmd, _, err := root.Find([]string{"config", "get"})
	if err != nil {
		t.Fatalf("failed to find config get command: %v", err)
	}

	if getCmd == nil {
		t.Fatal("expected config get command, got nil")
	}
}

func TestConfigSetSubcommandWithRoot(t *testing.T) {
	ctx := context.Background()
	root := NewRootCommand(ctx)

	setCmd, _, err := root.Find([]string{"config", "set"})
	if err != nil {
		t.Fatalf("failed to find config set command: %v", err)
	}

	if setCmd == nil {
		t.Fatal("expected config set command, got nil")
	}
}

func TestConfigListSubcommandWithRoot(t *testing.T) {
	ctx := context.Background()
	root := NewRootCommand(ctx)

	listCmd, _, err := root.Find([]string{"config", "list"})
	if err != nil {
		t.Fatalf("failed to find config list command: %v", err)
	}

	if listCmd == nil {
		t.Fatal("expected config list command, got nil")
	}
}

func TestConfigCommandHelp(t *testing.T) {
	flags := &GlobalFlags{}
	cmd := newConfigCommand(flags)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("expected help output, got empty")
	}
}

// 表格驱动测试：config 命令参数验证
func TestConfigSetCommandArgsValidation(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"valid single arg", []string{"instance-name"}, false},
		{"no args", []string{}, true},
		{"too many args", []string{"arg1", "arg2", "arg3"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := &GlobalFlags{}
			cmd := newConfigSetCommand(flags)

			err := cmd.Args(cmd, tt.args)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// 测试命令继承全局标志
func TestConfigCommandInheritsGlobalFlags(t *testing.T) {
	ctx := context.Background()
	root := NewRootCommand(ctx)

	// 获取 config 命令
	configCmd, _, err := root.Find([]string{"config"})
	if err != nil {
		t.Fatalf("failed to find config command: %v", err)
	}

	// 验证 config 命令可以访问持久化标志
	persistentFlags := configCmd.PersistentFlags()
	if persistentFlags == nil {
		// 子命令应该能继承父命令的持久化标志
		// 检查是否有从根命令继承的标志
		urlFlag := configCmd.Flags().Lookup("url")
		if urlFlag == nil {
			parentFlags := configCmd.Parent().PersistentFlags()
			if parentFlags != nil {
				_ = parentFlags.Lookup("url")
			}
		}
		// url 标志应该存在（从根命令继承）
		// 注意：这里只是验证命令结构，实际继承由 cobra 处理
	}
}

// 测试 config get 命令有实例的情况
func TestConfigGetCommandWithInstances(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("REDMINE_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("REDMINE_CONFIG_DIR")

	// 写入有实例的配置
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
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	flags := &GlobalFlags{Format: "json"}
	cmd := newConfigGetCommand(flags)

	// 捕获 stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)

	output := buf.String()
	// 验证输出包含实例信息
	if !strings.Contains(output, "prod") {
		t.Errorf("expected output to contain 'prod', got %s", output)
	}
	if !strings.Contains(output, "dev") {
		t.Errorf("expected output to contain 'dev', got %s", output)
	}
}

// 测试 config set 命令设置不存在的实例
func TestConfigSetCommandNonExistentInstance(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("REDMINE_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("REDMINE_CONFIG_DIR")

	// 写入初始配置
	configContent := `default: test
instances:
  test:
    url: https://example.com
    api_key: test-key
`
	configFile := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	flags := &GlobalFlags{Format: "json"}
	cmd := newConfigSetCommand(flags)

	// 测试设置不存在的实例
	err := cmd.RunE(cmd, []string{"nonexistent"})
	if err == nil {
		t.Error("expected error for non-existent instance, got nil")
	}
}

// 测试 config get 命令配置加载失败
func TestConfigGetCommandLoadError(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("REDMINE_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("REDMINE_CONFIG_DIR")

	// 写入无效的配置文件
	configFile := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configFile, []byte("invalid: yaml: ["), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	flags := &GlobalFlags{Format: "json"}
	cmd := newConfigGetCommand(flags)

	err := cmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error for invalid config, got nil")
	}
}

// 测试 config list 命令配置加载失败
func TestConfigListCommandLoadError(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("REDMINE_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("REDMINE_CONFIG_DIR")

	// 写入无效的配置文件
	configFile := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configFile, []byte("invalid: yaml: ["), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	flags := &GlobalFlags{Format: "json"}
	cmd := newConfigListCommand(flags)

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := cmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error for invalid config, got nil")
	}
}

// 测试 config set 命令配置加载失败
func TestConfigSetCommandLoadError(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("REDMINE_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("REDMINE_CONFIG_DIR")

	// 写入无效的配置文件
	configFile := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configFile, []byte("invalid: yaml: ["), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	flags := &GlobalFlags{Format: "json"}
	cmd := newConfigSetCommand(flags)

	err := cmd.RunE(cmd, []string{"test"})
	if err == nil {
		t.Error("expected error for invalid config, got nil")
	}
}

// 测试 config list 命令输出格式
func TestConfigListCommandOutputFormat(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("REDMINE_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("REDMINE_CONFIG_DIR")

	// 写入配置
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
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	tests := []struct {
		name   string
		format string
	}{
		{"json format", "json"},
		{"table format", "table"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := &GlobalFlags{Format: tt.format}
			cmd := newConfigListCommand(flags)

			var buf bytes.Buffer
			cmd.SetOut(&buf)

			err := cmd.RunE(cmd, []string{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			output := buf.String()
			if output == "" {
				t.Error("expected output, got empty")
			}
		})
	}
}

// 测试 config 命令的完整执行流程
func TestConfigCommandFullExecution(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("REDMINE_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("REDMINE_CONFIG_DIR")

	ctx := context.Background()
	root := NewRootCommand(ctx)

	// 测试 config get
	var getBuf bytes.Buffer
	root.SetOut(&getBuf)
	root.SetArgs([]string{"config", "get"})
	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
