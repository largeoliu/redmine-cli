# Sprint 列表详情实现计划

> **面向智能体工作器：** 必需子技能：使用 superpowers:subagent-driven-development（推荐）或 superpowers:executing-plans 来逐步实现此计划任务。步骤使用复选框（`- [ ]`）语法进行跟踪。

**目标：** 添加 `redmine sprint list <project> --details` 功能，使 sprint 列表可以选择性地扩展为完整的 sprint 记录。

**架构设计：** 保持 `sprint` 作为精简的 Cobra 包装器。命令获取一次 sprint 索引，然后要么返回轻量级切片，要么在启用 `--details` 时通过 `client.BatchGetFunc` 扇出每个 sprint 的 `GetSprint` 调用。共享的输出管道仍然负责 JSON/table/raw 格式化，因此命令只需要决定将哪个 `[]agile.Sprint` 切片传递给下游。

**技术栈：** Go 1.23, Cobra, `internal/client.BatchGetFunc`, 现有的 `resolver.WriteOutput`, `gofmt`

---

## File Structure

- Modify: `internal/resources/agile/types.go`
- Modify: `internal/resources/agile/client_test.go`
- Modify: `internal/resources/sprints/commands.go`
- Modify: `internal/resources/sprints/commands_test.go`
- Modify: `docs/commands.md`
- Modify: `docs/README.md`

### 任务 1：扩展 sprint 模型

**文件：**
- 修改：`internal/resources/agile/types.go`
- 修改：`internal/resources/agile/client_test.go`

- [ ] **步骤 1：先添加失败的断言**

更新 `TestClient_GetSprint` 使其期望详情负载包含描述字段：

```go
response := struct {
	AgileSprint Sprint `json:"agile_sprint"`
}{
	AgileSprint: Sprint{
		ID:          7,
		Name:        "Sprint 7",
		Description: "Release hardening",
		Status:      "active",
		StartDate:   "2026-04-01",
		EndDate:     "2026-04-14",
		Goal:        "Finish release scope",
		IsDefault:   true,
	},
}

if result.Description != "Release hardening" {
	t.Fatalf("expected description to be populated, got %q", result.Description)
}
```

- [ ] **步骤 2：验证测试因正确原因失败**

运行：

```bash
GOCACHE=/tmp/redmine-go-cache go test ./internal/resources/agile -run '^TestClient_GetSprint$' -v
```

预期：失败，因为 `Sprint` 尚未携带 `Description` 字段。

- [ ] **步骤 3：添加模型字段**

更改 `Sprint` 以包含缺失的字段：

```go
type Sprint struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status,omitempty"`
	StartDate   string `json:"start_date,omitempty"`
	EndDate     string `json:"end_date,omitempty"`
	Goal        string `json:"goal,omitempty"`
	IsDefault   bool   `json:"is_default,omitempty"`
	IsClosed    bool   `json:"is_closed,omitempty"`
	IsArchived  bool   `json:"is_archived,omitempty"`
}
```

保持现有的 `SprintList` 兼容性垫片不变，以便 `agile_sprints` 和 `sprints` 负载都能正确解码。

- [ ] **步骤 4：验证测试通过**

运行：

```bash
GOCACHE=/tmp/redmine-go-cache go test ./internal/resources/agile -run '^TestClient_GetSprint$' -v
```

预期：通过。

- [ ] **步骤 5：提交模型更改**

```bash
git add internal/resources/agile/types.go internal/resources/agile/client_test.go
git commit -m "feat(agile): 在 sprint 详情中包含描述字段"
```

---

### 任务 2：为 `sprint list` 添加 `--details` 扩展

**文件：**
- 修改：`internal/resources/sprints/commands.go`
- 修改：`internal/resources/sprints/commands_test.go`

- [ ] **步骤 1：编写失败的测试**

添加一个成功测试、一个表格渲染测试和一个错误测试。

```go
func newSprintDetailsClient(t *testing.T) *client.Client {
	return client.NewClient("https://example.com", "test-key", client.WithHTTPClient(&http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/projects/42.json":
				return jsonHTTPResponse(t, http.StatusOK, map[string]any{
					"project": projectspkg.Project{ID: 42, Name: "City", Identifier: "city"},
				}), nil
			case "/projects/42/agile_sprints.json":
				return jsonHTTPResponse(t, http.StatusOK, map[string]any{
					"project_id":   42,
					"project_name": "City",
					"sprints": []map[string]any{
						{"id": 7, "name": "Sprint 7", "status": "active"},
						{"id": 8, "name": "Sprint 8", "status": "open"},
					},
				}), nil
			case "/projects/42/agile_sprints/7.json":
				return jsonHTTPResponse(t, http.StatusOK, map[string]any{
					"agile_sprint": map[string]any{
						"id":          7,
						"name":        "Sprint 7",
						"description": "Release hardening",
						"status":      "active",
						"start_date":  "2026-04-01",
						"end_date":    "2026-04-14",
						"goal":        "Finish release scope",
						"is_default":  true,
					},
				}), nil
			case "/projects/42/agile_sprints/8.json":
				return jsonHTTPResponse(t, http.StatusOK, map[string]any{
					"agile_sprint": map[string]any{
						"id":          8,
						"name":        "Sprint 8",
						"description": "Stabilization",
						"status":      "open",
						"start_date":  "2026-04-15",
						"end_date":    "2026-04-28",
						"goal":        "Polish release",
						"is_default":  false,
					},
				}), nil
			default:
				t.Fatalf("unexpected path: %s", req.URL.Path)
				return nil, nil
			}
		}),
	}))
}

