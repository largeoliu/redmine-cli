package helpers

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestConfirmDelete(t *testing.T) {
	tests := []struct {
		name         string
		resourceName string
		id           int
		yes          bool
		input        string
		want         bool
	}{
		{
			name:         "skip confirmation with yes flag",
			resourceName: "issue",
			id:           123,
			yes:          true,
			want:         true,
		},
		{
			name:         "confirm with y input",
			resourceName: "project",
			id:           456,
			yes:          false,
			input:        "y\n",
			want:         true,
		},
		{
			name:         "confirm with Y input",
			resourceName: "user",
			id:           789,
			yes:          false,
			input:        "Y\n",
			want:         true,
		},
		{
			name:         "cancel with n input",
			resourceName: "issue",
			id:           123,
			yes:          false,
			input:        "n\n",
			want:         false,
		},
		{
			name:         "cancel with empty input",
			resourceName: "project",
			id:           456,
			yes:          false,
			input:        "\n",
			want:         false,
		},
		{
			name:         "cancel with other input",
			resourceName: "user",
			id:           789,
			yes:          false,
			input:        "abc\n",
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.yes {
				got := ConfirmDelete(tt.resourceName, tt.id, tt.yes)
				if got != tt.want {
					t.Errorf("ConfirmDelete() = %v, want %v", got, tt.want)
				}
				return
			}

			// Save and restore stdin
			oldStdin := os.Stdin
			defer func() { os.Stdin = oldStdin }()

			// Create pipe for simulated input
			r, w, _ := os.Pipe()
			os.Stdin = r

			// Write input in goroutine
			go func() {
				w.WriteString(tt.input)
				w.Close()
			}()

			got := ConfirmDelete(tt.resourceName, tt.id, tt.yes)
			if got != tt.want {
				t.Errorf("ConfirmDelete() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDryRunCreate(t *testing.T) {
	tests := []struct {
		name         string
		resourceName string
		req          any
		wantContain  string
	}{
		{
			name:         "create issue dry run",
			resourceName: "issue",
			req:          map[string]string{"subject": "Test Issue"},
			wantContain:  "[dry-run] Would create issue",
		},
		{
			name:         "create project dry run",
			resourceName: "project",
			req:          struct{ Name string }{Name: "Test Project"},
			wantContain:  "[dry-run] Would create project",
		},
		{
			name:         "create user dry run",
			resourceName: "user",
			req:          nil,
			wantContain:  "[dry-run] Would create user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			got := DryRunCreate(tt.resourceName, tt.req)

			// Restore stdout and read captured output
			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			if !got {
				t.Errorf("DryRunCreate() = %v, want true", got)
			}

			if !strings.Contains(output, tt.wantContain) {
				t.Errorf("DryRunCreate() output = %q, want to contain %q", output, tt.wantContain)
			}
		})
	}
}

func TestDryRunUpdate(t *testing.T) {
	tests := []struct {
		name         string
		resourceName string
		id           int
		req          any
		wantContain  string
	}{
		{
			name:         "update issue dry run",
			resourceName: "issue",
			id:           123,
			req:          map[string]string{"subject": "Updated Issue"},
			wantContain:  "[dry-run] Would update issue #123",
		},
		{
			name:         "update project dry run",
			resourceName: "project",
			id:           456,
			req:          struct{ Name string }{Name: "Updated Project"},
			wantContain:  "[dry-run] Would update project #456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			got := DryRunUpdate(tt.resourceName, tt.id, tt.req)

			// Restore stdout and read captured output
			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			if !got {
				t.Errorf("DryRunUpdate() = %v, want true", got)
			}

			if !strings.Contains(output, tt.wantContain) {
				t.Errorf("DryRunUpdate() output = %q, want to contain %q", output, tt.wantContain)
			}
		})
	}
}

func TestDryRunDelete(t *testing.T) {
	tests := []struct {
		name         string
		resourceName string
		id           int
		wantContain  string
	}{
		{
			name:         "delete issue dry run",
			resourceName: "issue",
			id:           123,
			wantContain:  "[dry-run] Would delete issue #123",
		},
		{
			name:         "delete project dry run",
			resourceName: "project",
			id:           456,
			wantContain:  "[dry-run] Would delete project #456",
		},
		{
			name:         "delete user dry run",
			resourceName: "user",
			id:           789,
			wantContain:  "[dry-run] Would delete user #789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			got := DryRunDelete(tt.resourceName, tt.id)

			// Restore stdout and read captured output
			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			if !got {
				t.Errorf("DryRunDelete() = %v, want true", got)
			}

			if !strings.Contains(output, tt.wantContain) {
				t.Errorf("DryRunDelete() output = %q, want to contain %q", output, tt.wantContain)
			}
		})
	}
}

func TestParseID(t *testing.T) {
	tests := []struct {
		name         string
		arg          string
		resourceName string
		want         int
		wantErr      bool
	}{
		{
			name:         "valid positive id",
			arg:          "123",
			resourceName: "issue",
			want:         123,
			wantErr:      false,
		},
		{
			name:         "valid zero id",
			arg:          "0",
			resourceName: "project",
			want:         0,
			wantErr:      false,
		},
		{
			name:         "valid negative id",
			arg:          "-1",
			resourceName: "user",
			want:         -1,
			wantErr:      false,
		},
		{
			name:         "invalid string",
			arg:          "abc",
			resourceName: "issue",
			want:         0,
			wantErr:      true,
		},
		{
			name:         "invalid float",
			arg:          "1.5",
			resourceName: "project",
			want:         0,
			wantErr:      true,
		},
		{
			name:         "invalid empty string",
			arg:          "",
			resourceName: "user",
			want:         0,
			wantErr:      true,
		},
		{
			name:         "invalid mixed string",
			arg:          "123abc",
			resourceName: "issue",
			want:         0,
			wantErr:      true,
		},
		{
			name:         "valid large id",
			arg:          "999999999",
			resourceName: "project",
			want:         999999999,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseID(tt.arg, tt.resourceName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseID() = %v, want %v", got, tt.want)
			}
			if err != nil {
				expectedMsg := "invalid " + tt.resourceName + " ID"
				if !strings.Contains(err.Error(), expectedMsg) {
					t.Errorf("ParseID() error message = %v, want to contain %v", err.Error(), expectedMsg)
				}
			}
		})
	}
}
