// internal/resources/versions/commands_test.go
package versions

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/largeoliu/redmine-cli/internal/client"
	"github.com/largeoliu/redmine-cli/internal/testutil"
	"github.com/largeoliu/redmine-cli/internal/types"
)

// mockResolver 实现 types.Resolver 接口
type mockResolver struct {
	resolveClientFunc func(flags *types.GlobalFlags) (*client.Client, error)
	writeOutputFunc   func(w io.Writer, flags *types.GlobalFlags, payload any) error
}

func (m *mockResolver) ResolveClient(flags *types.GlobalFlags) (*client.Client, error) {
	if m.resolveClientFunc != nil {
		return m.resolveClientFunc(flags)
	}
	return nil, nil
}

func (m *mockResolver) WriteOutput(w io.Writer, flags *types.GlobalFlags, payload any) error {
	if m.writeOutputFunc != nil {
		return m.writeOutputFunc(w, flags, payload)
	}
	return nil
}

func TestNewCommand(t *testing.T) {
	flags := &types.GlobalFlags{}
	resolver := &mockResolver{}

	cmd := NewCommand(flags, resolver)

	if cmd == nil {
		t.Fatal("expected command, got nil")
	}

	if cmd.Use != "version" {
		t.Errorf("expected Use 'version', got %s", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description, got empty")
	}

	// 检查子命令
	expectedCommands := []string{"list", "get", "create", "update", "delete"}
	commands := cmd.Commands()
	commandNames := make(map[string]bool)
	for _, c := range commands {
		commandNames[c.Name()] = true
	}

	for _, expected := range expectedCommands {
		if !commandNames[expected] {
			t.Errorf("expected subcommand %s not found", expected)
		}
	}
}

func TestListCommand_Success(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	response := VersionList{
		Versions: []Version{
			{ID: 1, Name: "v1.0", Status: "open"},
			{ID: 2, Name: "v2.0", Status: "closed"},
		},
		TotalCount: 2,
	}
	mock.HandleJSON("/projects/1/versions.json", response)

	flags := &types.GlobalFlags{}
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) {
			return client.NewClient(mock.URL, "test-key"), nil
		},
		writeOutputFunc: func(_ io.Writer, _ *types.GlobalFlags, _ any) error {
			return nil
		},
	}

	cmd := newListCommand(flags, resolver)
	cmd.SetArgs([]string{"--project-id", "1"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListCommand_MissingProjectID(t *testing.T) {
	flags := &types.GlobalFlags{}
	resolver := &mockResolver{}

	cmd := newListCommand(flags, resolver)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing project-id, got nil")
	}
}

func TestListCommand_ZeroProjectID(t *testing.T) {
	flags := &types.GlobalFlags{}
	resolver := &mockResolver{}

	cmd := newListCommand(flags, resolver)
	cmd.SetArgs([]string{"--project-id", "0"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for zero project-id, got nil")
	}
}

func TestListCommand_ResolveClientError(t *testing.T) {
	flags := &types.GlobalFlags{}
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) {
			return nil, context.Canceled
		},
	}

	cmd := newListCommand(flags, resolver)
	cmd.SetArgs([]string{"--project-id", "1"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error from ResolveClient, got nil")
	}
}

func TestListCommand_WithLimit(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	response := VersionList{
		Versions: []Version{
			{ID: 1, Name: "v1.0", Status: "open"},
		},
		TotalCount: 1,
	}
	mock.HandleJSON("/projects/1/versions.json", response)

	flags := &types.GlobalFlags{}
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) {
			return client.NewClient(mock.URL, "test-key"), nil
		},
		writeOutputFunc: func(_ io.Writer, _ *types.GlobalFlags, _ any) error {
			return nil
		},
	}

	cmd := newListCommand(flags, resolver)
	cmd.SetArgs([]string{"--project-id", "1", "--limit", "10"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListCommand_WithOffset(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	response := VersionList{
		Versions: []Version{
			{ID: 1, Name: "v1.0", Status: "open"},
		},
		TotalCount: 1,
	}
	mock.HandleJSON("/projects/1/versions.json", response)

	flags := &types.GlobalFlags{}
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) {
			return client.NewClient(mock.URL, "test-key"), nil
		},
		writeOutputFunc: func(_ io.Writer, _ *types.GlobalFlags, _ any) error {
			return nil
		},
	}

	cmd := newListCommand(flags, resolver)
	cmd.SetArgs([]string{"--project-id", "1", "--offset", "10"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetCommand_Success(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	response := struct {
		Version Version `json:"version"`
	}{
		Version: Version{
			ID:          1,
			Name:        "v1.0",
			Status:      "open",
			Description: "First version",
		},
	}
	mock.HandleJSON("/versions/1.json", response)

	flags := &types.GlobalFlags{}
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) {
			return client.NewClient(mock.URL, "test-key"), nil
		},
		writeOutputFunc: func(_ io.Writer, _ *types.GlobalFlags, _ any) error {
			return nil
		},
	}

	cmd := newGetCommand(flags, resolver)
	cmd.SetArgs([]string{"1"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetCommand_InvalidID(t *testing.T) {
	flags := &types.GlobalFlags{}
	resolver := &mockResolver{}

	cmd := newGetCommand(flags, resolver)
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for invalid ID, got nil")
	}
}

func TestCreateCommand_Success(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	response := struct {
		Version Version `json:"version"`
	}{
		Version: Version{
			ID:     1,
			Name:   "v1.0",
			Status: "open",
		},
	}
	mock.HandleJSON("/projects/1/versions.json", response)

	flags := &types.GlobalFlags{}
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) {
			return client.NewClient(mock.URL, "test-key"), nil
		},
		writeOutputFunc: func(_ io.Writer, _ *types.GlobalFlags, _ any) error {
			return nil
		},
	}

	var buf bytes.Buffer
	cmd := newCreateCommand(flags, resolver)
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--project-id", "1", "--name", "v1.0"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateCommand_MissingProjectID(t *testing.T) {
	flags := &types.GlobalFlags{}
	resolver := &mockResolver{}

	cmd := newCreateCommand(flags, resolver)
	cmd.SetArgs([]string{"--name", "v1.0"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing project-id, got nil")
	}
}

func TestCreateCommand_ZeroProjectID(t *testing.T) {
	flags := &types.GlobalFlags{}
	resolver := &mockResolver{}

	cmd := newCreateCommand(flags, resolver)
	cmd.SetArgs([]string{"--project-id", "0", "--name", "v1.0"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for zero project-id, got nil")
	}
}

func TestCreateCommand_MissingName(t *testing.T) {
	flags := &types.GlobalFlags{}
	resolver := &mockResolver{}

	cmd := newCreateCommand(flags, resolver)
	cmd.SetArgs([]string{"--project-id", "1"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing name, got nil")
	}
}

func TestCreateCommand_DryRun(t *testing.T) {
	flags := &types.GlobalFlags{DryRun: true}
	resolver := &mockResolver{}

	cmd := newCreateCommand(flags, resolver)
	cmd.SetArgs([]string{"--project-id", "1", "--name", "v1.0"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateCommand_ResolveClientError(t *testing.T) {
	flags := &types.GlobalFlags{}
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) {
			return nil, context.Canceled
		},
	}

	cmd := newCreateCommand(flags, resolver)
	cmd.SetArgs([]string{"--project-id", "1", "--name", "v1.0"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error from ResolveClient, got nil")
	}
}

func TestUpdateCommand_Success(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.Handle("/versions/1.json", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	flags := &types.GlobalFlags{}
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) {
			return client.NewClient(mock.URL, "test-key"), nil
		},
	}

	var buf bytes.Buffer
	cmd := newUpdateCommand(flags, resolver)
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"1", "--name", "v1.1"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateCommand_InvalidID(t *testing.T) {
	flags := &types.GlobalFlags{}
	resolver := &mockResolver{}

	cmd := newUpdateCommand(flags, resolver)
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for invalid ID, got nil")
	}
}

func TestUpdateCommand_DryRun(t *testing.T) {
	flags := &types.GlobalFlags{DryRun: true}
	resolver := &mockResolver{}

	cmd := newUpdateCommand(flags, resolver)
	cmd.SetArgs([]string{"1", "--name", "v1.1"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateCommand_ResolveClientError(t *testing.T) {
	flags := &types.GlobalFlags{}
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) {
			return nil, context.Canceled
		},
	}

	cmd := newUpdateCommand(flags, resolver)
	cmd.SetArgs([]string{"1", "--name", "v1.1"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error from ResolveClient, got nil")
	}
}

func TestDeleteCommand_Success(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.Handle("/versions/1.json", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	flags := &types.GlobalFlags{Yes: true}
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) {
			return client.NewClient(mock.URL, "test-key"), nil
		},
	}

	var buf bytes.Buffer
	cmd := newDeleteCommand(flags, resolver)
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"1"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteCommand_InvalidID(t *testing.T) {
	flags := &types.GlobalFlags{Yes: true}
	resolver := &mockResolver{}

	cmd := newDeleteCommand(flags, resolver)
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for invalid ID, got nil")
	}
}

func TestDeleteCommand_DryRun(t *testing.T) {
	flags := &types.GlobalFlags{DryRun: true, Yes: true}
	resolver := &mockResolver{}

	cmd := newDeleteCommand(flags, resolver)
	cmd.SetArgs([]string{"1"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteCommand_ConfirmationPrompt(t *testing.T) {
	// Skip this test as it requires interactive input
	// The confirmation prompt functionality is tested through integration tests
	t.Skip("skipping test that requires interactive input")
}

func TestDeleteCommand_ConfirmationPromptUppercaseY(t *testing.T) {
	// Skip this test as it requires interactive input
	// The confirmation prompt functionality is tested through integration tests
	t.Skip("skipping test that requires interactive input")
}

func TestDeleteCommand_ConfirmationPromptCancelled(t *testing.T) {
	// Skip this test as it requires interactive input
	// The confirmation prompt functionality is tested through integration tests
	t.Skip("skipping test that requires interactive input")
}

func TestDeleteCommand_ResolveClientError(t *testing.T) {
	flags := &types.GlobalFlags{Yes: true}
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) {
			return nil, context.Canceled
		},
	}

	cmd := newDeleteCommand(flags, resolver)
	cmd.SetArgs([]string{"1"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error from ResolveClient, got nil")
	}
}

func TestListCommand_APIError(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.HandleError("/projects/1/versions.json", http.StatusInternalServerError, "Internal Server Error")

	flags := &types.GlobalFlags{}
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) {
			return client.NewClient(mock.URL, "test-key"), nil
		},
	}

	cmd := newListCommand(flags, resolver)
	cmd.SetArgs([]string{"--project-id", "1"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error from API, got nil")
	}
}

func TestListCommand_WriteOutputError(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	response := VersionList{
		Versions:   []Version{},
		TotalCount: 0,
	}
	mock.HandleJSON("/projects/1/versions.json", response)

	flags := &types.GlobalFlags{}
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) {
			return client.NewClient(mock.URL, "test-key"), nil
		},
		writeOutputFunc: func(_ io.Writer, _ *types.GlobalFlags, _ any) error {
			return context.Canceled
		},
	}

	cmd := newListCommand(flags, resolver)
	cmd.SetArgs([]string{"--project-id", "1"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error from WriteOutput, got nil")
	}
}

func TestGetCommand_ResolveClientError(t *testing.T) {
	flags := &types.GlobalFlags{}
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) {
			return nil, context.Canceled
		},
	}

	cmd := newGetCommand(flags, resolver)
	cmd.SetArgs([]string{"1"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error from ResolveClient, got nil")
	}
}

func TestGetCommand_APIError(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.HandleError("/versions/1.json", http.StatusNotFound, "Not found")

	flags := &types.GlobalFlags{}
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) {
			return client.NewClient(mock.URL, "test-key"), nil
		},
	}

	cmd := newGetCommand(flags, resolver)
	cmd.SetArgs([]string{"1"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error from API, got nil")
	}
}

func TestGetCommand_WriteOutputError(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	response := struct {
		Version Version `json:"version"`
	}{
		Version: Version{
			ID:     1,
			Name:   "v1.0",
			Status: "open",
		},
	}
	mock.HandleJSON("/versions/1.json", response)

	flags := &types.GlobalFlags{}
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) {
			return client.NewClient(mock.URL, "test-key"), nil
		},
		writeOutputFunc: func(_ io.Writer, _ *types.GlobalFlags, _ any) error {
			return context.Canceled
		},
	}

	cmd := newGetCommand(flags, resolver)
	cmd.SetArgs([]string{"1"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error from WriteOutput, got nil")
	}
}

func TestCreateCommand_APIError(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.HandleError("/projects/1/versions.json", http.StatusBadRequest, "Bad Request")

	flags := &types.GlobalFlags{}
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) {
			return client.NewClient(mock.URL, "test-key"), nil
		},
	}

	cmd := newCreateCommand(flags, resolver)
	cmd.SetArgs([]string{"--project-id", "1", "--name", "v1.0"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error from API, got nil")
	}
}

func TestCreateCommand_WriteOutputError(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	response := struct {
		Version Version `json:"version"`
	}{
		Version: Version{
			ID:     1,
			Name:   "v1.0",
			Status: "open",
		},
	}
	mock.HandleJSON("/projects/1/versions.json", response)

	flags := &types.GlobalFlags{}
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) {
			return client.NewClient(mock.URL, "test-key"), nil
		},
		writeOutputFunc: func(_ io.Writer, _ *types.GlobalFlags, _ any) error {
			return context.Canceled
		},
	}

	cmd := newCreateCommand(flags, resolver)
	cmd.SetArgs([]string{"--project-id", "1", "--name", "v1.0"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error from WriteOutput, got nil")
	}
}

func TestUpdateCommand_APIError(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.HandleError("/versions/1.json", http.StatusNotFound, "Not found")

	flags := &types.GlobalFlags{}
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) {
			return client.NewClient(mock.URL, "test-key"), nil
		},
	}

	cmd := newUpdateCommand(flags, resolver)
	cmd.SetArgs([]string{"1", "--name", "v1.1"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error from API, got nil")
	}
}

