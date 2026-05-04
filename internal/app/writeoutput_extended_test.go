package app

import (
	"bytes"
	"strings"
	"testing"
)

// TestWriteOutputSelectFieldsError tests the error handling branch in WriteOutput
// when SelectFields returns an error
func TestWriteOutputSelectFieldsError(t *testing.T) {
	// Create a payload that cannot be marshaled to JSON
	// Channels cannot be marshaled to JSON
	badPayload := make(chan int)

	flags := &GlobalFlags{
		Format: "json",
		Fields: "id,name", // This will trigger SelectFields
	}

	var buf bytes.Buffer
	err := WriteOutput(&buf, flags, badPayload)

	// Should return an error because the payload cannot be marshaled
	if err == nil {
		t.Error("Expected error for unmashalable payload, got nil")
	}
}

// TestWriteOutputWithFieldsAndFormat tests WriteOutput with both fields and format
func TestWriteOutputWithFieldsAndFormat(t *testing.T) {
	tests := []struct {
		name    string
		format  string
		fields  string
		payload any
		wantErr bool
	}{
		{
			name:    "json format with fields",
			format:  "json",
			fields:  "id,name",
			payload: map[string]any{"id": 1, "name": "test", "extra": "ignored"},
			wantErr: false,
		},
		{
			name:    "table format with fields",
			format:  "table",
			fields:  "id,name",
			payload: map[string]any{"id": 1, "name": "test", "extra": "ignored"},
			wantErr: false,
		},
		{
			name:    "raw format with fields",
			format:  "raw",
			fields:  "id,name",
			payload: map[string]any{"id": 1, "name": "test", "extra": "ignored"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := &GlobalFlags{
				Format: tt.format,
				Fields: tt.fields,
			}

			var buf bytes.Buffer
			err := WriteOutput(&buf, flags, tt.payload)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if buf.String() == "" {
					t.Error("Expected output, got empty")
				}
			}
		})
	}
}

// TestWriteOutputJQPriority tests that JQ takes priority over Fields
func TestWriteOutputJQPriority(t *testing.T) {
	payload := map[string]any{
		"id":   1,
		"name": "test",
	}

	flags := &GlobalFlags{
		Format: "json",
		JQ:     ".name",
		Fields: "id", // This should be ignored because JQ is set
	}

	var buf bytes.Buffer
	err := WriteOutput(&buf, flags, payload)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("Expected output, got empty")
	}

	// Output should contain "test" (from JQ filter) not the full object
	if !bytes.Contains(buf.Bytes(), []byte("test")) {
		t.Errorf("Expected output to contain 'test', got: %s", output)
	}
}

// TestWriteOutputEmptyFields tests WriteOutput with empty fields
func TestWriteOutputEmptyFields(t *testing.T) {
	payload := map[string]any{
		"id":   1,
		"name": "test",
	}

	flags := &GlobalFlags{
		Format: "json",
		Fields: "",
	}

	var buf bytes.Buffer
	err := WriteOutput(&buf, flags, payload)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if buf.String() == "" {
		t.Error("Expected output, got empty")
	}
}

// TestWriteOutputNilPayload tests WriteOutput with nil payload
func TestWriteOutputNilPayload(t *testing.T) {
	flags := &GlobalFlags{
		Format: "json",
	}

	var buf bytes.Buffer
	err := WriteOutput(&buf, flags, nil)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if buf.String() == "" {
		t.Error("Expected output, got empty")
	}
}

// TestWriteOutputComplexPayload tests WriteOutput with complex nested payload
func TestWriteOutputComplexPayload(t *testing.T) {
	payload := map[string]any{
		"id": 1,
		"user": map[string]any{
			"name":  "test",
			"email": "test@example.com",
		},
		"items": []any{
			map[string]any{"id": 1, "name": "item1"},
			map[string]any{"id": 2, "name": "item2"},
		},
	}

	flags := &GlobalFlags{
		Format: "json",
		Fields: "id,user",
	}

	var buf bytes.Buffer
	err := WriteOutput(&buf, flags, payload)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("Expected output, got empty")
	}

	// Should contain id and user, but not items
	if !bytes.Contains(buf.Bytes(), []byte("id")) {
		t.Error("Expected output to contain 'id'")
	}
	if !bytes.Contains(buf.Bytes(), []byte("user")) {
		t.Error("Expected output to contain 'user'")
	}
}

