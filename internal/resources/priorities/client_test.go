// internal/resources/priorities/client_test.go
package priorities

import (
	"context"
	"net/http"
	"testing"

	"github.com/largeoliu/redmine-cli/internal/client"
	"github.com/largeoliu/redmine-cli/internal/testutil"
)

func TestNewClient(t *testing.T) {
	c := client.NewClient("https://example.com", "test-key")
	priorityClient := NewClient(c)
	if priorityClient == nil {
		t.Error("expected client to be created")
	}
}

func TestClient_List_Success(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	response := PriorityList{
		Priorities: []Priority{
			{ID: 1, Name: "Low", IsDefault: false, Position: 1},
			{ID: 2, Name: "Normal", IsDefault: true, Position: 2},
			{ID: 3, Name: "High", IsDefault: false, Position: 3},
		},
	}
	mock.HandleJSON("/enumerations/issue_priorities.json", response)

	c := client.NewClient(mock.URL, "test-key")
	priorityClient := NewClient(c)

	result, err := priorityClient.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Priorities) != 3 {
		t.Errorf("expected 3 priorities, got %d", len(result.Priorities))
	}

	// Verify default priority
	foundDefault := false
	for _, p := range result.Priorities {
		if p.IsDefault {
			foundDefault = true
			if p.Name != "Normal" {
				t.Errorf("expected default priority name 'Normal', got %s", p.Name)
			}
		}
	}
	if !foundDefault {
		t.Error("expected to find a default priority")
	}
}

func TestClient_List_Empty(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	response := PriorityList{
		Priorities: []Priority{},
	}
	mock.HandleJSON("/enumerations/issue_priorities.json", response)

	c := client.NewClient(mock.URL, "test-key")
	priorityClient := NewClient(c)

	result, err := priorityClient.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Priorities) != 0 {
		t.Errorf("expected 0 priorities, got %d", len(result.Priorities))
	}
}

func TestClient_List_Error(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.HandleError("/enumerations/issue_priorities.json", http.StatusUnauthorized, "Unauthorized")

	c := client.NewClient(mock.URL, "test-key")
	priorityClient := NewClient(c)

	_, err := priorityClient.List(context.Background())
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestClient_List_ServerError(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.HandleError("/enumerations/issue_priorities.json", http.StatusInternalServerError, "Internal server error")

	c := client.NewClient(mock.URL, "test-key")
	priorityClient := NewClient(c)

	_, err := priorityClient.List(context.Background())
	if err == nil {
		t.Error("expected error, got nil")
	}
}
