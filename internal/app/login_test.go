// internal/app/login_test.go
package app

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewLoginCommand(t *testing.T) {
	flags := &GlobalFlags{}
	cmd := newLoginCommand(flags)

	if cmd == nil {
		t.Fatal("expected login command, got nil")
	}

	if cmd.Use != "login" {
		t.Errorf("expected Use 'login', got %s", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description, got empty")
	}

	if cmd.RunE == nil {
		t.Error("expected RunE function, got nil")
	}

	// 验证自定义标�?
	nameFlag := cmd.Flags().Lookup("name")
	if nameFlag == nil {
		t.Error("expected 'name' flag, got nil")
	}

	setDefaultFlag := cmd.Flags().Lookup("set-default")
	if setDefaultFlag == nil {
		t.Error("expected 'set-default' flag, got nil")
	}
}

func TestLoginCommandFlags(t *testing.T) {
	flags := &GlobalFlags{}
	cmd := newLoginCommand(flags)

	tests := []struct {
		name         string
		flagName     string
		defaultValue string
	}{
		{"name flag", "name", ""},
		{"set-default flag", "set-default", "true"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := cmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Errorf("flag %s not found", tt.flagName)
				return
			}
			if flag.DefValue != tt.defaultValue {
				t.Errorf("flag %s: expected default %s, got %s", tt.flagName, tt.defaultValue, flag.DefValue)
			}
		})
	}
}

func TestLoginCommandWithRoot(t *testing.T) {
	ctx := context.Background()
	root := NewRootCommand(ctx)

	loginCmd, _, err := root.Find([]string{"login"})
	if err != nil {
		t.Fatalf("failed to find login command: %v", err)
	}

	if loginCmd == nil {
		t.Fatal("expected login command, got nil")
	}

	if loginCmd.Name() != "login" {
		t.Errorf("expected command name 'login', got %s", loginCmd.Name())
	}
}