// TestWriteOutputArrayPayload tests WriteOutput with array payload
func TestWriteOutputArrayPayload(t *testing.T) {
	payload := []map[string]any{
		{"id": 1, "name": "first"},
		{"id": 2, "name": "second"},
	}

	flags := &GlobalFlags{
		Format: "json",
		Fields: "id",
	}

	var buf bytes.Buffer
	err := WriteOutput(&buf, flags, payload)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if buf.String() == "" {
		t.Error("Expected output, got empty")
	}
}

// TestWriteOutputStringPayload tests WriteOutput with string payload
func TestWriteOutputStringPayload(t *testing.T) {
	payload := "simple string"

	flags := &GlobalFlags{
		Format: "json",
		Fields: "id", // Fields should be ignored for non-map payloads
	}

	var buf bytes.Buffer
	err := WriteOutput(&buf, flags, payload)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if buf.String() == "" {
		t.Error("Expected output, got empty")
	}
}

// TestWriteOutputIntegerPayload tests WriteOutput with integer payload
func TestWriteOutputIntegerPayload(t *testing.T) {
	payload := 42

	flags := &GlobalFlags{
		Format: "json",
	}

	var buf bytes.Buffer
	err := WriteOutput(&buf, flags, payload)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if buf.String() == "" {
		t.Error("Expected output, got empty")
	}
}

// TestWriteOutputScalarPayloadWithFields tests WriteOutput with scalar payload and fields set
// This covers the case where SelectFieldsNormalized returns input unchanged for non-collection types
func TestWriteOutputScalarPayloadWithFields(t *testing.T) {
	tests := []struct {
		name    string
		payload any
		fields  string
	}{
		{"string with fields", "hello", "id"},
		{"int with fields", 42, "id"},
		{"float with fields", 3.14, "id"},
		{"bool with fields", true, "id"},
		{"nil with fields", nil, "id"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := &GlobalFlags{
				Format: "json",
				Fields: tt.fields,
			}

			var buf bytes.Buffer
			err := WriteOutput(&buf, flags, tt.payload)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if buf.Len() == 0 {
				t.Error("expected output, got empty")
			}
		})
	}
}

// TestWriteOutputFieldsError tests WriteOutput when SelectFieldsNormalized returns an error
// This is distinct from NormalizePayload errors
func TestWriteOutputFieldsSelectFieldsError(t *testing.T) {
	flags := &GlobalFlags{
		Format: "json",
		Fields: "id",
	}

	ch := make(chan int)
	payload := map[string]any{
		"id": 1,
		"ch": ch,
	}

	var buf bytes.Buffer
	err := WriteOutput(&buf, flags, payload)

	if err == nil {
		t.Error("expected error from SelectFieldsNormalized with unmarshallable payload, got nil")
	}
}

// TestWriteOutputTableFormatWithFields tests WriteOutput with table format and fields
// Covers line 227 when format is FormatTable
func TestWriteOutputTableFormatWithFields(t *testing.T) {
	payload := map[string]any{
		"id":   1,
		"name": "test",
		"tags": []any{"a", "b"},
	}

	flags := &GlobalFlags{
		Format: "table",
		Fields: "id,name",
	}

	var buf bytes.Buffer
	err := WriteOutput(&buf, flags, payload)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected output, got empty")
	}
}

// TestWriteOutputRawFormatWithFields tests WriteOutput with raw format and fields
// Covers line 227 when format is FormatRaw
func TestWriteOutputRawFormatWithFields(t *testing.T) {
	payload := map[string]any{
		"id":   1,
		"name": "test",
	}

	flags := &GlobalFlags{
		Format: "raw",
		Fields: "id,name",
	}

	var buf bytes.Buffer
	err := WriteOutput(&buf, flags, payload)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected output, got empty")
	}
}

// TestWriteOutputTableFormatNonMapArray tests WriteOutput with table format and array of non-maps
// Covers output.Write with table format and non-map/slice input
func TestWriteOutputTableFormatNonMapArray(t *testing.T) {
	payload := []any{1, 2, 3, "four", true}

	flags := &GlobalFlags{
		Format: "table",
	}

	var buf bytes.Buffer
	err := WriteOutput(&buf, flags, payload)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected output, got empty")
	}
}

