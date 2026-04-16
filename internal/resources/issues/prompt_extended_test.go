// internal/resources/issues/prompt_extended_test.go
// promptCustomFields 扩展测试
package issues

import (
	"errors"
	"strings"
	"testing"

	"github.com/largeoliu/redmine-cli/internal/resources/trackers"
)

// mockTerminalReader 模拟终端读取器
type mockTerminalReader struct {
	inputs []string
	index  int
	err    error
}

func (m *mockTerminalReader) ReadLine() (string, error) {
	if m.err != nil {
		return "", m.err
	}
	if m.index >= len(m.inputs) {
		return "", errors.New("no more inputs")
	}
	input := m.inputs[m.index]
	m.index++
	return input, nil
}

// TestPromptCustomFieldsWithReader_EmptyCustomFields 测试空自定义字段
func TestPromptCustomFieldsWithReader_EmptyCustomFields(t *testing.T) {
	tracker := &trackers.Tracker{
		ID:           1,
		Name:         "Bug",
		CustomFields: []trackers.TrackerCustomField{},
	}

	reader := &mockTerminalReader{}
	result, err := promptCustomFieldsWithReader(tracker, nil, reader, true)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result for empty custom fields, got %v", result)
	}
}

// TestPromptCustomFieldsWithReader_NonTerminal 测试非终端环境
func TestPromptCustomFieldsWithReader_NonTerminal(t *testing.T) {
	tracker := &trackers.Tracker{
		ID:   1,
		Name: "Bug",
		CustomFields: []trackers.TrackerCustomField{
			{ID: 1, Name: "Severity", FieldFormat: "string"},
		},
	}

	reader := &mockTerminalReader{}
	result, err := promptCustomFieldsWithReader(tracker, nil, reader, false)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result for non-terminal, got %v", result)
	}
}

// TestPromptCustomFieldsWithReader_StringType 测试 string 类型字段
func TestPromptCustomFieldsWithReader_StringType(t *testing.T) {
	tracker := &trackers.Tracker{
		ID:   1,
		Name: "Bug",
		CustomFields: []trackers.TrackerCustomField{
			{ID: 1, Name: "Description", FieldFormat: "string"},
		},
	}

	reader := &mockTerminalReader{inputs: []string{"test value"}}
	result, err := promptCustomFieldsWithReader(tracker, nil, reader, true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0].ID != 1 || result[0].Value != "test value" {
		t.Errorf("expected ID=1, Value='test value', got ID=%d, Value='%v'", result[0].ID, result[0].Value)
	}
}

// TestPromptCustomFieldsWithReader_StringTypeWithInitialValue 测试 string 类型带初始值
func TestPromptCustomFieldsWithReader_StringTypeWithInitialValue(t *testing.T) {
	tracker := &trackers.Tracker{
		ID:   1,
		Name: "Bug",
		CustomFields: []trackers.TrackerCustomField{
			{ID: 1, Name: "Description", FieldFormat: "string"},
		},
	}

	initialValues := map[int]CustomField{
		1: {ID: 1, Value: "initial value"},
	}

	// 空输入应该保留初始值
	reader := &mockTerminalReader{inputs: []string{""}}
	result, err := promptCustomFieldsWithReader(tracker, initialValues, reader, true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0].Value != "initial value" {
		t.Errorf("expected initial value preserved, got '%v'", result[0].Value)
	}
}

// TestPromptCustomFieldsWithReader_BoolTypeYes 测试 bool 类型输入 y
func TestPromptCustomFieldsWithReader_BoolTypeYes(t *testing.T) {
	tracker := &trackers.Tracker{
		ID:   1,
		Name: "Bug",
		CustomFields: []trackers.TrackerCustomField{
			{ID: 1, Name: "Confirmed", FieldFormat: "bool"},
		},
	}

	reader := &mockTerminalReader{inputs: []string{"y"}}
	result, err := promptCustomFieldsWithReader(tracker, nil, reader, true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0].Value != "1" {
		t.Errorf("expected Value='1' for 'y', got '%v'", result[0].Value)
	}
}

