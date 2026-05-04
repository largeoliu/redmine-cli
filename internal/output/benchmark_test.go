// internal/output/benchmark_test.go
package output

import (
	"bytes"
	"testing"
)

// BenchmarkWriteJSON 测试 JSON 输出性能
func BenchmarkWriteJSON(b *testing.B) {
	data := map[string]any{
		"id":          1,
		"name":        "Test Issue",
		"description": "This is a test issue description",
		"status":      "open",
		"priority":    3,
		"author": map[string]any{
			"id":   1,
			"name": "John Doe",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		_ = WriteJSON(&buf, data)
	}
}

// BenchmarkWriteJSONLarge 测试大型数据结构�?JSON 输出性能
func BenchmarkWriteJSONLarge(b *testing.B) {
	// 创建包含 100 个项目的数组
	items := make([]map[string]any, 100)
	for i := 0; i < 100; i++ {
		items[i] = map[string]any{
			"id":          i + 1,
			"name":        "Issue " + itoaOutput(i+1),
			"description": "Description for issue " + itoaOutput(i+1),
			"status":      "open",
			"priority":    i % 5,
		}
	}
	data := map[string]any{
		"issues":      items,
		"total_count": 100,
		"limit":       100,
		"offset":      0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		_ = WriteJSON(&buf, data)
	}
}

// BenchmarkWriteTable 测试 Table 输出性能
func BenchmarkWriteTable(b *testing.B) {
	data := []map[string]any{
		{"id": 1, "name": "Alice", "age": 30, "email": "alice@example.com"},
		{"id": 2, "name": "Bob", "age": 25, "email": "bob@example.com"},
		{"id": 3, "name": "Charlie", "age": 35, "email": "charlie@example.com"},
		{"id": 4, "name": "Diana", "age": 28, "email": "diana@example.com"},
		{"id": 5, "name": "Eve", "age": 32, "email": "eve@example.com"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		_ = WriteTable(&buf, data)
	}
}

// BenchmarkWriteTableLarge 测试大型数据集的 Table 输出性能
func BenchmarkWriteTableLarge(b *testing.B) {
	// 创建 100 行数�?
	data := make([]map[string]any, 100)
	for i := 0; i < 100; i++ {
		data[i] = map[string]any{
			"id":          i + 1,
			"name":        "Issue " + itoaOutput(i+1),
			"status":      "open",
			"priority":    i % 5,
			"author":      "User " + itoaOutput(i%10+1),
			"description": "A longer description text for testing table output performance",
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		_ = WriteTable(&buf, data)
	}
}

// BenchmarkWriteTableSingleRow 测试单行数据�?Table 输出性能
func BenchmarkWriteTableSingleRow(b *testing.B) {
	data := []map[string]any{
		{"id": 1, "name": "Single Item", "status": "open"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		_ = WriteTable(&buf, data)
	}
}

// BenchmarkWriteTableMap 测试单个 map �?Table 输出性能
func BenchmarkWriteTableMap(b *testing.B) {
	data := map[string]any{
		"id":          1,
		"name":        "Test Issue",
		"description": "This is a detailed description",
		"status":      "open",
		"priority":    3,
		"author":      "John Doe",
		"created_on":  "2024-01-15T10:30:00Z",
		"updated_on":  "2024-01-16T14:45:00Z",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		_ = WriteTable(&buf, data)
	}
}

// BenchmarkApplyJQ 测试 JQ 过滤性能
func BenchmarkApplyJQ(b *testing.B) {
	data := map[string]any{
		"items": []map[string]any{
			{"id": 1, "name": "first", "value": 100},
			{"id": 2, "name": "second", "value": 200},
			{"id": 3, "name": "third", "value": 300},
		},
		"total": 3,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		_ = ApplyJQ(&buf, data, ".items[].id")
	}
}

// BenchmarkApplyJQComplex 测试复杂 JQ 表达式性能
func BenchmarkApplyJQComplex(b *testing.B) {
	items := make([]map[string]any, 50)
	for i := 0; i < 50; i++ {
		items[i] = map[string]any{
			"id":       i + 1,
			"name":     "Item " + itoaOutput(i+1),
			"value":    i * 10,
			"active":   i%2 == 0,
			"category": "cat-" + itoaOutput(i%5+1),
		}
	}
	data := map[string]any{
		"items": items,
		"meta": map[string]any{
			"total":   50,
			"page":    1,
			"perPage": 50,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		_ = ApplyJQ(&buf, data, ".items[] | select(.active == true) | {id, name}")
	}
}

// BenchmarkApplyJQSimpleFilter 测试简�?JQ 过滤性能
func BenchmarkApplyJQSimpleFilter(b *testing.B) {
	data := map[string]any{
		"id":          1,
		"name":        "Test",
		"description": "Description",
		"status":      "open",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		_ = ApplyJQ(&buf, data, ".name")
	}
}

// BenchmarkApplyJQArray 测试数组 JQ 过滤性能
func BenchmarkApplyJQArray(b *testing.B) {
	items := make([]any, 100)
	for i := 0; i < 100; i++ {
		items[i] = i + 1
	}
	data := map[string]any{
		"numbers": items,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		_ = ApplyJQ(&buf, data, ".numbers[] | select(. > 50)")
	}
}

// BenchmarkSelectFields 测试字段选择性能
func BenchmarkSelectFields(b *testing.B) {
	data := map[string]any{
		"id":          1,
		"name":        "Test",
		"description": "Description",
		"status":      "open",
		"priority":    3,
		"author":      "John",
		"extra1":      "value1",
		"extra2":      "value2",
		"extra3":      "value3",
	}
	fields := []string{"id", "name", "status"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = SelectFields(data, fields)
	}
}

// BenchmarkSelectFieldsAll 测试选择所有字段的性能
func BenchmarkSelectFieldsAll(b *testing.B) {
	data := map[string]any{
		"id":          1,
		"name":        "Test",
		"description": "Description",
		"status":      "open",
	}
	fields := []string{"id", "name", "description", "status"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = SelectFields(data, fields)
	}
}

// BenchmarkSanitize 测试字符串清理性能
func BenchmarkSanitize(b *testing.B) {
	input := "hello\x00world\twith\ncontrol\rcharacters"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Sanitize(input)
	}
}

// BenchmarkSanitizeClean 测试干净字符串的清理性能
func BenchmarkSanitizeClean(b *testing.B) {
	input := "hello world without control characters"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Sanitize(input)
	}
}

// BenchmarkWrite 测试通用 Write 函数性能 (JSON 格式)
func BenchmarkWriteJSONFormat(b *testing.B) {
	data := map[string]any{
		"id":   1,
		"name": "Test",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		_ = Write(&buf, FormatJSON, data)
	}
}

// BenchmarkWrite 测试通用 Write 函数性能 (Table 格式)
func BenchmarkWriteTableFormat(b *testing.B) {
	data := []map[string]any{
		{"id": 1, "name": "Test"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		_ = Write(&buf, FormatTable, data)
	}
}

// BenchmarkFormatValue 测试值格式化性能
func BenchmarkFormatValue(b *testing.B) {
	values := []any{
		"string value",
		123,
		3.14,
		true,
		nil,
		[]int{1, 2, 3},
		map[string]string{"key": "value"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, v := range values {
			_ = formatValue(v)
		}
	}
}

// BenchmarkApplyJQNormalized tests ApplyJQ with pre-compiled query performance
func BenchmarkApplyJQNormalized(b *testing.B) {
	items := make([]map[string]any, 100)
	for i := 0; i < 100; i++ {
		items[i] = map[string]any{
			"id":      i + 1,
			"subject": "Issue " + itoaOutput(i+1),
			"status":  "open",
		}
	}
	data := map[string]any{"issues": items}
	query, _ := ParseJQ(".issues[] | {id, subject, status}")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		_ = ApplyJQNormalized(&buf, data, query)
	}
}

// BenchmarkSelectFieldsNormalized tests SelectFieldsNormalized performance
func BenchmarkSelectFieldsNormalized(b *testing.B) {
	items := make([]map[string]any, 100)
	for i := 0; i < 100; i++ {
		items[i] = map[string]any{
			"id":      i + 1,
			"subject": "Issue " + itoaOutput(i+1),
			"status":  "open",
		}
	}
	data := map[string]any{"issues": items}
	fields := []string{"id", "subject", "status"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = SelectFieldsNormalized(data, fields)
	}
}

// 辅助函数
func itoaOutput(n int) string {
	if n == 0 {
		return "0"
	}
	var result []byte
	for n > 0 {
		result = append([]byte{byte('0' + n%10)}, result...)
		n /= 10
	}
	return string(result)
}