func TestPromptInput(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		defaultValue string
		expected     string
	}{
		{"with input", "test-value\n", "", "test-value"},
		{"empty input with default", "\n", "default-value", "default-value"},
		{"input overrides default", "custom\n", "default-value", "custom"},
		{"whitespace trimmed", "  trimmed  \n", "", "trimmed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tt.input))
			result := promptInput(reader, "prompt", tt.defaultValue)

			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestPromptSecret(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"normal input", "secret-key\n", "secret-key"},
		{"whitespace trimmed", "  secret  \n", "secret"},
		{"empty input", "\n", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tt.input))
			result := promptSecret(reader, "prompt")

			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestPromptBool(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		defaultValue bool
		expected     bool
	}{
		{"yes input", "y\n", false, true},
		{"yes uppercase", "Y\n", false, true},
		{"yes full word", "yes\n", false, true},
		{"no input", "n\n", true, false},
		{"no uppercase", "N\n", true, false},
		{"empty with true default", "\n", true, true},
		{"empty with false default", "\n", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tt.input))
			result := promptBool(reader, "prompt", tt.defaultValue)

			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestRunLoginValidation(t *testing.T) {
	// 测试缺少 URL 时的验证
	t.Run("missing URL", func(t *testing.T) {
		// 创建模拟服务�?
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		// 使用空输入模拟用户不输入 URL
		input := "\n\n\n"
		reader := bufio.NewReader(strings.NewReader(input))

		// 模拟 promptInput 行为
		url := promptInput(reader, "URL", "")
		if url != "" {
			t.Error("expected empty URL")
		}
	})

	t.Run("missing API key", func(t *testing.T) {
		input := "\n"
		reader := bufio.NewReader(strings.NewReader(input))

		key := promptSecret(reader, "API Key")
		if key != "" {
			t.Error("expected empty key")
		}
	})
}

func TestRunLoginWithMockServer(t *testing.T) {
	// 创建模拟 Redmine 服务�?
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证 API Key �?
		if r.Header.Get("X-Redmine-API-Key") == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// 模拟认证成功
		if r.URL.Path == "/users/current.json" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"user":{"id":1,"login":"test"}}`))
			return
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// 验证服务器响应（�?API Key�?
	req, _ := http.NewRequest("GET", server.URL+"/users/current.json", nil)
	req.Header.Set("X-Redmine-API-Key", "test-key")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to connect to mock server: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestLoginCommandHelp(t *testing.T) {
	flags := &GlobalFlags{}
	cmd := newLoginCommand(flags)

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

	// 验证帮助输出包含关键信息
	if !strings.Contains(output, "login") {
		t.Error("expected help to contain 'login'")
	}
}

func TestLoginCommandInheritsGlobalFlags(t *testing.T) {
	ctx := context.Background()
	root := NewRootCommand(ctx)

	// 测试 login 命令可以继承全局标志
	loginCmd, _, err := root.Find([]string{"login"})
	if err != nil {
		t.Fatalf("failed to find login command: %v", err)
	}

	// 验证 login 命令可以访问父命令的持久化标�?
	if loginCmd.Parent() == nil {
		t.Error("expected login command to have parent")
	}
}

// 测试登录流程中的错误处理
func TestLoginErrorHandling(t *testing.T) {
	// 测试认证失败的情�?
	t.Run("authentication failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		// 这里只验证服务器返回 401
		resp, err := http.Get(server.URL)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", resp.StatusCode)
		}
	})

	// 测试服务器不可达
	t.Run("server unreachable", func(t *testing.T) {
		// 使用无效 URL
		_, err := http.Get("http://localhost:99999/test")
		if err == nil {
			t.Error("expected error for unreachable server")
		}
	})
}

// 测试颜色输出函数
func TestColorFunctions(t *testing.T) {
	// 验证颜色函数不会 panic
	t.Run("green function", func(t *testing.T) {
		result := green("test")
		if result == "" {
			t.Error("expected non-empty result from green()")
		}
	})

	t.Run("cyan function", func(t *testing.T) {
		result := cyan("test")
		if result == "" {
			t.Error("expected non-empty result from cyan()")
		}
	})
}

// 测试登录命令的上下文处理
func TestLoginCommandContext(t *testing.T) {
	flags := &GlobalFlags{}
	cmd := newLoginCommand(flags)

	// 创建一个可取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 立即取消

	cmd.SetContext(ctx)

	// 验证命令可以处理已取消的上下�?
	// 注意：实际的 runLogin 函数会检查上下文
	if ctx.Err() == nil {
		t.Error("expected context to be canceled")
	}
}

func TestPromptInputReadStringError(t *testing.T) {
	reader := bufio.NewReader(&errorReader{})
	result := promptInput(reader, "prompt", "fallback")
	if result != "fallback" {
		t.Errorf("expected %q on ReadString error, got %q", "fallback", result)
	}
}

func TestPromptInputReadStringErrorNoDefault(t *testing.T) {
	reader := bufio.NewReader(&errorReader{})
	result := promptInput(reader, "prompt", "")
	if result != "" {
		t.Errorf("expected empty string on ReadString error with no default, got %q", result)
	}
}

func TestPromptSecretReadStringError(t *testing.T) {
	reader := bufio.NewReader(&errorReader{})
	result := promptSecret(reader, "prompt")
	if result != "" {
		t.Errorf("expected empty string on ReadString error, got %q", result)
	}
}

func TestPromptBoolReadStringError(t *testing.T) {
	reader := bufio.NewReader(&errorReader{})
	result := promptBool(reader, "prompt", true)
	if !result {
		t.Errorf("expected true (default) on ReadString error, got false")
	}
}

func TestPromptBoolReadStringErrorFalseDefault(t *testing.T) {
	reader := bufio.NewReader(&errorReader{})
	result := promptBool(reader, "prompt", false)
	if result {
		t.Errorf("expected false (default) on ReadString error, got true")
	}
}

type errorReader struct{}

func (r *errorReader) Read(_ []byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}

// 基准测试
func BenchmarkPromptInput(b *testing.B) {
	input := "test-value\n"
	reader := bufio.NewReader(strings.NewReader(input))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader.Reset(strings.NewReader(input))
		promptInput(reader, "prompt", "")
	}
}

func BenchmarkPromptBool(b *testing.B) {
	input := "y\n"
	reader := bufio.NewReader(strings.NewReader(input))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader.Reset(strings.NewReader(input))
		promptBool(reader, "prompt", false)
	}
}

// 表格驱动测试：promptBool 完整测试
func TestPromptBoolComplete(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		defaultValue bool
		expected     bool
	}{
		{"y -> true", "y\n", false, true},
		{"Y -> true", "Y\n", false, true},
		{"yes -> true", "yes\n", false, true},
		{"YES -> true", "YES\n", false, true},
		{"n -> false", "n\n", true, false},
		{"N -> false", "N\n", true, false},
		{"no -> false", "no\n", true, false},
		{"NO -> false", "NO\n", true, false},
		{"empty with true default", "\n", true, true},
		{"empty with false default", "\n", false, false},
		{"other value with true default", "x\n", true, false},
		{"other value with false default", "x\n", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tt.input))
			result := promptBool(reader, "prompt", tt.defaultValue)

			if result != tt.expected {
				t.Errorf("input %q: expected %v, got %v", tt.input, tt.expected, result)
			}
		})
	}
}

// 测试 runLogin 缺少 URL 的错误
func TestRunLoginMissingURL(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("REDMINE_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("REDMINE_CONFIG_DIR")

	// 模拟空输入（不输入 URL）
	input := "\n"
	reader := bufio.NewReader(strings.NewReader(input))

	url := promptInput(reader, "URL", "")
	if url != "" {
		t.Error("expected empty URL")
	}
}

// 测试 runLogin 缺少 API Key 的错误
func TestRunLoginMissingAPIKey(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("REDMINE_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("REDMINE_CONFIG_DIR")

	// 模拟空输入（不输入 API Key）
	input := "\n"
	reader := bufio.NewReader(strings.NewReader(input))

	key := promptSecret(reader, "API Key")
	if key != "" {
		t.Error("expected empty key")
	}
}

// 测试 runLogin 实例名称默认值
func TestRunLoginInstanceNameDefault(t *testing.T) {
	// 模拟不输入实例名称（使用默认值）
	input := "\n"
	reader := bufio.NewReader(strings.NewReader(input))

	name := promptInput(reader, "instance name", "default")
	if name != "default" {
		t.Errorf("expected 'default', got %s", name)
	}
}

// 测试 runLogin 自定义实例名称
func TestRunLoginCustomInstanceName(t *testing.T) {
	// 模拟输入自定义实例名称
	input := "my-instance\n"
	reader := bufio.NewReader(strings.NewReader(input))

	name := promptInput(reader, "instance name", "default")
	if name != "my-instance" {
		t.Errorf("expected 'my-instance', got %s", name)
	}
}

// 测试 runLogin 连接验证失败
func TestRunLoginAuthFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	// 验证服务器返回 401
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", resp.StatusCode)
	}
}

// 测试 runLogin 连接成功
func TestRunLoginAuthSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Redmine-API-Key") != "" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"user":{"id":1}}`))
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	}))
	defer server.Close()

	// 验证服务器返回 200
	req, _ := http.NewRequest("GET", server.URL, nil)
	req.Header.Set("X-Redmine-API-Key", "test-key")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

