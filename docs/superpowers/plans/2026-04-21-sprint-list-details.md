# Sprint List Details Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `redmine sprint list <project> --details` so sprint listings can optionally expand to full sprint records.

**Architecture:** Keep `sprint` as a thin Cobra wrapper. The command fetches the sprint index once, then either returns the lightweight slice or fans out per-sprint `GetSprint` calls through `client.BatchGetFunc` when `--details` is enabled. The shared output pipeline still owns JSON/table/raw formatting, so the command only decides which `[]agile.Sprint` slice to pass downstream.

**Tech Stack:** Go 1.23, Cobra, `internal/client.BatchGetFunc`, existing `resolver.WriteOutput`, `gofmt`

---

## File Structure

- Modify: `internal/resources/agile/types.go`
- Modify: `internal/resources/agile/client_test.go`
- Modify: `internal/resources/sprints/commands.go`
- Modify: `internal/resources/sprints/commands_test.go`
- Modify: `docs/commands.md`
- Modify: `docs/README.md`

### Task 1: Expand the sprint model

**Files:**
- Modify: `internal/resources/agile/types.go`
- Modify: `internal/resources/agile/client_test.go`

- [ ] **Step 1: Add the failing assertion first**

Update `TestClient_GetSprint` so it expects the detail payload to include a description:

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

- [ ] **Step 2: Verify the test fails for the right reason**

Run:

```bash
GOCACHE=/tmp/redmine-go-cache go test ./internal/resources/agile -run '^TestClient_GetSprint$' -v
```

Expected: fail because `Sprint` does not yet carry `Description`.

- [ ] **Step 3: Add the model field**

Change `Sprint` to include the missing field:

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

Keep the existing `SprintList` compatibility shim intact so both `agile_sprints` and `sprints` payloads still decode.

- [ ] **Step 4: Verify the test passes**

Run:

```bash
GOCACHE=/tmp/redmine-go-cache go test ./internal/resources/agile -run '^TestClient_GetSprint$' -v
```

Expected: PASS.

- [ ] **Step 5: Commit the model change**

```bash
git add internal/resources/agile/types.go internal/resources/agile/client_test.go
git commit -m "feat(agile): include sprint description in sprint detail"
```

---

### Task 2: Add `--details` expansion to `sprint list`

**Files:**
- Modify: `internal/resources/sprints/commands.go`
- Modify: `internal/resources/sprints/commands_test.go`

- [ ] **Step 1: Write the failing tests**

Add one success test, one table-render test, and one error test.

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

Use a transport-backed client in the test so these paths are covered without `httptest`:

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

- [ ] **Step 2: Verify the tests fail**

Run:

```bash
GOCACHE=/tmp/redmine-go-cache go test ./internal/resources/sprints -run 'TestListCommand_DetailsExpandsSprintPayload|TestListCommand_DetailsPropagatesDetailError|TestListCommand_DetailsRendersTablePayload' -v
```

Expected: fail because `--details` is not wired up yet and the detail payload is still lightweight.

- [ ] **Step 3: Implement the expansion path**

Add a `--details` flag and branch the command output:

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

Use `client.BatchGetFunc` so detail requests run concurrently and preserve list order:

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

- [ ] **Step 4: Verify the command tests pass**

Run:

```bash
GOCACHE=/tmp/redmine-go-cache go test ./internal/resources/sprints -run 'TestListCommand_DetailsExpandsSprintPayload|TestListCommand_DetailsPropagatesDetailError' -v
```

Expected: PASS.

- [ ] **Step 5: Commit the command change**

```bash
git add internal/resources/sprints/commands.go internal/resources/sprints/commands_test.go
git commit -m "feat(sprint): add optional detailed sprint listing"
```

---

### Task 3: Update docs and final verification

**Files:**
- Modify: `docs/commands.md`
- Modify: `docs/README.md`

- [ ] **Step 1: Update the user docs**

Add `--details` to the sprint command docs:

```md
redmine sprint list city --details --format table

| `--details` | Expand each sprint to full detail before output |

- `json`：sprint 数组，`--details` 时包含完整字段
- `table`：展示 sprint 的全部字段
- `raw`：单行 JSON
```

Also add the same example to `docs/README.md` quick links.

- [ ] **Step 2: Verify the documentation-linked tests pass**

Run:

```bash
GOCACHE=/tmp/redmine-go-cache go test ./internal/app ./internal/resources/agile ./internal/resources/sprints
rg -n "sprint list|--details" docs/commands.md docs/README.md
```

Expected: PASS.

- [ ] **Step 3: Commit the docs update**

```bash
git add docs/commands.md docs/README.md
git commit -m "docs: document detailed sprint listing"
```
