# AGENTS.md

## 目的
- 本仓库是一个 Go 语言编写的 Redmine CLI 工具，发布的二进制文件名为 `redmine`。
- 代码主入口在 `cmd/`、`internal/` 和 `test/`。
- `package.json` 仅用于 npm 分发和安装辅助脚本，不包含应用逻辑。

## 目录导航
- `cmd/`：程序入口，`cmd/main.go` 仅调用 `app.Execute()`。
- `internal/app/`：根命令、全局 flags、登录流程、版本输出。
- `internal/client/`：共享 HTTP 客户端、认证、重试、批量操作。
- `internal/resources/`：各类 Redmine 资源命令与 API 调用。
- `internal/config/`、`internal/errors/`、`internal/output/`：配置、错误模型、输出格式化。
- `test/`：测试规则与运行方式，细节见 `test/AGENTS.md`。
- `docs/`：面向用户的命令与配置文档。

## 全局硬约束
- 只在用户请求范围内工作，优先沿用现有代码模式，不扩散改动。
- 开始验证时先跑最小相关测试，不要一开始就跑完整测试套件。
- Go 代码遵循 `gofmt`、`goimports` 和 `golangci-lint` 约定。

## 使用方式
- 根 `AGENTS.md` 只保留跨目录索引与少量全局约束，不再充当工程手册。
- 同一条规则只保留一个权威位置；进入具体区域时，优先看就近代码、测试和局部文档。
- 工程约定、构建命令和代码组织规则统一写在 `docs/engineering/`，入口见 `docs/engineering/README.md`。
- 测试相关的环境变量、命令和边界统一写在 `test/AGENTS.md`。
