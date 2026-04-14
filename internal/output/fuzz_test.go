// internal/output/fuzz_test.go
package output

import (
	"bytes"
	"encoding/json"
	"testing"
	"unicode/utf8"
)

func FuzzSanitize(f *testing.F) {
	testcases := []string{
		"hello world",
		"test\x00null",
		"tab\there",
		"new\nline",
		"carriage\rreturn",
		"",
		"normal text without control characters",
		"\x00\x01\x02\x03\x04\x05",
		"mixed\x00content\twith\ncontrol\rcharacters",
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, input string) {
		result := Sanitize(input)
		for _, r := range result {
			if r < 0x20 && r != '\n' && r != '\t' {
				t.Errorf("Sanitize produced control character: %d", r)
			}
		}
	})
}

func FuzzWriteJSON(f *testing.F) {
	testcases := []string{
		`{"key": "value"}`,
		`{"nested": {"key": "value"}}`,
		`{"array": [1, 2, 3]}`,
		`{"number": 123}`,
		`{"bool": true}`,
		`{"null": null}`,
		`{}`,
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, input string) {
		var buf bytes.Buffer
		var data any
		if err := json.Unmarshal([]byte(input), &data); err != nil {
			return
		}

		err := WriteJSON(&buf, data)
		if err != nil {
			t.Errorf("WriteJSON failed: %v", err)
		}
	})
}

func FuzzWriteTable(f *testing.F) {
	testcases := []string{
		`{"id": 1, "name": "test"}`,
		`{"items": [{"id": 1}, {"id": 2}]}`,
		`{"nested": {"key": "value"}}`,
		`{}`,
		`[]`,
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(_ *testing.T, input string) {
		var buf bytes.Buffer
		var data any
		if err := json.Unmarshal([]byte(input), &data); err != nil {
			return
		}

		_ = WriteTable(&buf, data)
	})
}

func FuzzApplyJQ(f *testing.F) {
	testcases := []struct {
		json    string
		jqQuery string
	}{
		{`{"key": "value"}`, ".key"},
		{`{"nested": {"key": "value"}}`, ".nested.key"},
		{`{"array": [1, 2, 3]}`, ".array[]"},
		{`{"items": [{"id": 1}, {"id": 2}]}`, ".items[].id"},
		{`{"number": 123}`, ".number"},
	}

	for _, tc := range testcases {
		f.Add(tc.json, tc.jqQuery)
	}

	f.Fuzz(func(_ *testing.T, jsonInput, jqQuery string) {
		var buf bytes.Buffer
		var data any
		if err := json.Unmarshal([]byte(jsonInput), &data); err != nil {
			return
		}

		_ = ApplyJQ(&buf, data, jqQuery)
	})
}

func FuzzSelectFields(f *testing.F) {
	testcases := []struct {
		json   string
		fields string
	}{
		{`{"id": 1, "name": "test", "value": 100}`, "id,name"},
		{`{"nested": {"key": "value"}}`, "nested"},
		{`{"a": 1, "b": 2, "c": 3}`, "a,b,c"},
		{`{}`, ""},
	}

	for _, tc := range testcases {
		f.Add(tc.json, tc.fields)
	}

	f.Fuzz(func(_ *testing.T, jsonInput, fieldsStr string) {
		var data map[string]any
		if err := json.Unmarshal([]byte(jsonInput), &data); err != nil {
			return
		}

		var fields []string
		if fieldsStr != "" {
			fields = append(fields, splitFields(fieldsStr)...)
		}

		_, _ = SelectFields(data, fields)
	})
}

func FuzzFormatValue(f *testing.F) {
	testcases := []string{
		"string value",
		"123",
		"true",
		"null",
		`{"key": "value"}`,
		`[1, 2, 3]`,
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(_ *testing.T, input string) {
		var data any
		if err := json.Unmarshal([]byte(input), &data); err != nil {
			data = input
		}

		_ = formatValue(data)
	})
}

func FuzzTruncate(f *testing.F) {
	testcases := []struct {
		input string
		width int
	}{
		{"hello", 10},
		{"hello world", 5},
		{"short", 3},
		{"", 5},
		{"a", 1},
	}

	for _, tc := range testcases {
		f.Add(tc.input, tc.width)
	}

	f.Fuzz(func(t *testing.T, input string, width int) {
		if width < 0 || width > 1000 {
			return
		}

		result := truncate(input, width)
		runeCount := utf8.RuneCountInString(result)
		if runeCount > width {
			t.Errorf("truncate result rune count %d exceeds width %d", runeCount, width)
		}
	})
}

func FuzzWriteRaw(f *testing.F) {
	testcases := []string{
		"plain text",
		`{"json": "data"}`,
		"",
		"special\x00characters",
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(_ *testing.T, input string) {
		var buf bytes.Buffer
		_ = WriteRaw(&buf, input)
	})
}

func FuzzNormalizePayload(f *testing.F) {
	testcases := []string{
		`{"key": "value"}`,
		`[1, 2, 3]`,
		`"string"`,
		`123`,
		`true`,
		`null`,
		``,
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(_ *testing.T, input string) {
		var data any
		if input != "" {
			if err := json.Unmarshal([]byte(input), &data); err != nil {
				return
			}
		} else {
			data = nil
		}

		_, _ = normalizePayload(data)
	})
}

func splitFields(s string) []string {
	var result []string
	var current string
	for _, r := range s {
		if r == ',' {
			if current != "" {
				result = append(result, current)
			}
			current = ""
		} else {
			current += string(r)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}
