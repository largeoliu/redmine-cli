# 登录流程优化实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 优化登录流程的交互精简度和错误提示友好度

**Architecture:**
- 在 `client.Client` 新增 `Ping(ctx context.Context) error` 方法用于 URL 连通性校验
- 修改 `login.go` 中的交互流程，移除 Step 前缀和空行
- 复用现有错误类型和网络错误校验

**Tech Stack:** Go, cobra, http client

---

## 文件变更概览

| 文件 | 变更类型 |
|-----|---------|
| `internal/client/client.go` | 新增 Ping 方法 |
| `internal/app/login.go` | 重构交互流程和错误处理 |
| `internal/app/login_test.go` | 更新测试以匹配新行为 |

---

## Task 1: 在 client.Client 添加 Ping 方法

**Files:**
- Modify: `internal/client/client.go`

- [ ] **Step 1: 查看现有 client.go 结构**

```go
// 查看 Client 结构体定义和现有方法
```

- [ ] **Step 2: 添加 Ping 方法**

在 `client.go` 中添加以下方法：

```go
// Ping tests the connectivity to the given URL using HTTP HEAD request.
func (c *Client) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, c.baseURL, nil)
	if err != nil {
		return errors.NewNetwork("URL 格式无效", errors.WithCause(err))
	}

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return errors.NewNetwork("URL 无法访问", errors.WithActions(
			"1) URL 正确",
			"2) 网络畅通",
			"3) 服务正常运行",
		), errors.WithCause(err))
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return errors.NewAPI("服务器返回错误状态码", errors.WithCause(fmt.Errorf("status: %d", resp.StatusCode)))
	}

	return nil
}
```

- [ ] **Step 3: 验证编译通过**

Run: `go build ./internal/client`

---

## Task 2: 更新 client.go import

**Files:**
- Modify: `internal/client/client.go`

- [ ] **Step 1: 添加 time 和 fmt import**

确保 import 部分包含：

```go
import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/largeoliu/redmine-cli/internal/errors"
)
```

Run: `go build ./internal/client`

---

## Task 3: 重构 login.go 交互流程

**Files:**
- Modify: `internal/app/login.go`

- [ ] **Step 1: 移除 printStep 函数，改用直接提示**

移除 `printStep` 函数（第95-98行）

- [ ] **Step 2: 重构 runLogin 函数**

将 `runLogin` 函数改为：

```go
func runLogin(ctx context.Context, flags *GlobalFlags) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Redmine URL: ")
	url := promptInput(reader, "", flags.URL)
	if url == "" {
		return errors.NewValidation("URL 不能为空")
	}

	fmt.Print("正在检查连通性... ")
	c := client.NewClient(url, "")
	if err := c.Ping(ctx); err != nil {
		fmt.Println(color.RedString("✗"))
		return err
	}
	fmt.Println(green("✓"))

	fmt.Print("API Key: ")
	apiKey := promptSecret(reader, "")
	if apiKey == "" {
		return errors.NewValidation("API Key 不能为空")
	}

	fmt.Print("正在验证连接... ")
	c.SetAPIKey(apiKey)
	if err := c.TestAuth(ctx); err != nil {
		fmt.Println(color.RedString("✗"))
		return errors.NewAuth("API Key 无效", errors.WithActions(
			"1) API Key 正确",
			"2) 前往 Settings → API access 查看",
		), errors.WithCause(err))
	}
	fmt.Println(green("✓"))

	fmt.Print("实例名称 [default]: ")
	name := promptInput(reader, "", "")
	if name == "" {
		name = "default"
	}

	store := config.NewStore("")
	if err := store.SaveInstance(name, config.Instance{
		URL:    url,
		APIKey: apiKey,
	}); err != nil {
		return errors.NewInternal("配置保存失败", errors.WithCause(err))
	}

	setDefault := true
	cfg, err := store.Load()
	if err == nil && cfg != nil && cfg.Default != "" && cfg.Default != name {
		fmt.Print("设为默认实例? [Y/n]: ")
		setDefault = promptBool(reader, "", true)
	}

	if setDefault {
		if err := store.SetDefault(name); err != nil {
			return errors.NewInternal("设置默认实例失败", errors.WithCause(err))
		}
	}

	fmt.Println()
	fmt.Printf("%s 登录成功！配置已保存到 %s\n", green("✓"), config.Path())
	return nil
}
```

- [ ] **Step 3: 更新 promptInput 函数**

移除 prompt 参数中的 prefix 显示逻辑，因为提示直接在输入前打印：

```go
func promptInput(reader *bufio.Reader, prompt, defaultValue string) string {
	if defaultValue != "" {
		fmt.Printf("[%s]: ", cyan(defaultValue))
	} else {
		fmt.Print(": ")
	}
	input, err := reader.ReadString('\n')
	if err != nil {
		return defaultValue
	}
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue
	}
	return input
}
```

- [ ] **Step 4: 验证编译**

Run: `go build ./internal/app`

---

## Task 4: 更新 login_test.go

**Files:**
- Modify: `internal/app/login_test.go`

- [ ] **Step 1: 更新 TestPrintStep 测试**

移除 `TestPrintStep` 测试（函数已删除）

- [ ] **Step 2: 添加 Ping 方法测试**

添加新测试函数：

```go
func TestClientPing(t *testing.T) {
	t.Run("valid URL", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		c := client.NewClient(server.URL, "")
		err := c.Ping(context.Background())
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("unreachable URL", func(t *testing.T) {
		c := client.NewClient("http://localhost:99999", "")
		err := c.Ping(context.Background())
		if err == nil {
			t.Error("expected error for unreachable URL")
		}
	})

	t.Run("invalid URL format", func(t *testing.T) {
		c := client.NewClient("://invalid", "")
		err := c.Ping(context.Background())
		if err == nil {
			t.Error("expected error for invalid URL")
		}
	})
}
```

- [ ] **Step 3: 运行测试**

Run: `go test -v -run 'TestClientPing' ./internal/client`

Run: `go test -v ./internal/app`

---

## Task 5: 运行完整测试套件

- [ ] **Step 1: 运行 linter**

Run: `make lint`

- [ ] **Step 2: 运行所有测试**

Run: `go test -v ./internal/client ./internal/app`

- [ ] **Step 3: 提交代码**

```bash
git add internal/client/client.go internal/app/login.go internal/app/login_test.go
git commit -m "feat: optimize login flow UX - compact prompts, URL validation, better error messages"
```
