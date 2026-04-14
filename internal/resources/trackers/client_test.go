// internal/resources/trackers/client_test.go
package trackers

import (
	"context"
	"net/http"
	"testing"

	"github.com/largeoliu/redmine-cli/internal/client"
	"github.com/largeoliu/redmine-cli/internal/testutil"
)

func TestNewClient(t *testing.T) {
	c := client.NewClient("https://example.com", "test-key")
	trackerClient := NewClient(c)
	if trackerClient == nil {
		t.Error("expected client to be created")
	}
}

func TestClient_List_Success(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	defaultStatus := 1
	response := TrackerList{
		Trackers: []Tracker{
			{ID: 1, Name: "Bug", DefaultStatus: &defaultStatus, Description: "Bug tracking"},
			{ID: 2, Name: "Feature", DefaultStatus: nil, Description: "Feature requests"},
			{ID: 3, Name: "Support", DefaultStatus: &defaultStatus, Description: ""},
		},
	}
	mock.HandleJSON("/trackers.json", response)

	c := client.NewClient(mock.URL, "test-key")
	trackerClient := NewClient(c)

	result, err := trackerClient.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Trackers) != 3 {
		t.Errorf("expected 3 trackers, got %d", len(result.Trackers))
	}

	// Verify first tracker
	if result.Trackers[0].Name != "Bug" {
		t.Errorf("expected first tracker name 'Bug', got %s", result.Trackers[0].Name)
	}
	if result.Trackers[0].Description != "Bug tracking" {
		t.Errorf("expected first tracker description 'Bug tracking', got %s", result.Trackers[0].Description)
	}
	if result.Trackers[0].DefaultStatus == nil || *result.Trackers[0].DefaultStatus != 1 {
		t.Error("expected first tracker default status to be 1")
	}

	// Verify second tracker has nil default status
	if result.Trackers[1].DefaultStatus != nil {
		t.Error("expected second tracker default status to be nil")
	}
}

func TestClient_List_Empty(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	response := TrackerList{
		Trackers: []Tracker{},
	}
	mock.HandleJSON("/trackers.json", response)

	c := client.NewClient(mock.URL, "test-key")
	trackerClient := NewClient(c)

	result, err := trackerClient.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Trackers) != 0 {
		t.Errorf("expected 0 trackers, got %d", len(result.Trackers))
	}
}

func TestClient_List_Error(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.HandleError("/trackers.json", http.StatusUnauthorized, "Unauthorized")

	c := client.NewClient(mock.URL, "test-key")
	trackerClient := NewClient(c)

	_, err := trackerClient.List(context.Background())
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestClient_List_Forbidden(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.HandleError("/trackers.json", http.StatusForbidden, "Permission denied")

	c := client.NewClient(mock.URL, "test-key")
	trackerClient := NewClient(c)

	_, err := trackerClient.List(context.Background())
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestClient_List_ServerError(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.HandleError("/trackers.json", http.StatusInternalServerError, "Internal server error")

	c := client.NewClient(mock.URL, "test-key")
	trackerClient := NewClient(c)

	_, err := trackerClient.List(context.Background())
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestClient_List_NotFound(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.HandleError("/trackers.json", http.StatusNotFound, "Not found")

	c := client.NewClient(mock.URL, "test-key")
	trackerClient := NewClient(c)

	_, err := trackerClient.List(context.Background())
	if err == nil {
		t.Error("expected error, got nil")
	}
}