// TestPromptCustomFieldsWithReader_BoolTypeNo 测试 bool 类型输入 n
func TestPromptCustomFieldsWithReader_BoolTypeNo(t *testing.T) {
	tracker := &trackers.Tracker{
		ID:   1,
		Name: "Bug",
		CustomFields: []trackers.TrackerCustomField{
			{ID: 1, Name: "Confirmed", FieldFormat: "bool"},
		},
	}

	reader := &mockTerminalReader{inputs: []string{"n"}}
	result, err := promptCustomFieldsWithReader(tracker, nil, reader, true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0].Value != "0" {
		t.Errorf("expected Value='0' for 'n', got '%v'", result[0].Value)
	}
}

// TestPromptCustomFieldsWithReader_BoolTypeEmptyWithInitial 测试 bool 类型空输入保留初始值
func TestPromptCustomFieldsWithReader_BoolTypeEmptyWithInitial(t *testing.T) {
	tracker := &trackers.Tracker{
		ID:   1,
		Name: "Bug",
		CustomFields: []trackers.TrackerCustomField{
			{ID: 1, Name: "Confirmed", FieldFormat: "bool"},
		},
	}

	initialValues := map[int]CustomField{
		1: {ID: 1, Value: "1"},
	}

	reader := &mockTerminalReader{inputs: []string{""}}
	result, err := promptCustomFieldsWithReader(tracker, initialValues, reader, true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0].Value != "1" {
		t.Errorf("expected initial value preserved, got '%v'", result[0].Value)
	}
}

// TestPromptCustomFieldsWithReader_DateType 测试 date 类型
func TestPromptCustomFieldsWithReader_DateType(t *testing.T) {
	tracker := &trackers.Tracker{
		ID:   1,
		Name: "Bug",
		CustomFields: []trackers.TrackerCustomField{
			{ID: 1, Name: "Due Date", FieldFormat: "date"},
		},
	}

	reader := &mockTerminalReader{inputs: []string{"2024-12-31"}}
	result, err := promptCustomFieldsWithReader(tracker, nil, reader, true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0].Value != "2024-12-31" {
		t.Errorf("expected Value='2024-12-31', got '%v'", result[0].Value)
	}
}

// TestPromptCustomFieldsWithReader_DateTypeEmptyWithDefault 测试 date 类型空输入使用默认值
func TestPromptCustomFieldsWithReader_DateTypeEmptyWithDefault(t *testing.T) {
	tracker := &trackers.Tracker{
		ID:   1,
		Name: "Bug",
		CustomFields: []trackers.TrackerCustomField{
			{ID: 1, Name: "Due Date", FieldFormat: "date"},
		},
	}

	initialValues := map[int]CustomField{
		1: {ID: 1, Value: "2024-01-01"},
	}

	reader := &mockTerminalReader{inputs: []string{""}}
	result, err := promptCustomFieldsWithReader(tracker, initialValues, reader, true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0].Value != "2024-01-01" {
		t.Errorf("expected default value used, got '%v'", result[0].Value)
	}
}

// TestPromptCustomFieldsWithReader_ListType 测试 list 类型
func TestPromptCustomFieldsWithReader_ListType(t *testing.T) {
	tracker := &trackers.Tracker{
		ID:   1,
		Name: "Bug",
		CustomFields: []trackers.TrackerCustomField{
			{
				ID:          1,
				Name:        "Severity",
				FieldFormat: "list",
				PossibleValues: []trackers.ValueLabel{
					{Value: "1", Label: "Low"},
					{Value: "2", Label: "Medium"},
					{Value: "3", Label: "High"},
				},
			},
		},
	}

	reader := &mockTerminalReader{inputs: []string{"2"}}
	result, err := promptCustomFieldsWithReader(tracker, nil, reader, true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0].Value != "2" {
		t.Errorf("expected Value='2', got '%v'", result[0].Value)
	}
}