// TestWriteOutputRawFormatScalar tests WriteOutput with raw format and scalar payload
// Covers output.Write with raw format and string input
func TestWriteOutputRawFormatScalar(t *testing.T) {
	payload := "simple string value"

	flags := &GlobalFlags{
		Format: "raw",
	}

	var buf bytes.Buffer
	err := WriteOutput(&buf, flags, payload)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected output, got empty")
	}
}

// TestWriteOutputSelectFieldsArrayPayload tests WriteOutput with fields on array payload
// Covers SelectFieldsNormalized with array input
func TestWriteOutputSelectFieldsArrayPayload(t *testing.T) {
	payload := []map[string]any{
		{"id": 1, "name": "first", "extra": "ignored"},
		{"id": 2, "name": "second", "extra": "ignored"},
	}

	flags := &GlobalFlags{
		Format: "json",
		Fields: "id,name",
	}

	var buf bytes.Buffer
	err := WriteOutput(&buf, flags, payload)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected output, got empty")
	}

	output := buf.String()
	if !strings.Contains(output, `"id"`) && !strings.Contains(output, `"name"`) {
		t.Error("expected output to contain id and name fields")
	}
}

// TestWriteOutputSelectFieldsNestedArrayInMap tests the nested array handling in SelectFieldsNormalized
// This covers the code path where SelectFieldsNormalized processes map with array values
func TestWriteOutputSelectFieldsNestedArrayInMap(t *testing.T) {
	payload := map[string]any{
		"id": 1,
		"items": []map[string]any{
			{"id": 10, "name": "item1", "extra": "ignored"},
			{"id": 20, "name": "item2", "extra": "ignored"},
		},
		"nested": map[string]any{
			"inner": "value",
		},
	}

	flags := &GlobalFlags{
		Format: "json",
		Fields: "id,items",
	}

	var buf bytes.Buffer
	err := WriteOutput(&buf, flags, payload)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected output, got empty")
	}
}

// TestWriteOutputNoFieldsNoJQ tests WriteOutput with no fields and no JQ
// This is the fallback path at line 230
func TestWriteOutputNoFieldsNoJQ(t *testing.T) {
	payload := map[string]any{"id": 1, "name": "test"}

	flags := &GlobalFlags{
		Format: "json",
	}

	var buf bytes.Buffer
	err := WriteOutput(&buf, flags, payload)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected output, got empty")
	}
}

// TestWriteOutputTableFormatWithJQ tests that JQ takes priority and returns early
// This covers line 218 (ApplyJQNormalized return)
func TestWriteOutputTableFormatWithJQ(t *testing.T) {
	payload := map[string]any{
		"id":   1,
		"name": "test",
		"items": []map[string]any{
			{"id": 10, "name": "item1"},
		},
	}

	flags := &GlobalFlags{
		Format: "table",
		JQ:     ".items[]",
		Fields: "id,name",
	}

	var buf bytes.Buffer
	err := WriteOutput(&buf, flags, payload)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected output, got empty")
	}

	output := buf.String()
	if !strings.Contains(output, "10") && !strings.Contains(output, "item1") {
		t.Error("expected JQ-filtered output")
	}
}

// TestWriteOutputEmptyPayload tests WriteOutput with empty/nil payload
func TestWriteOutputEmptyPayload(t *testing.T) {
	flags := &GlobalFlags{
		Format: "json",
	}

	var buf bytes.Buffer

	err := WriteOutput(&buf, flags, map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error for empty map: %v", err)
	}

	err = WriteOutput(&buf, flags, []any{})
	if err != nil {
		t.Fatalf("unexpected error for empty slice: %v", err)
	}
}

// TestWriteOutputMapWithNestedMap tests WriteOutput with nested map structure
func TestWriteOutputMapWithNestedMap(t *testing.T) {
	payload := map[string]any{
		"id": 1,
		"nested": map[string]any{
			"level": 2,
			"data":  "value",
		},
	}

	flags := &GlobalFlags{
		Format: "table",
	}

	var buf bytes.Buffer
	err := WriteOutput(&buf, flags, payload)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected output, got empty")
	}
}
