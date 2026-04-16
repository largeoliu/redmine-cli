// internal/resources/priorities/commands_test.go
package priorities

import (
	"context"
	"io"
	"net/http"
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

	if cmd.Use != "priority" {
		t.Errorf("expected Use 'priority', got %s", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description, got empty")
	}

	// 检查子命令
	expectedCommands := []string{"list"}
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

	response := PriorityList{
		Priorities: []Priority{
			{ID: 1, Name: "Normal"},
			{ID: 2, Name: "High"},
			{ID: 3, Name: "Low"},
		},
	}
	mock.HandleJSON("/enumerations/issue_priorities.json", response)

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

func TestListCommand_APIError(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.HandleError("/enumerations/issue_priorities.json", http.StatusInternalServerError, "Internal Server Error")

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

func TestListCommand_WriteOutputError(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	response := PriorityList{
		Priorities: []Priority{
			{ID: 1, Name: "Normal"},
		},
	}
	mock.HandleJSON("/enumerations/issue_priorities.json", response)

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
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error from WriteOutput, got nil")
	}
}
