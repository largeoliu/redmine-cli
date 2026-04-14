# AGENTS.md

## 作用范围
- 本文件覆盖 `test/` 目录。
- 这里只写测试目录特有的规则；通用约束仍以仓库根 `AGENTS.md` 为准。

## 目录说明
- `test/e2e/`：编译二进制后运行的命令行端到端测试。
- `test/integration/`：针对真实 Redmine 的集成测试。

## 运行约定
- E2E 测试会在 `TestMain` 中重新编译二进制，运行时使用 `-count=1` 避免缓存干扰。
- 集成测试需要 `REDMINE_URL`、`REDMINE_API_KEY`、`REDMINE_PROJECT_ID`；缺少任一值时不能视为有效验证。
- 修改 CLI 根命令、输出、共享 client 或真实 API 流程后，按改动范围补跑最小有用的 E2E 或集成测试。

## 编写约定
- 单元测试尽量与被测包同目录放置。
- HTTP 驱动单元测试复用 `internal/testutil.NewMockServer(t)`。
- 可能产生 goroutine 泄漏的包使用 `testutil.LeakTestMain(m)`。
- 行为矩阵明显时优先写表格驱动 `t.Run(...)`。
- CLI 命令测试要断言退出行为、stdout/stderr 和参数校验。
- HTTP 客户端测试覆盖状态码、重试、取消和畸形响应体。

## 常用命令
- 单个 E2E：`go test -v -count=1 -run '^TestHelpCommand$' ./test/e2e`
- 单个集成：`go test -v -count=1 -run '^TestIssueCRUD$' ./test/integration`
- E2E 套件：`make e2e`
- 集成套件：`make integration`
