// internal/resources/issues/commands_test.go
package issues

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/largeoliu/redmine-cli/internal/client"
	"github.com/largeoliu/redmine-cli/internal/resources/trackers"
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

	if cmd.Use != "issue" {
		t.Errorf("expected Use 'issue', got %s", cmd.Use)
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

	response := IssueList{
		Issues: []Issue{
			{ID: 1, Subject: "Issue 1", Project: &Reference{ID: 1, Name: "Project A"}},
			{ID: 2, Subject: "Issue 2", Project: &Reference{ID: 1, Name: "Project A"}},
		},
		TotalCount: 2,
	}
	mock.HandleJSON("/issues.json", response)

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
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListCommand_TrackerNameFilter(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.HandleJSON("/trackers.json", trackers.TrackerList{
		Trackers: []trackers.Tracker{
			{ID: 1, Name: "需求"},
			{ID: 2, Name: "缺陷"},
		},
	})
	mock.Handle("/issues.json", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("tracker_id") != "1" {
			t.Fatalf("expected tracker_id=1, got %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(sampleIssueList())
	})

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
	cmd.SetArgs([]string{"--tracker", "需求"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListCommand_TrackerAllSkipsLookup(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.Handle("/issues.json", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("tracker_id"); got != "" {
			t.Fatalf("expected no tracker_id, got %s", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(sampleIssueList())
	})

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
	cmd.SetArgs([]string{"--tracker", "全部"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
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
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error from ResolveClient, got nil")
	}
}

func TestGetCommand_Success(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	response := struct {
		Issue Issue `json:"issue"`
	}{
		Issue: Issue{
			ID:          1,
			Subject:     "Test Issue",
			Description: "Test description",
			Project:     &Reference{ID: 1, Name: "Project A"},
		},
	}
	mock.HandleJSON("/issues/1.json", response)

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
		Issue Issue `json:"issue"`
	}{
		Issue: Issue{
			ID:      1,
			Subject: "New Issue",
			Project: &Reference{ID: 1, Name: "Project A"},
		},
	}
	mock.HandleJSON("/issues.json", response)

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
	cmd.SetArgs([]string{
		"--project-id", "1",
		"--subject", "New Issue",
	})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateCommand_MissingProjectID(t *testing.T) {
	flags := &types.GlobalFlags{}
	resolver := &mockResolver{}

	cmd := newCreateCommand(flags, resolver)
	cmd.SetArgs([]string{"--subject", "New Issue"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing project-id, got nil")
	}
}

func TestCreateCommand_MissingSubject(t *testing.T) {
	flags := &types.GlobalFlags{}
	resolver := &mockResolver{}

	cmd := newCreateCommand(flags, resolver)
	cmd.SetArgs([]string{"--project-id", "1"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing subject, got nil")
	}
}

func TestCreateCommand_DryRun(t *testing.T) {
	flags := &types.GlobalFlags{DryRun: true}
	resolver := &mockResolver{}

	cmd := newCreateCommand(flags, resolver)
	cmd.SetArgs([]string{
		"--project-id", "1",
		"--subject", "New Issue",
	})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateCommand_Success(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.Handle("/issues/1.json", func(w http.ResponseWriter, r *http.Request) {
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
	cmd.SetArgs([]string{"1", "--subject", "Updated Issue"})

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
	cmd.SetArgs([]string{"1", "--subject", "Updated Issue"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteCommand_Success(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.Handle("/issues/1.json", func(w http.ResponseWriter, r *http.Request) {
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

// 补充测试用例以达到100%覆盖率

func TestListCommand_APIError(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.HandleError("/issues.json", http.StatusInternalServerError, "Internal Server Error")

	flags := &types.GlobalFlags{}
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) {
			return client.NewClient(mock.URL, "test-key"), nil
		},
	}

	cmd := newListCommand(flags, resolver)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error from API, got nil")
	}
}

func TestListCommand_WithGlobalLimit(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	response := IssueList{
		Issues:     []Issue{},
		TotalCount: 0,
	}
	mock.HandleJSON("/issues.json", response)

	flags := &types.GlobalFlags{Limit: 10}
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) {
			return client.NewClient(mock.URL, "test-key"), nil
		},
		writeOutputFunc: func(_ io.Writer, _ *types.GlobalFlags, _ any) error {
			return nil
		},
	}

	cmd := newListCommand(flags, resolver)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListCommand_WithGlobalOffset(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	response := IssueList{
		Issues:     []Issue{},
		TotalCount: 0,
	}
	mock.HandleJSON("/issues.json", response)

	flags := &types.GlobalFlags{Offset: 20}
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) {
			return client.NewClient(mock.URL, "test-key"), nil
		},
		writeOutputFunc: func(_ io.Writer, _ *types.GlobalFlags, _ any) error {
			return nil
		},
	}

	cmd := newListCommand(flags, resolver)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetCommand_WithInclude(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	response := struct {
		Issue Issue `json:"issue"`
	}{
		Issue: Issue{
			ID:      1,
			Subject: "Test Issue",
		},
	}
	mock.HandleJSON("/issues/1.json", response)

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
	cmd.SetArgs([]string{"1", "--include", "children,attachments"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
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

	mock.HandleError("/issues/1.json", http.StatusNotFound, "Issue not found")

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

func TestCreateCommand_ResolveClientError(t *testing.T) {
	flags := &types.GlobalFlags{}
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) {
			return nil, context.Canceled
		},
	}

	cmd := newCreateCommand(flags, resolver)
	cmd.SetArgs([]string{"--project-id", "1", "--subject", "New Issue"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error from ResolveClient, got nil")
	}
}

func TestCreateCommand_APIError(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.HandleError("/issues.json", http.StatusBadRequest, "Bad Request")

	flags := &types.GlobalFlags{}
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) {
			return client.NewClient(mock.URL, "test-key"), nil
		},
	}

	cmd := newCreateCommand(flags, resolver)
	cmd.SetArgs([]string{"--project-id", "1", "--subject", "New Issue"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error from API, got nil")
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
	cmd.SetArgs([]string{"1", "--subject", "Updated Issue"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error from ResolveClient, got nil")
	}
}

func TestUpdateCommand_APIError(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.HandleError("/issues/1.json", http.StatusNotFound, "Issue not found")

	flags := &types.GlobalFlags{}
	resolver := &mockResolver{
		resolveClientFunc: func(_ *types.GlobalFlags) (*client.Client, error) {
			return client.NewClient(mock.URL, "test-key"), nil
		},
	}

	cmd := newUpdateCommand(flags, resolver)
	cmd.SetArgs([]string{"1", "--subject", "Updated Issue"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error from API, got nil")
	}
}

func TestDeleteCommand_ConfirmationWithYes(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.Handle("/issues/1.json", func(w http.ResponseWriter, r *http.Request) {
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

	// 模拟用户输入 'y'
	input := "y\n"
	r, w, _ := os.Pipe()
	_, _ = w.WriteString(input)
	w.Close()

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	os.Stdin = r

	cmd := newDeleteCommand(flags, resolver)
	cmd.SetArgs([]string{"1"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteCommand_ConfirmationWithNo(t *testing.T) {
	flags := &types.GlobalFlags{Yes: false}
	resolver := &mockResolver{}

	// 模拟用户输入 'n'
	input := "n\n"
	r, w, _ := os.Pipe()
	_, _ = w.WriteString(input)
	w.Close()

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	os.Stdin = r

	cmd := newDeleteCommand(flags, resolver)
	cmd.SetArgs([]string{"1"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteCommand_ConfirmationWithYesUppercase(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.Handle("/issues/1.json", func(w http.ResponseWriter, r *http.Request) {
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

	// 模拟用户输入 'Y'
	input := "Y\n"
	r, w, _ := os.Pipe()
	_, _ = w.WriteString(input)
	w.Close()

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	os.Stdin = r

	cmd := newDeleteCommand(flags, resolver)
	cmd.SetArgs([]string{"1"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
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

func TestDeleteCommand_APIError(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.HandleError("/issues/1.json", http.StatusNotFound, "Issue not found")

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
