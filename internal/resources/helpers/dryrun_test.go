package helpers

import (
	"io"
	"os"
	"strings"
	"testing"
)

func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = old

	return string(out)
}

func TestDryRunCreate(t *testing.T) {
	req := struct {
		Name string
	}{Name: "test"}

	out := captureStdout(func() {
		result := DryRunCreate("issue", req)
		if !result {
			t.Error("expected true")
		}
	})

	if !strings.Contains(out, "[dry-run] Would create issue:") {
		t.Errorf("unexpected output: %s", out)
	}
}

func TestDryRunUpdate(t *testing.T) {
	req := map[string]string{"subject": "updated"}

	out := captureStdout(func() {
		result := DryRunUpdate("issue", 10, req)
		if !result {
			t.Error("expected true")
		}
	})

	if !strings.Contains(out, "[dry-run] Would update issue #10:") {
		t.Errorf("unexpected output: %s", out)
	}
}

func TestDryRunDelete(t *testing.T) {
	out := captureStdout(func() {
		result := DryRunDelete("project", 5)
		if !result {
			t.Error("expected true")
		}
	})

	if !strings.Contains(out, "[dry-run] Would delete project #5") {
		t.Errorf("unexpected output: %s", out)
	}
}
