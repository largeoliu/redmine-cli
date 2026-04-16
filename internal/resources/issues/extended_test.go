// internal/resources/issues/extended_test.go
// 扩展测试用例 - 覆盖边界条件和异常场景
package issues

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/largeoliu/redmine-cli/internal/client"
	"github.com/largeoliu/redmine-cli/internal/resources/trackers"
	"github.com/largeoliu/redmine-cli/internal/testutil"
	"github.com/largeoliu/redmine-cli/internal/types"
)

// TestCreateCommand_WithTrackerAndCustomFields 测试创建命令带 tracker 和自定义字段
func TestCreateCommand_WithTrackerAndCustomFields(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	// 模拟 trackers 接口
	trackerResponse := trackers.Tracker{
		ID:   1,
		Name: "Bug",
		CustomFields: []trackers.TrackerCustomField{
			{ID: 1, Name: "Severity", FieldFormat: "string"},
		},
	}
	mock.HandleJSON("/trackers/1.json", map[string]any{"tracker": trackerResponse})

	// 模拟 issues 创建接口
	issueResponse := struct {
		Issue Issue `json:"issue"`
	}{
		Issue: Issue{
			ID:      1,
			Subject: "Test Issue with Custom Fields",
			Project: &Reference{ID: 1, Name: "Project A"},
		},
	}
	mock.HandleJSON("/issues.json", issueResponse)

	flags := &types.GlobalFlags{}
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) {
			return client.NewClient(mock.URL, "test-key"), nil
		},
		writeOutputFunc: func(_ io.Writer, _ *types.GlobalFlags, _ any) error {
			return nil
		},
	}

	var buf bytes.Buffer
	cmd := newCreateCommand(flags, resolver)
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{
		"--project-id", "1",
		"--subject", "Test Issue with Custom Fields",
		"--tracker-id", "1",
		"--custom-field", "id:1:High",
	})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestCreateCommand_WithCustomFieldsError 测试创建命令带错误的自定义字段格式
func TestCreateCommand_WithCustomFieldsError(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	// 模拟 trackers 接口
	trackerResponse := trackers.Tracker{
		ID:   1,
		Name: "Bug",
		CustomFields: []trackers.TrackerCustomField{
			{ID: 1, Name: "Severity", FieldFormat: "string"},
		},
	}
	mock.HandleJSON("/trackers/1.json", map[string]any{"tracker": trackerResponse})

	flags := &types.GlobalFlags{}
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) {
			return client.NewClient(mock.URL, "test-key"), nil
		},
	}

	cmd := newCreateCommand(flags, resolver)
	cmd.SetArgs([]string{
		"--project-id", "1",
		"--subject", "Test Issue",
		"--tracker-id", "1",
		"--custom-field", "invalid-format", // 错误的格式
	})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for invalid custom field format, got nil")
	}
}

// TestCreateCommand_TrackerGetError 测试创建命令时获取 tracker 失败
func TestCreateCommand_TrackerGetError(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	// 模拟 trackers 接口返回错误
	mock.HandleError("/trackers/1.json", http.StatusNotFound, "Tracker not found")

	// 模拟 issues 创建接口
	issueResponse := struct {
		Issue Issue `json:"issue"`
	}{
		Issue: Issue{
			ID:      1,
			Subject: "Test Issue",
			Project: &Reference{ID: 1, Name: "Project A"},
		},
	}
	mock.HandleJSON("/issues.json", issueResponse)

	flags := &types.GlobalFlags{}
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) {
			return client.NewClient(mock.URL, "test-key"), nil
		},
		writeOutputFunc: func(_ io.Writer, _ *types.GlobalFlags, _ any) error {
			return nil
		},
	}

	var buf bytes.Buffer
	cmd := newCreateCommand(flags, resolver)
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{
		"--project-id", "1",
		"--subject", "Test Issue",
		"--tracker-id", "1",
	})

	// 即使 tracker 获取失败，也应该能继续创建 issue
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestCreateCommand_WriteOutputError 测试创建命令写输出错误
func TestCreateCommand_WriteOutputError(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	issueResponse := struct {
		Issue Issue `json:"issue"`
	}{
		Issue: Issue{
			ID:      1,
			Subject: "Test Issue",
			Project: &Reference{ID: 1, Name: "Project A"},
		},
	}
	mock.HandleJSON("/issues.json", issueResponse)

	flags := &types.GlobalFlags{}
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) {
			return client.NewClient(mock.URL, "test-key"), nil
		},
		writeOutputFunc: func(_ io.Writer, _ *types.GlobalFlags, _ any) error {
			return errors.New("write output error")
		},
	}

	cmd := newCreateCommand(flags, resolver)
	cmd.SetArgs([]string{
		"--project-id", "1",
		"--subject", "Test Issue",
	})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error from WriteOutput, got nil")
	}
}

