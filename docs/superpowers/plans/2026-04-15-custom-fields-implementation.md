# Custom Fields Support Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Enable users to query tracker custom fields and fill them when creating/updating issues via interactive selection or --custom-field flags.

**Architecture:** Add tracker get command for discovering custom fields. Extend issue create/update with --custom-field flags and interactive mode. Custom field values sent via Redmine's custom_fields JSON array.

**Tech Stack:** Go, cobra, spf13/cast for type conversion.

---

## File Structure

```
internal/
├── resources/
│   ├── trackers/
│   │   ├── types.go           # MODIFY: Add CustomFields to Tracker struct
│   │   ├── client.go           # MODIFY: Add Get(id) method
│   │   ├── commands.go         # MODIFY: Add newGetCommand
│   │   └── client_test.go      # MODIFY: Add test for Get
│   └── issues/
│       ├── types.go           # MODIFY: Add CustomFields to Create/Update requests
│       ├── commands.go        # MODIFY: Add --custom-field flags, interactive flow
│       └── client_test.go     # MODIFY: Add custom field in create test
```

---

## Task 1: Extend Tracker Types and Client with Get and CustomFields

**Files:**
- Modify: `internal/resources/trackers/types.go:1-15`
- Modify: `internal/resources/trackers/client.go:1-27`
- Modify: `internal/resources/trackers/commands.go:1-38`

- [ ] **Step 1: Update Tracker types.go - Add CustomField and PossibleValues types**

Read and modify `internal/resources/trackers/types.go`:

```go
package trackers

// ValueLabel represents a list field option.
type ValueLabel struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// TrackerCustomField represents a custom field definition on a tracker.
type TrackerCustomField struct {
	ID             int           `json:"id"`
	Name           string        `json:"name"`
	FieldFormat    string        `json:"field_format"`  // string, list, bool, date, text
	PossibleValues []ValueLabel  `json:"possible_values,omitempty"`
}

// Tracker represents a Redmine tracker.
type Tracker struct {
	ID            int                 `json:"id"`
	Name          string              `json:"name"`
	DefaultStatus *int               `json:"default_status,omitempty"`
	Description   string              `json:"description,omitempty"`
	CustomFields  []TrackerCustomField `json:"custom_fields,omitempty"`
}

// TrackerList represents a list of trackers.
type TrackerList struct {
	Trackers []Tracker `json:"trackers"`
}
```

- [ ] **Step 2: Update client.go - Add Get(id) method**

Read and modify `internal/resources/trackers/client.go`:

Add after `List` method:

```go
// Get retrieves a tracker by ID.
func (c *Client) Get(ctx context.Context, id int) (*Tracker, error) {
	var result struct {
		Tracker *Tracker `json:"tracker"`
	}
	if err := c.client.Get(ctx, "/trackers/"+strconv.Itoa(id)+".json", &result); err != nil {
		return nil, err
	}
	return result.Tracker, nil
}
```

Add import: `strconv` to the existing import block.

- [ ] **Step 3: Update commands.go - Add newGetCommand**

Read and modify `internal/resources/trackers/commands.go`:

Add to imports:
```go
"github.com/largeoliu/redmine-cli/internal/resources/helpers"
```

Add to `NewCommand`:
```go
cmd.AddCommand(newGetCommand(flags, resolver))
```

Add new function after `newListCommand`:

```go
func newGetCommand(flags *types.GlobalFlags, resolver types.Resolver) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get tracker details including custom fields",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := helpers.ParseID(args[0], "tracker")
			if err != nil {
				return err
			}
			c, err := resolver.ResolveClient(flags)
			if err != nil {
				return err
			}
			result, err := NewClient(c).Get(cmd.Context(), id)
			if err != nil {
				return err
			}
			return resolver.WriteOutput(cmd.OutOrStdout(), flags, result)
		},
	}
	return cmd
}
```

- [ ] **Step 4: Add test for Get method in client_test.go**

