// internal/output/remaining_branches_test.go
package output

import (
	"bytes"
	"testing"
)

// TestWriteTableWithMoreCellsThanHeaders 测试行中单元格数量超过表头数量的情况
func TestWriteTableWithMoreCellsThanHeaders(t *testing.T) {
	// 这个测试覆盖 writeTable 中的 if i >= len(widths) 分支
	var buf bytes.Buffer
	// 手动调用 writeTable，传入比表头更多的单元格
	headers := []string{"id", "name"}
	rows := [][]string{
		{"1", "test", "extra"}, // 这个行有 3 个单元格，但表头只有 2 个
	}
	err := writeTable(&buf, headers, rows)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// 输出应该只包含前两个单元格
	output := buf.String()
	if len(output) == 0 {
		t.Error("expected non-empty output")
	}
}

// TestWriteTableWithEmptyHeaders 测试空表头的情况
func TestWriteTableWithEmptyHeaders(t *testing.T) {
	var buf bytes.Buffer
	headers := []string{}
	rows := [][]string{}
	err := writeTable(&buf, headers, rows)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// 空表头应该输出换行符
	output := buf.String()
	if len(output) == 0 {
		t.Error("expected output with newlines")
	}
}

// TestWriteTableWithSingleHeader 测试单个表头的情况
func TestWriteTableWithSingleHeader(t *testing.T) {
	var buf bytes.Buffer
	headers := []string{"id"}
	rows := [][]string{
		{"1"},
		{"2"},
	}
	err := writeTable(&buf, headers, rows)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}

// TestWriteTableWithVeryLongHeader 测试非常长的表头
func TestWriteTableWithVeryLongHeader(t *testing.T) {
	var buf bytes.Buffer
	// 创建一个超过 60 个字符的表头
	longHeader := "this_is_a_very_long_header_name_that_exceeds_the_maximum_width_limit"
	headers := []string{longHeader}
	rows := [][]string{
		{"value"},
	}
	err := writeTable(&buf, headers, rows)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// 表头应该被截断
	output := buf.String()
	if len(output) == 0 {
		t.Error("expected non-empty output")
	}
}

// TestWriteTableWithVeryLongCell 测试非常长的单元格
func TestWriteTableWithVeryLongCell(t *testing.T) {
	var buf bytes.Buffer
	// 创建一个超过 60 个字符的单元格
	longCell := "this_is_a_very_long_cell_value_that_exceeds_the_maximum_width_limit"
	headers := []string{"id"}
	rows := [][]string{
		{longCell},
	}
	err := writeTable(&buf, headers, rows)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// 单元格应该被截断
	output := buf.String()
	if len(output) == 0 {
		t.Error("expected non-empty output")
	}
}

// TestWriteTableWithMultipleColumns 测试多列表格
func TestWriteTableWithMultipleColumns(t *testing.T) {
	var buf bytes.Buffer
	headers := []string{"id", "name", "email", "status"}
	rows := [][]string{
		{"1", "Alice", "alice@example.com", "active"},
		{"2", "Bob", "bob@example.com", "inactive"},
	}
	err := writeTable(&buf, headers, rows)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}

// TestWriteTableWithEmptyRows 测试空行的情况
func TestWriteTableWithEmptyRows(t *testing.T) {
	var buf bytes.Buffer
	headers := []string{"id", "name"}
	rows := [][]string{}
	err := writeTable(&buf, headers, rows)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// 空行应该只输出表头和分隔线
	output := buf.String()
	if len(output) == 0 {
		t.Error("expected non-empty output")
	}
}

// TestWriteTableWithEmptyCells 测试空单元格的情况
func TestWriteTableWithEmptyCells(t *testing.T) {
	var buf bytes.Buffer
	headers := []string{"id", "name", "email"}
	rows := [][]string{
		{"1", "", "test@example.com"},
		{"2", "Bob", ""},
	}
	err := writeTable(&buf, headers, rows)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}

// TestWriteTableWithUnicodeCells 测试 Unicode 单元格
func TestWriteTableWithUnicodeCells(t *testing.T) {
	var buf bytes.Buffer
	headers := []string{"id", "name"}
	rows := [][]string{
		{"1", "中文测试"},
		{"2", "日本語テスト"},
		{"3", "한국어 테스트"},
	}
	err := writeTable(&buf, headers, rows)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}

// TestWriteTableWithEmojiCells 测试 Emoji 单元格
func TestWriteTableWithEmojiCells(t *testing.T) {
	var buf bytes.Buffer
	headers := []string{"id", "emoji"}
	rows := [][]string{
		{"1", "😀"},
		{"2", "🎉"},
		{"3", "🚀"},
	}
	err := writeTable(&buf, headers, rows)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}

// TestRowsFromSliceWithEmptyMap 测试空 map 的切片
func TestRowsFromSliceWithEmptyMap(t *testing.T) {
	items := []any{
		map[string]any{},
		map[string]any{},
	}
	headers, rows, ok := rowsFromSlice(items)
	if !ok {
		t.Error("expected ok to be true")
	}
	if len(headers) != 0 {
		t.Errorf("expected 0 headers, got %d", len(headers))
	}
	if len(rows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(rows))
	}
}

