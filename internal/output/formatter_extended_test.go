// internal/output/formatter_extended_test.go
package output

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

// TestWrite 测试 Write 函数的所有格式
func TestWrite(t *testing.T) {
	tests := []struct {
		name    string
		format  Format
		payload any
		wantErr bool
		check   func(t *testing.T, output string)
	}{
		{
			name:    "json format",
			format:  FormatJSON,
			payload: map[string]string{"key": "value"},
			wantErr: false,
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, `"key"`) {
					t.Error("expected JSON output to contain key")
				}
			},
		},
		{
			name:    "table format with slice",
			format:  FormatTable,
			payload: []map[string]any{{"id": 1, "name": "test"}},
			wantErr: false,
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "id") || !strings.Contains(output, "name") {
					t.Error("expected table output to contain headers")
				}
			},
		},
		{
			name:    "table format with map",
			format:  FormatTable,
			payload: map[string]any{"id": 1, "name": "test"},
			wantErr: false,
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "id") || !strings.Contains(output, "name") {
					t.Error("expected table output to contain keys")
				}
			},
		},
		{
			name:    "raw format with string",
			format:  FormatRaw,
			payload: "plain text",
			wantErr: false,
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "plain text") {
					t.Error("expected raw output to contain text")
				}
			},
		},
		{
			name:    "raw format with object",
			format:  FormatRaw,
			payload: map[string]int{"count": 42},
			wantErr: false,
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "count") {
					t.Error("expected raw output to contain key")
				}
			},
		},
		{
			name:    "default format (unknown)",
			format:  Format("unknown"),
			payload: map[string]string{"key": "value"},
			wantErr: false,
			check: func(t *testing.T, output string) {
				// 默认应该使用 JSON 格式
				if !strings.Contains(output, `"key"`) {
					t.Error("expected default to use JSON format")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := Write(&buf, tt.format, tt.payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.check != nil {
				tt.check(t, buf.String())
			}
		})
	}
}

// TestWriteRaw 测试 WriteRaw 函数
func TestWriteRaw(t *testing.T) {
	tests := []struct {
		name    string
		payload any
		wantErr bool
		check   func(t *testing.T, output string)
	}{
		{
			name:    "string input",
			payload: "hello world",
			wantErr: false,
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "hello world") {
					t.Error("expected output to contain 'hello world'")
				}
			},
		},
		{
			name:    "string with control characters",
			payload: "hello\x00world",
			wantErr: false,
			check: func(t *testing.T, output string) {
				// 控制字符应该被清理
				if strings.Contains(output, "\x00") {
					t.Error("expected control characters to be sanitized")
				}
			},
		},
		{
			name:    "object input",
			payload: map[string]int{"count": 42},
			wantErr: false,
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "count") || !strings.Contains(output, "42") {
					t.Error("expected output to contain object data")
				}
			},
		},
		{
			name:    "slice input",
			payload: []string{"a", "b", "c"},
			wantErr: false,
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "a") || !strings.Contains(output, "b") {
					t.Error("expected output to contain slice data")
				}
			},
		},
		{
			name:    "nil input",
			payload: nil,
			wantErr: false,
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "null") {
					t.Error("expected output to contain 'null'")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := WriteRaw(&buf, tt.payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteRaw() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.check != nil {
				tt.check(t, buf.String())
			}
		})
	}
}