Read `internal/resources/trackers/client_test.go` and add:

```go
func TestClient_Get(t *testing.T) {
	// ... similar pattern to existing tests
	// Test that Get calls /trackers/{id}.json and parses response
}
```

- [ ] **Step 5: Commit**

```bash
git add internal/resources/trackers/types.go internal/resources/trackers/client.go internal/resources/trackers/commands.go internal/resources/trackers/client_test.go
git commit -m "feat(tracker): add get command with custom fields support"
```

---

## Task 2: Add CustomFields to Issue Create/Update Requests

**Files:**
- Modify: `internal/resources/issues/types.go:53-85`

- [ ] **Step 1: Add CustomFields to IssueCreateRequest**

Read `internal/resources/issues/types.go` and add to `IssueCreateRequest`:

```go
type IssueCreateRequest struct {
	ProjectID      int    `json:"project_id,omitempty"`
	Subject        string `json:"subject,omitempty"`
	Description    string `json:"description,omitempty"`
	TrackerID      int    `json:"tracker_id,omitempty"`
	StatusID       int    `json:"status_id,omitempty"`
	PriorityID     int    `json:"priority_id,omitempty"`
	AssignedToID   int    `json:"assigned_to_id,omitempty"`
	CategoryID     int    `json:"category_id,omitempty"`
	FixedVersionID int    `json:"fixed_version_id,omitempty"`
	ParentIssueID  int    `json:"parent_issue_id,omitempty"`
	StartDate      string `json:"start_date,omitempty"`
	DueDate        string `json:"due_date,omitempty"`
	DoneRatio      int    `json:"done_ratio,omitempty"`
	WatcherUserIDs []int  `json:"watcher_user_ids,omitempty"`
	CustomFields   []CustomField `json:"custom_fields,omitempty"`  // NEW
}
```

- [ ] **Step 2: Add CustomFields to IssueUpdateRequest**

Add same field to `IssueUpdateRequest`:

```go
type IssueUpdateRequest struct {
	Subject        string `json:"subject,omitempty"`
	Description    string `json:"description,omitempty"`
	StatusID       int    `json:"status_id,omitempty"`
	PriorityID     int    `json:"priority_id,omitempty"`
	AssignedToID   int    `json:"assigned_to_id,omitempty"`
	CategoryID     int    `json:"category_id,omitempty"`
	FixedVersionID int    `json:"fixed_version_id,omitempty"`
	StartDate      string `json:"start_date,omitempty"`
	DueDate        string `json:"due_date,omitempty"`
	DoneRatio      int    `json:"done_ratio,omitempty"`
	Notes          string `json:"notes,omitempty"`
	PrivateNotes   bool   `json:"private_notes,omitempty"`
	CustomFields   []CustomField `json:"custom_fields,omitempty"`  // NEW
}
```

- [ ] **Step 3: Verify client.go sends CustomFields in request body**

Read `internal/resources/issues/client.go` to verify Create/Update already send the full struct as JSON (they should).

- [ ] **Step 4: Add test for CustomFields in create request**

Read `internal/resources/issues/client_test.go`, find the TestCreate function, and verify the request JSON includes custom_fields when provided.

- [ ] **Step 5: Commit**

```bash
git add internal/resources/issues/types.go
git commit -m "feat(issues): add CustomFields to create/update requests"
```

---

## Task 3: Add --custom-field Flag Parsing

**Files:**
- Modify: `internal/resources/issues/commands.go`

- [ ] **Step 1: Add customFieldFlags variable and parsing function**

Read `internal/resources/issues/commands.go`. Add at package level (after imports):

