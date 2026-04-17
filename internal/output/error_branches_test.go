// internal/output/error_branches_test.go
package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
)

// TestNormalizePayloadErrors 测试 normalizePayload 的错误分支
func TestNormalizePayloadErrors(t *testing.T) {
	// 创建一个会导致 JSON 编码错误的情况
	// Go 的 json.Marshal 无法序列化 channel 等不可序列化的类型
	circular := make(chan int)
	_, err := normalizePayload(circular)
	if err == nil {
		t.Error("expected error for channel type")
	}
}

// TestWriteJSONErrors 测试 WriteJSON 的错误分支
func TestWriteJSONErrors(t *testing.T) {
	// 测试无法序列化的数据类型
	circular := make(chan int)
	err := WriteJSON(&bytes.Buffer{}, circular)
	if err == nil {
		t.Error("expected error for channel type")
	}
}

// TestWriteRawErrors 测试 WriteRaw 的错误分支
func TestWriteRawErrors(t *testing.T) {
	// 测试无法序列化的数据类型
	circular := make(chan int)
	err := WriteRaw(&bytes.Buffer{}, circular)
	if err == nil {
		t.Error("expected error for channel type")
	}
}

// TestApplyJQWithJSONMarshalError 测试 ApplyJQ 的 JSON 编码错误
func TestApplyJQWithJSONMarshalError(t *testing.T) {
	// 测试无法序列化的数据类型
	circular := make(chan int)
	err := ApplyJQ(&bytes.Buffer{}, circular, ".key")
	if err == nil {
		t.Error("expected error for channel type")
	}
}

// TestSelectFieldsWithJSONMarshalError 测试 SelectFields 的 JSON 编码错误
func TestSelectFieldsWithJSONMarshalError(t *testing.T) {
	circular := make(chan int)
	_, err := SelectFields(circular, []string{"key"})
	if err == nil {
		t.Error("expected error for unmarshalable data")
	}
}

