# 开发约定

测试运行入口、环境变量和测试边界统一见 `test/AGENTS.md`，本文不重复测试目录规则。

## 构建与校验命令

- 本地构建：`go build -o bin/redmine ./cmd`
- Make 构建入口：`make build`
- 直接运行：`go run ./cmd`
- 主要 lint：`make lint`，底层为 `golangci-lint run ./...`
- 自动修复：`make lint-fix`，底层为 `golangci-lint run --fix ./...`
- 格式化检查：`gofmt -s -l .`
- 依赖卫生检查：`go mod verify`、`go mod tidy -diff`
- 修改 import 或格式后，先跑 `make lint-fix`，再跑 `make lint`

## 格式与命名约定

- 使用 `gofmt` 格式化，由 `goimports` 管理 import 分组
- 文件保持 UTF-8、LF 结尾，并保留末尾换行
- Go 文件使用 tab 缩进，不保留行尾空格
- import 分组遵循标准库、空行、项目与第三方的顺序
- 导出构造函数使用 `NewX` 或 `NewCommand`
- 未导出的 Cobra 工厂函数使用 `newXCommand`
- receiver 使用现有短名风格，如 `c`、`r`、`m`
- 类型名优先清晰名词，如 `Client`、`Config`、`BatchResult`
- API 层命名遵循现有 tag 风格，如 `project_id`、`api_key`、`total_count`
- Cobra `Use` 名称默认使用单数，复数通过 alias 处理

## 错误处理与控制流

- 用户可见的命令和客户端错误优先使用 `internal/errors`
- 使用既有错误类别：`validation`、`auth`、`api`、`network`、`internal`、`timeout`、`rate_limit`
- 需要改善 CLI 恢复体验时，使用 `errors.WithHint(...)`、`errors.WithActions(...)`
- 校验失败尽早返回，避免深层嵌套
- 错误字符串保持简短直接，除非沿用现有用户消息风格，否则避免尾部标点
- 根命令统一负责错误打印和退出码映射，命令代码直接返回带类型错误
- `context.Context` 贯穿请求路径和长操作，并作为函数第一个参数
- Cobra handler 调用客户端时使用 `cmd.Context()`
- 如 API 形态强制不同顺序，使用带原因说明的 `//nolint:revive`

## CLI 约定

- 命令实现保持为返回 `*cobra.Command` 的小工厂函数
- 常见流程是：校验参数和 flags、解析 client、调用资源 client、写入输出
- 全局 flag 绑定集中在 `internal/app/root.go`
- 结构化输出通过 `resolver.WriteOutput(...)` 统一处理
- 涉及破坏性或写操作时，尊重 `--dry-run` 和 `--yes` 语义

## Commit Message 规范

- CI 使用 `commitlint` + `@commitlint/config-conventional` 校验，不符合规范的 commit 会导致 CI 失败
- 格式：`<type>(<scope>): <description>`
  - type 必须是以下之一：`feat`、`fix`、`perf`、`refactor`、`docs`、`style`、`test`、`build`、`ci`、`chore`、`revert`
  - scope 可选，表示影响范围，如 `issues`、`client`、`output`、`deps`、`release`
  - description 使用小写开头，不加句号
- 示例：
  - `feat(issues): add sprint filter support`
  - `fix(client): retry on 429 status`
  - `perf: reduce redundant JSON marshal`
  - `chore(deps): bump golangci-lint to v2`
  - `test: improve coverage for output package`
- 禁止事项：
  - 禁止使用规范外的 type（如 `add`、`update`、`change`）
  - 禁止 description 为空
  - 禁止首字母大写或末尾加句号
  - 禁止多行 commit 的首行超过 72 字符
- `release-please` 依赖 conventional commits 生成 changelog 和自动发版，保持格式一致至关重要

## Agent 实用建议

- 优先运行最小相关验证，不要起手跑完整测试套件
- 修改共享 client、output 或根命令行为后，补跑受影响包测试，并按需要运行 `make e2e`
- 修改真实 API 流程后，先跑受影响单元测试，再跑最小有用的集成验证
- 声称完成前，运行能够直接证明修改正确性的验证命令