// TestNormalizePayload 测试 normalizePayload 函数
func TestNormalizePayload(t *testing.T) {
	tests := []struct {
		name    string
		payload any
		wantErr bool
		check   func(t *testing.T, result any)
	}{
		{
			name:    "nil payload",
			payload: nil,
			wantErr: false,
			check: func(t *testing.T, result any) {
				if result != nil {
					t.Error("expected nil result for nil payload")
				}
			},
		},
		{
			name:    "string payload",
			payload: "test string",
			wantErr: false,
			check: func(t *testing.T, result any) {
				if result != "test string" {
					t.Error("expected string to be returned as-is")
				}
			},
		},
		{
			name:    "map payload",
			payload: map[string]any{"key": "value"},
			wantErr: false,
			check: func(t *testing.T, result any) {
				m, ok := result.(map[string]any)
				if !ok {
					t.Error("expected map result")
					return
				}
				if m["key"] != "value" {
					t.Error("expected key to have value")
				}
			},
		},
		{
			name:    "slice payload",
			payload: []int{1, 2, 3},
			wantErr: false,
			check: func(t *testing.T, result any) {
				slice, ok := result.([]any)
				if !ok {
					t.Error("expected slice result")
					return
				}
				if len(slice) != 3 {
					t.Errorf("expected slice length 3, got %d", len(slice))
				}
			},
		},
		{
			name:    "struct payload",
			payload: struct{ Name string }{Name: "test"},
			wantErr: false,
			check: func(t *testing.T, result any) {
				m, ok := result.(map[string]any)
				if !ok {
					t.Error("expected map result from struct")
					return
				}
				if m["Name"] != "test" {
					t.Error("expected Name field to be 'test'")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := normalizePayload(tt.payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("normalizePayload() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

// TestWriteKeyValues 测试 writeKeyValues 函数
func TestWriteKeyValues(t *testing.T) {
	tests := []struct {
		name    string
		payload map[string]any
		wantErr bool
		check   func(t *testing.T, output string)
	}{
		{
			name:    "simple map",
			payload: map[string]any{"id": 1, "name": "test"},
			wantErr: false,
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "id") || !strings.Contains(output, "name") {
					t.Error("expected output to contain keys")
				}
				if !strings.Contains(output, "1") || !strings.Contains(output, "test") {
					t.Error("expected output to contain values")
				}
			},
		},
		{
			name:    "map with long key",
			payload: map[string]any{"very_long_key_name_that_exceeds_limit": "value"},
			wantErr: false,
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "value") {
					t.Error("expected output to contain value")
				}
			},
		},
		{
			name:    "map with nested value",
			payload: map[string]any{"nested": map[string]string{"key": "value"}},
			wantErr: false,
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "nested") {
					t.Error("expected output to contain nested key")
				}
			},
		},
		{
			name:    "map with nil value",
			payload: map[string]any{"nil_value": nil},
			wantErr: false,
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "nil_value") {
					t.Error("expected output to contain nil_value key")
				}
			},
		},
		{
			name:    "map with slice value",
			payload: map[string]any{"items": []int{1, 2, 3}},
			wantErr: false,
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "items") {
					t.Error("expected output to contain items key")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := writeKeyValues(&buf, tt.payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("writeKeyValues() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.check != nil {
				tt.check(t, buf.String())
			}
		})
	}
}

// TestWriteTableExtended 测试 WriteTable 函数的更多场景
func TestWriteTableExtended(t *testing.T) {
	tests := []struct {
		name    string
		payload any
		wantErr bool
		check   func(t *testing.T, output string)
	}{
		{
			name:    "empty slice",
			payload: []map[string]any{},
			wantErr: false,
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "value") {
					t.Error("expected header 'value' for empty slice")
				}
			},
		},
		{
			name:    "slice of non-maps",
			payload: []any{1, 2, 3},
			wantErr: false,
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "value") {
					t.Error("expected header 'value' for non-map slice")
				}
			},
		},
		{
			name:    "slice of strings",
			payload: []any{"a", "b", "c"},
			wantErr: false,
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "a") || !strings.Contains(output, "b") {
					t.Error("expected output to contain string values")
				}
			},
		},
		{
			name:    "single map",
			payload: map[string]any{"id": 1, "name": "test"},
			wantErr: false,
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "id") || !strings.Contains(output, "name") {
					t.Error("expected output to contain keys")
				}
			},
		},
		{
			name:    "string payload",
			payload: "plain string",
			wantErr: false,
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "plain string") {
					t.Error("expected output to contain string")
				}
			},
		},
		{
			name:    "nil payload",
			payload: nil,
			wantErr: false,
			check: func(_ *testing.T, _ string) {
				// nil 应该被处理
			},
		},
		{
			name:    "slice with different keys",
			payload: []map[string]any{{"id": 1}, {"id": 2, "name": "test"}},
			wantErr: false,
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "id") || !strings.Contains(output, "name") {
					t.Error("expected output to contain all keys")
				}
			},
		},
		{
			name:    "slice with long values",
			payload: []map[string]any{{"description": strings.Repeat("x", 100)}},
			wantErr: false,
			check: func(t *testing.T, output string) {
				// 长值应该被截断
				if len(output) > 200 {
					t.Error("expected long values to be truncated")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := WriteTable(&buf, tt.payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteTable() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.check != nil {
				tt.check(t, buf.String())
			}
		})
	}
}

