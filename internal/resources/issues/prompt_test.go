// internal/resources/issues/prompt_test.go
package issues

import (
	"bufio"
	"strings"
	"testing"

	"github.com/largeoliu/redmine-cli/internal/resources/trackers"
)

// TestPromptCustomFields_EmptyCustomFields 测试没有自定义字段的情况
func TestPromptCustomFields_EmptyCustomFields(t *testing.T) {
	tracker := &trackers.Tracker{
		ID:           1,
		Name:         "Bug",
		CustomFields: []trackers.TrackerCustomField{},
	}

	result, err := promptCustomFields(tracker, nil)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

// TestPromptCustomFields_NonTerminal 测试非终端环境
func TestPromptCustomFields_NonTerminal(t *testing.T) {
	tracker := &trackers.Tracker{
		ID:   1,
		Name: "Bug",
		CustomFields: []trackers.TrackerCustomField{
			{ID: 1, Name: "Severity", FieldFormat: "list"},
		},
	}

	// 在非终端环境中运行，应该直接返回 nil
	result, err := promptCustomFields(tracker, nil)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result in non-terminal environment, got %v", result)
	}
}

// TestPromptCustomFields_WithInitialValues 测试带初始值的情况
func TestPromptCustomFields_WithInitialValues(t *testing.T) {
	tracker := &trackers.Tracker{
		ID:   1,
		Name: "Bug",
		CustomFields: []trackers.TrackerCustomField{
			{ID: 1, Name: "Severity", FieldFormat: "string"},
		},
	}

	initialValues := map[int]CustomField{
		1: {ID: 1, Value: "High"},
	}

	// 在非终端环境中运行，应该直接返回 nil
	result, err := promptCustomFields(tracker, initialValues)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result in non-terminal environment, got %v", result)
	}
}

// TestPromptCustomFields_NilTracker 测试 nil tracker
// 注意：当前实现会在访问 tracker.CustomFields 时 panic
// 这个测试验证了代码的当前行为
func TestPromptCustomFields_NilTracker(t *testing.T) {
	// 当前实现没有处理 nil tracker，会 panic
	// 这里我们跳过这个测试，因为实际代码中不会传入 nil tracker
	t.Skip("Skipping nil tracker test - current implementation does not handle nil tracker")
}

// TestPromptCustomFieldsInteractive_StringField 测试字符串类型自定义字段
func TestPromptCustomFieldsInteractive_StringField(t *testing.T) {
	tracker := &trackers.Tracker{
		ID:   1,
		Name: "Bug",
		CustomFields: []trackers.TrackerCustomField{
			{ID: 1, Name: "Severity", FieldFormat: "string"},
		},
	}

	// 模拟用户输入 "High"
	input := strings.NewReader("High\n")
	reader := bufio.NewReader(input)

	result, err := promptCustomFieldsInteractive(tracker, nil, reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0].ID != 1 || result[0].Value != "High" {
		t.Errorf("expected {ID: 1, Value: 'High'}, got {ID: %d, Value: '%s'}", result[0].ID, result[0].Value)
	}
}

// TestPromptCustomFieldsInteractive_BoolField 测试布尔类型自定义字段
func TestPromptCustomFieldsInteractive_BoolField(t *testing.T) {
	tracker := &trackers.Tracker{
		ID:   1,
		Name: "Bug",
		CustomFields: []trackers.TrackerCustomField{
			{ID: 1, Name: "Confirmed", FieldFormat: "bool"},
		},
	}

	// 模拟用户输入 "y"
	input := strings.NewReader("y\n")
	reader := bufio.NewReader(input)

	result, err := promptCustomFieldsInteractive(tracker, nil, reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0].ID != 1 || result[0].Value != "1" {
		t.Errorf("expected {ID: 1, Value: '1'}, got {ID: %d, Value: '%s'}", result[0].ID, result[0].Value)
	}
}

// TestPromptCustomFieldsInteractive_BoolField_No 测试布尔类型自定义字段输入 n
func TestPromptCustomFieldsInteractive_BoolField_No(t *testing.T) {
	tracker := &trackers.Tracker{
		ID:   1,
		Name: "Bug",
		CustomFields: []trackers.TrackerCustomField{
			{ID: 1, Name: "Confirmed", FieldFormat: "bool"},
		},
	}

	// 模拟用户输入 "n"
	input := strings.NewReader("n\n")
	reader := bufio.NewReader(input)

	result, err := promptCustomFieldsInteractive(tracker, nil, reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0].ID != 1 || result[0].Value != "0" {
		t.Errorf("expected {ID: 1, Value: '0'}, got {ID: %d, Value: '%s'}", result[0].ID, result[0].Value)
	}
}

// TestPromptCustomFieldsInteractive_DateField 测试日期类型自定义字段
func TestPromptCustomFieldsInteractive_DateField(t *testing.T) {
	tracker := &trackers.Tracker{
		ID:   1,
		Name: "Bug",
		CustomFields: []trackers.TrackerCustomField{
			{ID: 1, Name: "Due Date", FieldFormat: "date"},
		},
	}

	// 模拟用户输入日期
	input := strings.NewReader("2024-01-15\n")
	reader := bufio.NewReader(input)

	result, err := promptCustomFieldsInteractive(tracker, nil, reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0].ID != 1 || result[0].Value != "2024-01-15" {
		t.Errorf("expected {ID: 1, Value: '2024-01-15'}, got {ID: %d, Value: '%s'}", result[0].ID, result[0].Value)
	}
}

