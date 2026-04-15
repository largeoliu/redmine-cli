# Custom Fields Support Design

**Date:** 2026-04-15
**Status:** Approved

## Overview

Add custom field support to the Redmine CLI, enabling users to:
1. Query tracker definitions to discover available custom fields
2. Fill in custom fields interactively when creating/updating issues
3. Pass custom fields via flags for scripted/automated workflows

## Problem Statement

Redmine supports custom fields that vary by instance, tracker, and issue type. Users need a way to:
- Discover what custom fields are available for a given tracker
- Fill in custom field values when creating or updating issues
- Support both interactive (manual) and flag-based (scripted) workflows

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                     CLI 入口                            │
├─────────────────────────────────────────────────────────┤
│  issue create/update    │    tracker get/list           │
│  (交互式 + flag)         │    (查询命令)                 │
├─────────────────────────────────────────────────────────┤
│           Custom Field 处理器                           │
│  ┌─────────────────┐  ┌─────────────────┐            │
│  │  交互式选择器    │  │  Flag 解析器    │            │
│  │  (自动触发)      │  │  --custom-field │            │
│  └─────────────────┘  └─────────────────┘            │
├─────────────────────────────────────────────────────────┤
│           Redmine API 客户端                            │
│  • 查询 tracker 定义 (含 custom_fields)                │
│  • 创建/更新 issue (含 custom_fields)                  │
└─────────────────────────────────────────────────────────┘
```

## Part 1: Tracker Subcommand

### New Commands

```
redmine tracker list                                  # 列出所有 tracker
redmine tracker get <tracker-id>                       # 获取详情，含 custom_fields
```

### Tracker Struct

```go
// internal/resources/trackers/types.go
type Tracker struct {
    ID           int           `json:"id"`
    Name         string        `json:"name"`
    Description  string        `json:"description,omitempty"`
    CustomFields []CustomField `json:"custom_fields,omitempty"`
}
```

### CustomField Struct (already exists)

```go
// internal/resources/issues/types.go
type CustomField struct {
    ID         int           `json:"id"`
    Name       string        `json:"name"`
    Value      any           `json:"value,omitempty"`
    PossibleValues []ValueLabel `json:"possible_values,omitempty"` // For list-type fields
}

type ValueLabel struct {
    Value string `json:"value"`
    Label string `json:"label"`
}
```

### Output Example

```
$ redmine tracker get 1
ID:        1
Name:      Bug
Custom Fields:
  ─────────────────────────────────────────────
  ID   Name          Type      Possible Values
  ─────────────────────────────────────────────
  5    优先级        List      [高, 中, 低]
  6    修复版本      List      [v1.0, v1.1, v2.0]
  7    详细说明      Text      (free text)
  8    是否紧急      Boolean   [true, false]
  9    预计日期      Date      (date)
  ─────────────────────────────────────────────
```

## Part 2: Interactive Custom Field Selection

### Trigger

When creating/updating an issue and the tracker has custom fields defined:
- **Interactive shell (TTY)**: Automatically enter interactive selection mode
- **Non-interactive (piped/redirected/CI)**: Skip interactive mode, use only `--custom-field` flags; if required fields are missing, show error

**Note**: If `--custom-field` flags are provided, they are pre-populated in the interactive prompt (user can modify them).

### Flow

```
1. User executes: redmine issue create --project-id 123 --subject "新问题"
2. CLI fetches tracker definition (includes custom_fields)
3. If custom_fields exist, display interactive prompt:
```

```
📋 检测到 3 个自定义字段，请填写：

[5] 优先级
    类型: 列表
    选项: [1. 高] [2. 中] [3. 低]
    选择: 1

[6] 修复版本
    类型: 列表
    选项: [1. v1.0] [2. v1.1] [3. v2.0]
    选择: 2

[7] 详细说明
    类型: 文本
    输入: ________________

按 Enter 确认提交，或 Ctrl+C 取消
```

### Field Type Handling

| Type | Interactive Input | Flag Value Format |
|------|-------------------|-------------------|
| String | Free text input | `name:value` |
| List | Numbered selection | `name:v1,v2` |
| Bool | Y/N prompt | `name:true/false` |
| Date | Date input (YYYY-MM-DD) | `name:2024-01-15` |
| Text | Multi-line input | `name:value` |

### Exit

- All fields filled + Enter = confirm and submit
- Ctrl+C = cancel operation, do not create/update issue

## Part 3: Flag Support (Scripted/Automation)

### Command Format

```bash
# Single field
redmine issue create --project-id 123 --subject "..." \
  --custom-field "优先级:高" \
  --custom-field "是否紧急:true"

# List field with multiple values (comma-separated)
redmine issue create --project-id 123 --subject "..." \
  --custom-field "修复版本:v1.0,v1.1"

# Mix of multiple fields
```

### Parsing Rules

- `name:value` - Match by field name
- For list fields, comma-separated values indicate multi-select
- Exact match by ID: `id:5:value` format (useful when field names are ambiguous)

### Implementation

```go
// Parse --custom-field flags into CustomField slice
func parseCustomFieldFlags(flags []string) ([]CustomField, error) {
    fields := make([]CustomField, 0, len(flags))
    for _, f := range flags {
        // Format: "name:value" or "id:5:value"
        parts := strings.SplitN(f, ":", 3)
        if len(parts) < 2 {
            return nil, fmt.Errorf("invalid custom field format: %s", f)
        }
        cf := CustomField{}
        if parts[0] == "id" {
            cf.ID, _ = strconv.Atoi(parts[1])
            cf.Value = parts[2]
        } else {
            cf.Name = parts[0]
            cf.Value = parts[1]
        }
        fields = append(fields, cf)
    }
    return fields, nil
}
```

## Part 4: API Layer Changes

### IssueCreateRequest / IssueUpdateRequest

Add `CustomFields` field:

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

### API Payload Format

```json
{
  "issue": {
    "project_id": 123,
    "subject": "...",
    "custom_fields": [
      {"id": 5, "value": "高"},
      {"id": 6, "value": ["v1.0", "v1.1"]}
    ]
  }
}
```

## Part 5: File Structure

```
internal/
├── resources/
│   ├── trackers/              # NEW - tracker resource
│   │   ├── commands.go        # tracker list, tracker get commands
│   │   ├── client.go          # API client methods
│   │   └── types.go           # Tracker, CustomField types
│   └── issues/
│       ├── types.go           # Add CustomFields to Create/Update requests
│       ├── commands.go        # Add --custom-field flag, interactive mode
│       └── client.go          # Ensure custom_fields sent in requests
```

## Implementation Order

1. **Tracker resource** (new)
   - `internal/resources/trackers/types.go`
   - `internal/resources/trackers/client.go`
   - `internal/resources/trackers/commands.go`
   - Add `tracker` command to root

2. **API layer** (issues)
   - Add `CustomFields` to `IssueCreateRequest` and `IssueUpdateRequest`
   - Verify client sends custom_fields in JSON payload

3. **Flag support** (issues create/update)
   - Add `--custom-field` flag to issue create/update commands
   - Implement flag parsing function

4. **Interactive mode** (issues create/update)
   - Fetch tracker definition when creating/updating
   - If custom_fields exist, prompt interactively
   - Merge interactive values with flag values

5. **Testing**
   - Unit tests for custom field parsing
   - Integration tests for API payload format
   - Manual testing of interactive flow