// TestRowsFromSlice 测试 rowsFromSlice 函数
func TestRowsFromSlice(t *testing.T) {
	tests := []struct {
		name        string
		items       []any
		wantHeaders []string
		wantRows    int
	}{
		{
			name:        "empty slice",
			items:       []any{},
			wantHeaders: []string{"value"},
			wantRows:    0,
		},
		{
			name:        "slice of maps",
			items:       []any{map[string]any{"id": 1, "name": "a"}, map[string]any{"id": 2, "name": "b"}},
			wantHeaders: []string{"id", "name"},
			wantRows:    2,
		},
		{
			name:        "slice of non-maps",
			items:       []any{1, 2, 3},
			wantHeaders: []string{"value"},
			wantRows:    3,
		},
		{
			name:        "slice of strings",
			items:       []any{"a", "b", "c"},
			wantHeaders: []string{"value"},
			wantRows:    3,
		},
		{
			name:        "slice with different keys",
			items:       []any{map[string]any{"id": 1}, map[string]any{"id": 2, "name": "test"}},
			wantHeaders: []string{"id", "name"},
			wantRows:    2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers, rows, ok := rowsFromSlice(tt.items)
			if !ok {
				t.Error("expected ok to be true")
				return
			}
			if len(headers) != len(tt.wantHeaders) {
				t.Errorf("expected %d headers, got %d", len(tt.wantHeaders), len(headers))
			}
			if len(rows) != tt.wantRows {
				t.Errorf("expected %d rows, got %d", tt.wantRows, len(rows))
			}
		})
	}
}

// TestTruncate 测试 truncate 函数
func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxWidth int
		want     string
	}{
		{
			name:     "empty string",
			input:    "",
			maxWidth: 10,
			want:     "",
		},
		{
			name:     "short string",
			input:    "hello",
			maxWidth: 10,
			want:     "hello",
		},
		{
			name:     "exact length",
			input:    "hello",
			maxWidth: 5,
			want:     "hello",
		},
		{
			name:     "needs truncation",
			input:    "hello world",
			maxWidth: 5,
			want:     "hell…",
		},
		{
			name:     "maxWidth 0",
			input:    "hello",
			maxWidth: 0,
			want:     "",
		},
		{
			name:     "maxWidth 1",
			input:    "hello",
			maxWidth: 1,
			want:     "…",
		},
		{
			name:     "maxWidth negative",
			input:    "hello",
			maxWidth: -1,
			want:     "",
		},
		{
			name:     "unicode string",
			input:    "你好世界",
			maxWidth: 3,
			want:     "你好…",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncate(tt.input, tt.maxWidth)
			if got != tt.want {
				t.Errorf("truncate() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestFormatValue 测试 formatValue 函数
func TestFormatValue(t *testing.T) {
	tests := []struct {
		name  string
		value any
		check func(t *testing.T, result string)
	}{
		{
			name:  "nil value",
			value: nil,
			check: func(t *testing.T, result string) {
				if result != "" {
					t.Errorf("expected empty string for nil, got %q", result)
				}
			},
		},
		{
			name:  "string value",
			value: "hello",
			check: func(t *testing.T, result string) {
				if result != "hello" {
					t.Errorf("expected 'hello', got %q", result)
				}
			},
		},
		{
			name:  "string with control characters",
			value: "hello\x00world",
			check: func(t *testing.T, result string) {
				if strings.Contains(result, "\x00") {
					t.Error("expected control characters to be sanitized")
				}
			},
		},
		{
			name:  "int value",
			value: 42,
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "42") {
					t.Errorf("expected '42', got %q", result)
				}
			},
		},
		{
			name:  "bool value",
			value: true,
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "true") {
					t.Errorf("expected 'true', got %q", result)
				}
			},
		},
		{
			name:  "slice value",
			value: []int{1, 2, 3},
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "[") || !strings.Contains(result, "1") {
					t.Errorf("expected JSON array, got %q", result)
				}
			},
		},
		{
			name:  "map value",
			value: map[string]string{"key": "value"},
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "key") || !strings.Contains(result, "value") {
					t.Errorf("expected JSON object, got %q", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatValue(tt.value)
			tt.check(t, result)
		})
	}
}

