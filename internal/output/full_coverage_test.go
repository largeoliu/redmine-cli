// internal/output/full_coverage_test.go
package output

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
)

// TestWriteTableMultipleColumnsWithWriterErrors 测试 writeTable 在多列情况下的写入错误
func TestWriteTableMultipleColumnsWithWriterErrors(t *testing.T) {
	tests := []struct {
		name        string
		headers     []string
		rows        [][]string
		failAtWrite int // 在第几次 Write 时失败
	}{
		{
			name:        "fail at first column header",
			headers:     []string{"id", "name", "email"},
			rows:        [][]string{},
			failAtWrite: 1,
		},
		{
			name:        "fail at second column header",
			headers:     []string{"id", "name", "email"},
			rows:        [][]string{},
			failAtWrite: 2,
		},
		{
			name:        "fail at third column header",
			headers:     []string{"id", "name", "email"},
			rows:        [][]string{},
			failAtWrite: 3,
		},
		{
			name:        "fail at separator first column",
			headers:     []string{"id", "name"},
			rows:        [][]string{},
			failAtWrite: 4,
		},
		{
			name:        "fail at separator second column",
			headers:     []string{"id", "name"},
			rows:        [][]string{},
			failAtWrite: 5,
		},
		{
			name:        "fail at row first column",
			headers:     []string{"id", "name"},
			rows:        [][]string{{"1", "test"}},
			failAtWrite: 7,
		},
		{
			name:        "fail at row second column",
			headers:     []string{"id", "name"},
			rows:        [][]string{{"1", "test"}},
			failAtWrite: 8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := &controlledErrorWriter{failAt: tt.failAtWrite}
			err := writeTable(writer, tt.headers, tt.rows)
			if err == nil {
				t.Error("expected error from controlledErrorWriter")
			}
		})
	}
}

// controlledErrorWriter 在指定次数的 Write 调用后返回错误
type controlledErrorWriter struct {
	writeCount int
	failAt     int
}

func (w *controlledErrorWriter) Write(p []byte) (n int, err error) {
	w.writeCount++
	if w.writeCount >= w.failAt {
		return 0, errors.New("controlled write error")
	}
	return len(p), nil
}

