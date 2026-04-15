package helpers

import (
	"testing"

	"github.com/largeoliu/redmine-cli/internal/errors"
)

func TestParseID(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		resource    string
		wantID      int
		wantErr     bool
		errCategory errors.Category
	}{
		{"valid id", "123", "issue", 123, false, ""},
		{"invalid id", "abc", "project", 0, true, errors.CategoryValidation},
		{"negative id", "-1", "issue", -1, false, ""},
		{"zero id", "0", "issue", 0, false, ""},
		{"empty string", "", "project", 0, true, errors.CategoryValidation},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := ParseID(tt.input, tt.resource)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				var appErr *errors.Error
				if !errors.As(err, &appErr) {
					t.Fatal("expected *errors.Error")
				}
				if appErr.Category != tt.errCategory {
					t.Errorf("category = %v, want %v", appErr.Category, tt.errCategory)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if id != tt.wantID {
				t.Errorf("id = %d, want %d", id, tt.wantID)
			}
		})
	}
}
