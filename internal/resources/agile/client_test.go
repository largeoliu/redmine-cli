package agile

import (
	"context"
	"net/http"
	"testing"

	"github.com/largeoliu/redmine-cli/internal/client"
	"github.com/largeoliu/redmine-cli/internal/testutil"
)

func TestNewClient(t *testing.T) {
	c := client.NewClient("https://example.com", "test-key")
	agileClient := NewClient(c)
	if agileClient == nil {
		t.Fatal("expected client, got nil")
	}
}

func TestClient_ListSprints(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	response := SprintList{
		AgileSprints: []Sprint{
			{ID: 7, Name: "Sprint 7", Status: "active"},
			{ID: 8, Name: "Sprint 8", Status: "open"},
		},
	}
	mock.HandleJSON("/projects/42/agile_sprints.json", response)

	c := client.NewClient(mock.URL, "test-key")
	agileClient := NewClient(c)

	result, err := agileClient.ListSprints(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.AgileSprints) != 2 {
		t.Fatalf("expected 2 sprints, got %d", len(result.AgileSprints))
	}
	if result.AgileSprints[0].Name != "Sprint 7" {
		t.Fatalf("expected first sprint to be Sprint 7, got %s", result.AgileSprints[0].Name)
	}
}

func TestClient_GetSprint(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	response := struct {
		AgileSprint Sprint `json:"agile_sprint"`
	}{
		AgileSprint: Sprint{
			ID:         7,
			Name:       "Sprint 7",
			Status:     "active",
			StartDate:  "2026-04-01",
			EndDate:    "2026-04-14",
			Goal:       "Finish release scope",
			IsDefault:  true,
			IsClosed:   false,
			IsArchived: false,
		},
	}
	mock.HandleJSON("/projects/42/agile_sprints/7.json", response)

	c := client.NewClient(mock.URL, "test-key")
	agileClient := NewClient(c)

	result, err := agileClient.GetSprint(context.Background(), 42, 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != 7 {
		t.Fatalf("expected sprint ID 7, got %d", result.ID)
	}
	if result.StartDate != "2026-04-01" {
		t.Fatalf("expected start date 2026-04-01, got %s", result.StartDate)
	}
}

func TestClient_GetIssueAgileData(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	response := struct {
		AgileData AgileData `json:"agile_data"`
	}{
		AgileData: AgileData{
			AgileSprintID: intPtr(7),
			StoryPoints:   5,
			Position:      12,
		},
	}
	mock.HandleJSON("/issues/9/agile_data.json", response)

	c := client.NewClient(mock.URL, "test-key")
	agileClient := NewClient(c)

	result, err := agileClient.GetIssueAgileData(context.Background(), 9)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AgileSprintID == nil || *result.AgileSprintID != 7 {
		t.Fatalf("expected sprint ID 7, got %v", result.AgileSprintID)
	}
	if result.StoryPoints != 5 {
		t.Fatalf("expected story points 5, got %v", result.StoryPoints)
	}
}

func TestClient_ListSprintsNotFound(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.HandleError("/projects/42/agile_sprints.json", http.StatusNotFound, "Not found")

	c := client.NewClient(mock.URL, "test-key")
	agileClient := NewClient(c)

	_, err := agileClient.ListSprints(context.Background(), 42)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func intPtr(v int) *int {
	return &v
}