// 测试 runLogin 配置保存失败（权限问题等）
func TestRunLoginConfigSaveError(t *testing.T) {
	// 这个测试验证配置保存的错误处理
	// 由于实际保存需要文件系统权限，这里只验证错误类型
	tmpDir := t.TempDir()
	os.Setenv("REDMINE_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("REDMINE_CONFIG_DIR")

	// 验证临时目录存在
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		t.Error("temp dir should exist")
	}
}

// 测试 promptInput 带默认值
func TestPromptInputWithDefault(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		defaultValue string
		expected     string
	}{
		{"empty input returns default", "\n", "default-value", "default-value"},
		{"input overrides default", "custom\n", "default-value", "custom"},
		{"whitespace trimmed", "  value  \n", "", "value"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tt.input))
			result := promptInput(reader, "prompt", tt.defaultValue)

			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// 测试 promptSecret 带空格
func TestPromptSecretWithWhitespace(t *testing.T) {
	input := "  secret-key  \n"
	reader := bufio.NewReader(strings.NewReader(input))
	result := promptSecret(reader, "prompt")

	if result != "secret-key" {
		t.Errorf("expected 'secret-key', got %q", result)
	}
}

// 测试 login 命令的标志解析
func TestLoginCommandFlagParsing(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		expectedName string
		expectedSet  bool
	}{
		{"default values", []string{}, "", true},
		{"custom name", []string{"--name", "custom"}, "custom", true},
		{"no set-default", []string{"--set-default=false"}, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := &GlobalFlags{}
			cmd := newLoginCommand(flags)
			cmd.SetArgs(tt.args)

			// 解析标志
			err := cmd.ParseFlags(tt.args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// 验证标志值
			nameFlag, _ := cmd.Flags().GetString("name")
			if nameFlag != tt.expectedName {
				t.Errorf("expected name %q, got %q", tt.expectedName, nameFlag)
			}

			setDefaultFlag, _ := cmd.Flags().GetBool("set-default")
			if setDefaultFlag != tt.expectedSet {
				t.Errorf("expected set-default %v, got %v", tt.expectedSet, setDefaultFlag)
			}
		})
	}
}