func TestDeleteCommand_APIError(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.HandleError("/versions/1.json", http.StatusNotFound, "Not found")

	flags := &types.GlobalFlags{Yes: true}
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) {
			return client.NewClient(mock.URL, "test-key"), nil
		},
	}

	cmd := newDeleteCommand(flags, resolver)
	cmd.SetArgs([]string{"1"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error from API, got nil")
	}
}

func TestDeleteCommand_ConfirmDecline(t *testing.T) {
	flags := &types.GlobalFlags{Yes: false}
	resolver := &mockResolver{}

	input := "n\n"
	r, w, _ := os.Pipe()
	_, _ = w.WriteString(input)
	w.Close()

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	os.Stdin = r

	var buf bytes.Buffer
	cmd := newDeleteCommand(flags, resolver)
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"1"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteCommand_ConfirmAccept(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.Handle("/versions/1.json", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	flags := &types.GlobalFlags{Yes: false}
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) {
			return client.NewClient(mock.URL, "test-key"), nil
		},
	}

	input := "y\n"
	r, w, _ := os.Pipe()
	_, _ = w.WriteString(input)
	w.Close()

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	os.Stdin = r

	var buf bytes.Buffer
	cmd := newDeleteCommand(flags, resolver)
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"1"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListCommand_WithLimitAndOffset(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	response := VersionList{
		Versions:   []Version{},
		TotalCount: 0,
	}
	mock.HandleJSON("/projects/1/versions.json", response)

	flags := &types.GlobalFlags{}
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) {
			return client.NewClient(mock.URL, "test-key"), nil
		},
		writeOutputFunc: func(_ io.Writer, _ *types.GlobalFlags, _ any) error {
			return nil
		},
	}

	cmd := newListCommand(flags, resolver)
	cmd.SetArgs([]string{"--project-id", "1", "--limit", "10", "--offset", "5"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