// TestUpdateCommand_WithCustomFields 测试更新命令带自定义字段
func TestUpdateCommand_WithCustomFields(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.Handle("/issues/1.json", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT method, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	flags := &types.GlobalFlags{}
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) {
			return client.NewClient(mock.URL, "test-key"), nil
		},
	}

	var buf bytes.Buffer
	cmd := newUpdateCommand(flags, resolver)
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{
		"1",
		"--subject", "Updated Issue",
		"--custom-field", "id:1:High",
	})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestUpdateCommand_WithCustomFieldsError 测试更新命令带错误的自定义字段格式
func TestUpdateCommand_WithCustomFieldsError(t *testing.T) {
	flags := &types.GlobalFlags{}
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) {
			return client.NewClient("http://example.com", "test-key"), nil
		},
	}

	cmd := newUpdateCommand(flags, resolver)
	cmd.SetArgs([]string{
		"1",
		"--subject", "Updated Issue",
		"--custom-field", "invalid", // 错误的格式
	})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for invalid custom field format, got nil")
	}
}

// TestUpdateCommand_NoCustomFields 测试更新命令不带自定义字段
func TestUpdateCommand_NoCustomFields(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.Handle("/issues/1.json", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT method, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	flags := &types.GlobalFlags{}
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) {
			return client.NewClient(mock.URL, "test-key"), nil
		},
	}

	cmd := newUpdateCommand(flags, resolver)
	cmd.SetArgs([]string{
		"1",
		"--subject", "Updated Issue",
	})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestParseCustomFieldFlags_EdgeCases 测试自定义字段解析的边界情况
func TestParseCustomFieldFlags_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		flags   []string
		tracker *trackers.Tracker
		wantErr bool
	}{
		{
			name:    "空字符串",
			flags:   []string{""},
			tracker: nil,
			wantErr: true,
		},
		{
			name:    "只有冒号",
			flags:   []string{":"},
			tracker: nil,
			wantErr: true,
		},
		{
			name:    "id格式缺少值部分",
			flags:   []string{"id:5"},
			tracker: nil,
			wantErr: true,
		},
		{
			name:    "id格式负数id",
			flags:   []string{"id:-1:value"},
			tracker: nil,
			wantErr: false, // 负数 id 也是有效的整数格式
		},
		{
			name:    "id格式零值id",
			flags:   []string{"id:0:value"},
			tracker: nil,
			wantErr: false, // 0 是有效的整数
		},
		{
			name:    "name格式空tracker",
			flags:   []string{"field:value"},
			tracker: nil,
			wantErr: true,
		},
		{
			name:  "name格式空custom fields",
			flags: []string{"field:value"},
			tracker: &trackers.Tracker{
				ID:           1,
				Name:         "Bug",
				CustomFields: []trackers.TrackerCustomField{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseCustomFieldFlags(tt.flags, tt.tracker)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseCustomFieldFlags() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestListCommand_WithAllFlags 测试 list 命令带所有 flags
func TestListCommand_WithAllFlags(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	response := IssueList{
		Issues: []Issue{
			{ID: 1, Subject: "Issue 1", Project: &Reference{ID: 1, Name: "Project A"}},
		},
		TotalCount: 1,
	}
	mock.HandleJSON("/issues.json", response)

	flags := &types.GlobalFlags{}
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) {
			return client.NewClient(mock.URL, "test-key"), nil
		},
		writeOutputFunc: func(_ io.Writer, _ *types.GlobalFlags, _ any) error {
			return nil
		},
	}

	cmd := newListCommand(flags, resolver)
	cmd.SetArgs([]string{
		"--project-id", "1",
		"--tracker-id", "2",
		"--status-id", "3",
		"--assigned-to-id", "4",
		"--query", "5",
		"--sort", "updated_on:desc",
	})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestGetCommand_WithoutInclude 测试 get 命令不带 include 参数
func TestGetCommand_WithoutInclude(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	response := struct {
		Issue Issue `json:"issue"`
	}{
		Issue: Issue{
			ID:      1,
			Subject: "Test Issue",
			Project: &Reference{ID: 1, Name: "Project A"},
		},
	}
	mock.HandleJSON("/issues/1.json", response)

	flags := &types.GlobalFlags{}
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) {
			return client.NewClient(mock.URL, "test-key"), nil
		},
		writeOutputFunc: func(_ io.Writer, _ *types.GlobalFlags, _ any) error {
			return nil
		},
	}

	cmd := newGetCommand(flags, resolver)
	cmd.SetArgs([]string{"1"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestGetCommand_WriteOutputError 测试 get 命令写输出错误
func TestGetCommand_WriteOutputError(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	response := struct {
		Issue Issue `json:"issue"`
	}{
		Issue: Issue{
			ID:      1,
			Subject: "Test Issue",
		},
	}
	mock.HandleJSON("/issues/1.json", response)

	flags := &types.GlobalFlags{}
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) {
			return client.NewClient(mock.URL, "test-key"), nil
		},
		writeOutputFunc: func(_ io.Writer, _ *types.GlobalFlags, _ any) error {
			return errors.New("write output error")
		},
	}

	cmd := newGetCommand(flags, resolver)
	cmd.SetArgs([]string{"1"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error from WriteOutput, got nil")
	}
}

// TestListCommand_WriteOutputError 测试 list 命令写输出错误
func TestListCommand_WriteOutputError(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	response := IssueList{
		Issues:     []Issue{},
		TotalCount: 0,
	}
	mock.HandleJSON("/issues.json", response)

	flags := &types.GlobalFlags{}
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) {
			return client.NewClient(mock.URL, "test-key"), nil
		},
		writeOutputFunc: func(_ io.Writer, _ *types.GlobalFlags, _ any) error {
			return errors.New("write output error")
		},
	}

	cmd := newListCommand(flags, resolver)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error from WriteOutput, got nil")
	}
}