```go
// customFieldFlag holds parsed --custom-field flag values
type customFieldFlags struct {
	Fields []string
}

// parseCustomFieldFlags parses --custom-field "name:value" or "id:X:value" flags into issues.CustomField slice
func parseCustomFieldFlags(fields []string, tracker *trackers.Tracker) ([]issues.CustomField, error) {
	if len(fields) == 0 {
		return nil, nil
	}

	cfMap := make(map[int]issues.CustomField) // keyed by field ID

	for _, f := range fields {
		parts := strings.SplitN(f, ":", 3)
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid custom field format: %s (expected name:value or id:X:value)", f)
		}

		cf := issues.CustomField{}
		if parts[0] == "id" {
			id, err := strconv.Atoi(parts[1])
			if err != nil {
				return nil, fmt.Errorf("invalid custom field id: %s", parts[1])
			}
			cf.ID = id
			cf.Value = parts[2]
		} else {
			// Match by name - need to find ID from tracker definition
			if tracker == nil {
				return nil, fmt.Errorf("tracker required to match custom field by name: %s", parts[0])
			}
			found := false
			for _, tf := range tracker.CustomFields {
				if tf.Name == parts[0] {
					cf.ID = tf.ID
					cf.Value = parts[1]
					found = true
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("custom field not found: %s", parts[0])
			}
		}
		cfMap[cf.ID] = cf
	}

	result := make([]issues.CustomField, 0, len(cfMap))
	for _, cf := range cfMap {
		result = append(result, cf)
	}
	return result, nil
}
```

Add to imports: `"strings"`, `"strconv"`, `"fmt"`, and `"github.com/largeoliu/redmine-cli/internal/resources/trackers"` (aliased as `trackers` if not conflicting).

- [ ] **Step 2: Add --custom-field flag to newCreateCommand**

Modify `newCreateCommand` in `commands.go`:

Change:
```go
req := &IssueCreateRequest{}
```

To:
```go
req := &IssueCreateRequest{}
createFlags := &customFieldFlags{}
```

Add flag before `cmd.Flags().IntVar(&req.ProjectID...`:
```go
cmd.Flags().StringSliceVar(&createFlags.Fields, "custom-field", nil, "Custom field value (format: name:value or id:X:value, can be specified multiple times)")
```

Modify the RunE function - before calling Create, parse custom fields and attach to request. Add after `c, err := resolver.ResolveClient(flags)`:

```go
// Parse custom fields if provided
if len(createFlags.Fields) > 0 {
	cfs, err := parseCustomFieldFlags(createFlags.Fields, nil) // tracker not available in create
	if err != nil {
		return err
	}
	req.CustomFields = cfs
}
```

Note: In create flow, we don't have tracker info yet, so name-based matching won't work. Only id:X:value format works for create.

- [ ] **Step 3: Add --custom-field flag to newUpdateCommand**

Similar changes to `newUpdateCommand`:

```go
req := &IssueUpdateRequest{}
updateFlags := &customFieldFlags{}
```

Add flag:
```go
cmd.Flags().StringSliceVar(&updateFlags.Fields, "custom-field", nil, "Custom field value (format: name:value or id:X:value, can be specified multiple times)")
```

Add in RunE after client resolution:
```go
if len(updateFlags.Fields) > 0 {
	cfs, err := parseCustomFieldFlags(updateFlags.Fields, nil)
	if err != nil {
		return err
	}
	req.CustomFields = cfs
}
```

- [ ] **Step 4: Write test for parseCustomFieldFlags**

Create test in a new file `internal/resources/issues/custom_field_test.go`:

```go
package issues

import (
	"testing"
)

func TestParseCustomFieldFlags(t *testing.T) {
	tests := []struct {
		name    string
		flags   []string
		want    []issues.CustomField
		wantErr bool
	}{
		{
			name:  "id format single",
			flags: []string{"id:5:高"},
			want: []issues.CustomField{{ID: 5, Value: "高"}},
		},
		{
			name:  "id format multiple",
			flags: []string{"id:5:高", "id:6:v1.0"},
			want: []issues.CustomField{{ID: 5, Value: "高"}, {ID: 6, Value: "v1.0"}},
		},
		{
			name:  "invalid format",
			flags: []string{"invalid"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCustomFieldFlags(tt.flags, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseCustomFieldFlags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !equalCustomFields(got, tt.want) {
				t.Errorf("parseCustomFieldFlags() = %v, want %v", got, tt.want)
			}
		})
	}
}

func equalCustomFields(a, b []issues.CustomField) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].ID != b[i].ID || a[i].Value != b[i].Value {
			return false
		}
	}
	return true
}
```