func TestSelectFieldsWithJSONUnmarshalError(t *testing.T) {
	slice := []int{1, 2, 3}
	result, err := SelectFields(slice, []string{"id"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Error("expected original data to be returned for non-map type")
	}
}

// TestWriteTableWithNormalizeError 测试 WriteTable 的 normalize 错误
func TestWriteTableWithNormalizeError(t *testing.T) {
	// 测试无法序列化的数据类型
	circular := make(chan int)
	err := WriteTable(&bytes.Buffer{}, circular)
	if err == nil {
		t.Error("expected error for channel type")
	}
}

// TestFormatValueWithMarshalError 测试 formatValue 的 JSON 编码错误
func TestFormatValueWithMarshalError(t *testing.T) {
	// 测试无法序列化的数据类型
	circular := make(chan int)
	result := formatValue(circular)
	// 对于无法序列化的数据，应该使用 fmt.Sprintf
	if result == "" {
		t.Error("expected non-empty result")
	}
}

// TestWriteTableWithRowsFromSliceError 测试 WriteTable 的 rowsFromSlice 分支
func TestWriteTableWithRowsFromSliceError(t *testing.T) {
	// 测试空数组的情况
	var buf bytes.Buffer
	err := WriteTable(&buf, []any{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestApplyJQWithJSONUnmarshalError 测试 ApplyJQ 的 JSON 解码错误
func TestApplyJQWithJSONUnmarshalError(t *testing.T) {
	// 测试无法序列化的数据类型
	circular := make(chan int)
	err := ApplyJQ(&bytes.Buffer{}, circular, ".key")
	if err == nil {
		t.Error("expected error for channel type")
	}
}

// TestApplyJQWithInvalidQuery 测试 ApplyJQ 的无效查询
func TestApplyJQWithInvalidQuery(t *testing.T) {
	data := map[string]any{"key": "value"}
	err := ApplyJQ(&bytes.Buffer{}, data, ".[invalid")
	if err == nil {
		t.Error("expected error for invalid query")
	}
}

// TestApplyJQWithQueryError 测试 ApplyJQ 查询执行时的错误
func TestApplyJQWithQueryError(t *testing.T) {
	data := map[string]any{"key": "value"}
	err := ApplyJQ(&bytes.Buffer{}, data, ".key | error")
	if err == nil {
		t.Error("expected error from error function")
	}
}

// TestWriteTableWithMapInput 测试 WriteTable 的 map 输入
func TestWriteTableWithMapInput(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]any{"id": 1, "name": "test"}
	err := WriteTable(&buf, data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}

// TestWriteTableWithNonMapNonSliceInput 测试 WriteTable 的非 map 非 slice 输入
func TestWriteTableWithNonMapNonSliceInput(t *testing.T) {
	var buf bytes.Buffer
	data := "plain string"
	err := WriteTable(&buf, data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}

// TestWriteKeyValuesWithEmptyMap 测试 writeKeyValues 的空 map 输入
func TestWriteKeyValuesWithEmptyMap(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]any{}
	err := writeKeyValues(&buf, data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// 空 map 应该产生空输出
	if buf.Len() > 0 {
		t.Error("expected empty output for empty map")
	}
}

// TestWriteTableWithComplexData 测试 WriteTable 的复杂数据
func TestWriteTableWithComplexData(t *testing.T) {
	var buf bytes.Buffer
	data := []map[string]any{
		{
			"id":          1,
			"name":        "test",
			"nested":      map[string]any{"key": "value"},
			"slice":       []int{1, 2, 3},
			"nil_value":   nil,
			"bool_value":  true,
			"float_value": 3.14,
		},
	}
	err := WriteTable(&buf, data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}

// TestWriteWithAllFormats 测试 Write 的所有格式
func TestWriteWithAllFormats(t *testing.T) {
	data := map[string]any{"id": 1, "name": "test"}

	formats := []Format{FormatJSON, FormatTable, FormatRaw, Format("unknown")}
	for _, format := range formats {
		var buf bytes.Buffer
		err := Write(&buf, format, data)
		if err != nil {
			t.Errorf("unexpected error for format %s: %v", format, err)
		}
		if buf.Len() == 0 {
			t.Errorf("expected non-empty output for format %s", format)
		}
	}
}

// TestWriteRawWithNonString 测试 WriteRaw 的非字符串输入
func TestWriteRawWithNonString(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]any{"id": 1, "name": "test"}
	err := WriteRaw(&buf, data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}

// TestNormalizePayloadWithInvalidJSON 测试 normalizePayload 的无效 JSON
func TestNormalizePayloadWithInvalidJSON(t *testing.T) {
	// 测试无法序列化的数据类型
	circular := make(chan int)
	_, err := normalizePayload(circular)
	if err == nil {
		t.Error("expected error for channel type")
	}
}

// TestApplyJQWithEmptyQuery 测试 ApplyJQ 的空查询
func TestApplyJQWithEmptyQuery(t *testing.T) {
	data := map[string]any{"key": "value"}
	err := ApplyJQ(&bytes.Buffer{}, data, "")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestApplyJQWithIdentityQuery 测试 ApplyJQ 的恒等查询
func TestApplyJQWithIdentityQuery(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]any{"key": "value"}
	err := ApplyJQ(&buf, data, ".")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}

// TestApplyJQWithRecursiveDescent 测试 ApplyJQ 的递归下降
func TestApplyJQWithRecursiveDescent(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]any{
		"a": map[string]any{
			"b": map[string]any{
				"c": "value",
			},
		},
	}
	err := ApplyJQ(&buf, data, "..")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}

// TestSelectFieldsWithEmptyPayload 测试 SelectFields 的空数据
func TestSelectFieldsWithEmptyPayload(t *testing.T) {
	result, _ := SelectFields(nil, []string{"id"})
	if result != nil {
		t.Error("expected nil for nil payload")
	}
}

// TestSelectFieldsWithEmptyFields 测试 SelectFields 的空字段列表
func TestSelectFieldsWithEmptyFields(t *testing.T) {
	data := map[string]any{"id": 1, "name": "test"}
	result, _ := SelectFields(data, []string{})
	// 空字段列表应该返回原始数据
	if result == nil {
		t.Error("expected original data to be returned")
	}
}

// TestWriteTableWithNilPayload 测试 WriteTable 的 nil 输入
func TestWriteTableWithNilPayload(t *testing.T) {
	var buf bytes.Buffer
	err := WriteTable(&buf, nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestNormalizePayloadWithNil 测试 normalizePayload 的 nil 输入
func TestNormalizePayloadWithNil(t *testing.T) {
	result, err := normalizePayload(nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != nil {
		t.Error("expected nil result for nil input")
	}
}

// TestNormalizePayloadWithString 测试 normalizePayload 的字符串输入
func TestNormalizePayloadWithString(t *testing.T) {
	result, err := normalizePayload("test string")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "test string" {
		t.Errorf("expected 'test string', got %v", result)
	}
}

// TestRowsFromSliceWithMixedTypes 测试 rowsFromSlice 的混合类型
func TestRowsFromSliceWithMixedTypes(t *testing.T) {
	items := []any{
		map[string]any{"id": 1},
		"string",
		123,
	}
	headers, rows, ok := rowsFromSlice(items)
	if !ok {
		t.Error("expected ok to be true")
	}
	if len(headers) != 1 || headers[0] != "value" {
		t.Errorf("expected headers to be ['value'], got %v", headers)
	}
	if len(rows) != 3 {
		t.Errorf("expected 3 rows, got %d", len(rows))
	}
}

// TestTruncateWithNegativeWidth 测试 truncate 的负宽度
func TestTruncateWithNegativeWidth(t *testing.T) {
	result := truncate("hello", -1)
	if result != "" {
		t.Errorf("expected empty string for negative width, got %q", result)
	}
}

// TestTruncateWithZeroWidth 测试 truncate 的零宽度
func TestTruncateWithZeroWidth(t *testing.T) {
	result := truncate("hello", 0)
	if result != "" {
		t.Errorf("expected empty string for zero width, got %q", result)
	}
}

// TestTruncateWithOneWidth 测试 truncate 的宽度为 1
func TestTruncateWithOneWidth(t *testing.T) {
	result := truncate("hello", 1)
	if result != "…" {
		t.Errorf("expected '…', got %q", result)
	}
}

// TestFormatValueWithNil 测试 formatValue 的 nil 输入
func TestFormatValueWithNil(t *testing.T) {
	result := formatValue(nil)
	if result != "" {
		t.Errorf("expected empty string for nil, got %q", result)
	}
}

// TestFormatValueWithString 测试 formatValue 的字符串输入
func TestFormatValueWithString(t *testing.T) {
	result := formatValue("hello")
	if result != "hello" {
		t.Errorf("expected 'hello', got %q", result)
	}
}

// TestFormatValueWithInt 测试 formatValue 的整数输入
func TestFormatValueWithInt(t *testing.T) {
	result := formatValue(42)
	if !json.Valid([]byte(result)) {
		t.Errorf("expected valid JSON, got %q", result)
	}
}

// TestFormatValueWithBool 测试 formatValue 的布尔输入
func TestFormatValueWithBool(t *testing.T) {
	result := formatValue(true)
	if !json.Valid([]byte(result)) {
		t.Errorf("expected valid JSON, got %q", result)
	}
}

// TestFormatValueWithSlice 测试 formatValue 的切片输入
func TestFormatValueWithSlice(t *testing.T) {
	result := formatValue([]int{1, 2, 3})
	if !json.Valid([]byte(result)) {
		t.Errorf("expected valid JSON, got %q", result)
	}
}

// TestFormatValueWithMap 测试 formatValue 的 map 输入
func TestFormatValueWithMap(t *testing.T) {
	result := formatValue(map[string]string{"key": "value"})
	if !json.Valid([]byte(result)) {
		t.Errorf("expected valid JSON, got %q", result)
	}
}

// TestWriteTableWithWriteError 测试 WriteTable 的写入错误
func TestWriteTableWithWriteError(t *testing.T) {
	errWriter := &errorWriter{}
	data := []map[string]any{{"id": 1, "name": "test"}}
	err := WriteTable(errWriter, data)
	if err == nil {
		t.Error("expected error from errorWriter")
	}
}

// TestWriteKeyValuesWithWriteError 测试 writeKeyValues 的写入错误
func TestWriteKeyValuesWithWriteError(t *testing.T) {
	errWriter := &errorWriter{}
	data := map[string]any{"id": 1, "name": "test"}
	err := writeKeyValues(errWriter, data)
	if err == nil {
		t.Error("expected error from errorWriter")
	}
}

// limitedErrorWriter succeeds for first N writes, then fails
type limitedErrorWriter struct {
	count   int
	limit   int
	failErr error
}

func (w *limitedErrorWriter) Write(p []byte) (n int, err error) {
	w.count++
	if w.count > w.limit {
		return 0, w.failErr
	}
	return len(p), nil
}

// TestWriteDataRowsWithWriteError 测试 writeDataRows 的写入错误
func TestWriteDataRowsWithWriteError(t *testing.T) {
	// For row[0]: Fprintf (i=0), WriteString+Fprintf (i=1), Fprintln = 3 writes
	// So we need limit=3 to fail on Fprintln
	writer := &limitedErrorWriter{limit: 3, failErr: fmt.Errorf("write error at row")}
	rows := [][]string{{"v1", "v2"}, {"v3", "v4"}}
	widths := []int{10, 10}
	err := writeDataRows(writer, rows, widths)
	if err == nil {
		t.Error("expected error from limitedErrorWriter")
	}
}

// TestWriteTableWithNonMapNonSliceTypes tests the default branch in WriteTable
func TestWriteTableWithDefaultCase(t *testing.T) {
	var buf bytes.Buffer
	// float64 would be the default case after normalizePayload
	var data float64 = 3.14159
	err := WriteTable(&buf, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty output for float64")
	}
}

// TestWriteTableWithSliceNonMaps tests []any that isn't all maps
func TestWriteTableWithSliceNonMaps(t *testing.T) {
	var buf bytes.Buffer
	data := []any{"string", 123, true}
	err := WriteTable(&buf, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty output for non-map slice")
	}
}

// TestApplyJQWithWriteError 测试 ApplyJQ 的写入错误
func TestApplyJQWithWriteError(t *testing.T) {
	errWriter := &errorWriter{}
	data := map[string]any{"key": "value"}
	err := ApplyJQ(errWriter, data, ".key")
	if err == nil {
		t.Error("expected error from errorWriter")
	}
}

// 确保 errorWriter 实现了 io.Writer 接口
var _ interface {
	Write(p []byte) (n int, err error)
} = (*errorWriter)(nil)

// TestWriteJSONWithMarshalError 测试 WriteJSON 的序列化错误
func TestWriteJSONWithMarshalError(t *testing.T) {
	// 创建一个无法序列化的值
	unmarshalable := func() {}
	err := WriteJSON(&bytes.Buffer{}, unmarshalable)
	if err == nil {
		t.Error("expected error for function type")
	}
}

// TestWriteRawWithMarshalError 测试 WriteRaw 的序列化错误
func TestWriteRawWithMarshalError(t *testing.T) {
	// 创建一个无法序列化的值
	unmarshalable := func() {}
	err := WriteRaw(&bytes.Buffer{}, unmarshalable)
	if err == nil {
		t.Error("expected error for function type")
	}
}

// TestNormalizePayloadWithMarshalError 测试 normalizePayload 的序列化错误
func TestNormalizePayloadWithMarshalError(t *testing.T) {
	// 创建一个无法序列化的值
	unmarshalable := func() {}
	_, err := normalizePayload(unmarshalable)
	if err == nil {
		t.Error("expected error for function type")
	}
}

// TestApplyJQWithMarshalError 测试 ApplyJQ 的序列化错误
func TestApplyJQWithMarshalError(t *testing.T) {
	// 创建一个无法序列化的值
	unmarshalable := func() {}
	err := ApplyJQ(&bytes.Buffer{}, unmarshalable, ".")
	if err == nil {
		t.Error("expected error for function type")
	}
}

// TestSelectFieldsWithMarshalError 测试 SelectFields 的序列化错误
func TestSelectFieldsWithMarshalError(t *testing.T) {
	unmarshalable := func() {}
	_, err := SelectFields(unmarshalable, []string{"key"})
	if err == nil {
		t.Error("expected error for unmarshalable data")
	}
}

// TestWriteTableWithMarshalError 测试 WriteTable 的序列化错误
func TestWriteTableWithMarshalError(t *testing.T) {
	// 创建一个无法序列化的值
	unmarshalable := func() {}
	err := WriteTable(&bytes.Buffer{}, unmarshalable)
	if err == nil {
		t.Error("expected error for function type")
	}
}

// TestFormatValueWithUnmarshalable 测试 formatValue 的无法序列化值
func TestFormatValueWithUnmarshalable(t *testing.T) {
	// 创建一个无法序列化的值
	unmarshalable := func() {}
	result := formatValue(unmarshalable)
	// 对于无法序列化的数据，应该使用 fmt.Sprintf
	if result == "" {
		t.Error("expected non-empty result")
	}
}

// TestWriteWithMarshalError 测试 Write 的序列化错误
func TestWriteWithMarshalError(t *testing.T) {
	// 创建一个无法序列化的值
	unmarshalable := func() {}
	err := Write(&bytes.Buffer{}, FormatJSON, unmarshalable)
	if err == nil {
		t.Error("expected error for function type")
	}
}
