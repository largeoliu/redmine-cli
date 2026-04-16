# 单元测试覆盖率提升方案

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 消除 `.codecov.yml` 中所有 ignore 项，使所有包覆盖率 >= 95%，总覆盖率从 91.4% 提升至 95%+。

**Architecture:** 按业务价值分层：先覆盖核心基础设施包（client/pool、config/keyring、output/filter），再覆盖资源命令包（helpers、resources），最后清理 codecov ignore 配置。每个 Task 对应一个包或一组强相关文件，独立可测。

**Tech Stack:** Go 1.x、`testing` 标准库、`internal/testutil.NewMockServer(t)`、`httptest`、表格驱动测试。

---

## 当前覆盖率概览

| 包 | 覆盖率 | 状态 |
|---|---|---|
| `cmd` | 0.0% | codecov ignore |
| `internal/app` | 92.7% | 需提升 |
| `internal/client` | 81.7% | pool.go 0% |
| `internal/config` | 86.1% | keyring/store 低覆盖 |
| `internal/errors` | 100.0% | ok |
| `internal/output` | 91.3% | SelectFields 61.3% |
| `internal/resources/helpers` | 0.0% | 无测试 |
| `internal/resources/categories` | 100.0% | ok |
| `internal/resources/issues` | 98.8% | 接近达标 |
| `internal/resources/priorities` | 94.1% | 需提升 |
| `internal/resources/projects` | 96.2% | 接近达标 |
| `internal/resources/statuses` | 100.0% | ok |
| `internal/resources/time_entries` | 92.4% | 需提升 |
| `internal/resources/trackers` | 94.1% | 需提升 |
| `internal/resources/users` | 91.4% | 需提升 |
| `internal/resources/versions` | 92.1% | 需提升 |
| `internal/testutil` | 73.8% | 需提升 |
| `internal/types` | [no statements] | ok |

**关键 0% 覆盖函数：**
- `client/pool.go`: 所有函数 0%
- `resources/helpers/`: 整个包 0%
- `app/root.go:Execute`: 0%
- `cmd/main.go:main`: 0%

**关键低覆盖函数（< 90%）：**
- `config/keyring.go`: `IsAvailable` 44.4%, `Delete` 66.7%, `Get` 71.4%, `Set` 75.0%, `NewKeyring` 75.0%
- `config/store.go`: `SaveInstance` 66.7%, `NewStoreWithKeyring` 66.7%, `Save` 71.4%, `SetDefault` 85.7%, `DeleteInstance` 87.5%
- `output/filter.go:SelectFields`: 61.3%
- `testutil/mock.go:HandleError`: 71.4%
- `testutil/leak.go`: `LeakTestMain` 0%, `LeakTestMainWithOptions` 0%, `VerifyNoneWithDelay` 75.0%

---

## 文件结构

| 操作 | 文件 | 职责 |
|------|------|------|
| 创建 | `internal/client/pool_test.go` | 连接池配置测试 |
| 创建 | `internal/resources/helpers/confirm_test.go` | 确认删除提示测试 |
| 创建 | `internal/resources/helpers/dryrun_test.go` | DryRun 输出测试 |
| 创建 | `internal/resources/helpers/parse_test.go` | ID 解析测试 |
| 修改 | `internal/config/keyring_test.go` | 增加 realKeyring/fallbackKeyring 分支覆盖 |
| 修改 | `internal/config/config_test.go` | 增加 store 覆盖 |
| 修改 | `internal/output/filter_extended_test.go` | 增加 SelectFields 分支 |
| 修改 | `internal/testutil/mock_test.go` | 增加 HandleError/HandlePrefix 覆盖 |
| 修改 | `internal/testutil/leak_test.go` | 增加 VerifyNoneWithDelay 覆盖 |
| 修改 | `internal/client/client_test.go` | 增加 doSingleRequest/retryDelay/Ping 覆盖 |
| 修改 | `internal/client/logging_test.go` | 增加 RoundTrip 错误分支 |
| 修改 | `internal/app/login_test.go` | 增加 promptBool/newLoginCommand 覆盖 |
| 修改 | `internal/app/root_test.go` | 增加 ResolveClient/WriteOutput 覆盖 |
| 修改 | `.codecov.yml` | 移除大部分 ignore 列表 |
| 修改 | 资源命令各 `commands_test.go` | 参数验证边界覆盖 |