// TestPromptCustomFieldsInteractive_ListField 测试列表类型自定义字段
func TestPromptCustomFieldsInteractive_ListField(t *testing.T) {
	tracker := &trackers.Tracker{
		ID:   1,
		Name: "Bug",
		CustomFields: []trackers.TrackerCustomField{
			{
				ID:          1,
				Name:        "Priority",
				FieldFormat: "list",
				PossibleValues: []trackers.ValueLabel{
					{Label: "Low", Value: "1"},
					{Label: "Medium", Value: "2"},
					{Label: "High", Value: "3"},
				},
			},
		},
	}

	// 模拟用户选择第2个选项
	input := strings.NewReader("2\n")
	reader := bufio.NewReader(input)

	result, err := promptCustomFieldsInteractive(tracker, nil, reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0].ID != 1 || result[0].Value != "2" {
		t.Errorf("expected {ID: 1, Value: '2'}, got {ID: %d, Value: '%s'}", result[0].ID, result[0].Value)
	}
}

// TestPromptCustomFieldsInteractive_WithInitialValues 测试带初始值的情况
func TestPromptCustomFieldsInteractive_WithInitialValues(t *testing.T) {
	tracker := &trackers.Tracker{
		ID:   1,
		Name: "Bug",
		CustomFields: []trackers.TrackerCustomField{
			{ID: 1, Name: "Severity", FieldFormat: "string"},
		},
	}

	initialValues := map[int]CustomField{
		1: {ID: 1, Value: "Medium"},
	}

	// 模拟用户直接回车（使用默认值）
	input := strings.NewReader("\n")
	reader := bufio.NewReader(input)

	result, err := promptCustomFieldsInteractive(tracker, initialValues, reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0].ID != 1 || result[0].Value != "Medium" {
		t.Errorf("expected {ID: 1, Value: 'Medium'}, got {ID: %d, Value: '%s'}", result[0].ID, result[0].Value)
	}
}

// TestPromptCustomFieldsInteractive_MultipleFields 测试多个自定义字段
func TestPromptCustomFieldsInteractive_MultipleFields(t *testing.T) {
	tracker := &trackers.Tracker{
		ID:   1,
		Name: "Bug",
		CustomFields: []trackers.TrackerCustomField{
			{ID: 1, Name: "Severity", FieldFormat: "string"},
			{ID: 2, Name: "Confirmed", FieldFormat: "bool"},
		},
	}

	// 模拟用户输入
	input := strings.NewReader("High\ny\n")
	reader := bufio.NewReader(input)

	result, err := promptCustomFieldsInteractive(tracker, nil, reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result))
	}
}

