// internal/resources/statuses/client_test.go
package statuses

import (
	"context"
	"net/http"
	"testing"

	"github.com/largeoliu/redmine-cli/internal/client"
	"github.com/largeoliu/redmine-cli/internal/testutil"
)

func TestNewClient(t *testing.T) {
	c := client.NewClient("https://example.com", "test-key")
	statusClient := NewClient(c)
	if statusClient == nil {
		t.Error("expected client to be created")
	}
}

func TestClient_List_Success(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	response := IssueStatusList{
		IssueStatuses: []IssueStatus{
			{ID: 1, Name: "New", IsClosed: false, IsDefault: true, Position: 1},
			{ID: 2, Name: "In Progress", IsClosed: false, IsDefault: false, Position: 2},
			{ID: 3, Name: "Resolved", IsClosed: true, IsDefault: false, Position: 3},
			{ID: 4, Name: "Closed", IsClosed: true, IsDefault: false, Position: 4},
		},
	}
	mock.HandleJSON("/issue_statuses.json", response)

	c := client.NewClient(mock.URL, "test-key")
	statusClient := NewClient(c)

	result, err := statusClient.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.IssueStatuses) != 4 {
		t.Errorf("expected 4 statuses, got %d", len(result.IssueStatuses))
	}

	// Verify default status
	foundDefault := false
	for _, s := range result.IssueStatuses {
		if s.IsDefault {
			foundDefault = true
			if s.Name != "New" {
				t.Errorf("expected default status name 'New', got %s", s.Name)
			}
		}
	}
	if !foundDefault {
		t.Error("expected to find a default status")
	}

	// Verify closed statuses
	closedCount := 0
	for _, s := range result.IssueStatuses {
		if s.IsClosed {
			closedCount++
		}
	}
	if closedCount != 2 {
		t.Errorf("expected 2 closed statuses, got %d", closedCount)
	}
}

func TestClient_List_Empty(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	response := IssueStatusList{
		IssueStatuses: []IssueStatus{},
	}
	mock.HandleJSON("/issue_statuses.json", response)

	c := client.NewClient(mock.URL, "test-key")
	statusClient := NewClient(c)

	result, err := statusClient.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.IssueStatuses) != 0 {
		t.Errorf("expected 0 statuses, got %d", len(result.IssueStatuses))
	}
}

func TestClient_List_Error(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.HandleError("/issue_statuses.json", http.StatusUnauthorized, "Unauthorized")

	c := client.NewClient(mock.URL, "test-key")
	statusClient := NewClient(c)

	_, err := statusClient.List(context.Background())
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestClient_List_Forbidden(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.HandleError("/issue_statuses.json", http.StatusForbidden, "Permission denied")

	c := client.NewClient(mock.URL, "test-key")
	statusClient := NewClient(c)

	_, err := statusClient.List(context.Background())
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestClient_List_ServerError(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.HandleError("/issue_statuses.json", http.StatusInternalServerError, "Internal server error")

	c := client.NewClient(mock.URL, "test-key")
	statusClient := NewClient(c)

	_, err := statusClient.List(context.Background())
	if err == nil {
		t.Error("expected error, got nil")
	}
}