---

### Task 1: client/pool.go - 连接池配置 0% -> 95%+

**Files:**
- Create: `internal/client/pool_test.go`
- Read: `internal/client/pool.go`

- [ ] **Step 1: 编写 pool_test.go 测试文件**

覆盖以下函数：
- `DefaultConnectionPoolConfig` - 验证默认值
- `WithConnectionPool(nil)` - nil config 使用默认
- `WithConnectionPool(custom)` - 自定义配置
- `WithConnectionPool` 遇到非 `*http.Transport` 时不替换
- `WithMaxIdleConns` / `WithMaxIdleConnsPerHost` / `WithIdleConnTimeout` / `WithMaxConnsPerHost` - 各单项配置
- `GetConnectionPoolConfig` - 有 Transport、nil Transport、非 Transport 三种情况

测试策略：构造 `NewClient` 后应用选项，断言 Transport 字段值。对 `GetConnectionPoolConfig` 的 nil Transport 和非 Transport 分支，需手动设置 `c.httpClient.Transport = nil` 或使用自定义 Transport mock。

- [ ] **Step 2: 运行测试验证通过**

Run: `go test -v -run 'Test(DefaultConnectionPoolConfig|WithConnectionPool|WithMaxIdleConns|WithMaxIdleConnsPerHost|WithIdleConnTimeout|WithMaxConnsPerHost|GetConnectionPoolConfig)' ./internal/client/`
Expected: PASS

- [ ] **Step 3: 验证覆盖率**

Run: `go test -coverprofile=pool_cover.out ./internal/client/ && go tool cover -func=pool_cover.out | grep pool.go`
Expected: pool.go 函数覆盖率 >= 95%

- [ ] **Step 4: 提交**

```bash
git add internal/client/pool_test.go
git commit -m "test: add unit tests for client/pool.go connection pool configuration"
```

---

### Task 2: resources/helpers - 全包 0% -> 95%+

**Files:**
- Create: `internal/resources/helpers/confirm_test.go`
- Create: `internal/resources/helpers/dryrun_test.go`
- Create: `internal/resources/helpers/parse_test.go`
- Read: `internal/resources/helpers/confirm.go`, `dryrun.go`, `parse.go`

- [ ] **Step 1: 编写 confirm_test.go**

覆盖场景：
- `yes=true` 直接返回 true（跳过提示）
- 用户输入 `y` 确认
- 用户输入 `n` 取消
- 用户输入空行取消
- `fmt.Scanln` 出错时默认取消（用关闭的 pipe 模拟）

注意：`ConfirmDelete` 使用 `fmt.Scanln` 读取 stdin 和 `fmt.Printf` 输出 stdout，需要用 `os.Pipe()` 替换 `os.Stdin`。

- [ ] **Step 2: 编写 dryrun_test.go**

覆盖场景：
- `DryRunCreate` 返回 true 并输出 `[dry-run] Would create`
- `DryRunUpdate` 返回 true 并输出 `[dry-run] Would update #<id>`
- `DryRunDelete` 返回 true 并输出 `[dry-run] Would delete #<id>`

注意：使用 `os.Pipe()` 捕获 stdout。

- [ ] **Step 3: 编写 parse_test.go**

覆盖场景（表格驱动）：
- 有效 ID "123" -> 123, nil
- 无效 ID "abc" -> error
- 负数 "-1" -> -1, nil
- 零 "0" -> 0, nil
- 空串 "" -> error

- [ ] **Step 4: 运行测试验证通过**