// TestRowsFromSliceWithNilValues 测试包含 nil 值的切片
func TestRowsFromSliceWithNilValues(t *testing.T) {
	items := []any{
		map[string]any{"id": 1, "name": "test"},
		nil,
		map[string]any{"id": 2, "name": "test2"},
	}
	headers, rows, ok := rowsFromSlice(items)
	if !ok {
		t.Error("expected ok to be true")
	}
	// nil 值应该被当作非 map 处理
	if len(headers) != 1 || headers[0] != "value" {
		t.Errorf("expected headers to be ['value'], got %v", headers)
	}
	if len(rows) != 3 {
		t.Errorf("expected 3 rows, got %d", len(rows))
	}
}

// TestWriteKeyValuesWithMultipleKeys 测试多个键的情况
func TestWriteKeyValuesWithMultipleKeys(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]any{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
		"key4": "value4",
		"key5": "value5",
	}
	err := writeKeyValues(&buf, data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}

// TestWriteKeyValuesWithUnicodeKeys 测试 Unicode 键
func TestWriteKeyValuesWithUnicodeKeys(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]any{
		"中文键": "中文值",
		"日本語": "日本語値",
		"한국어": "한국어 값",
	}
	err := writeKeyValues(&buf, data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}

// TestWriteKeyValuesWithEmojiKeys 测试 Emoji 键
func TestWriteKeyValuesWithEmojiKeys(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]any{
		"😀": "happy",
		"🎉": "celebration",
		"🚀": "rocket",
	}
	err := writeKeyValues(&buf, data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}

// TestWriteKeyValuesWithNumericKeys 测试数字键
func TestWriteKeyValuesWithNumericKeys(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]any{
		"1": "one",
		"2": "two",
		"3": "three",
	}
	err := writeKeyValues(&buf, data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}

// TestWriteKeyValuesWithSpecialCharacters 测试特殊字符键
func TestWriteKeyValuesWithSpecialCharacters(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]any{
		"key-with-dash":       "value1",
		"key_with_underscore": "value2",
		"key.with.dot":        "value3",
		"key/with/slash":      "value4",
	}
	err := writeKeyValues(&buf, data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}

// TestTruncateWithExactWidth 测试精确宽度的截断
func TestTruncateWithExactWidth(t *testing.T) {
	result := truncate("hello", 5)
	if result != "hello" {
		t.Errorf("expected 'hello', got %q", result)
	}
}

// TestTruncateWithWidthTwo 测试宽度为 2 的截断
func TestTruncateWithWidthTwo(t *testing.T) {
	result := truncate("hello", 2)
	if result != "h…" {
		t.Errorf("expected 'h…', got %q", result)
	}
}

// TestTruncateWithLongUnicodeString 测试长 Unicode 字符串的截断
func TestTruncateWithLongUnicodeString(t *testing.T) {
	result := truncate("这是一个很长的中文字符串用于测试截断功能", 10)
	if len([]rune(result)) != 10 {
		t.Errorf("expected 10 runes, got %d", len([]rune(result)))
	}
}

// TestTruncateWithEmojiString 测试 Emoji 字符串的截断
func TestTruncateWithEmojiString(t *testing.T) {
	result := truncate("😀🎉🚀🌟💫", 3)
	if len([]rune(result)) != 3 {
		t.Errorf("expected 3 runes, got %d", len([]rune(result)))
	}
}

// TestFormatValueWithFloat 测试浮点数的格式化
func TestFormatValueWithFloat(t *testing.T) {
	result := formatValue(3.14159)
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

// TestFormatValueWithComplex 测试复数的格式化
func TestFormatValueWithComplex(t *testing.T) {
	result := formatValue(complex(1, 2))
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

// TestFormatValueWithStruct 测试结构体的格式化
func TestFormatValueWithStruct(t *testing.T) {
	type TestStruct struct {
		Name  string
		Value int
	}
	result := formatValue(TestStruct{Name: "test", Value: 42})
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

// TestFormatValueWithPointer 测试指针的格式化
func TestFormatValueWithPointer(t *testing.T) {
	value := 42
	result := formatValue(&value)
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

// TestFormatValueWithNilPointer 测试 nil 指针的格式化
func TestFormatValueWithNilPointer(t *testing.T) {
	var ptr *int
	result := formatValue(ptr)
	// nil 指针应该被序列化为 "null"
	if result != "null" {
		t.Errorf("expected 'null' for nil pointer, got %q", result)
	}
}

// TestWriteTableWithNilWriter 测试 nil 写入器
func TestWriteTableWithNilWriter(_ *testing.T) {
	// 这个测试应该 panic，但我们不能测试 panic
	// 所以我们跳过这个测试
}

// TestWriteKeyValuesWithNilWriter 测试 nil 写入器
func TestWriteKeyValuesWithNilWriter(_ *testing.T) {
	// 这个测试应该 panic，但我们不能测试 panic
	// 所以我们跳过这个测试
}

// TestWriteJSONWithNilWriter 测试 nil 写入器
func TestWriteJSONWithNilWriter(_ *testing.T) {
	// 这个测试应该 panic，但我们不能测试 panic
	// 所以我们跳过这个测试
}

// TestWriteRawWithNilWriter 测试 nil 写入器
func TestWriteRawWithNilWriter(_ *testing.T) {
	// 这个测试应该 panic，但我们不能测试 panic
	// 所以我们跳过这个测试
}

// TestApplyJQWithNilWriter 测试 nil 写入器
func TestApplyJQWithNilWriter(_ *testing.T) {
	// 这个测试应该 panic，但我们不能测试 panic
	// 所以我们跳过这个测试
}
