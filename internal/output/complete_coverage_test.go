// internal/output/complete_coverage_test.go
package output

import (
	"bytes"
	"testing"
)

// TestWriteTableAllBranches 测试 writeTable 的所有分支
func TestWriteTableAllBranches(t *testing.T) {
	tests := []struct {
		name    string
		headers []string
		rows    [][]string
		wantErr bool
	}{
		{
			name:    "empty headers and rows",
			headers: []string{},
			rows:    [][]string{},
			wantErr: false,
		},
		{
			name:    "single header single row",
			headers: []string{"id"},
			rows:    [][]string{{"1"}},
			wantErr: false,
		},
		{
			name:    "multiple headers multiple rows",
			headers: []string{"id", "name", "email"},
			rows: [][]string{
				{"1", "Alice", "alice@example.com"},
				{"2", "Bob", "bob@example.com"},
			},
			wantErr: false,
		},
		{
			name:    "rows with more cells than headers",
			headers: []string{"id", "name"},
			rows: [][]string{
				{"1", "Alice", "extra1", "extra2"},
			},
			wantErr: false,
		},
		{
			name:    "rows with fewer cells than headers",
			headers: []string{"id", "name", "email"},
			rows: [][]string{
				{"1", "Alice"},
			},
			wantErr: false,
		},
		{
			name:    "very long header",
			headers: []string{"this_is_a_very_long_header_name_that_exceeds_sixty_characters_limit_for_testing"},
			rows:    [][]string{{"value"}},
			wantErr: false,
		},
		{
			name:    "very long cell value",
			headers: []string{"id"},
			rows: [][]string{
				{"this_is_a_very_long_cell_value_that_exceeds_sixty_characters_limit_for_testing"},
			},
			wantErr: false,
		},
		{
			name:    "empty cells",
			headers: []string{"id", "name", "email"},
			rows: [][]string{
				{"", "", ""},
			},
			wantErr: false,
		},
		{
			name:    "unicode headers and cells",
			headers: []string{"编号", "姓名", "邮箱"},
			rows: [][]string{
				{"1", "张三", "zhangsan@example.com"},
				{"2", "李四", "lisi@example.com"},
			},
			wantErr: false,
		},
		{
			name:    "emoji headers and cells",
			headers: []string{"😀", "🎉", "🚀"},
			rows: [][]string{
				{"happy", "celebrate", "launch"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := writeTable(&buf, tt.headers, tt.rows)
			if (err != nil) != tt.wantErr {
				t.Errorf("writeTable() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestWriteKeyValuesAllBranches 测试 writeKeyValues 的所有分支
func TestWriteKeyValuesAllBranches(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string]any
		wantErr bool
	}{
		{
			name:    "empty map",
			data:    map[string]any{},
			wantErr: false,
		},
		{
			name:    "single key",
			data:    map[string]any{"id": 1},
			wantErr: false,
		},
		{
			name:    "multiple keys",
			data:    map[string]any{"id": 1, "name": "test", "active": true},
			wantErr: false,
		},
		{
			name:    "very long key",
			data:    map[string]any{"this_is_a_very_long_key_name_that_exceeds_twenty_four_characters": "value"},
			wantErr: false,
		},
		{
			name:    "nil value",
			data:    map[string]any{"nil_value": nil},
			wantErr: false,
		},
		{
			name:    "nested map value",
			data:    map[string]any{"nested": map[string]any{"key": "value"}},
			wantErr: false,
		},
		{
			name:    "slice value",
			data:    map[string]any{"items": []int{1, 2, 3}},
			wantErr: false,
		},
		{
			name:    "unicode keys",
			data:    map[string]any{"编号": 1, "姓名": "测试"},
			wantErr: false,
		},
		{
			name:    "emoji keys",
			data:    map[string]any{"😀": "happy", "🎉": "celebrate"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := writeKeyValues(&buf, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("writeKeyValues() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestRowsFromSliceAllBranches 测试 rowsFromSlice 的所有分支
func TestRowsFromSliceAllBranches(t *testing.T) {
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
			name:        "slice of maps with same keys",
			items:       []any{map[string]any{"id": 1}, map[string]any{"id": 2}},
			wantHeaders: []string{"id"},
			wantRows:    2,
		},
		{
			name:        "slice of maps with different keys",
			items:       []any{map[string]any{"id": 1}, map[string]any{"id": 2, "name": "test"}},
			wantHeaders: []string{"id", "name"},
			wantRows:    2,
		},
		{
			name:        "slice of non-maps",
			items:       []any{1, "test", true},
			wantHeaders: []string{"value"},
			wantRows:    3,
		},
		{
			name:        "slice with nil",
			items:       []any{map[string]any{"id": 1}, nil, map[string]any{"id": 2}},
			wantHeaders: []string{"value"},
			wantRows:    3,
		},
		{
			name:        "slice of empty maps",
			items:       []any{map[string]any{}, map[string]any{}},
			wantHeaders: []string{},
			wantRows:    2,
		},
		{
			name:        "slice of maps with nested values",
			items:       []any{map[string]any{"id": 1, "nested": map[string]any{"key": "value"}}},
			wantHeaders: []string{"id", "nested"},
			wantRows:    1,
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

// TestTruncateAllBranches 测试 truncate 的所有分支
func TestTruncateAllBranches(t *testing.T) {
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
			name:     "maxWidth 2",
			input:    "hello",
			maxWidth: 2,
			want:     "h…",
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
		{
			name:     "emoji string",
			input:    "😀🎉🚀🌟💫",
			maxWidth: 3,
			want:     "😀🎉…",
		},
		{
			name:     "mixed unicode and ascii",
			input:    "hello世界",
			maxWidth: 6,
			want:     "hello…",
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

// TestFormatValueAllBranches 测试 formatValue 的所有分支
func TestFormatValueAllBranches(t *testing.T) {
	tests := []struct {
		name  string
		value any
		check func(t *testing.T, result string)
	}{
		{
			name:  "nil",
			value: nil,
			check: func(t *testing.T, result string) {
				if result != "" {
					t.Errorf("expected empty string for nil, got %q", result)
				}
			},
		},
		{
			name:  "string",
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
				if result != "hello world" {
					t.Errorf("expected 'hello world', got %q", result)
				}
			},
		},
		{
			name:  "int",
			value: 42,
			check: func(t *testing.T, result string) {
				if result != "42" {
					t.Errorf("expected '42', got %q", result)
				}
			},
		},
		{
			name:  "float",
			value: 3.14,
			check: func(t *testing.T, result string) {
				if len(result) == 0 {
					t.Error("expected non-empty result")
				}
			},
		},
		{
			name:  "bool true",
			value: true,
			check: func(t *testing.T, result string) {
				if result != "true" {
					t.Errorf("expected 'true', got %q", result)
				}
			},
		},
		{
			name:  "bool false",
			value: false,
			check: func(t *testing.T, result string) {
				if result != "false" {
					t.Errorf("expected 'false', got %q", result)
				}
			},
		},
		{
			name:  "slice",
			value: []int{1, 2, 3},
			check: func(t *testing.T, result string) {
				if result != "[1,2,3]" {
					t.Errorf("expected '[1,2,3]', got %q", result)
				}
			},
		},
		{
			name:  "map",
			value: map[string]string{"key": "value"},
			check: func(t *testing.T, result string) {
				if result != `{"key":"value"}` {
					t.Errorf("expected '{\"key\":\"value\"}', got %q", result)
				}
			},
		},
		{
			name:  "nil pointer",
			value: (*int)(nil),
			check: func(t *testing.T, result string) {
				if result != "null" {
					t.Errorf("expected 'null', got %q", result)
				}
			},
		},
		{
			name:  "non-nil pointer",
			value: func() any { i := 42; return &i }(),
			check: func(t *testing.T, result string) {
				if result != "42" {
					t.Errorf("expected '42', got %q", result)
				}
			},
		},
		{
			name:  "struct",
			value: struct{ Name string }{Name: "test"},
			check: func(t *testing.T, result string) {
				if result != `{"Name":"test"}` {
					t.Errorf("expected '{\"Name\":\"test\"}', got %q", result)
				}
			},
		},
		{
			name:  "complex",
			value: complex(1, 2),
			check: func(t *testing.T, result string) {
				if len(result) == 0 {
					t.Error("expected non-empty result")
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

// TestWriteAllFormats 测试 Write 的所有格式
func TestWriteAllFormats(t *testing.T) {
	data := map[string]any{"id": 1, "name": "test"}

	tests := []struct {
		name   string
		format Format
	}{
		{name: "json", format: FormatJSON},
		{name: "table", format: FormatTable},
		{name: "raw", format: FormatRaw},
		{name: "unknown", format: Format("unknown")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := Write(&buf, tt.format, data)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if buf.Len() == 0 {
				t.Error("expected non-empty output")
			}
		})
	}
}

// TestNormalizePayloadAllTypes 测试 normalizePayload 的所有类型
func TestNormalizePayloadAllTypes(t *testing.T) {
	tests := []struct {
		name    string
		payload any
		wantErr bool
	}{
		{name: "nil", payload: nil, wantErr: false},
		{name: "string", payload: "test", wantErr: false},
		{name: "int", payload: 42, wantErr: false},
		{name: "float", payload: 3.14, wantErr: false},
		{name: "bool", payload: true, wantErr: false},
		{name: "slice", payload: []int{1, 2, 3}, wantErr: false},
		{name: "map", payload: map[string]string{"key": "value"}, wantErr: false},
		{name: "struct", payload: struct{ Name string }{Name: "test"}, wantErr: false},
		{name: "channel", payload: make(chan int), wantErr: true},
		{name: "function", payload: func() {}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := normalizePayload(tt.payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("normalizePayload() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestWriteRawAllTypes 测试 WriteRaw 的所有类型
func TestWriteRawAllTypes(t *testing.T) {
	tests := []struct {
		name    string
		payload any
		wantErr bool
	}{
		{name: "string", payload: "test", wantErr: false},
		{name: "int", payload: 42, wantErr: false},
		{name: "map", payload: map[string]string{"key": "value"}, wantErr: false},
		{name: "nil", payload: nil, wantErr: false},
		{name: "channel", payload: make(chan int), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := WriteRaw(&buf, tt.payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteRaw() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestWriteJSONAllTypes 测试 WriteJSON 的所有类型
func TestWriteJSONAllTypes(t *testing.T) {
	tests := []struct {
		name    string
		payload any
		wantErr bool
	}{
		{name: "map", payload: map[string]string{"key": "value"}, wantErr: false},
		{name: "slice", payload: []int{1, 2, 3}, wantErr: false},
		{name: "nil", payload: nil, wantErr: false},
		{name: "channel", payload: make(chan int), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := WriteJSON(&buf, tt.payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestWriteTableAllTypes 测试 WriteTable 的所有类型
func TestWriteTableAllTypes(t *testing.T) {
	tests := []struct {
		name    string
		payload any
		wantErr bool
	}{
		{name: "slice of maps", payload: []map[string]any{{"id": 1}}, wantErr: false},
		{name: "single map", payload: map[string]any{"id": 1}, wantErr: false},
		{name: "string", payload: "test", wantErr: false},
		{name: "nil", payload: nil, wantErr: false},
		{name: "channel", payload: make(chan int), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := WriteTable(&buf, tt.payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteTable() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestApplyJQAllTypes 测试 ApplyJQ 的所有类型
func TestApplyJQAllTypes(t *testing.T) {
	tests := []struct {
		name    string
		payload any
		expr    string
		wantErr bool
	}{
		{name: "map", payload: map[string]any{"key": "value"}, expr: ".key", wantErr: false},
		{name: "slice", payload: []int{1, 2, 3}, expr: ".[]", wantErr: false},
		{name: "nil", payload: nil, expr: ".", wantErr: false},
		{name: "invalid query", payload: map[string]any{"key": "value"}, expr: ".[invalid", wantErr: true},
		{name: "channel", payload: make(chan int), expr: ".", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := ApplyJQ(&buf, tt.payload, tt.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ApplyJQ() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestSelectFieldsAllTypes 测试 SelectFields 的所有类型
func TestSelectFieldsAllTypes(t *testing.T) {
	tests := []struct {
		name    string
		data    any
		fields  []string
		check   func(t *testing.T, result any)
		wantErr bool
	}{
		{
			name:   "map with fields",
			data:   map[string]any{"id": 1, "name": "test"},
			fields: []string{"id"},
			check: func(t *testing.T, result any) {
				m, ok := result.(map[string]any)
				if !ok {
					t.Error("expected map result")
					return
				}
				if len(m) != 1 {
					t.Errorf("expected 1 field, got %d", len(m))
				}
			},
		},
		{
			name:   "map with empty fields",
			data:   map[string]any{"id": 1, "name": "test"},
			fields: []string{},
			check: func(t *testing.T, result any) {
				m, ok := result.(map[string]any)
				if !ok {
					t.Error("expected map result")
					return
				}
				if len(m) != 2 {
					t.Errorf("expected 2 fields, got %d", len(m))
				}
			},
		},
		{
			name:   "non-map",
			data:   []int{1, 2, 3},
			fields: []string{"id"},
			check: func(t *testing.T, result any) {
				slice, ok := result.([]int)
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
			name:   "nil",
			data:   nil,
			fields: []string{"id"},
			check: func(t *testing.T, result any) {
				if result != nil {
					t.Error("expected nil result")
				}
			},
		},
		{
			name:   "channel",
			data:   make(chan int),
			fields: []string{"id"},
			check: func(t *testing.T, result any) {
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SelectFields(tt.data, tt.fields)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			tt.check(t, result)
		})
	}
}
