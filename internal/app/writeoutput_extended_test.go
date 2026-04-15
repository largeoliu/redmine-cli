package app

import (
	"bytes"
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

// TestWriteOutputBooleanPayload tests WriteOutput with boolean payload
func TestWriteOutputBooleanPayload(t *testing.T) {
	payload := true

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
