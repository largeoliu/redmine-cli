// internal/output/formatter_test.go
package output

import (
	"bytes"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]string{"key": "value"}
	err := WriteJSON(&buf, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "{\n  \"key\": \"value\"\n}\n"
	if buf.String() != expected {
		t.Errorf("expected %q, got %q", expected, buf.String())
	}
}

func TestWriteTable(t *testing.T) {
	var buf bytes.Buffer
	data := []map[string]any{
		{"name": "Alice", "age": 30},
		{"name": "Bob", "age": 25},
	}
	err := WriteTable(&buf, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("Alice")) {
		t.Error("expected output to contain 'Alice'")
	}
}

func TestApplyJQ(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]any{
		"items": []map[string]any{
			{"id": 1, "name": "first"},
			{"id": 2, "name": "second"},
		},
	}
	err := ApplyJQ(&buf, data, ".items[].id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "1\n2\n"
	if buf.String() != expected {
		t.Errorf("expected %q, got %q", expected, buf.String())
	}
}

func TestSelectFields(t *testing.T) {
	data := map[string]any{
		"id":    1,
		"name":  "test",
		"extra": "should be removed",
	}
	result, _ := SelectFields(data, []string{"id", "name"})
	m, ok := result.(map[string]any)
	if !ok {
		t.Fatal("expected map result")
	}
	if len(m) != 2 {
		t.Errorf("expected 2 fields, got %d", len(m))
	}
	if _, exists := m["extra"]; exists {
		t.Error("expected 'extra' to be removed")
	}
}

func TestSanitize(t *testing.T) {
	input := "hello\x00world"
	result := Sanitize(input)
	if result != "hello world" {
		t.Errorf("expected 'hello world', got %q", result)
	}
}

func TestNormalizePayloadNil(t *testing.T) {
	result, err := normalizePayload(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestNormalizePayloadString(t *testing.T) {
	result, err := normalizePayload("hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "hello" {
		t.Errorf("expected 'hello', got %v", result)
	}
}

func TestNormalizePayloadUnmarshalError(t *testing.T) {
	bad := make(chan int)
	_, err := normalizePayload(bad)
	if err == nil {
		t.Fatal("expected error for unmarshalable type, got nil")
	}
}

func TestWriteDataRowsWriteError(t *testing.T) {
	w := &errorWriter{}
	rows := [][]string{{"a", "b"}}
	widths := []int{5, 5}
	err := writeDataRows(w, rows, widths)
	if err == nil {
		t.Fatal("expected error from errorWriter, got nil")
	}
}

func TestWriteTableWriteError(t *testing.T) {
	w := &errorWriter{}
	headers := []string{"H1", "H2"}
	rows := [][]string{{"a", "b"}}
	err := writeTable(w, headers, rows)
	if err == nil {
		t.Fatal("expected error from errorWriter, got nil")
	}
}

func TestWriteKeyValuesWriteError(t *testing.T) {
	w := &errorWriter{}
	payload := map[string]any{"key": "value"}
	err := writeKeyValues(w, payload)
	if err == nil {
		t.Fatal("expected error from errorWriter, got nil")
	}
}