- [ ] **Step 5: Commit**

```bash
git add internal/resources/issues/commands.go internal/resources/issues/custom_field_test.go
git commit -m "feat(issues): add --custom-field flag parsing"
```

---

## Task 4: Add Interactive Custom Field Mode

**Files:**
- Modify: `internal/resources/issues/commands.go`

- [ ] **Step 1: Add interactive prompt function**

Add to `commands.go` (after the parseCustomFieldFlags function):

```go
// promptCustomFields interactively asks user for custom field values.
// tracker param uses trackers.TrackerCustomField for field definitions.
func promptCustomFields(tracker *trackers.Tracker, initialValues map[int]issues.CustomField) ([]issues.CustomField, error) {
	if len(tracker.CustomFields) == 0 {
		return nil, nil
	}

	// Check if terminal is interactive using go-isatty
	if !isatty.IsTerminal(os.Stdout.Fd()) && !isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		return nil, nil // Skip in non-interactive mode
	}

	values := make(map[int]issues.CustomField)
	for k, v := range initialValues {
		values[k] = v
	}

	fmt.Println("\n📋 检测到自定义字段，请填写：\n")

	for _, cf := range tracker.CustomFields {
		current, hasCurrent := values[cf.ID]
		defaultVal := ""
		if hasCurrent {
			defaultVal = fmt.Sprintf("%v", current.Value)
		}

		fmt.Printf("[%d] %s\n", cf.ID, cf.Name)
		fmt.Printf("    类型: %s\n", cf.FieldFormat)

		var input string

		switch cf.FieldFormat {
		case "list":
			if len(cf.PossibleValues) > 0 {
				fmt.Print("    选项: ")
				for i, pv := range cf.PossibleValues {
					fmt.Printf("[%d. %s] ", i+1, pv.Label)
				}
				fmt.Println()
				fmt.Print("    选择: ")
				fmt.Scanln(&input)
				if input != "" {
					idx, err := strconv.Atoi(input)
					if err == nil && idx >= 1 && idx <= len(cf.PossibleValues) {
						values[cf.ID] = issues.CustomField{ID: cf.ID, Value: cf.PossibleValues[idx-1].Value}
					}
				} else if hasCurrent {
					values[cf.ID] = current
				}
			}
		case "bool":
			fmt.Print("    (y/n): ")
			fmt.Scanln(&input)
			if input == "y" || input == "Y" {
				values[cf.ID] = issues.CustomField{ID: cf.ID, Value: "1"}
			} else if input == "n" || input == "N" {
				values[cf.ID] = issues.CustomField{ID: cf.ID, Value: "0"}
			} else if hasCurrent {
				values[cf.ID] = current
			}
		case "date":
			fmt.Printf("    输入 (YYYY-MM-DD) [%s]: ", defaultVal)
			fmt.Scanln(&input)
			if input == "" && defaultVal != "" {
				values[cf.ID] = issues.CustomField{ID: cf.ID, Value: defaultVal}
			} else if input != "" {
				values[cf.ID] = issues.CustomField{ID: cf.ID, Value: input}
			}
		default: // string, text
			fmt.Printf("    输入 [%s]: ", defaultVal)
			fmt.Scanln(&input)
			if input != "" {
				values[cf.ID] = issues.CustomField{ID: cf.ID, Value: input}
			} else if hasCurrent {
				values[cf.ID] = current
			}
		}
		fmt.Println()
	}

	result := make([]issues.CustomField, 0, len(values))
	for _, v := range values {
		result = append(result, v)
	}
	return result, nil
}
```