// TestBuildListParams_ZeroValues 测试 BuildListParams 零值处理
func TestBuildListParams_ZeroValues(t *testing.T) {
	// 测试所有零值都不应该出现在结果中
	flags := ListFlags{
		ProjectID:    0,
		TrackerID:    0,
		StatusID:     0,
		AssignedToID: 0,
		Limit:        0,
		Offset:       0,
		Query:        "",
		Sort:         "",
	}

	params := BuildListParams(flags)

	if len(params) != 0 {
		t.Errorf("expected empty params, got %v", params)
	}
}

// TestBuildListParams_NegativeValues 测试 BuildListParams 负值处理
func TestBuildListParams_NegativeValues(t *testing.T) {
	// 测试负值不会出现在结果中（因为 BuildListParams 只处理 > 0 的值）
	flags := ListFlags{
		ProjectID: -1,
		Limit:     -10,
	}

	params := BuildListParams(flags)

	// 负值应该被忽略，不会出现在结果中
	if len(params) != 0 {
		t.Errorf("expected empty params for negative values, got %v", params)
	}
}

// TestIssue_WithAllFields 测试 Issue 结构体所有字段
func TestIssue_WithAllFields(t *testing.T) {
	createdOn := sampleTime()
	updatedOn := sampleTime()
	closedOn := sampleTime()

	issue := Issue{
		ID:          1,
		Subject:     "Test Issue",
		Description: "Test Description",
		Project: &Reference{
			ID:   1,
			Name: "Test Project",
		},
		Tracker: &Reference{
			ID:   1,
			Name: "Bug",
		},
		Status: &Reference{
			ID:   1,
			Name: "New",
		},
		Priority: &Reference{
			ID:   2,
			Name: "Normal",
		},
		Author: &Reference{
			ID:   1,
			Name: "John Doe",
		},
		AssignedTo: &Reference{
			ID:   2,
			Name: "Jane Smith",
		},
		Category: &Reference{
			ID:   1,
			Name: "Development",
		},
		FixedVersion: &Reference{
			ID:   1,
			Name: "v1.0",
		},
		Parent: &Reference{
			ID:   10,
			Name: "Parent Issue",
		},
		StartDate:    "2024-01-01",
		DueDate:      "2024-12-31",
		DoneRatio:    50,
		CreatedOn:    &createdOn,
		UpdatedOn:    &updatedOn,
		ClosedOn:     &closedOn,
		Notes:        "Some notes",
		PrivateNotes: true,
		Watchers: []Reference{
			{ID: 1, Name: "Watcher 1"},
			{ID: 2, Name: "Watcher 2"},
		},
		CustomFields: []CustomField{
			{ID: 1, Name: "Severity", Value: "High"},
			{ID: 2, Name: "Version", Value: "v1.0"},
		},
	}

	// 验证序列化和反序列化
	data, err := json.Marshal(issue)
	if err != nil {
		t.Fatalf("failed to marshal issue: %v", err)
	}

	var unmarshaled Issue
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal issue: %v", err)
	}

	if unmarshaled.ID != issue.ID {
		t.Errorf("expected ID %d, got %d", issue.ID, unmarshaled.ID)
	}
	if unmarshaled.Subject != issue.Subject {
		t.Errorf("expected Subject %s, got %s", issue.Subject, unmarshaled.Subject)
	}
	if unmarshaled.Parent == nil || unmarshaled.Parent.ID != issue.Parent.ID {
		t.Errorf("expected Parent.ID %d, got %v", issue.Parent.ID, unmarshaled.Parent)
	}
	if unmarshaled.ClosedOn == nil {
		t.Error("expected ClosedOn to be set")
	}
	if len(unmarshaled.Watchers) != len(issue.Watchers) {
		t.Errorf("expected %d watchers, got %d", len(issue.Watchers), len(unmarshaled.Watchers))
	}
	if len(unmarshaled.CustomFields) != len(issue.CustomFields) {
		t.Errorf("expected %d custom fields, got %d", len(issue.CustomFields), len(unmarshaled.CustomFields))
	}
}