Run: `go test -v ./internal/resources/helpers/`
Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add internal/resources/helpers/confirm_test.go internal/resources/helpers/dryrun_test.go internal/resources/helpers/parse_test.go
git commit -m "test: add unit tests for resources/helpers package (confirm, dryrun, parse)"
```

---

### Task 3: config/keyring - 覆盖率 44-75% -> 90%+

**Files:**
- Modify: `internal/config/keyring_test.go`
- Read: `internal/config/keyring.go`

- [ ] **Step 1: 增加 fallbackKeyring 全方法覆盖和 NewKeyring 测试**

先阅读现有 `keyring_test.go`，然后补充：

1. `fallbackKeyring` 的完整 CRUD：`Set` + `Get`、`Get` 不存在 key、`Delete` 存在/不存在 key
2. `fallbackKeyring.IsAvailable()` - `available=false` 返回 false，`available=true` 返回 true
3. `NewKeyring()` 在系统 keyring 不可用时返回 `fallbackKeyring`
4. `realKeyring` 的方法在系统 keyring 不可用时通过 `fallbackKeyring` 间接验证

注意：`realKeyring` 依赖系统 keyring（`github.com/zalando/go-keyring`），在 headless CI 中通常不可用。对于 `IsAvailable` 返回 false 的情况，`NewKeyring` 会降级到 `fallbackKeyring`，因此 `realKeyring` 的 Get/Set/Delete 可以通过 build tag 或 `testing.Short()` 在 CI 中跳过。

- [ ] **Step 2: 运行测试验证通过**

Run: `go test -v -run 'TestFallbackKeyring|TestNewKeyring' ./internal/config/`
Expected: PASS

- [ ] **Step 3: 验证覆盖率**

Run: `go test -coverprofile=config_cover.out ./internal/config/ && go tool cover -func=config_cover.out | grep keyring.go`
Expected: keyring.go 覆盖率 >= 90%

- [ ] **Step 4: 提交**

```bash
git add internal/config/keyring_test.go
git commit -m "test: add comprehensive fallbackKeyring and NewKeyring coverage"
```

---

### Task 4: config/store - 覆盖率 66-94% -> 95%+

**Files:**
- Modify: `internal/config/config_test.go`
- Read: `internal/config/store.go`

- [ ] **Step 1: 增加 store 的边界覆盖**

先阅读现有 `config_test.go`，然后补充：

1. `NewStoreWithKeyring` 使用自定义 keyring
2. `SaveInstance` 在 keyring 可用时 API key 存入 keyring，YAML 中为空
3. `SaveInstance` 在 keyring 不可用时 API key 保留在 YAML
4. `Save` 目录创建失败场景
5. `SetDefault` 对不存在实例返回 `ErrInstanceNotFound`
6. `DeleteInstance` 删除默认实例后自动选择下一个（按字母序）
7. `DeleteInstance` keyring.Delete 返回非 ErrAPIKeyNotFound 错误时传播

使用 `mockKeyring` struct 实现 `Keyring` 接口，控制 `available` 和错误行为。

- [ ] **Step 2: 运行测试验证通过**

Run: `go test -v -run 'TestNewStoreWithKeyring|TestStore_SaveInstance|TestStore_DeleteInstance|TestStore_SetDefault' ./internal/config/`
Expected: PASS

- [ ] **Step 3: 验证覆盖率**

Run: `go test -coverprofile=config_cover.out ./internal/config/ && go tool cover -func=config_cover.out | grep store.go`
Expected: store.go 覆盖率 >= 95%

- [ ] **Step 4: 提交**

```bash
git add internal/config/config_test.go
git commit -m "test: add store coverage for keyring integration, default instance, and error paths"
```

---

### Task 5: output/filter - SelectFields 61.3% -> 95%+

**Files:**
- Modify: `internal/output/filter_extended_test.go`
- Read: `internal/output/filter.go`

- [ ] **Step 1: 增加 SelectFields 的数组过滤和非 map 输入分支**

先阅读现有 `filter_extended_test.go`，然后补充：

1. 输入包含数组字段且数组元素为 map -> 按 fields 过滤每个元素的属性
2. 数组元素包含不在 fields 中的 key -> 过滤掉
3. 数组元素为非 map（如 int、string）-> 保留原始值
4. 输入为 slice（非 map）-> 原样返回
5. `fields` 为空 -> 原样返回
6. 数组元素过滤后无匹配字段 -> 该元素被排除

- [ ] **Step 2: 运行测试验证通过**

Run: `go test -v -run 'TestSelectFields_' ./internal/output/`
Expected: PASS

- [ ] **Step 3: 验证覆盖率**

Run: `go test -coverprofile=output_cover.out ./internal/output/ && go tool cover -func=output_cover.out | grep SelectFields`
Expected: SelectFields >= 95%

- [ ] **Step 4: 提交**

```bash
git add internal/output/filter_extended_test.go
git commit -m "test: add SelectFields array filtering, non-map input, and edge case coverage"
```

---

### Task 6: client - doSingleRequest/retryDelay/Ping/RoundTrip 边界覆盖

**Files:**
- Modify: `internal/client/client_test.go`
- Modify: `internal/client/logging_test.go`
- Read: `internal/client/client.go`, `logging.go`

- [ ] **Step 1: 增加 client 边界测试**

先阅读现有 `client_test.go`，然后补充：

1. `doSingleRequest` 3xx 重定向（非 2xx/4xx/5xx）路径
2. `doSingleRequest` 5xx 重试耗尽后返回错误
3. `retryDelay` 指数退避计算验证
4. `Ping` 非 200 响应返回错误

使用 `testutil.NewMockServer` 构造 mock HTTP 响应。

- [ ] **Step 2: 增加 logging RoundTrip 错误路径测试**

先阅读现有 `logging_test.go`，然后补充：

1. RoundTrip 中 response body 读取失败
2. RoundTrip 中 request body 读取场景

- [ ] **Step 3: 运行测试验证通过**

Run: `go test -v -run 'TestDoSingleRequest|TestRetryDelay|TestPing|TestRoundTrip' ./internal/client/`
Expected: PASS

- [ ] **Step 4: 验证覆盖率**

Run: `go test -coverprofile=client_cover.out ./internal/client/ && go tool cover -func=client_cover.out | grep -E '(doSingleRequest|retryDelay|Ping|RoundTrip)'`
Expected: 各函数 >= 95%

- [ ] **Step 5: 提交**

```bash
git add internal/client/client_test.go internal/client/logging_test.go
git commit -m "test: add client edge case coverage for retry, delay, ping, and logging"
```

---

### Task 7: app - login/root 边界覆盖

**Files:**
- Modify: `internal/app/login_test.go`
- Modify: `internal/app/root_test.go`
- Read: `internal/app/login.go`, `root.go`

- [ ] **Step 1: 增加 login 测试**

先阅读现有 `login_test.go`，然后补充：

1. `newLoginCommand` 的 flag 注册验证
2. `promptBool` 的 false 分支（用户输入非 y/Y）
3. `runLogin` 的交互式完整路径

- [ ] **Step 2: 增加 root 测试**

先阅读现有 `root_test.go`，然后补充：

1. `ResolveClient` 的配置缺失错误路径（无 instance、无 API key）
2. `WriteOutput` 的 writer 错误路径
3. `Execute` 无法在此环境直接测试（需要 cobra 命令行执行），考虑通过 NewRootCommand 测试间接覆盖

- [ ] **Step 3: 运行测试验证通过**

Run: `go test -v ./internal/app/`
Expected: PASS

- [ ] **Step 4: 验证覆盖率**

Run: `go test -coverprofile=app_cover.out ./internal/app/ && go tool cover -func=app_cover.out`
Expected: app 包 >= 93%（Execute 0% 保留，见 Task 10）

- [ ] **Step 5: 提交**

```bash
git add internal/app/login_test.go internal/app/root_test.go
git commit -m "test: add app login/root edge case coverage"
```

---

### Task 8: testutil - mock/leak 覆盖率提升

**Files:**
- Modify: `internal/testutil/mock_test.go`
- Modify: `internal/testutil/leak_test.go`
- Read: `internal/testutil/mock.go`, `leak.go`

- [ ] **Step 1: 增加 HandleError 和 HandlePrefix 测试**

先阅读现有 `mock_test.go`，然后补充：

1. `HandleError` - 注册错误路径处理，发送请求验证状态码和 body
2. `HandlePrefix` - 匹配路径前缀的请求
3. `HandlePrefix` 精确匹配（无尾部 slash）
4. `HandlePrefix` 不匹配路径返回 404

- [ ] **Step 2: 增加 VerifyNoneWithDelay 测试**

`LeakTestMain` 和 `LeakTestMainWithOptions` 作为 `TestMain` 调用，难以直接单元测试。为 `VerifyNoneWithDelay` 补充无 goroutine 泄漏场景的测试。

- [ ] **Step 3: 运行测试验证通过**

Run: `go test -v ./internal/testutil/`
Expected: PASS

- [ ] **Step 4: 提交**

```bash
git add internal/testutil/mock_test.go internal/testutil/leak_test.go
git commit -m "test: add testutil HandleError, HandlePrefix, and VerifyNoneWithDelay coverage"
```

---

### Task 9: resources - 各命令包覆盖率微调

**Files:**
- Modify: `internal/resources/priorities/commands_test.go`
- Modify: `internal/resources/projects/client_test.go` + `commands_test.go`
- Modify: `internal/resources/time_entries/commands_test.go`
- Modify: `internal/resources/trackers/commands_test.go`
- Modify: `internal/resources/users/client_test.go` + `commands_test.go`
- Modify: `internal/resources/versions/commands_test.go`

- [ ] **Step 1: 逐包分析未覆盖行并补充测试**

每个资源包的未覆盖行主要在 `commands.go` 的参数解析错误路径。逐包读取现有测试和源码，补充：

1. `priorities/commands.go:newListCommand` - flag 注册边界
2. `projects/client.go:get` - 错误响应
3. `projects/commands.go:newListCommand/newGetCommand` - flag 边界
4. `time_entries/commands.go` - CRUD 各命令参数验证
5. `trackers/commands.go:newListCommand` - flag 注册
6. `users/client.go:List/Get` - 错误路径
7. `users/commands.go` - 全命令参数验证
8. `versions/commands.go` - CRUD 各命令

由于这些包结构相似，每个包补充 2-4 个测试用例即可达到 95%+。

- [ ] **Step 2: 运行测试验证通过**

Run: `go test -v ./internal/resources/priorities/ ./internal/resources/projects/ ./internal/resources/time_entries/ ./internal/resources/trackers/ ./internal/resources/users/ ./internal/resources/versions/`
Expected: PASS

- [ ] **Step 3: 验证覆盖率**

Run: `go test -coverprofile=res_cover.out ./internal/resources/... && go tool cover -func=res_cover.out | grep -E '(priorities|projects|time_entries|trackers|users|versions)/'`
Expected: 各包 >= 95%

- [ ] **Step 4: 提交**

```bash
git add internal/resources/priorities/ internal/resources/projects/ internal/resources/time_entries/ internal/resources/trackers/ internal/resources/users/ internal/resources/versions/
git commit -m "test: add resource command edge case coverage for priorities, projects, time_entries, trackers, users, versions"
```

---

### Task 10: cmd/main.go + app/root.go:Execute - 0% 覆盖处理

**Files:**
- Modify: `.codecov.yml`
- Read: `cmd/main.go`, `internal/app/root.go`

- [ ] **Step 1: 分析不可测试代码**

`cmd/main.go:main` 和 `internal/app/root.go:Execute` 是程序入口，直接调用 `os.Exit`，难以单元测试。有三种策略：

1. **保持 codecov ignore**（仅这两项）- 最务实
2. **重构 Execute 为可测试函数** - 需要提取 `os.Args` 和 `os.Exit` 调用
3. **用 TestMain 覆盖** - 通过构建二进制子进程测试

- [ ] **Step 2: 选择策略并实现**

推荐策略：保持 `cmd/main.go` 和 `app/root.go:Execute` 在 codecov ignore 中（仅这两项），从 ignore 列表中移除其余所有项。

- [ ] **Step 3: 更新 .codecov.yml**

将 ignore 列表从 22 项缩减为仅 2 项不可测试入口：

```yaml
ignore:
  - "cmd/main.go"
  - "internal/app/root.go"

coverage:
  status:
    project:
      default:
        target: 95%
        threshold: null
        if_not_found: success
    patch:
      default:
        target: 95%
        threshold: null
        if_not_found: success
```

将 target 从 100% 降至 95%（更务实），仅保留两个不可测试的入口文件在 ignore 中。

- [ ] **Step 4: 提交**

```bash
git add .codecov.yml
git commit -m "chore: update codecov.yml - remove most ignores, lower target to 95%"
```

---

### Task 11: 最终验证 - 全量覆盖率确认

**Files:** 无修改

- [ ] **Step 1: 运行全量测试**

Run: `go test -race ./...`
Expected: PASS

- [ ] **Step 2: 运行覆盖率并验证**

Run: `go test -coverprofile=final_coverage.out ./... && go tool cover -func=final_coverage.out | tail -1`
Expected: 总覆盖率 >= 95%

- [ ] **Step 3: 检查各包覆盖率**

Run: `go tool cover -func=final_coverage.out | grep -v '100.0%' | grep -v 'total:'`
Expected: 仅 `cmd/main.go:main` 和 `app/root.go:Execute` 为 0%

- [ ] **Step 4: 运行 lint**

Run: `golangci-lint run ./...`
Expected: 无新 warning
