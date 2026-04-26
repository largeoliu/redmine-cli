package agile

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
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

func TestClient_ListSprintsWithSprintsField(t *testing.T) {
	c := client.NewClient("https://example.com", "test-key", client.WithHTTPClient(&http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.Path != "/projects/42/agile_sprints.json" {
				t.Fatalf("unexpected path: %s", req.URL.Path)
			}
			return jsonHTTPResponse(t, http.StatusOK, map[string]any{
				"project_id":   42,
				"project_name": "City",
				"sprints": []map[string]any{
					{"id": 7, "name": "Sprint 7", "status": "active"},
					{"id": 8, "name": "Sprint 8", "status": "open"},
				},
			}), nil
		}),
	}))
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

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func jsonHTTPResponse(t *testing.T, status int, payload any) *http.Response {
	t.Helper()

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	return &http.Response{
		StatusCode: status,
		Status:     http.StatusText(status),
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(data)),
	}
}

func TestClient_GetSprint(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	response := struct {
		AgileSprint Sprint `json:"agile_sprint"`
	}{
		AgileSprint: Sprint{
			ID:          7,
			Name:        "Sprint 7",
			Description: "Release hardening",
			Status:      "active",
			StartDate:   "2026-04-01",
			EndDate:     "2026-04-14",
			Goal:        "Finish release scope",
			IsDefault:   true,
			IsClosed:    false,
			IsArchived:  false,
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
	if result.Description != "Release hardening" {
		t.Fatalf("expected description to be populated, got %q", result.Description)
	}
}

func TestClient_GetIssueAgileData(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	response := struct {
		AgileData Data `json:"agile_data"`
	}{
		AgileData: Data{
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

func TestClient_GetSprintNotFound(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.HandleError("/projects/42/agile_sprints/7.json", http.StatusNotFound, "Not found")

	c := client.NewClient(mock.URL, "test-key")
	agileClient := NewClient(c)

	_, err := agileClient.GetSprint(context.Background(), 42, 7)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_GetIssueAgileDataNotFound(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.HandleError("/issues/9/agile_data.json", http.StatusNotFound, "Not found")

	c := client.NewClient(mock.URL, "test-key")
	agileClient := NewClient(c)

	_, err := agileClient.GetIssueAgileData(context.Background(), 9)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func intPtr(v int) *int {
	return &v
}

func TestSprintList_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name       string
		data       []byte
		wantErr    bool
		verifyFunc func(t *testing.T, list *SprintList)
	}{
		{
			name:    "invalid JSON",
			data:    []byte(`not valid json`),
			wantErr: true,
		},
		{
			name:    "no sprints or agile_sprints field",
			data:    []byte(`{"total_count": 0, "limit": 25, "offset": 0}`),
			wantErr: false,
			verifyFunc: func(t *testing.T, list *SprintList) {
				if len(list.AgileSprints) != 0 {
					t.Errorf("expected 0 sprints, got %d", len(list.AgileSprints))
				}
			},
		},
		{
			name:    "invalid agile_sprints format",
			data:    []byte(`{"agile_sprints": "not an array"}`),
			wantErr: true,
		},
		{
			name:    "invalid sprints format",
			data:    []byte(`{"sprints": "not an array"}`),
			wantErr: true,
		},
		{
			name:    "invalid total_count format",
			data:    []byte(`{"agile_sprints": [], "total_count": "not a number"}`),
			wantErr: true,
		},
		{
			name:    "invalid limit format",
			data:    []byte(`{"agile_sprints": [], "total_count": 0, "limit": "not a number"}`),
			wantErr: true,
		},
		{
			name:    "invalid offset format",
			data:    []byte(`{"agile_sprints": [], "total_count": 0, "limit": 25, "offset": "not a number"}`),
			wantErr: true,
		},
		{
			name:    "with all metadata fields",
			data:    []byte(`{"agile_sprints": [{"id": 1, "name": "Sprint 1"}], "total_count": 100, "limit": 25, "offset": 50}`),
			wantErr: false,
			verifyFunc: func(t *testing.T, list *SprintList) {
				if len(list.AgileSprints) != 1 {
					t.Errorf("expected 1 sprint, got %d", len(list.AgileSprints))
				}
				if list.TotalCount != 100 {
					t.Errorf("expected total_count 100, got %d", list.TotalCount)
				}
				if list.Limit != 25 {
					t.Errorf("expected limit 25, got %d", list.Limit)
				}
				if list.Offset != 50 {
					t.Errorf("expected offset 50, got %d", list.Offset)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var list SprintList
			err := list.UnmarshalJSON(tt.data)

			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && tt.verifyFunc != nil {
				tt.verifyFunc(t, &list)
			}
		})
	}
}