Add to imports: `"fmt"`, `"os"`, `"strconv"`, `"github.com/mattn/go-isatty"`, `"github.com/largeoliu/redmine-cli/internal/resources/trackers"` (aliased as `trackers`), `"github.com/largeoliu/redmine-cli/internal/resources/issues"` (aliased as `issues`).

- [ ] **Step 2: Check if IsInteractive helper exists**

Run: `grep -r "IsInteractive" internal/` to see if this helper exists. If not, we'll need to use `isatty` from `golang.org/x/crypto/ssh/terminal` or similar.

Actually, simpler approach - use `github.com/mattn/go-isatty`:

```go
import "github.com/mattn/go-isatty"

func isInteractive() bool {
	return isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
}
```

If this package isn't in go.mod yet, add it.

- [ ] **Step 3: Wire interactive flow into newCreateCommand**

Modify `newCreateCommand` RunE:

After resolving client, BEFORE calling Create:

```go
// Fetch tracker definition if tracker-id is provided
var trackerDef *trackers.Tracker
if req.TrackerID > 0 {
	trackerDef, _ = trackers.NewClient(c).Get(cmd.Context(), req.TrackerID)
}

// Merge flag-based custom fields with interactive input
flagCFs, _ := parseCustomFieldFlags(createFlags.Fields, trackerDef)

// Interactive prompt if tracker has custom fields and we're in TTY
interactiveCFs, _ := promptCustomFields(trackerDef, nil)

// Merge: flag values take precedence over interactive defaults
mergedCFs := mergeCustomFields(interactiveCFs, flagCFs)
if len(mergedCFs) > 0 {
	req.CustomFields = mergedCFs
}
```

Note: Since we may not have tracker info (tracker-id not provided), the interactive flow only works when tracker-id IS provided. This is acceptable.

Add mergeCustomFields helper:

```go
func mergeCustomFields(interactive, flags []issues.CustomField) []issues.CustomField {
	cfMap := make(map[int]issues.CustomField)
	// Interactive first (lower priority)
	for _, cf := range interactive {
		cfMap[cf.ID] = cf
	}
	// Flags override interactive
	for _, cf := range flags {
		cfMap[cf.ID] = cf
	}
	result := make([]issues.CustomField, 0, len(cfMap))
	for _, cf := range cfMap {
		result = append(result, cf)
	}
	return result
}
```

- [ ] **Step 4: Wire interactive flow into newUpdateCommand**

Similar changes to `newUpdateCommand`. For update, we may want to first GET the issue to find current custom field values, but that's a separate API call. For now, start with no initial values.

- [ ] **Step 5: Commit**

```bash
git add internal/resources/issues/commands.go
git commit -m "feat(issues): add interactive custom field prompting"
```

---

## Task 5: Full Integration Test

- [ ] **Step 1: Build and test locally**

```bash
go build -o bin/redmine ./cmd/main.go
./bin/redmine tracker list
./bin/redmine tracker get 1
./bin/redmine issue create --help  # Should show --custom-field flag
```

- [ ] **Step 2: Test with mock server or actual Redmine (if available)**

---

## Implementation Order Summary

1. **Task 1** - Extend Tracker with Get and TrackerCustomField (types, client, commands)
2. **Task 2** - Add CustomFields to IssueCreateRequest/IssueUpdateRequest
3. **Task 3** - Add --custom-field flag parsing
4. **Task 4** - Add interactive prompting
5. **Task 5** - Integration test

---

## Spec Coverage Check

- [x] Tracker list/get commands - Task 1
- [x] CustomFields in tracker response - Task 1
- [x] CustomFields in issue create/update API - Task 2
- [x] --custom-field flag parsing - Task 3
- [x] Interactive mode (auto-trigger in TTY) - Task 4
- [x] Non-interactive fallback (skip prompt) - Task 4
- [x] Name and id:format support - Task 3
