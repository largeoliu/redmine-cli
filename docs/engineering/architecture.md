# 架构与结构

## 工具链版本

- Go 版本：`1.23`，以 `go.mod` 和 CI 配置为准
- Node.js：`>=14`，发布 CI 使用 Node `20`
- 代码检查工具：`golangci-lint`
- 代码格式化工具：`gofmt`、`goimports`
- `goimports` 本地导入前缀：`github.com/largeoliu/redmine-cli`

## 仓库结构

- `cmd/main.go` 保持简洁，只调用 `app.Execute()`
- `internal/app/` 负责根命令注册、登录流程、全局 flags 和版本输出
- `internal/client/` 负责共享 HTTP 客户端、认证、重试、批量操作和传输层行为
- `internal/config/` 负责配置结构和磁盘持久化
- `internal/errors/` 负责带类型的错误模型和退出码映射
- `internal/output/` 负责 JSON、table、raw 输出以及 jq / 字段过滤
- `internal/resources/<resource>/` 是资源命令与 API 调用的主要组织方式
- `internal/testutil/` 提供 mock server 和 goroutine 泄漏检测辅助函数

## 资源包模式

- 遵循 `issues`、`projects`、`users`、`versions` 等现有拆分方式
- `commands.go`：Cobra 命令工厂和 flag 校验
- `client.go`：该资源的 Redmine API 调用
- `types.go`：请求和响应结构体
- `main_test.go`：需要 goroutine 泄漏检测时放包级 `TestMain`
- `commands_test.go`、`client_test.go`：资源包单元测试