// TestWriteTableWithWriterError 测试写入器错误情况
func TestWriteTableWithWriterError(t *testing.T) {
	// 创建一个总是返回错误的写入器
	errWriter := &errorWriter{}
	data := []map[string]any{{"id": 1, "name": "test"}}
	err := WriteTable(errWriter, data)
	if err == nil {
		t.Error("expected error from errorWriter")
	}
}

// TestWriteKeyValuesWithWriterError 测试写入器错误情况
func TestWriteKeyValuesWithWriterError(t *testing.T) {
	errWriter := &errorWriter{}
	data := map[string]any{"id": 1, "name": "test"}
	err := writeKeyValues(errWriter, data)
	if err == nil {
		t.Error("expected error from errorWriter")
	}
}

// errorWriter 是一个总是返回错误的写入器
type errorWriter struct{}

func (w *errorWriter) Write(_ []byte) (n int, err error) {
	return 0, errors.New("write error")
}

// TestWriteJSONWithWriterError 测试写入器错误情况
func TestWriteJSONWithWriterError(t *testing.T) {
	errWriter := &errorWriter{}
	data := map[string]string{"key": "value"}
	err := WriteJSON(errWriter, data)
	if err == nil {
		t.Error("expected error from errorWriter")
	}
}

// TestWriteRawWithWriterError 测试写入器错误情况
func TestWriteRawWithWriterError(t *testing.T) {
	errWriter := &errorWriter{}
	data := "test string"
	err := WriteRaw(errWriter, data)
	if err == nil {
		t.Error("expected error from errorWriter")
	}
}

// TestWriteWithWriterError 测试写入器错误情况
func TestWriteWithWriterError(t *testing.T) {
	errWriter := &errorWriter{}
	data := map[string]string{"key": "value"}
	err := Write(errWriter, FormatJSON, data)
	if err == nil {
		t.Error("expected error from errorWriter")
	}
}

// TestWriteTableWithLongHeaders 测试长表头的情况
func TestWriteTableWithLongHeaders(t *testing.T) {
	var buf bytes.Buffer
	data := []map[string]any{
		{"very_long_header_name_that_should_be_truncated": "value1"},
		{"very_long_header_name_that_should_be_truncated": "value2"},
	}
	err := WriteTable(&buf, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "very_long_header_name") {
		t.Error("expected output to contain header")
	}
}

// TestWriteTableWithLongValues 测试长值的情况
func TestWriteTableWithLongValues(t *testing.T) {
	var buf bytes.Buffer
	data := []map[string]any{
		{"description": strings.Repeat("x", 100)},
	}
	err := WriteTable(&buf, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()
	// 长值应该被截断到 60 个字符
	if len(output) > 200 {
		t.Error("expected long values to be truncated")
	}
}

// TestWriteKeyValuesWithLongKey 测试长键的情况
func TestWriteKeyValuesWithLongKey(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]any{
		"very_long_key_name_that_exceeds_the_maximum_width_limit": "value",
	}
	err := writeKeyValues(&buf, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "value") {
		t.Error("expected output to contain value")
	}
}

// TestSortedKeys 测试 sortedKeys 函数
func TestSortedKeys(t *testing.T) {
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
	for i := 1; i < len(result); i++ {
		if result[i-1] > result[i] {
			t.Error("expected keys to be sorted")
		}
	}
}

// TestWriteTableWithEmptyMap 测试空 map 的情况
func TestWriteTableWithEmptyMap(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]any{}
	err := WriteTable(&buf, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 空 map 应该产生空输出
	if buf.Len() > 0 {
		t.Error("expected empty output for empty map")
	}
}

// TestWriteTableWithMixedTypes 测试混合类型的情况
func TestWriteTableWithMixedTypes(t *testing.T) {
	var buf bytes.Buffer
	data := []map[string]any{
		{"id": 1, "name": "test", "active": true, "count": 42.5},
	}
	err := WriteTable(&buf, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "id") || !strings.Contains(output, "name") {
		t.Error("expected output to contain headers")
	}
}

// 确保 errorWriter 实现了 io.Writer 接口
var _ io.Writer = (*errorWriter)(nil)