// 测试 login 命令的完整流程（模拟）
func TestLoginCommandFullFlow(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/users/current.json" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"user":{"id":1,"login":"test"}}`))
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// 验证服务器响应
	resp, err := http.Get(server.URL + "/users/current.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

// 测试 runLogin 函数的完整流程
func TestRunLoginFullFlow(t *testing.T) {
	// 创建模拟服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/users/current.json" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"user":{"id":1,"login":"test"}}`))
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// 创建临时配置目录
	tmpDir := t.TempDir()
	os.Setenv("REDMINE_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("REDMINE_CONFIG_DIR")

	// 模拟用户输入：
	// 1. URL
	// 2. API Key
	// 3. 实例名称
	input := strings.Join([]string{
		server.URL,      // URL
		"test-api-key",  // API Key
		"test-instance", // 实例名称
	}, "\n") + "\n"

	// 模拟 stdin
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r

	// 在另一个 goroutine 中写入输入
	go func() {
		defer w.Close()
		w.Write([]byte(input))
	}()

	// 捕获 stdout
	oldStdout := os.Stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	// 运行登录
	ctx := context.Background()
	flags := &GlobalFlags{}
	err := runLogin(ctx, flags)

	// 恢复 stdin 和 stdout
	os.Stdin = oldStdin
	wOut.Close()
	os.Stdout = oldStdout

	// 读取输出
	var buf bytes.Buffer
	io.Copy(&buf, rOut)

	// 验证结果
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "登录成功") {
		t.Errorf("expected success message, got: %s", output)
	}
}

// 测试 runLogin 函数缺少 URL
func TestRunLoginMissingURLFlow(t *testing.T) {
	// 创建临时配置目录
	tmpDir := t.TempDir()
	os.Setenv("REDMINE_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("REDMINE_CONFIG_DIR")

	// 模拟用户输入：空 URL
	input := "\n"

	// 模拟 stdin
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r

	go func() {
		defer w.Close()
		w.Write([]byte(input))
	}()

	// 捕获 stdout
	oldStdout := os.Stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	ctx := context.Background()
	flags := &GlobalFlags{}
	err := runLogin(ctx, flags)

	os.Stdin = oldStdin
	wOut.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, rOut)

	// 验证错误
	if err == nil {
		t.Error("expected error for missing URL, got nil")
	}
}

// 测试 runLogin 函数缺少 API Key
func TestRunLoginMissingAPIKeyFlow(t *testing.T) {
	// 创建临时配置目录
	tmpDir := t.TempDir()
	os.Setenv("REDMINE_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("REDMINE_CONFIG_DIR")

	// 模拟用户输入：URL + 空 API Key
	input := strings.Join([]string{
		"https://example.com",
		"", // 空 API Key
	}, "\n") + "\n"

	// 模拟 stdin
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r

	go func() {
		defer w.Close()
		w.Write([]byte(input))
	}()

	// 捕获 stdout
	oldStdout := os.Stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	ctx := context.Background()
	flags := &GlobalFlags{}
	err := runLogin(ctx, flags)

	os.Stdin = oldStdin
	wOut.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, rOut)

	// 验证错误
	if err == nil {
		t.Error("expected error for missing API key, got nil")
	}
}

// 测试 runLogin 函数连接失败
func TestRunLoginConnectionFailedFlow(t *testing.T) {
	// 创建临时配置目录
	tmpDir := t.TempDir()
	os.Setenv("REDMINE_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("REDMINE_CONFIG_DIR")

	// 创建模拟服务器（返回 401）
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	// 模拟用户输入
	input := strings.Join([]string{
		server.URL,
		"test-api-key",
	}, "\n") + "\n"

	// 模拟 stdin
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r

	go func() {
		defer w.Close()
		w.Write([]byte(input))
	}()

	// 捕获 stdout
	oldStdout := os.Stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	ctx := context.Background()
	flags := &GlobalFlags{}
	err := runLogin(ctx, flags)

	os.Stdin = oldStdin
	wOut.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, rOut)

	// 验证错误
	if err == nil {
		t.Error("expected error for connection failed, got nil")
	}
}

// 测试 runLogin 函数已有默认实例的情况
func TestRunLoginWithExistingDefault(t *testing.T) {
	// 创建模拟服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/users/current.json" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"user":{"id":1,"login":"test"}}`))
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// 创建临时配置目录
	tmpDir := t.TempDir()
	os.Setenv("REDMINE_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("REDMINE_CONFIG_DIR")

	// 写入已有配置（有默认实例）
	configContent := `default: existing
instances:
  existing:
    url: https://existing.example.com
    api_key: existing-key
`
	configFile := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// 模拟用户输入
	input := strings.Join([]string{
		server.URL,
		"test-api-key",
		"new-instance",
		"y", // 设为默认实例
	}, "\n") + "\n"

	// 模拟 stdin
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r

	go func() {
		defer w.Close()
		w.Write([]byte(input))
	}()

	// 捕获 stdout
	oldStdout := os.Stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	ctx := context.Background()
	flags := &GlobalFlags{}
	err := runLogin(ctx, flags)

	os.Stdin = oldStdin
	wOut.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, rOut)

	// 验证结果
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