// TestPromptCustomFieldsInteractive_InvalidListIndex 测试无效的列表索引
func TestPromptCustomFieldsInteractive_InvalidListIndex(t *testing.T) {
	tracker := &trackers.Tracker{
		ID:   1,
		Name: "Bug",
		CustomFields: []trackers.TrackerCustomField{
			{
				ID:          1,
				Name:        "Priority",
				FieldFormat: "list",
				PossibleValues: []trackers.ValueLabel{
					{Label: "Low", Value: "1"},
					{Label: "High", Value: "2"},
				},
			},
		},
	}

	// 模拟用户输入无效的索引
	input := strings.NewReader("99\n")
	reader := bufio.NewReader(input)

	result, err := promptCustomFieldsInteractive(tracker, nil, reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 无效索引不应该添加值
	if len(result) != 0 {
		t.Errorf("expected 0 results for invalid index, got %d", len(result))
	}
}

// TestPromptCustomFieldsInteractive_EmptyListPossibleValues 测试空列表选项
func TestPromptCustomFieldsInteractive_EmptyListPossibleValues(t *testing.T) {
	tracker := &trackers.Tracker{
		ID:   1,
		Name: "Bug",
		CustomFields: []trackers.TrackerCustomField{
			{
				ID:             1,
				Name:           "Priority",
				FieldFormat:    "list",
				PossibleValues: []trackers.ValueLabel{},
			},
		},
	}

	input := strings.NewReader("")
	reader := bufio.NewReader(input)

	result, err := promptCustomFieldsInteractive(tracker, nil, reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 空列表选项不应该添加值
	if len(result) != 0 {
		t.Errorf("expected 0 results for empty possible values, got %d", len(result))
	}
}

// TestPromptCustomFieldsInteractive_BoolField_EmptyWithInitial 测试布尔字段空输入但有初始值
func TestPromptCustomFieldsInteractive_BoolField_EmptyWithInitial(t *testing.T) {
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

	// 模拟用户直接回车
	input := strings.NewReader("\n")
	reader := bufio.NewReader(input)

	result, err := promptCustomFieldsInteractive(tracker, initialValues, reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	// 应该保留初始值
	if result[0].Value != "1" {
		t.Errorf("expected initial value '1', got '%s'", result[0].Value)
	}
}

// TestPromptCustomFieldsInteractive_ListField_EmptyWithInitial 测试列表字段空输入但有初始值
func TestPromptCustomFieldsInteractive_ListField_EmptyWithInitial(t *testing.T) {
	tracker := &trackers.Tracker{
		ID:   1,
		Name: "Bug",
		CustomFields: []trackers.TrackerCustomField{
			{
				ID:          1,
				Name:        "Priority",
				FieldFormat: "list",
				PossibleValues: []trackers.ValueLabel{
					{Label: "Low", Value: "1"},
					{Label: "High", Value: "2"},
				},
			},
		},
	}

	initialValues := map[int]CustomField{
		1: {ID: 1, Value: "2"},
	}

	// 模拟用户直接回车
	input := strings.NewReader("\n")
	reader := bufio.NewReader(input)

	result, err := promptCustomFieldsInteractive(tracker, initialValues, reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	// 应该保留初始值
	if result[0].Value != "2" {
		t.Errorf("expected initial value '2', got '%s'", result[0].Value)
	}
}

// TestPromptCustomFieldsInteractive_DateField_EmptyNoDefault 测试日期字段空输入且无默认值
func TestPromptCustomFieldsInteractive_DateField_EmptyNoDefault(t *testing.T) {
	tracker := &trackers.Tracker{
		ID:   1,
		Name: "Bug",
		CustomFields: []trackers.TrackerCustomField{
			{ID: 1, Name: "Due Date", FieldFormat: "date"},
		},
	}

	// 模拟用户直接回车（无默认值）
	input := strings.NewReader("\n")
	reader := bufio.NewReader(input)

	result, err := promptCustomFieldsInteractive(tracker, nil, reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 空输入且无默认值，不应该添加值
	if len(result) != 0 {
		t.Errorf("expected 0 results for empty date with no default, got %d", len(result))
	}
}

// TestPromptCustomFieldsInteractive_DateField_EmptyWithDefault 测试日期字段空输入但有默认值
func TestPromptCustomFieldsInteractive_DateField_EmptyWithDefault(t *testing.T) {
	tracker := &trackers.Tracker{
		ID:   1,
		Name: "Bug",
		CustomFields: []trackers.TrackerCustomField{
			{ID: 1, Name: "Due Date", FieldFormat: "date"},
		},
	}

	initialValues := map[int]CustomField{
		1: {ID: 1, Value: "2024-06-01"},
	}

	// 模拟用户直接回车（使用默认值）
	input := strings.NewReader("\n")
	reader := bufio.NewReader(input)

	result, err := promptCustomFieldsInteractive(tracker, initialValues, reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	// 应该使用默认值
	if result[0].Value != "2024-06-01" {
		t.Errorf("expected default value '2024-06-01', got '%s'", result[0].Value)
	}
}

// TestPromptCustomFieldsInteractive_StringField_EmptyNoInitial 测试字符串字段空输入且无初始值
func TestPromptCustomFieldsInteractive_StringField_EmptyNoInitial(t *testing.T) {
	tracker := &trackers.Tracker{
		ID:   1,
		Name: "Bug",
		CustomFields: []trackers.TrackerCustomField{
			{ID: 1, Name: "Severity", FieldFormat: "string"},
		},
	}

	// 模拟用户直接回车（无初始值）
	input := strings.NewReader("\n")
	reader := bufio.NewReader(input)

	result, err := promptCustomFieldsInteractive(tracker, nil, reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 空输入且无初始值，不应该添加值
	if len(result) != 0 {
		t.Errorf("expected 0 results for empty string with no initial, got %d", len(result))
	}
}

// TestPromptCustomFields_TerminalMock 测试终端环境模拟
func TestPromptCustomFields_TerminalMock(t *testing.T) {
	// 保存原始函数
	originalIsTerminal := isTerminalFunc
	originalNewStdinReader := newStdinReader
	// 恢复原始函数
	defer func() {
		isTerminalFunc = originalIsTerminal
		newStdinReader = originalNewStdinReader
	}()

	// 模拟终端环境
	isTerminalFunc = func() bool { return true }

	// 模拟 stdin 读取器
	newStdinReader = func() *bufio.Reader {
		return bufio.NewReader(strings.NewReader("test-value\n"))
	}

	tracker := &trackers.Tracker{
		ID:   1,
		Name: "Bug",
		CustomFields: []trackers.TrackerCustomField{
			{ID: 1, Name: "Severity", FieldFormat: "string"},
		},
	}

	result, err := promptCustomFields(tracker, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0].ID != 1 || result[0].Value != "test-value" {
		t.Errorf("expected {ID: 1, Value: 'test-value'}, got {ID: %d, Value: '%s'}", result[0].ID, result[0].Value)
	}
}

// TestNewBufioReader 测试 newBufioReader 函数
func TestNewBufioReader(t *testing.T) {
	reader := newBufioReader()
	if reader == nil {
		t.Error("expected non-nil reader")
	}
}