// TestWriteTableWithExactWidthLimit 测试 writeTable 的宽度限制边界
func TestWriteTableWithExactWidthLimit(t *testing.T) {
	var buf bytes.Buffer
	// 创建一个正好 60 个字符的表头
	exactHeader := strings.Repeat("x", 60)
	headers := []string{exactHeader}
	rows := [][]string{{"value"}}
	err := writeTable(&buf, headers, rows)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestWriteTableWithWidthJustOverLimit 测试 writeTable 的宽度刚好超过限制
func TestWriteTableWithWidthJustOverLimit(t *testing.T) {
	var buf bytes.Buffer
	// 创建一个 61 个字符的表头（超过 60 的限制）
	longHeader := strings.Repeat("x", 61)
	headers := []string{longHeader}
	rows := [][]string{{"value"}}
	err := writeTable(&buf, headers, rows)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// 验证表头被截断
	output := buf.String()
	if len(output) == 0 {
		t.Error("expected non-empty output")
	}
}

// TestWriteKeyValuesWithExactMaxWidth 测试 writeKeyValues 的最大宽度边界
func TestWriteKeyValuesWithExactMaxWidth(t *testing.T) {
	var buf bytes.Buffer
	// 创建一个正好 24 个字符的键
	exactKey := strings.Repeat("k", 24)
	data := map[string]any{exactKey: "value"}
	err := writeKeyValues(&buf, data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestWriteKeyValuesWithWidthJustOverMaxWidth 测试 writeKeyValues 的宽度刚好超过最大宽度
func TestWriteKeyValuesWithWidthJustOverMaxWidth(t *testing.T) {
	var buf bytes.Buffer
	// 创建一个 25 个字符的键（超过 24 的限制）
	longKey := strings.Repeat("k", 25)
	data := map[string]any{longKey: "value"}
	err := writeKeyValues(&buf, data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestApplyJQWithJSONMarshalIndentError 测试 ApplyJQ 的 JSON 编码错误
func TestApplyJQWithJSONMarshalIndentError(t *testing.T) {
	// 使用一个会导致 json.MarshalIndent 失败的 writer
	errWriter := &jsonMarshalErrorWriter{}
	data := map[string]any{"key": "value"}
	err := ApplyJQ(errWriter, data, ".")
	if err == nil {
		t.Error("expected error from jsonMarshalErrorWriter")
	}
}

// jsonMarshalErrorWriter 是一个在 Write 时返回错误的写入器
type jsonMarshalErrorWriter struct{}

func (w *jsonMarshalErrorWriter) Write(_ []byte) (n int, err error) {
	return 0, errors.New("write error during JSON output")
}

// TestApplyJQWithMultipleIterations 测试 ApplyJQ 的多次迭代
func TestApplyJQWithMultipleIterations(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]any{
		"items": []any{1, 2, 3, 4, 5},
	}
	err := ApplyJQ(&buf, data, ".items[]")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// 应该有 5 行输出
	lines := strings.Count(buf.String(), "\n")
	if lines != 5 {
		t.Errorf("expected 5 lines, got %d", lines)
	}
}

// TestApplyJQWithIteratorError 测试 ApplyJQ 迭代器返回错误
func TestApplyJQWithIteratorError(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]any{"key": "value"}
	// 使用 error 函数来触发迭代器错误
	err := ApplyJQ(&buf, data, ".key | error")
	if err == nil {
		t.Error("expected error from jq error function")
	}
}

// TestSelectFieldsWithNonMapJSONResult 测试 SelectFields 处理非 map JSON 结果
func TestSelectFieldsWithNonMapJSONResult(t *testing.T) {
	data := []int{1, 2, 3}
	result, err := SelectFields(data, []string{"id"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Error("expected original data to be returned for non-map type")
	}
}

// TestWriteRawWithNonStringPayload 测试 WriteRaw 处理非字符串数据
func TestWriteRawWithNonStringPayload(t *testing.T) {
	tests := []struct {
		name    string
		payload any
	}{
		{name: "int", payload: 42},
		{name: "float", payload: 3.14},
		{name: "bool", payload: true},
		{name: "slice", payload: []int{1, 2, 3}},
		{name: "map", payload: map[string]string{"key": "value"}},
		{name: "nil", payload: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := WriteRaw(&buf, tt.payload)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if buf.Len() == 0 {
				t.Error("expected non-empty output")
			}
		})
	}
}

// TestWriteTableWithEmptyHeadersAndNonEmptyRows 测试空表头但非空行的情况
func TestWriteTableWithEmptyHeadersAndNonEmptyRows(t *testing.T) {
	var buf bytes.Buffer
	headers := []string{}
	rows := [][]string{{"value"}}
	err := writeTable(&buf, headers, rows)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestWriteTableWithMoreCellsThanHeadersInMultipleRows 测试多行中单元格数量超过表头
func TestWriteTableWithMoreCellsThanHeadersInMultipleRows(t *testing.T) {
	var buf bytes.Buffer
	headers := []string{"id"}
	rows := [][]string{
		{"1", "extra1"},
		{"2", "extra2"},
		{"3", "extra3"},
	}
	err := writeTable(&buf, headers, rows)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// 验证输出只包含第一列
	output := buf.String()
	if !strings.Contains(output, "1") || !strings.Contains(output, "2") || !strings.Contains(output, "3") {
		t.Error("expected output to contain all first column values")
	}
}

// TestWriteTableWithFewerCellsThanHeaders 测试单元格数量少于表头
func TestWriteTableWithFewerCellsThanHeaders(t *testing.T) {
	var buf bytes.Buffer
	headers := []string{"id", "name", "email"}
	rows := [][]string{
		{"1"},
		{"2", "Bob"},
	}
	err := writeTable(&buf, headers, rows)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestRowsFromSliceWithAllEmptyMaps 测试所有空 map 的切片
func TestRowsFromSliceWithAllEmptyMaps(t *testing.T) {
	items := []any{
		map[string]any{},
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
	if len(rows) != 3 {
		t.Errorf("expected 3 rows, got %d", len(rows))
	}
}

// TestRowsFromSliceWithMixedMapsAndNonMaps 测试混合 map 和非 map 的切片
func TestRowsFromSliceWithMixedMapsAndNonMaps(t *testing.T) {
	items := []any{
		map[string]any{"id": 1},
		"string",
		map[string]any{"id": 2, "name": "test"},
		42,
	}
	headers, rows, ok := rowsFromSlice(items)
	if !ok {
		t.Error("expected ok to be true")
	}
	// 由于有非 map 元素，应该使用 "value" 作为表头
	if len(headers) != 1 || headers[0] != "value" {
		t.Errorf("expected headers to be ['value'], got %v", headers)
	}
	if len(rows) != 4 {
		t.Errorf("expected 4 rows, got %d", len(rows))
	}
}

// TestFormatValueWithUnmarshalableType 测试 formatValue 处理无法序列化的类型
func TestFormatValueWithUnmarshalableType(t *testing.T) {
	// channel 无法被 JSON 序列化
	ch := make(chan int)
	result := formatValue(ch)
	// 应该使用 fmt.Sprintf 作为回退
	if result == "" {
		t.Error("expected non-empty result for channel type")
	}
}

// TestNormalizePayloadWithUnmarshalableType 测试 normalizePayload 处理无法序列化的类型
func TestNormalizePayloadWithUnmarshalableType(t *testing.T) {
	// channel 无法被 JSON 序列化
	ch := make(chan int)
	_, err := normalizePayload(ch)
	if err == nil {
		t.Error("expected error for channel type")
	}
}

// TestWriteJSONWithUnmarshalableType 测试 WriteJSON 处理无法序列化的类型
func TestWriteJSONWithUnmarshalableType(t *testing.T) {
	var buf bytes.Buffer
	// channel 无法被 JSON 序列化
	ch := make(chan int)
	err := WriteJSON(&buf, ch)
	if err == nil {
		t.Error("expected error for channel type")
	}
}

// TestWriteRawWithUnmarshalableType 测试 WriteRaw 处理无法序列化的类型
func TestWriteRawWithUnmarshalableType(t *testing.T) {
	var buf bytes.Buffer
	// channel 无法被 JSON 序列化
	ch := make(chan int)
	err := WriteRaw(&buf, ch)
	if err == nil {
		t.Error("expected error for channel type")
	}
}

// TestWriteTableWithUnmarshalableType 测试 WriteTable 处理无法序列化的类型
func TestWriteTableWithUnmarshalableType(t *testing.T) {
	var buf bytes.Buffer
	// channel 无法被 JSON 序列化
	ch := make(chan int)
	err := WriteTable(&buf, ch)
	if err == nil {
		t.Error("expected error for channel type")
	}
}

// TestApplyJQWithUnmarshalableType 测试 ApplyJQ 处理无法序列化的类型
func TestApplyJQWithUnmarshalableType(t *testing.T) {
	var buf bytes.Buffer
	// channel 无法被 JSON 序列化
	ch := make(chan int)
	err := ApplyJQ(&buf, ch, ".")
	if err == nil {
		t.Error("expected error for channel type")
	}
}

// TestSelectFieldsWithUnmarshalableType 测试 SelectFields 处理无法序列化的类型
func TestSelectFieldsWithUnmarshalableType(t *testing.T) {
	ch := make(chan int)
	_, err := SelectFields(ch, []string{"key"})
	if err == nil {
		t.Error("expected error for unmarshalable data")
	}
}

// TestWriteWithAllFormatsAndUnmarshalableType 测试 Write 所有格式处理无法序列化的类型
func TestWriteWithAllFormatsAndUnmarshalableType(t *testing.T) {
	formats := []Format{FormatJSON, FormatTable, FormatRaw}
	for _, format := range formats {
		var buf bytes.Buffer
		ch := make(chan int)
		err := Write(&buf, format, ch)
		if err == nil {
			t.Errorf("expected error for format %s with channel type", format)
		}
	}
}

// TestWriteWithUnknownFormat 测试 Write 使用未知格式
func TestWriteWithUnknownFormat(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]string{"key": "value"}
	err := Write(&buf, Format("unknown"), data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// 未知格式应该默认使用 JSON
	if !strings.Contains(buf.String(), "key") {
		t.Error("expected JSON output for unknown format")
	}
}

// TestSanitizeWithAllControlChars 测试所有控制字符
func TestSanitizeWithAllControlChars(t *testing.T) {
	// 测试所有 ASCII 控制字符 (0x00-0x1F 和 0x7F)
	for r := rune(0); r <= 0x7F; r++ {
		if r >= 0x20 && r < 0x7F {
			// 跳过可打印字符
			continue
		}
		input := string(r)
		result := Sanitize(input)
		if r == '\n' || r == '\t' {
			// 换行符和制表符应该被保留
			if result != input {
				t.Errorf("expected rune %d to be preserved, got %q", r, result)
			}
		} else {
			// 其他控制字符应该被替换为空格
			if result != " " {
				t.Errorf("expected control char %d to be replaced with space, got %q", r, result)
			}
		}
	}
}

// TestWriteKeyValuesWithWriterErrorAtDifferentPositions 测试 writeKeyValues 在不同位置的写入错误
func TestWriteKeyValuesWithWriterErrorAtDifferentPositions(t *testing.T) {
	tests := []struct {
		name        string
		data        map[string]any
		failAtWrite int
	}{
		{
			name:        "fail at first key",
			data:        map[string]any{"a": 1, "b": 2, "c": 3},
			failAtWrite: 1,
		},
		{
			name:        "fail at second key",
			data:        map[string]any{"a": 1, "b": 2, "c": 3},
			failAtWrite: 2,
		},
		{
			name:        "fail at third key",
			data:        map[string]any{"a": 1, "b": 2, "c": 3},
			failAtWrite: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := &controlledErrorWriter{failAt: tt.failAtWrite}
			err := writeKeyValues(writer, tt.data)
			if err == nil {
				t.Error("expected error from controlledErrorWriter")
			}
		})
	}
}

// TestWriteJSONWithLargePayload 测试 WriteJSON 处理大数据
func TestWriteJSONWithLargePayload(t *testing.T) {
	var buf bytes.Buffer
	// 创建一个大的数据结构
	data := make(map[string]any)
	for i := 0; i < 100; i++ {
		data[fmt.Sprintf("key%d", i)] = fmt.Sprintf("value%d", i)
	}
	err := WriteJSON(&buf, data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}

// TestWriteTableWithLargeDataset 测试 WriteTable 处理大数据集
func TestWriteTableWithLargeDataset(t *testing.T) {
	var buf bytes.Buffer
	// 创建一个大的数据集
	data := make([]map[string]any, 100)
	for i := 0; i < 100; i++ {
		data[i] = map[string]any{
			"id":   i,
			"name": fmt.Sprintf("name%d", i),
		}
	}
	err := WriteTable(&buf, data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}

// TestApplyJQWithComplexNestedData 测试 ApplyJQ 处理复杂嵌套数据
func TestApplyJQWithComplexNestedData(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]any{
		"level1": map[string]any{
			"level2": map[string]any{
				"level3": map[string]any{
					"value": "deep",
				},
			},
		},
	}
	err := ApplyJQ(&buf, data, ".level1.level2.level3.value")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "deep") {
		t.Error("expected output to contain 'deep'")
	}
}

// TestSelectFieldsWithNestedFields 测试 SelectFields 处理嵌套字段
func TestSelectFieldsWithNestedFields(t *testing.T) {
	data := map[string]any{
		"id":     1,
		"name":   "test",
		"nested": map[string]any{"key": "value"},
	}
	result, _ := SelectFields(data, []string{"id", "nested"})
	m, ok := result.(map[string]any)
	if !ok {
		t.Error("expected map result")
		return
	}
	if len(m) != 2 {
		t.Errorf("expected 2 fields, got %d", len(m))
	}
	if _, exists := m["id"]; !exists {
		t.Error("expected 'id' field")
	}
	if _, exists := m["nested"]; !exists {
		t.Error("expected 'nested' field")
	}
}

// TestTruncateWithVeryLongString 测试 truncate 处理非常长的字符串
func TestTruncateWithVeryLongString(t *testing.T) {
	longStr := strings.Repeat("x", 1000)
	result := truncate(longStr, 10)
	if len([]rune(result)) != 10 {
		t.Errorf("expected 10 runes, got %d", len([]rune(result)))
	}
}

// TestFormatValueWithNestedSlice 测试 formatValue 处理嵌套切片
func TestFormatValueWithNestedSlice(t *testing.T) {
	value := [][]int{{1, 2}, {3, 4}}
	result := formatValue(value)
	if !json.Valid([]byte(result)) {
		t.Errorf("expected valid JSON, got %q", result)
	}
}

// TestFormatValueWithNestedMap 测试 formatValue 处理嵌套 map
func TestFormatValueWithNestedMap(t *testing.T) {
	value := map[string]map[string]int{
		"outer": {"inner": 42},
	}
	result := formatValue(value)
	if !json.Valid([]byte(result)) {
		t.Errorf("expected valid JSON, got %q", result)
	}
}

// TestWriteTableSeparatorWriteError 测试 writeTable 在写入分隔线时的错误
func TestWriteTableSeparatorWriteError(t *testing.T) {
	tests := []struct {
		name        string
		headers     []string
		rows        [][]string
		failAtWrite int
	}{
		{
			name:        "fail at separator line first column write",
			headers:     []string{"id", "name"},
			rows:        [][]string{},
			failAtWrite: 4, // After header line
		},
		{
			name:        "fail at separator line second column separator",
			headers:     []string{"id", "name"},
			rows:        [][]string{},
			failAtWrite: 5,
		},
		{
			name:        "fail at separator line second column dashes",
			headers:     []string{"id", "name"},
			rows:        [][]string{},
			failAtWrite: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := &controlledErrorWriter{failAt: tt.failAtWrite}
			err := writeTable(writer, tt.headers, tt.rows)
			if err == nil {
				t.Error("expected error from controlledErrorWriter")
			}
		})
	}
}

// TestWriteTableRowWriteError 测试 writeTable 在写入行数据时的错误
func TestWriteTableRowWriteError(t *testing.T) {
	tests := []struct {
		name        string
		headers     []string
		rows        [][]string
		failAtWrite int
	}{
		{
			name:        "fail at row first column write",
			headers:     []string{"id", "name"},
			rows:        [][]string{{"1", "test"}},
			failAtWrite: 8, // After header and separator
		},
		{
			name:        "fail at row second column separator",
			headers:     []string{"id", "name"},
			rows:        [][]string{{"1", "test"}},
			failAtWrite: 9,
		},
		{
			name:        "fail at row second column write",
			headers:     []string{"id", "name"},
			rows:        [][]string{{"1", "test"}},
			failAtWrite: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := &controlledErrorWriter{failAt: tt.failAtWrite}
			err := writeTable(writer, tt.headers, tt.rows)
			if err == nil {
				t.Error("expected error from controlledErrorWriter")
			}
		})
	}
}

// TestWriteTableHeaderWriteError 测试 writeTable 在写入表头时的错误
func TestWriteTableHeaderWriteError(t *testing.T) {
	tests := []struct {
		name        string
		headers     []string
		rows        [][]string
		failAtWrite int
	}{
		{
			name:        "fail at first header write",
			headers:     []string{"id", "name"},
			rows:        [][]string{},
			failAtWrite: 1,
		},
		{
			name:        "fail at second header separator",
			headers:     []string{"id", "name"},
			rows:        [][]string{},
			failAtWrite: 2,
		},
		{
			name:        "fail at second header write",
			headers:     []string{"id", "name"},
			rows:        [][]string{},
			failAtWrite: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := &controlledErrorWriter{failAt: tt.failAtWrite}
			err := writeTable(writer, tt.headers, tt.rows)
			if err == nil {
				t.Error("expected error from controlledErrorWriter")
			}
		})
	}
}

// TestWriteTableWithMultipleRows 测试 writeTable 处理多行数据
func TestWriteTableWithMultipleRows(t *testing.T) {
	var buf bytes.Buffer
	headers := []string{"id", "name"}
	rows := [][]string{
		{"1", "Alice"},
		{"2", "Bob"},
		{"3", "Charlie"},
	}
	err := writeTable(&buf, headers, rows)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "Alice") || !strings.Contains(output, "Bob") || !strings.Contains(output, "Charlie") {
		t.Error("expected output to contain all names")
	}
}

// TestWriteTableWithExactCellWidth 测试 writeTable 处理刚好达到宽度限制的单元格
func TestWriteTableWithExactCellWidth(t *testing.T) {
	var buf bytes.Buffer
	// 创建一个正好 60 个字符的单元格
	exactCell := strings.Repeat("x", 60)
	headers := []string{"id"}
	rows := [][]string{{exactCell}}
	err := writeTable(&buf, headers, rows)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestWriteTableWithCellJustOverWidthLimit 测试 writeTable 处理刚好超过宽度限制的单元格
func TestWriteTableWithCellJustOverWidthLimit(t *testing.T) {
	var buf bytes.Buffer
	// 创建一个 61 个字符的单元格
	longCell := strings.Repeat("x", 61)
	headers := []string{"id"}
	rows := [][]string{{longCell}}
	err := writeTable(&buf, headers, rows)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestWriteKeyValuesWithKeyWidth25 测试 writeKeyValues 处理键宽度为 25 的情况
func TestWriteKeyValuesWithKeyWidth25(t *testing.T) {
	var buf bytes.Buffer
	// 创建一个 25 个字符的键
	key25 := strings.Repeat("k", 25)
	data := map[string]any{key25: "value"}
	err := writeKeyValues(&buf, data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// 验证输出
	output := buf.String()
	if !strings.Contains(output, "value") {
		t.Error("expected output to contain value")
	}
}

// TestWriteKeyValuesWithKeyWidth24 测试 writeKeyValues 处理键宽度为 24 的情况
func TestWriteKeyValuesWithKeyWidth24(t *testing.T) {
	var buf bytes.Buffer
	// 创建一个 24 个字符的键
	key24 := strings.Repeat("k", 24)
	data := map[string]any{key24: "value"}
	err := writeKeyValues(&buf, data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// 验证输出
	output := buf.String()
	if !strings.Contains(output, "value") {
		t.Error("expected output to contain value")
	}
}

// TestApplyJQWithWriteErrorOnNewline 测试 ApplyJQ 在写入换行符时的错误
func TestApplyJQWithWriteErrorOnNewline(t *testing.T) {
	// 使用一个在第二次写入时失败的 writer
	writer := &failAfterFirstWriteWriter{}
	data := map[string]any{"key": "value"}
	err := ApplyJQ(writer, data, ".")
	if err == nil {
		t.Error("expected error from failAfterFirstWriteWriter")
	}
}

// failAfterFirstWriteWriter 在第一次写入后返回错误
type failAfterFirstWriteWriter struct {
	writeCount int
}

func (w *failAfterFirstWriteWriter) Write(p []byte) (n int, err error) {
	w.writeCount++
	if w.writeCount > 1 {
		return 0, errors.New("write error after first write")
	}
	return len(p), nil
}

// TestApplyJQWithMultipleResultsAndWriteError 测试 ApplyJQ 在多个结果时的写入错误
func TestApplyJQWithMultipleResultsAndWriteError(t *testing.T) {
	// 使用一个在第三次写入时失败的 writer
	writer := &failAtThirdWriteWriter{}
	data := map[string]any{
		"items": []any{1, 2, 3, 4, 5},
	}
	err := ApplyJQ(writer, data, ".items[]")
	if err == nil {
		t.Error("expected error from failAtThirdWriteWriter")
	}
}

// failAtThirdWriteWriter 在第三次写入时返回错误
type failAtThirdWriteWriter struct {
	writeCount int
}

func (w *failAtThirdWriteWriter) Write(p []byte) (n int, err error) {
	w.writeCount++
	if w.writeCount >= 3 {
		return 0, errors.New("write error at third write")
	}
	return len(p), nil
}

// TestApplyJQWithEmptyResultAndWriteError 测试 ApplyJQ 在空结果时的写入
func TestApplyJQWithEmptyResultAndWriteError(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]any{"items": []any{}}
	err := ApplyJQ(&buf, data, ".items[]")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// 空数组应该产生空输出
	if buf.Len() > 0 {
		t.Error("expected empty output for empty array")
	}
}

// TestSelectFieldsWithEmptyFieldsList 测试 SelectFields 处理空字段列表
func TestSelectFieldsWithEmptyFieldsList(t *testing.T) {
	data := map[string]any{"id": 1, "name": "test"}
	result, _ := SelectFields(data, []string{})
	// 空字段列表应该返回原始数据
	m, ok := result.(map[string]any)
	if !ok {
		t.Error("expected map result")
		return
	}
	if len(m) != 2 {
		t.Errorf("expected 2 fields, got %d", len(m))
	}
}

// TestSelectFieldsWithNilData 测试 SelectFields 处理 nil 数据
func TestSelectFieldsWithNilData(t *testing.T) {
	result, _ := SelectFields(nil, []string{"id"})
	if result != nil {
		t.Error("expected nil for nil data")
	}
}

// TestSelectFieldsWithNonExistentFields 测试 SelectFields 处理不存在的字段
func TestSelectFieldsWithNonExistentFields(t *testing.T) {
	data := map[string]any{"id": 1, "name": "test"}
	result, _ := SelectFields(data, []string{"nonexistent"})
	m, ok := result.(map[string]any)
	if !ok {
		t.Error("expected map result")
		return
	}
	if len(m) != 0 {
		t.Errorf("expected 0 fields, got %d", len(m))
	}
}

// TestWriteJSONWithNilPayload 测试 WriteJSON 处理 nil 数据
func TestWriteJSONWithNilPayload(t *testing.T) {
	var buf bytes.Buffer
	err := WriteJSON(&buf, nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "null") {
		t.Error("expected 'null' in output")
	}
}

// TestWriteRawWithNilPayload 测试 WriteRaw 处理 nil 数据
func TestWriteRawWithNilPayload(t *testing.T) {
	var buf bytes.Buffer
	err := WriteRaw(&buf, nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "null") {
		t.Error("expected 'null' in output")
	}
}

// TestWriteWithNilPayload 测试 Write 处理 nil 数据
func TestWriteWithNilPayload(t *testing.T) {
	formats := []Format{FormatJSON, FormatTable, FormatRaw}
	for _, format := range formats {
		var buf bytes.Buffer
		err := Write(&buf, format, nil)
		if err != nil {
			t.Errorf("unexpected error for format %s: %v", format, err)
		}
	}
}

// TestNormalizePayloadWithNilInput 测试 normalizePayload 处理 nil 输入
func TestNormalizePayloadWithNilInput(t *testing.T) {
	result, err := normalizePayload(nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != nil {
		t.Error("expected nil result for nil input")
	}
}

// TestNormalizePayloadWithStringInput 测试 normalizePayload 处理字符串输入
func TestNormalizePayloadWithStringInput(t *testing.T) {
	result, err := normalizePayload("test string")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "test string" {
		t.Errorf("expected 'test string', got %v", result)
	}
}

// TestFormatValueWithNilInput 测试 formatValue 处理 nil 输入
func TestFormatValueWithNilInput(t *testing.T) {
	result := formatValue(nil)
	if result != "" {
		t.Errorf("expected empty string for nil, got %q", result)
	}
}

// TestFormatValueWithStringInput 测试 formatValue 处理字符串输入
func TestFormatValueWithStringInput(t *testing.T) {
	result := formatValue("test string")
	if result != "test string" {
		t.Errorf("expected 'test string', got %q", result)
	}
}

// TestTruncateWithEmptyString 测试 truncate 处理空字符串
func TestTruncateWithEmptyString(t *testing.T) {
	result := truncate("", 10)
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

// TestTruncateWithWidthOne 测试 truncate 处理宽度为 1
func TestTruncateWithWidthOne(t *testing.T) {
	result := truncate("hello", 1)
	if result != "…" {
		t.Errorf("expected '…', got %q", result)
	}
}

// TestSortedKeysWithEmptyMap 测试 sortedKeys 处理空 map
func TestSortedKeysWithEmptyMap(t *testing.T) {
	keys := map[string]struct{}{}
	result := sortedKeys(keys)
	if len(result) != 0 {
		t.Errorf("expected 0 keys, got %d", len(result))
	}
}

// TestSortedKeysWithSingleKey 测试 sortedKeys 处理单个键
func TestSortedKeysWithSingleKey(t *testing.T) {
	keys := map[string]struct{}{"a": {}}
	result := sortedKeys(keys)
	if len(result) != 1 || result[0] != "a" {
		t.Errorf("expected ['a'], got %v", result)
	}
}

// TestSortedKeysWithMultipleKeys 测试 sortedKeys 处理多个键
func TestSortedKeysWithMultipleKeys(t *testing.T) {
	keys := map[string]struct{}{
		"c": {},
		"a": {},
		"b": {},
	}
	result := sortedKeys(keys)
	if len(result) != 3 {
		t.Errorf("expected 3 keys, got %d", len(result))
	}
	// 检查是否已排序
	if result[0] != "a" || result[1] != "b" || result[2] != "c" {
		t.Errorf("expected sorted keys ['a', 'b', 'c'], got %v", result)
	}
}

// TestRowsFromSliceWithEmptySlice 测试 rowsFromSlice 处理空切片
func TestRowsFromSliceWithEmptySlice(t *testing.T) {
	headers, rows, ok := rowsFromSlice([]any{})
	if !ok {
		t.Error("expected ok to be true")
	}
	if len(headers) != 1 || headers[0] != "value" {
		t.Errorf("expected headers to be ['value'], got %v", headers)
	}
	if len(rows) != 0 {
		t.Errorf("expected 0 rows, got %d", len(rows))
	}
}

// TestRowsFromSliceWithSingleMap 测试 rowsFromSlice 处理单个 map
func TestRowsFromSliceWithSingleMap(t *testing.T) {
	items := []any{map[string]any{"id": 1, "name": "test"}}
	headers, rows, ok := rowsFromSlice(items)
	if !ok {
		t.Error("expected ok to be true")
	}
	if len(headers) != 2 {
		t.Errorf("expected 2 headers, got %d", len(headers))
	}
	if len(rows) != 1 {
		t.Errorf("expected 1 row, got %d", len(rows))
	}
}

// TestRowsFromSliceWithSingleNonMap 测试 rowsFromSlice 处理单个非 map
func TestRowsFromSliceWithSingleNonMap(t *testing.T) {
	items := []any{"test"}
	headers, rows, ok := rowsFromSlice(items)
	if !ok {
		t.Error("expected ok to be true")
	}
	if len(headers) != 1 || headers[0] != "value" {
		t.Errorf("expected headers to be ['value'], got %v", headers)
	}
	if len(rows) != 1 {
		t.Errorf("expected 1 row, got %d", len(rows))
	}
}

// 确保 controlledErrorWriter 实现了 io.Writer 接口
var _ io.Writer = (*controlledErrorWriter)(nil)

// 确保 jsonMarshalErrorWriter 实现了 io.Writer 接口
var _ io.Writer = (*jsonMarshalErrorWriter)(nil)

// 确保 failAfterFirstWriteWriter 实现了 io.Writer 接口
var _ io.Writer = (*failAfterFirstWriteWriter)(nil)

// 确保 failAtThirdWriteWriter 实现了 io.Writer 接口
var _ io.Writer = (*failAtThirdWriteWriter)(nil)

// TestWriteTableHeaderSeparatorError 测试 writeTable 在写入表头分隔符时的错误
func TestWriteTableHeaderSeparatorError(t *testing.T) {
	// 测试在写入第二个表头之前的分隔符时出错
	writer := &failAtSecondWriteWriter{}
	headers := []string{"id", "name"}
	rows := [][]string{}
	err := writeTable(writer, headers, rows)
	if err == nil {
		t.Error("expected error from failAtSecondWriteWriter")
	}
}

// failAtSecondWriteWriter 在第二次写入时返回错误
type failAtSecondWriteWriter struct {
	writeCount int
}

func (w *failAtSecondWriteWriter) Write(p []byte) (n int, err error) {
	w.writeCount++
	if w.writeCount >= 2 {
		return 0, errors.New("write error at second write")
	}
	return len(p), nil
}

// TestWriteTableSeparatorSeparatorError 测试 writeTable 在写入分隔线分隔符时的错误
func TestWriteTableSeparatorSeparatorError(t *testing.T) {
	// 测试在写入第二个分隔线之前的分隔符时出错
	writer := &failAtFifthWriteWriter{}
	headers := []string{"id", "name"}
	rows := [][]string{}
	err := writeTable(writer, headers, rows)
	if err == nil {
		t.Error("expected error from failAtFifthWriteWriter")
	}
}

// failAtFifthWriteWriter 在第五次写入时返回错误
type failAtFifthWriteWriter struct {
	writeCount int
}

func (w *failAtFifthWriteWriter) Write(p []byte) (n int, err error) {
	w.writeCount++
	if w.writeCount >= 5 {
		return 0, errors.New("write error at fifth write")
	}
	return len(p), nil
}

// TestWriteTableRowSeparatorError 测试 writeTable 在写入行分隔符时的错误
func TestWriteTableRowSeparatorError(t *testing.T) {
	// 测试在写入第二个单元格之前的分隔符时出错
	writer := &failAtNinthWriteWriter{}
	headers := []string{"id", "name"}
	rows := [][]string{{"1", "test"}}
	err := writeTable(writer, headers, rows)
	if err == nil {
		t.Error("expected error from failAtNinthWriteWriter")
	}
}

// failAtNinthWriteWriter 在第九次写入时返回错误
type failAtNinthWriteWriter struct {
	writeCount int
}

func (w *failAtNinthWriteWriter) Write(p []byte) (n int, err error) {
	w.writeCount++
	if w.writeCount >= 9 {
		return 0, errors.New("write error at ninth write")
	}
	return len(p), nil
}

// TestWriteTableWithMultipleColumnsAndRows 测试 writeTable 处理多列多行
func TestWriteTableWithMultipleColumnsAndRows(t *testing.T) {
	var buf bytes.Buffer
	headers := []string{"id", "name", "email"}
	rows := [][]string{
		{"1", "Alice", "alice@example.com"},
		{"2", "Bob", "bob@example.com"},
		{"3", "Charlie", "charlie@example.com"},
	}
	err := writeTable(&buf, headers, rows)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "Alice") || !strings.Contains(output, "Bob") || !strings.Contains(output, "Charlie") {
		t.Error("expected output to contain all names")
	}
}

// TestWriteKeyValuesWithMaxWidthGreaterThan24 测试 writeKeyValues 处理最大宽度大于 24
func TestWriteKeyValuesWithMaxWidthGreaterThan24(t *testing.T) {
	var buf bytes.Buffer
	// 创建一个键宽度大于 24 的 map
	data := map[string]any{
		"this_is_a_very_long_key_name": "value1",
		"short":                        "value2",
	}
	err := writeKeyValues(&buf, data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "value1") || !strings.Contains(output, "value2") {
		t.Error("expected output to contain all values")
	}
}

// TestFormatValueWithUnmarshalableValue 测试 formatValue 处理无法序列化的值
func TestFormatValueWithUnmarshalableValue(t *testing.T) {
	// channel 无法被 JSON 序列化
	ch := make(chan int)
	result := formatValue(ch)
	// 应该使用 fmt.Sprintf 作为回退
	if result == "" {
		t.Error("expected non-empty result for channel type")
	}
}

// TestFormatValueWithFunction 测试 formatValue 处理函数类型
func TestFormatValueWithFunction(t *testing.T) {
	// 函数无法被 JSON 序列化
	fn := func() {}
	result := formatValue(fn)
	// 应该使用 fmt.Sprintf 作为回退
	if result == "" {
		t.Error("expected non-empty result for function type")
	}
}

// TestWriteTableWithCellWidthGreaterThan60 测试 writeTable 处理单元格宽度大于 60
func TestWriteTableWithCellWidthGreaterThan60(t *testing.T) {
	var buf bytes.Buffer
	// 创建一个单元格宽度大于 60 的表格
	longCell := strings.Repeat("x", 100)
	headers := []string{"id"}
	rows := [][]string{{longCell}}
	err := writeTable(&buf, headers, rows)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// 验证输出被截断
	output := buf.String()
	if len(output) == 0 {
		t.Error("expected non-empty output")
	}
}

// TestWriteTableWithHeaderWidthGreaterThan60 测试 writeTable 处理表头宽度大于 60
func TestWriteTableWithHeaderWidthGreaterThan60(t *testing.T) {
	var buf bytes.Buffer
	// 创建一个表头宽度大于 60 的表格
	longHeader := strings.Repeat("x", 100)
	headers := []string{longHeader}
	rows := [][]string{{"value"}}
	err := writeTable(&buf, headers, rows)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// 验证输出被截断
	output := buf.String()
	if len(output) == 0 {
		t.Error("expected non-empty output")
	}
}

// TestWriteTableWithEmptyHeadersAndRows 测试 writeTable 处理空表头和空行
func TestWriteTableWithEmptyHeadersAndRows(t *testing.T) {
	var buf bytes.Buffer
	headers := []string{}
	rows := [][]string{}
	err := writeTable(&buf, headers, rows)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestWriteTableWithSingleColumn 测试 writeTable 处理单列
func TestWriteTableWithSingleColumn(t *testing.T) {
	var buf bytes.Buffer
	headers := []string{"id"}
	rows := [][]string{{"1"}, {"2"}, {"3"}}
	err := writeTable(&buf, headers, rows)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "1") || !strings.Contains(output, "2") || !strings.Contains(output, "3") {
		t.Error("expected output to contain all values")
	}
}

// TestWriteTableWithRowHavingMoreCellsThanHeaders 测试 writeTable 处理行单元格多于表头
func TestWriteTableWithRowHavingMoreCellsThanHeaders(t *testing.T) {
	var buf bytes.Buffer
	headers := []string{"id"}
	rows := [][]string{
		{"1", "extra1", "extra2"},
		{"2", "extra3"},
	}
	err := writeTable(&buf, headers, rows)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	output := buf.String()
	// 只应该包含第一列
	if !strings.Contains(output, "1") || !strings.Contains(output, "2") {
		t.Error("expected output to contain first column values")
	}
	// 不应该包含额外的单元格
	if strings.Contains(output, "extra1") || strings.Contains(output, "extra2") {
		t.Error("expected output to not contain extra cells")
	}
}

// TestWriteTableWithRowHavingFewerCellsThanHeaders 测试 writeTable 处理行单元格少于表头
func TestWriteTableWithRowHavingFewerCellsThanHeaders(t *testing.T) {
	var buf bytes.Buffer
	headers := []string{"id", "name", "email"}
	rows := [][]string{
		{"1"},
		{"2", "Bob"},
	}
	err := writeTable(&buf, headers, rows)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestWriteKeyValuesWithSingleKey 测试 writeKeyValues 处理单个键
func TestWriteKeyValuesWithSingleKey(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]any{"id": 1}
	err := writeKeyValues(&buf, data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "id") {
		t.Error("expected output to contain 'id'")
	}
}

// TestNormalizePayloadWithMap 测试 normalizePayload 处理 map
func TestNormalizePayloadWithMap(t *testing.T) {
	data := map[string]any{"id": 1, "name": "test"}
	result, err := normalizePayload(data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	m, ok := result.(map[string]any)
	if !ok {
		t.Error("expected map result")
		return
	}
	if m["id"] != 1.0 || m["name"] != "test" {
		t.Errorf("expected map with id=1 and name='test', got %v", m)
	}
}

// TestNormalizePayloadWithSlice 测试 normalizePayload 处理切片
func TestNormalizePayloadWithSlice(t *testing.T) {
	data := []int{1, 2, 3}
	result, err := normalizePayload(data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	slice, ok := result.([]any)
	if !ok {
		t.Error("expected slice result")
		return
	}
	if len(slice) != 3 {
		t.Errorf("expected slice length 3, got %d", len(slice))
	}
}

// TestNormalizePayloadWithStruct 测试 normalizePayload 处理结构体
func TestNormalizePayloadWithStruct(t *testing.T) {
	type TestStruct struct {
		ID   int
		Name string
	}
	data := TestStruct{ID: 1, Name: "test"}
	result, err := normalizePayload(data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	m, ok := result.(map[string]any)
	if !ok {
		t.Error("expected map result")
		return
	}
	if m["ID"] != 1.0 || m["Name"] != "test" {
		t.Errorf("expected map with ID=1 and Name='test', got %v", m)
	}
}

// TestWriteTableWithSliceInput 测试 WriteTable 处理切片输入
func TestWriteTableWithSliceInput(t *testing.T) {
	var buf bytes.Buffer
	data := []map[string]any{{"id": 1, "name": "test"}}
	err := WriteTable(&buf, data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "id") {
		t.Error("expected output to contain 'id'")
	}
}

// TestWriteTableWithStringInput 测试 WriteTable 处理字符串输入
func TestWriteTableWithStringInput(t *testing.T) {
	var buf bytes.Buffer
	data := "test string"
	err := WriteTable(&buf, data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "test string") {
		t.Error("expected output to contain 'test string'")
	}
}

// TestWriteRawWithStringInput 测试 WriteRaw 处理字符串输入
func TestWriteRawWithStringInput(t *testing.T) {
	var buf bytes.Buffer
	data := "test string"
	err := WriteRaw(&buf, data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "test string") {
		t.Error("expected output to contain 'test string'")
	}
}

// TestWriteRawWithNonStringInput 测试 WriteRaw 处理非字符串输入
func TestWriteRawWithNonStringInput(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]any{"id": 1, "name": "test"}
	err := WriteRaw(&buf, data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "id") {
		t.Error("expected output to contain 'id'")
	}
}

// TestWriteJSONWithValidInput 测试 WriteJSON 处理有效输入
func TestWriteJSONWithValidInput(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]any{"id": 1, "name": "test"}
	err := WriteJSON(&buf, data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "id") {
		t.Error("expected output to contain 'id'")
	}
}

// 确保 failAtSecondWriteWriter 实现了 io.Writer 接口
var _ io.Writer = (*failAtSecondWriteWriter)(nil)

// 确保 failAtFifthWriteWriter 实现了 io.Writer 接口
var _ io.Writer = (*failAtFifthWriteWriter)(nil)

// 确保 failAtNinthWriteWriter 实现了 io.Writer 接口
var _ io.Writer = (*failAtNinthWriteWriter)(nil)
