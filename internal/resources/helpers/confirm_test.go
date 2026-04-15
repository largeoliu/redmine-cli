package helpers

import (
	"io"
	"os"
	"strings"
	"testing"
)

func TestConfirmDelete_YesFlag(t *testing.T) {
	result := ConfirmDelete("issue", 42, true)
	if !result {
		t.Error("expected true when yes=true")
	}
}

func TestConfirmDelete_UserConfirms(t *testing.T) {
	oldStdin := os.Stdin
	oldStdout := os.Stdout

	r, w, _ := os.Pipe()
	os.Stdin = r

	outR, outW, _ := os.Pipe()
	os.Stdout = outW

	go func() {
		w.WriteString("y\n")
		w.Close()
	}()

	result := ConfirmDelete("issue", 1, false)

	w.Close()
	outW.Close()
	out, _ := io.ReadAll(outR)

	os.Stdin = oldStdin
	os.Stdout = oldStdout

	if !result {
		t.Error("expected true for 'y' input")
	}
	if !strings.Contains(string(out), "Delete issue #1?") {
		t.Error("expected prompt output")
	}
}

func TestConfirmDelete_UserUpperY(t *testing.T) {
	oldStdin := os.Stdin
	oldStdout := os.Stdout

	r, w, _ := os.Pipe()
	os.Stdin = r

	_, outW, _ := os.Pipe()
	os.Stdout = outW

	go func() {
		w.WriteString("Y\n")
		w.Close()
	}()

	result := ConfirmDelete("project", 5, false)

	w.Close()
	outW.Close()

	os.Stdin = oldStdin
	os.Stdout = oldStdout

	if !result {
		t.Error("expected true for 'Y' input")
	}
}

func TestConfirmDelete_UserCancels(t *testing.T) {
	oldStdin := os.Stdin
	oldStdout := os.Stdout

	r, w, _ := os.Pipe()
	os.Stdin = r

	outR, outW, _ := os.Pipe()
	os.Stdout = outW

	go func() {
		w.WriteString("n\n")
		w.Close()
	}()

	result := ConfirmDelete("issue", 2, false)

	w.Close()
	outW.Close()
	out, _ := io.ReadAll(outR)

	os.Stdin = oldStdin
	os.Stdout = oldStdout

	if result {
		t.Error("expected false for 'n' input")
	}
	if !strings.Contains(string(out), "Canceled") {
		t.Error("expected 'Canceled' output")
	}
}

func TestConfirmDelete_EmptyInput(t *testing.T) {
	oldStdin := os.Stdin
	oldStdout := os.Stdout

	r, w, _ := os.Pipe()
	os.Stdin = r

	outR, outW, _ := os.Pipe()
	os.Stdout = outW

	go func() {
		w.Close()
	}()

	result := ConfirmDelete("issue", 3, false)

	outW.Close()
	out, _ := io.ReadAll(outR)

	os.Stdin = oldStdin
	os.Stdout = oldStdout

	if result {
		t.Error("expected false for empty input (Scanln error)")
	}
	if !strings.Contains(string(out), "Canceled") {
		t.Error("expected 'Canceled' output")
	}
}
