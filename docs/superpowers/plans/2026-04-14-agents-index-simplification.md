# AGENTS Index Simplification Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Turn the root `AGENTS.md` into a lightweight index page and add only the minimum local guidance needed under `test/`.

**Architecture:** Keep the change intentionally small. Replace the current root handbook with a short navigation document that contains only repository identity, directory map, and a few cross-cutting hard constraints. Add one local `AGENTS.md` under `test/` for the only area where local execution rules are worth isolating. Do not create additional local agent docs unless the codebase proves they are necessary.

**Tech Stack:** Markdown, existing repository documentation, Go test commands, Make targets

---

## File Structure

- Modify: `AGENTS.md` - shrink the root file into a real index with minimal global rules.
- Create: `test/AGENTS.md` - hold local test guidance that does not belong in the root index.
- Reference: `README.md` - keep terminology aligned with the repository overview.
- Reference: `docs/README.md` - avoid duplicating end-user documentation links.

### Task 1: Replace the root handbook with a minimal index

**Files:**
- Modify: `AGENTS.md`
- Reference: `README.md`

- [ ] **Step 1: Capture the current scope before editing**

Run: `wc -l AGENTS.md`
Expected: the file is much larger than an index page and should report well over 100 lines.

- [ ] **Step 2: Replace `AGENTS.md` with this exact minimal structure**

```md
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
- `test/`：测试约定与运行方式，细节见 `test/AGENTS.md`。
- `docs/`：面向用户的命令与配置文档。

## 全局硬约束
- 不要修改产品逻辑之外的用户未请求范围；优先沿用现有代码模式。
- 开始验证时先跑最小相关测试，不要一开始就跑完整测试套件。
- Go 代码使用 `gofmt` / `goimports` 约定；lint 使用 `golangci-lint`。
- 集成测试依赖 `REDMINE_URL`、`REDMINE_API_KEY`、`REDMINE_PROJECT_ID`。

## 使用方式
- 根 `AGENTS.md` 只保留跨目录索引与少量全局约束，不再充当工程手册。
- 进入具体区域时，优先看就近代码、测试和局部文档；测试相关规则看 `test/AGENTS.md`。
```

- [ ] **Step 3: Verify the root file is now index-sized**

Run: `wc -l AGENTS.md`
Expected: the file is now a short document, roughly a few dozen lines rather than a handbook.

- [ ] **Step 4: If the user explicitly requests a commit, create the root index commit**

```bash
git add AGENTS.md
git commit -m "docs: slim down root agent guide"
```

### Task 2: Add the only necessary local AGENTS file under `test/`

**Files:**
- Create: `test/AGENTS.md`
- Reference: `AGENTS.md`

- [ ] **Step 1: Create `test/AGENTS.md` with only local testing guidance**

```md
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

## 常用命令
- 单个 E2E：`go test -v -count=1 -run '^TestHelpCommand$' ./test/e2e`
- 单个集成：`go test -v -count=1 -run '^TestIssueCRUD$' ./test/integration`
- E2E 套件：`make e2e`
- 集成套件：`make integration`
```

- [ ] **Step 2: Verify the new local file is discoverable and minimal**

Run: `wc -l test/AGENTS.md && test -f test/AGENTS.md`
Expected: the file exists and is short, focused on test-only guidance.

- [ ] **Step 3: If the user explicitly requests a commit, create the test guidance commit**

```bash
git add test/AGENTS.md
git commit -m "docs: add test-specific agent guide"
```

### Task 3: Validate the simplified documentation shape

**Files:**
- Modify: `AGENTS.md`
- Create: `test/AGENTS.md`

- [ ] **Step 1: Check that only the intended AGENTS files exist**

Run: `rg --files -g 'AGENTS.md'`
Expected: only `AGENTS.md` and `test/AGENTS.md` are present unless the repo already contains other intentional AGENTS files.

- [ ] **Step 2: Read both files together to confirm the split matches the design**

Run: `python - <<'PY'
from pathlib import Path
for path in [Path('AGENTS.md'), Path('test/AGENTS.md')]:
    print(f'--- {path} ---')
    print(path.read_text())
PY`
Expected: the root file reads like an index, while `test/AGENTS.md` contains only test-local rules.

- [ ] **Step 3: If the user explicitly requests a commit, create the final docs commit**

```bash
git add AGENTS.md test/AGENTS.md
git commit -m "docs: reduce agent docs to a minimal index"
```