// TestPromptCustomFieldsWithReader_ListTypeInvalidIndex 测试 list 类型无效索引
func TestPromptCustomFieldsWithReader_ListTypeInvalidIndex(t *testing.T) {
	tracker := &trackers.Tracker{
		ID:   1,
		Name: "Bug",
		CustomFields: []trackers.TrackerCustomField{
			{
				ID:          1,
				Name:        "Severity",
				FieldFormat: "list",
				PossibleValues: []trackers.ValueLabel{
					{Value: "1", Label: "Low"},
					{Value: "2", Label: "Medium"},
					{Value: "3", Label: "High"},
				},
			},
		},
	}

	initialValues := map[int]CustomField{
		1: {ID: 1, Value: "1"},
	}

	// 无效索引应该保留初始值
	reader := &mockTerminalReader{inputs: []string{"99"}}
	result, err := promptCustomFieldsWithReader(tracker, initialValues, reader, true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0].Value != "1" {
		t.Errorf("expected initial value preserved for invalid index, got '%v'", result[0].Value)
	}
}

// TestPromptCustomFieldsWithReader_ListTypeEmptyPossibleValues 测试 list 类型空选项
func TestPromptCustomFieldsWithReader_ListTypeEmptyPossibleValues(t *testing.T) {
	tracker := &trackers.Tracker{
		ID:   1,
		Name: "Bug",
		CustomFields: []trackers.TrackerCustomField{
			{
				ID:             1,
				Name:           "Severity",
				FieldFormat:    "list",
				PossibleValues: []trackers.ValueLabel{},
			},
		},
	}

	reader := &mockTerminalReader{}
	result, err := promptCustomFieldsWithReader(tracker, nil, reader, true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 空选项列表不应该添加任何值
	if len(result) != 0 {
		t.Errorf("expected 0 results for empty possible values, got %d", len(result))
	}
}

// TestPromptCustomFieldsWithReader_MultipleFields 测试多个字段
func TestPromptCustomFieldsWithReader_MultipleFields(t *testing.T) {
	tracker := &trackers.Tracker{
		ID:   1,
		Name: "Bug",
		CustomFields: []trackers.TrackerCustomField{
			{ID: 1, Name: "Description", FieldFormat: "string"},
			{ID: 2, Name: "Confirmed", FieldFormat: "bool"},
			{ID: 3, Name: "Due Date", FieldFormat: "date"},
		},
	}

	reader := &mockTerminalReader{inputs: []string{"test desc", "y", "2024-12-31"}}
	result, err := promptCustomFieldsWithReader(tracker, nil, reader, true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 results, got %d", len(result))
	}

	values := make(map[int]string)
	for _, cf := range result {
		values[cf.ID] = cf.Value.(string)
	}

	if values[1] != "test desc" {
		t.Errorf("expected field 1 = 'test desc', got '%v'", values[1])
	}
	if values[2] != "1" {
		t.Errorf("expected field 2 = '1', got '%v'", values[2])
	}
	if values[3] != "2024-12-31" {
		t.Errorf("expected field 3 = '2024-12-31', got '%v'", values[3])
	}
}

// TestPromptCustomFieldsWithReader_ReaderError 测试读取错误
func TestPromptCustomFieldsWithReader_ReaderError(t *testing.T) {
	tracker := &trackers.Tracker{
		ID:   1,
		Name: "Bug",
		CustomFields: []trackers.TrackerCustomField{
			{ID: 1, Name: "Description", FieldFormat: "string"},
		},
	}

	initialValues := map[int]CustomField{
		1: {ID: 1, Value: "default"},
	}

	reader := &mockTerminalReader{err: errors.New("read error")}
	result, err := promptCustomFieldsWithReader(tracker, initialValues, reader, true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	// 读取错误应该保留初始值
	if result[0].Value != "default" {
		t.Errorf("expected initial value preserved on read error, got '%v'", result[0].Value)
	}
}

// TestPromptCustomFieldsWithReader_LinkType 测试 link 类型（使用 default 分支）
func TestPromptCustomFieldsWithReader_LinkType(t *testing.T) {
	tracker := &trackers.Tracker{
		ID:   1,
		Name: "Bug",
		CustomFields: []trackers.TrackerCustomField{
			{ID: 1, Name: "URL", FieldFormat: "link"},
		},
	}

	reader := &mockTerminalReader{inputs: []string{"https://example.com"}}
	result, err := promptCustomFieldsWithReader(tracker, nil, reader, true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0].Value != "https://example.com" {
		t.Errorf("expected Value='https://example.com', got '%v'", result[0].Value)
	}
}

// TestPromptCustomFieldsWithReader_FloatType 测试 float 类型（使用 default 分支）
func TestPromptCustomFieldsWithReader_FloatType(t *testing.T) {
	tracker := &trackers.Tracker{
		ID:   1,
		Name: "Bug",
		CustomFields: []trackers.TrackerCustomField{
			{ID: 1, Name: "Estimate", FieldFormat: "float"},
		},
	}

	reader := &mockTerminalReader{inputs: []string{"3.5"}}
	result, err := promptCustomFieldsWithReader(tracker, nil, reader, true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0].Value != "3.5" {
		t.Errorf("expected Value='3.5', got '%v'", result[0].Value)
	}
}

// TestPromptCustomFieldsWithReader_IntType 测试 int 类型（使用 default 分支）
func TestPromptCustomFieldsWithReader_IntType(t *testing.T) {
	tracker := &trackers.Tracker{
		ID:   1,
		Name: "Bug",
		CustomFields: []trackers.TrackerCustomField{
			{ID: 1, Name: "Count", FieldFormat: "int"},
		},
	}

	reader := &mockTerminalReader{inputs: []string{"42"}}
	result, err := promptCustomFieldsWithReader(tracker, nil, reader, true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0].Value != "42" {
		t.Errorf("expected Value='42', got '%v'", result[0].Value)
	}
}

// TestPromptCustomFieldsWithReader_TextType 测试 text 类型（使用 default 分支）
func TestPromptCustomFieldsWithReader_TextType(t *testing.T) {
	tracker := &trackers.Tracker{
		ID:   1,
		Name: "Bug",
		CustomFields: []trackers.TrackerCustomField{
			{ID: 1, Name: "Notes", FieldFormat: "text"},
		},
	}

	reader := &mockTerminalReader{inputs: []string{"multi\nline\ntext"}}
	result, err := promptCustomFieldsWithReader(tracker, nil, reader, true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if !strings.Contains(result[0].Value.(string), "multi") {
		t.Errorf("expected Value to contain 'multi', got '%v'", result[0].Value)
	}
}

// TestPromptCustomFieldsWithReader_UserType 测试 user 类型（使用 default 分支）
func TestPromptCustomFieldsWithReader_UserType(t *testing.T) {
	tracker := &trackers.Tracker{
		ID:   1,
		Name: "Bug",
		CustomFields: []trackers.TrackerCustomField{
			{ID: 1, Name: "Assignee", FieldFormat: "user"},
		},
	}

	reader := &mockTerminalReader{inputs: []string{"123"}}
	result, err := promptCustomFieldsWithReader(tracker, nil, reader, true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0].Value != "123" {
		t.Errorf("expected Value='123', got '%v'", result[0].Value)
	}
}

// TestPromptCustomFieldsWithReader_VersionType 测试 version 类型（使用 default 分支）
func TestPromptCustomFieldsWithReader_VersionType(t *testing.T) {
	tracker := &trackers.Tracker{
		ID:   1,
		Name: "Bug",
		CustomFields: []trackers.TrackerCustomField{
			{ID: 1, Name: "Target Version", FieldFormat: "version"},
		},
	}

	reader := &mockTerminalReader{inputs: []string{"v1.0"}}
	result, err := promptCustomFieldsWithReader(tracker, nil, reader, true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0].Value != "v1.0" {
		t.Errorf("expected Value='v1.0', got '%v'", result[0].Value)
	}
}

// TestPromptCustomFieldsWithReader_EnumerationType 测试 enumeration 类型（使用 default 分支）
func TestPromptCustomFieldsWithReader_EnumerationType(t *testing.T) {
	tracker := &trackers.Tracker{
		ID:   1,
		Name: "Bug",
		CustomFields: []trackers.TrackerCustomField{
			{ID: 1, Name: "Category", FieldFormat: "enumeration"},
		},
	}

	reader := &mockTerminalReader{inputs: []string{"category1"}}
	result, err := promptCustomFieldsWithReader(tracker, nil, reader, true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0].Value != "category1" {
		t.Errorf("expected Value='category1', got '%v'", result[0].Value)
	}
}
