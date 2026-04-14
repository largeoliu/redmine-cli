// internal/app/version_test.go
package app

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"
)

func TestNewVersionCommand(t *testing.T) {
	cmd := newVersionCommand()

	if cmd == nil {
		t.Fatal("expected version command, got nil")
	}

	if cmd.Use != "version" {
		t.Errorf("expected Use 'version', got %s", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description, got empty")
	}

	if cmd.Run == nil {
		t.Error("expected Run function, got nil")
	}
}

func TestVersionCommandExecution(t *testing.T) {
	cmd := newVersionCommand()

	// 捕获 stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd.Run(cmd, []string{})

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)

	output := buf.String()
	if output == "" {
		t.Error("expected output, got empty")
	}

	// 验证输出包含版本信息
	if !strings.Contains(output, "redmine version") {
		t.Errorf("expected output to contain 'redmine version', got %s", output)
	}

	// 验证输出包含 commit 信息
	if !strings.Contains(output, "commit:") {
		t.Errorf("expected output to contain 'commit:', got %s", output)
	}

	// 验证输出包含 built 信息
	if !strings.Contains(output, "built:") {
		t.Errorf("expected output to contain 'built:', got %s", output)
	}
}

func TestVersionCommandOutput(t *testing.T) {
	cmd := newVersionCommand()

	// 捕获 stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd.Run(cmd, []string{})

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)

	output := buf.String()

	// 验证版本变量被正确使�?
	if !strings.Contains(output, version) {
		t.Errorf("expected output to contain version %s, got %s", version, output)
	}

	// 验证 commit 变量被正确使�?
	if !strings.Contains(output, commit) {
		t.Errorf("expected output to contain commit %s, got %s", commit, output)
	}

	// 验证 date 变量被正确使�?
	if !strings.Contains(output, date) {
		t.Errorf("expected output to contain date %s, got %s", date, output)
	}
}

func TestVersionCommandWithRoot(t *testing.T) {
	ctx := context.Background()
	root := NewRootCommand(ctx)

	versionCmd, _, err := root.Find([]string{"version"})
	if err != nil {
		t.Fatalf("failed to find version command: %v", err)
	}

	if versionCmd == nil {
		t.Fatal("expected version command, got nil")
	}

	if versionCmd.Name() != "version" {
		t.Errorf("expected command name 'version', got %s", versionCmd.Name())
	}
}

func TestVersionCommandHelp(t *testing.T) {
	cmd := newVersionCommand()

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

	// 验证帮助输出包含命令�?
	if !strings.Contains(output, "version") {
		t.Error("expected help to contain 'version'")
	}
}

func TestVersionVariables(t *testing.T) {
	// 测试版本变量的默认�?
	if version == "" {
		t.Error("expected version to have a value")
	}

	if commit == "" {
		t.Error("expected commit to have a value")
	}

	if date == "" {
		t.Error("expected date to have a value")
	}
}

func TestVersionCommandRunVsRunE(t *testing.T) {
	cmd := newVersionCommand()

	// version 命令使用 Run 而不�?RunE
	// 因为它总是成功的，不需要返回错�?
	if cmd.RunE != nil {
		t.Error("expected RunE to be nil for version command")
	}

	if cmd.Run == nil {
		t.Error("expected Run to be set for version command")
	}
}

func TestVersionCommandNoArgs(t *testing.T) {
	cmd := newVersionCommand()

	// 捕获 stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// version 命令应该忽略任何参数
	cmd.Run(cmd, []string{"ignored", "args"})

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)

	output := buf.String()
	if output == "" {
		t.Error("expected output, got empty")
	}
}

func TestVersionCommandIntegration(t *testing.T) {
	ctx := context.Background()
	root := NewRootCommand(ctx)

	// 捕获 stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	root.SetArgs([]string{"version"})

	err := root.Execute()
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

	// 验证集成测试输出
	if !strings.Contains(output, "redmine version") {
		t.Errorf("expected output to contain 'redmine version', got %s", output)
	}
}

// 表格驱动测试：版本命令输出格�?
func TestVersionCommandOutputFormat(t *testing.T) {
	tests := []struct {
		name        string
		contains    string
		shouldExist bool
	}{
		{"version line", "redmine version", true},
		{"commit line", "commit:", true},
		{"built line", "built:", true},
		{"version value", version, true},
		{"commit value", commit, true},
		{"date value", date, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newVersionCommand()

			// 捕获 stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			cmd.Run(cmd, []string{})

			w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			io.Copy(&buf, r)

			output := buf.String()
			contains := strings.Contains(output, tt.contains)

			if tt.shouldExist && !contains {
				t.Errorf("expected output to contain %q, got %s", tt.contains, output)
			}
			if !tt.shouldExist && contains {
				t.Errorf("expected output NOT to contain %q, got %s", tt.contains, output)
			}
		})
	}
}

// 测试版本命令的输出写入器
func TestVersionCommandOutputWriter(t *testing.T) {
	cmd := newVersionCommand()

	// 捕获 stdout
	old := os.Stdout

	// 第一次运�?
	r1, w1, _ := os.Pipe()
	os.Stdout = w1
	cmd.Run(cmd, []string{})
	w1.Close()
	var buf1 bytes.Buffer
	io.Copy(&buf1, r1)

	// 第二次运�?
	r2, w2, _ := os.Pipe()
	os.Stdout = w2
	cmd.Run(cmd, []string{})
	w2.Close()
	var buf2 bytes.Buffer
	io.Copy(&buf2, r2)

	os.Stdout = old

	// 两次输出应该相同
	if buf1.String() != buf2.String() {
		t.Error("expected consistent output")
	}
}

// 基准测试
func BenchmarkVersionCommand(b *testing.B) {
	cmd := newVersionCommand()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 捕获 stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		cmd.Run(cmd, []string{})

		w.Close()
		os.Stdout = old
		io.Copy(io.Discard, r)
	}
}

// 测试版本命令是否正确注册到根命令
func TestVersionCommandRegistration(t *testing.T) {
	ctx := context.Background()
	root := NewRootCommand(ctx)

	// 获取所有子命令
	commands := root.Commands()

	// 查找 version 命令
	var found bool
	for _, cmd := range commands {
		if cmd.Name() == "version" {
			found = true
			break
		}
	}

	if !found {
		t.Error("version command not registered with root command")
	}
}

// 测试版本命令的上下文
func TestVersionCommandContext(t *testing.T) {
	cmd := newVersionCommand()

	// version 命令不需要上下文，因为它只是打印信息
	// 但我们验证它不会因为上下文而失�?
	ctx := context.Background()
	cmd.SetContext(ctx)

	// 捕获 stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd.Run(cmd, []string{})

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)

	output := buf.String()
	if output == "" {
		t.Error("expected output, got empty")
	}
}

// 测试版本命令的静默模�?
func TestVersionCommandSilenceMode(t *testing.T) {
	ctx := context.Background()
	root := NewRootCommand(ctx)

	versionCmd, _, err := root.Find([]string{"version"})
	if err != nil {
		t.Fatalf("failed to find version command: %v", err)
	}

	// version 命令应该继承根命令的静默设置
	// 或者有自己的静默设�?
	// 这里只验证命令可以正常执�?

	// 捕获 stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	versionCmd.Run(versionCmd, []string{})

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)

	output := buf.String()
	if output == "" {
		t.Error("expected output, got empty")
	}
}