func newSprintDetailsErrorClient(t *testing.T) *client.Client {
	return client.NewClient("https://example.com", "test-key", client.WithHTTPClient(&http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/projects/42.json":
				return jsonHTTPResponse(t, http.StatusOK, map[string]any{
					"project": projectspkg.Project{ID: 42, Name: "City", Identifier: "city"},
				}), nil
			case "/projects/42/agile_sprints.json":
				return jsonHTTPResponse(t, http.StatusOK, map[string]any{
					"project_id":   42,
					"project_name": "City",
					"sprints": []map[string]any{
						{"id": 7, "name": "Sprint 7", "status": "active"},
						{"id": 8, "name": "Sprint 8", "status": "open"},
					},
				}), nil
			case "/projects/42/agile_sprints/7.json":
				return jsonHTTPResponse(t, http.StatusOK, map[string]any{
					"agile_sprint": map[string]any{
						"id":          7,
						"name":        "Sprint 7",
						"description": "Release hardening",
						"status":      "active",
					},
				}), nil
			case "/projects/42/agile_sprints/8.json":
				return jsonHTTPResponse(t, http.StatusNotFound, map[string]any{
					"errors": []string{"Not found"},
				}), nil
			default:
				t.Fatalf("unexpected path: %s", req.URL.Path)
				return nil, nil
			}
		}),
	}))
}

func TestListCommand_DetailsExpandsSprintPayload(t *testing.T) {
	flags := &types.GlobalFlags{Format: "json"}
	var payload any
	c := newSprintDetailsClient(t)
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) { return c, nil },
		writeOutputFunc: func(_ io.Writer, _ *types.GlobalFlags, p any) error {
			payload = p
			return nil
		},
	}

	cmd := newListCommand(flags, resolver)
	cmd.SetArgs([]string{"--details", "42"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sprints, ok := payload.([]agilepkg.Sprint)
	if !ok {
		t.Fatalf("expected []agile.Sprint payload, got %T", payload)
	}
	if sprints[0].Description != "Release hardening" {
		t.Fatalf("expected expanded description, got %+v", sprints[0])
	}
}

func TestListCommand_DetailsRendersTablePayload(t *testing.T) {
	flags := &types.GlobalFlags{Format: "table"}
	var rendered bytes.Buffer
	c := newSprintDetailsClient(t)
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) { return c, nil },
		writeOutputFunc: func(_ io.Writer, _ *types.GlobalFlags, p any) error {
			return output.Write(&rendered, output.FormatTable, p)
		},
	}

	cmd := newListCommand(flags, resolver)
	cmd.SetArgs([]string{"--details", "42"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(rendered.String(), "Release hardening") {
		t.Fatalf("expected table rendering to include the description, got:\n%s", rendered.String())
	}
}

func TestListCommand_DetailsPropagatesDetailError(t *testing.T) {
	flags := &types.GlobalFlags{Format: "json"}
	c := newSprintDetailsErrorClient(t)
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) { return c, nil },
		writeOutputFunc: func(_ io.Writer, _ *types.GlobalFlags, _ any) error {
			return nil
		},
	}

	cmd := newListCommand(flags, resolver)
	cmd.SetArgs([]string{"--details", "42"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error from sprint detail fetch, got nil")
	}
}
```

使用传输层支持的客户端进行测试，这样可以在不使用 `httptest` 的情况下覆盖这些路径：

```go
Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
	switch req.URL.Path {
	case "/projects/42.json":
		// project lookup
	case "/projects/42/agile_sprints.json":
		// sprint index with IDs 7 and 8
	case "/projects/42/agile_sprints/7.json":
		// detailed sprint payload
	case "/projects/42/agile_sprints/8.json":
		// detailed sprint payload
	default:
		t.Fatalf("unexpected path: %s", req.URL.Path)
	}
	return nil, nil
})
```

- [ ] **步骤 2：验证测试失败**

运行：

```bash
GOCACHE=/tmp/redmine-go-cache go test ./internal/resources/sprints -run 'TestListCommand_DetailsExpandsSprintPayload|TestListCommand_DetailsPropagatesDetailError|TestListCommand_DetailsRendersTablePayload' -v
```

预期：失败，因为 `--details` 尚未连接，且详情负载仍然是轻量级的。

- [ ] **步骤 3：实现扩展路径**

添加 `--details` 标志并分支命令输出：

```go
details := false
cmd.Flags().BoolVar(&details, "details", false, "Expand each sprint with full details")

