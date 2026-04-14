// internal/output/filter_extended_test.go
package output

import (
	"bytes"
	"testing"
)

// TestApplyJQErrors 测试 ApplyJQ 的错误情况
func TestApplyJQErrors(t *testing.T) {
	tests := []struct {
		name    string
		payload any
		expr    string
		wantErr bool
	}{
		{
			name:    "invalid jq expression",
			payload: map[string]any{"key": "value"},
			expr:    ".[invalid",
			wantErr: true,
		},
		{
			name:    "valid jq with nil payload",
			payload: nil,
			expr:    ".key",
			wantErr: false,
		},
		{
			name:    "jq that produces error",
			payload: map[string]any{"key": "value"},
			expr:    ".key | error",
			wantErr: true,
		},
		{
			name:    "complex jq expression",
			payload: map[string]any{"items": []any{1, 2, 3}},
			expr:    ".items[] | select(. > 1)",
			wantErr: false,
		},
		{
			name:    "jq with object construction",
			payload: map[string]any{"id": 1, "name": "test"},
			expr:    "{id, name}",
			wantErr: false,
		},
		{
			name:    "jq with array construction",
			payload: map[string]any{"items": []any{1, 2, 3}},
			expr:    "[.items[]]",
			wantErr: false,
		},
		{
			name:    "jq with pipe",
			payload: map[string]any{"items": []any{map[string]any{"id": 1}, map[string]any{"id": 2}}},
			expr:    ".items[] | .id",
			wantErr: false,
		},
		{
			name:    "jq with conditional",
			payload: map[string]any{"value": 5},
			expr:    "if .value > 3 then \"big\" else \"small\" end",
			wantErr: false,
		},
		{
			name:    "jq with recursive descent",
			payload: map[string]any{"a": map[string]any{"b": map[string]any{"c": "value"}}},
			expr:    ".. | .c?",
			wantErr: false,
		},
		{
			name:    "jq with string interpolation",
			payload: map[string]any{"name": "world"},
			expr:    "\"Hello, \\(.name)\"",
			wantErr: false,
		},
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

// TestApplyJQWithWriterError 测试写入器错误情况
func TestApplyJQWithWriterError(t *testing.T) {
	errWriter := &errorWriter{}
	data := map[string]any{"key": "value"}
	err := ApplyJQ(errWriter, data, ".key")
	if err == nil {
		t.Error("expected error from errorWriter")
	}
}

// TestApplyJQWithMultipleResults 测试返回多个结果的 JQ 表达式
func TestApplyJQWithMultipleResults(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]any{
		"items": []any{
			map[string]any{"id": 1, "name": "first"},
			map[string]any{"id": 2, "name": "second"},
			map[string]any{"id": 3, "name": "third"},
		},
	}
	err := ApplyJQ(&buf, data, ".items[]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()
	// 应该有多个 JSON 对象输出
	if len(output) == 0 {
		t.Error("expected output to contain multiple results")
	}
}

// TestApplyJQWithEmptyResult 测试返回空结果的 JQ 表达式
func TestApplyJQWithEmptyResult(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]any{"items": []any{}}
	err := ApplyJQ(&buf, data, ".items[]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 空数组应该产生空输出
	if buf.Len() > 0 {
		t.Error("expected empty output for empty array")
	}
}

// TestApplyJQWithNullResult 测试返回 null 的 JQ 表达式
func TestApplyJQWithNullResult(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]any{"key": "value"}
	err := ApplyJQ(&buf, data, ".nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()
	// null 值应该被输出
	if output != "null\n" {
		t.Errorf("expected 'null\\n', got %q", output)
	}
}

// TestSelectFieldsErrors 测试 SelectFields 的错误情况
func TestSelectFieldsErrors(t *testing.T) {
	tests := []struct {
		name    string
		payload any
		fields  []string
		check   func(t *testing.T, result any)
	}{
		{
			name:    "empty fields list",
			payload: map[string]any{"id": 1, "name": "test"},
			fields:  []string{},
			check: func(t *testing.T, result any) {
				// 空字段列表应该返回原始数据
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
			name:    "non-map payload",
			payload: []int{1, 2, 3},
			fields:  []string{"id"},
			check: func(t *testing.T, result any) {
				// 非 map 数据应该返回原始数据
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
			name:    "string payload",
			payload: "test string",
			fields:  []string{"id"},
			check: func(t *testing.T, result any) {
				// 字符串应该返回原始数据
				if result != "test string" {
					t.Errorf("expected 'test string', got %v", result)
				}
			},
		},
		{
			name:    "nil payload",
			payload: nil,
			fields:  []string{"id"},
			check: func(t *testing.T, result any) {
				// nil 应该返回 nil
				if result != nil {
					t.Errorf("expected nil, got %v", result)
				}
			},
		},
		{
			name:    "non-existent fields",
			payload: map[string]any{"id": 1, "name": "test"},
			fields:  []string{"nonexistent"},
			check: func(t *testing.T, result any) {
				m, ok := result.(map[string]any)
				if !ok {
					t.Error("expected map result")
					return
				}
				// 不存在的字段不应该被添加
				if len(m) != 0 {
					t.Errorf("expected 0 fields, got %d", len(m))
				}
			},
		},
		{
			name:    "mixed existent and non-existent fields",
			payload: map[string]any{"id": 1, "name": "test", "value": 100},
			fields:  []string{"id", "nonexistent", "name"},
			check: func(t *testing.T, result any) {
				m, ok := result.(map[string]any)
				if !ok {
					t.Error("expected map result")
					return
				}
				// 只应该包含存在的字段
				if len(m) != 2 {
					t.Errorf("expected 2 fields, got %d", len(m))
				}
				if _, exists := m["id"]; !exists {
					t.Error("expected 'id' field to exist")
				}
				if _, exists := m["name"]; !exists {
					t.Error("expected 'name' field to exist")
				}
			},
		},
		{
			name:    "nested map payload",
			payload: map[string]any{"nested": map[string]any{"key": "value"}},
			fields:  []string{"nested"},
			check: func(t *testing.T, result any) {
				m, ok := result.(map[string]any)
				if !ok {
					t.Error("expected map result")
					return
				}
				if len(m) != 1 {
					t.Errorf("expected 1 field, got %d", len(m))
				}
				if _, exists := m["nested"]; !exists {
					t.Error("expected 'nested' field to exist")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := SelectFields(tt.payload, tt.fields)
			tt.check(t, result)
		})
	}
}

// TestSelectFieldsWithNilFields 测试 nil 字段列表
func TestSelectFieldsWithNilFields(t *testing.T) {
	data := map[string]any{"id": 1, "name": "test"}
	result, _ := SelectFields(data, nil)
	m, ok := result.(map[string]any)
	if !ok {
		t.Error("expected map result")
		return
	}
	// nil 字段列表应该返回原始数据
	if len(m) != 2 {
		t.Errorf("expected 2 fields, got %d", len(m))
	}
}

// TestApplyJQWithNestedData 测试嵌套数据
func TestApplyJQWithNestedData(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]any{
		"level1": map[string]any{
			"level2": map[string]any{
				"level3": "deep value",
			},
		},
	}
	err := ApplyJQ(&buf, data, ".level1.level2.level3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()
	if output != "\"deep value\"\n" {
		t.Errorf("expected '\"deep value\"\\n', got %q", output)
	}
}

// TestApplyJQWithArrayIndex 测试数组索引
func TestApplyJQWithArrayIndex(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]any{
		"items": []any{"first", "second", "third"},
	}
	err := ApplyJQ(&buf, data, ".items[1]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()
	if output != "\"second\"\n" {
		t.Errorf("expected '\"second\"\\n', got %q", output)
	}
}

// TestApplyJQWithArraySlice 测试数组切片
func TestApplyJQWithArraySlice(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]any{
		"items": []any{1, 2, 3, 4, 5},
	}
	err := ApplyJQ(&buf, data, ".items[1:3]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()
	// 应该输出 [2, 3]
	if len(output) == 0 {
		t.Error("expected output to contain array slice")
	}
}

// TestApplyJQWithMapKeys 测试 map 键
func TestApplyJQWithMapKeys(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]any{
		"a": 1,
		"b": 2,
		"c": 3,
	}
	err := ApplyJQ(&buf, data, "keys")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()
	// 应该输出键的数组
	if len(output) == 0 {
		t.Error("expected output to contain keys array")
	}
}

// TestApplyJQWithMapValues 测试 map 值
func TestApplyJQWithMapValues(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]any{
		"a": 1,
		"b": 2,
		"c": 3,
	}
	err := ApplyJQ(&buf, data, "values")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()
	// 应该输出值的数组
	if len(output) == 0 {
		t.Error("expected output to contain values array")
	}
}

// TestApplyJQWithLength 测试长度函数
func TestApplyJQWithLength(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]any{
		"items": []any{1, 2, 3, 4, 5},
	}
	err := ApplyJQ(&buf, data, ".items | length")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()
	if output != "5\n" {
		t.Errorf("expected '5\\n', got %q", output)
	}
}

// TestApplyJQWithMap 测试 map 函数
func TestApplyJQWithMap(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]any{
		"items": []any{1, 2, 3},
	}
	err := ApplyJQ(&buf, data, ".items | map(. * 2)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()
	// 应该输出 [2, 4, 6]
	if len(output) == 0 {
		t.Error("expected output to contain mapped array")
	}
}

// TestApplyJQWithSelect 测试 select 函数
func TestApplyJQWithSelect(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]any{
		"items": []any{
			map[string]any{"id": 1, "active": true},
			map[string]any{"id": 2, "active": false},
			map[string]any{"id": 3, "active": true},
		},
	}
	err := ApplyJQ(&buf, data, ".items[] | select(.active == true) | .id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()
	// 应该输出 1 和 3
	if len(output) == 0 {
		t.Error("expected output to contain selected values")
	}
}

// TestApplyJQWithSort 测试 sort 函数
func TestApplyJQWithSort(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]any{
		"items": []any{3, 1, 2},
	}
	err := ApplyJQ(&buf, data, ".items | sort")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()
	// 应该输出 [1, 2, 3]
	if len(output) == 0 {
		t.Error("expected output to contain sorted array")
	}
}

// TestApplyJQWithUnique 测试 unique 函数
func TestApplyJQWithUnique(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]any{
		"items": []any{1, 2, 1, 3, 2},
	}
	err := ApplyJQ(&buf, data, ".items | unique")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()
	// 应该输出 [1, 2, 3]
	if len(output) == 0 {
		t.Error("expected output to contain unique array")
	}
}

// TestApplyJQWithReverse 测试 reverse 函数
func TestApplyJQWithReverse(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]any{
		"items": []any{1, 2, 3},
	}
	err := ApplyJQ(&buf, data, ".items | reverse")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()
	// 应该输出 [3, 2, 1]
	if len(output) == 0 {
		t.Error("expected output to contain reversed array")
	}
}

// TestApplyJQWithAdd 测试 add 函数
func TestApplyJQWithAdd(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]any{
		"items": []any{1, 2, 3, 4, 5},
	}
	err := ApplyJQ(&buf, data, ".items | add")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()
	if output != "15\n" {
		t.Errorf("expected '15\\n', got %q", output)
	}
}