// TestIssueCreateRequest_WithAllFields 测试 IssueCreateRequest 所有字段
func TestIssueCreateRequest_WithAllFields(t *testing.T) {
	req := IssueCreateRequest{
		ProjectID:      1,
		Subject:        "Test Subject",
		Description:    "Test Description",
		TrackerID:      2,
		StatusID:       3,
		PriorityID:     4,
		AssignedToID:   5,
		CategoryID:     6,
		FixedVersionID: 7,
		ParentIssueID:  8,
		StartDate:      "2024-01-01",
		DueDate:        "2024-12-31",
		DoneRatio:      50,
		WatcherUserIDs: []int{1, 2, 3},
		CustomFields: []CustomField{
			{ID: 1, Name: "Field1", Value: "Value1"},
		},
	}

	// 包装为 issue 字段
	wrapped := map[string]any{"issue": req}

	data, err := json.Marshal(wrapped)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	var unmarshaled map[string]any
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal request: %v", err)
	}

	issue, ok := unmarshaled["issue"].(map[string]any)
	if !ok {
		t.Fatal("expected issue field in unmarshaled data")
	}

	if issue["project_id"] != float64(1) {
		t.Errorf("expected project_id 1, got %v", issue["project_id"])
	}
	if issue["parent_issue_id"] != float64(8) {
		t.Errorf("expected parent_issue_id 8, got %v", issue["parent_issue_id"])
	}
}

// TestIssueUpdateRequest_WithAllFields 测试 IssueUpdateRequest 所有字段
func TestIssueUpdateRequest_WithAllFields(t *testing.T) {
	req := IssueUpdateRequest{
		Subject:        "Updated Subject",
		Description:    "Updated Description",
		StatusID:       3,
		PriorityID:     4,
		AssignedToID:   5,
		CategoryID:     6,
		FixedVersionID: 7,
		StartDate:      "2024-01-01",
		DueDate:        "2024-12-31",
		DoneRatio:      75,
		Notes:          "Update notes",
		PrivateNotes:   true,
		CustomFields: []CustomField{
			{ID: 1, Name: "Field1", Value: "Value1"},
		},
	}

	// 包装为 issue 字段
	wrapped := map[string]any{"issue": req}

	data, err := json.Marshal(wrapped)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	var unmarshaled map[string]any
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal request: %v", err)
	}

	issue, ok := unmarshaled["issue"].(map[string]any)
	if !ok {
		t.Fatal("expected issue field in unmarshaled data")
	}

	if issue["notes"] != "Update notes" {
		t.Errorf("expected notes 'Update notes', got %v", issue["notes"])
	}
	if issue["private_notes"] != true {
		t.Errorf("expected private_notes true, got %v", issue["private_notes"])
	}
}