payload := result.AgileSprints
if details {
	payload, err = loadSprintDetails(cmd.Context(), c, project.ID, result.AgileSprints)
	if err != nil {
		return err
	}
}

return resolver.WriteOutput(cmd.OutOrStdout(), flags, payload)
```

使用 `client.BatchGetFunc` 使详情请求并发运行并保持列表顺序：

```go
func loadSprintDetails(ctx context.Context, c *client.Client, projectID int, sprints []agilepkg.Sprint) ([]agilepkg.Sprint, error) {
	ids := make([]int, len(sprints))
	for i, sprint := range sprints {
		ids[i] = sprint.ID
	}

	results := client.BatchGetFunc(ids, ctx, func(innerCtx context.Context, _ int, sprintID int) (agilepkg.Sprint, error) {
		sprint, err := agilepkg.NewClient(c).GetSprint(innerCtx, projectID, sprintID)
		if err != nil {
			return agilepkg.Sprint{}, err
		}
		return *sprint, nil
	}, 5)

	detailed := make([]agilepkg.Sprint, len(results))
	for i, result := range results {
		if result.Error != nil {
			return nil, result.Error
		}
		detailed[i] = result.Result
	}
	return detailed, nil
}
```

- [ ] **步骤 4：验证命令测试通过**

运行：

```bash
GOCACHE=/tmp/redmine-go-cache go test ./internal/resources/sprints -run 'TestListCommand_DetailsExpandsSprintPayload|TestListCommand_DetailsPropagatesDetailError' -v
```

预期：通过。

- [ ] **步骤 5：提交命令更改**

```bash
git add internal/resources/sprints/commands.go internal/resources/sprints/commands_test.go
git commit -m "feat(sprint): 添加可选的详细 sprint 列表功能"
```

---

### 任务 3：更新文档和最终验证

**文件：**
- 修改：`docs/commands.md`
- 修改：`docs/README.md`

- [ ] **步骤 1：更新用户文档**

将 `--details` 添加到 sprint 命令文档中：

```md
redmine sprint list city --details --format table

| `--details` | Expand each sprint to full detail before output |

- `json`：sprint 数组，`--details` 时包含完整字段
- `table`：展示 sprint 的全部字段
- `raw`：单行 JSON
```

同时将相同的示例添加到 `docs/README.md` 的快速链接中。

- [ ] **步骤 2：验证与文档相关的测试通过**

运行：

```bash
GOCACHE=/tmp/redmine-go-cache go test ./internal/app ./internal/resources/agile ./internal/resources/sprints
rg -n "sprint list|--details" docs/commands.md docs/README.md
```

预期：通过。

- [ ] **步骤 3：提交文档更新**

```bash
git add docs/commands.md docs/README.md
git commit -m "docs: 记录详细 sprint 列表功能"
```