// TestReference_JSONSerialization 测试 Reference JSON 序列化
func TestReference_JSONSerialization(t *testing.T) {
	ref := Reference{
		ID:   1,
		Name: "Test Reference",
	}

	data, err := json.Marshal(ref)
	if err != nil {
		t.Fatalf("failed to marshal reference: %v", err)
	}

	var unmarshaled Reference
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal reference: %v", err)
	}

	if unmarshaled.ID != ref.ID {
		t.Errorf("expected ID %d, got %d", ref.ID, unmarshaled.ID)
	}
	if unmarshaled.Name != ref.Name {
		t.Errorf("expected Name %s, got %s", ref.Name, unmarshaled.Name)
	}
}

// TestCustomField_JSONSerialization 测试 CustomField JSON 序列化
func TestCustomField_JSONSerialization(t *testing.T) {
	tests := []struct {
		name  string
		field CustomField
	}{
		{
			name:  "字符串值",
			field: CustomField{ID: 1, Name: "String", Value: "test"},
		},
		{
			name:  "整数值",
			field: CustomField{ID: 2, Name: "Int", Value: 123},
		},
		{
			name:  "布尔值",
			field: CustomField{ID: 3, Name: "Bool", Value: true},
		},
		{
			name:  "数组值",
			field: CustomField{ID: 4, Name: "Array", Value: []string{"a", "b"}},
		},
		{
			name:  "nil值",
			field: CustomField{ID: 5, Name: "Nil", Value: nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.field)
			if err != nil {
				t.Fatalf("failed to marshal custom field: %v", err)
			}

			var unmarshaled CustomField
			if err := json.Unmarshal(data, &unmarshaled); err != nil {
				t.Fatalf("failed to unmarshal custom field: %v", err)
			}

			if unmarshaled.ID != tt.field.ID {
				t.Errorf("expected ID %d, got %d", tt.field.ID, unmarshaled.ID)
			}
			if unmarshaled.Name != tt.field.Name {
				t.Errorf("expected Name %s, got %s", tt.field.Name, unmarshaled.Name)
			}
		})
	}
}

// TestClient_List_InvalidJSON 测试 List 方法返回无效 JSON
func TestClient_List_InvalidJSON(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.Handle("/issues.json", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	})

	baseClient := client.NewClient(mock.URL, "test-key")
	issueClient := NewClient(baseClient)

	_, err := issueClient.List(context.Background(), nil)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

// TestClient_Get_InvalidJSON 测试 Get 方法返回无效 JSON
func TestClient_Get_InvalidJSON(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.Handle("/issues/1.json", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	})

	baseClient := client.NewClient(mock.URL, "test-key")
	issueClient := NewClient(baseClient)

	_, err := issueClient.Get(context.Background(), 1, nil)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

// TestClient_Create_InvalidJSON 测试 Create 方法返回无效 JSON
func TestClient_Create_InvalidJSON(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.Handle("/issues.json", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	})

	baseClient := client.NewClient(mock.URL, "test-key")
	issueClient := NewClient(baseClient)

	req := &IssueCreateRequest{
		ProjectID: 1,
		Subject:   "Test",
	}

	_, err := issueClient.Create(context.Background(), req)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

// TestClient_List_EmptyResponse 测试 List 方法返回空响应
func TestClient_List_EmptyResponse(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.Handle("/issues.json", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{}"))
	})

	baseClient := client.NewClient(mock.URL, "test-key")
	issueClient := NewClient(baseClient)

	result, err := issueClient.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if len(result.Issues) != 0 {
		t.Errorf("expected 0 issues, got %d", len(result.Issues))
	}
}

// TestClient_Get_MissingIssueField 测试 Get 方法返回缺少 issue 字段的响应
func TestClient_Get_MissingIssueField(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.Handle("/issues/1.json", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// 返回一个没有 issue 字段的 JSON
		w.Write([]byte(`{"message": "success"}`))
	})

	baseClient := client.NewClient(mock.URL, "test-key")
	issueClient := NewClient(baseClient)

	result, err := issueClient.Get(context.Background(), 1, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 应该返回一个零值的 Issue
	if result == nil {
		t.Fatal("expected result, got nil")
	}
}

// TestClient_Create_MissingIssueField 测试 Create 方法返回缺少 issue 字段的响应
func TestClient_Create_MissingIssueField(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.Handle("/issues.json", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// 返回一个没有 issue 字段的 JSON
		w.Write([]byte(`{"message": "created"}`))
	})

	baseClient := client.NewClient(mock.URL, "test-key")
	issueClient := NewClient(baseClient)

	req := &IssueCreateRequest{
		ProjectID: 1,
		Subject:   "Test",
	}

	result, err := issueClient.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
}

// 辅助函数 - 返回示例时间
func sampleTime() time.Time {
	return time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
}
