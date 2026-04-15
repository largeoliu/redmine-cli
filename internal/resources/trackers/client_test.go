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

func TestClient_Get_Success(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	response := map[string]any{
		"tracker": map[string]any{
			"id":          1,
			"name":        "Bug",
			"description": "Bug tracking",
			"custom_fields": []map[string]any{
				{
					"id":           5,
					"name":         "优先级",
					"field_format": "list",
					"possible_values": []map[string]string{
						{"value": "high", "label": "高"},
						{"value": "medium", "label": "中"},
						{"value": "low", "label": "低"},
					},
				},
				{
					"id":           6,
					"name":         "修复版本",
					"field_format": "list",
					"possible_values": []map[string]string{
						{"value": "v1.0", "label": "v1.0"},
						{"value": "v2.0", "label": "v2.0"},
					},
				},
				{
					"id":           7,
					"name":         "详细说明",
					"field_format": "text",
				},
				{
					"id":           8,
					"name":         "是否紧急",
					"field_format": "bool",
				},
				{
					"id":           9,
					"name":         "预计日期",
					"field_format": "date",
				},
			},
		},
	}
	mock.HandleJSON("/trackers/1.json", response)

	c := client.NewClient(mock.URL, "test-key")
	trackerClient := NewClient(c)

	result, err := trackerClient.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Name != "Bug" {
		t.Errorf("expected name 'Bug', got %s", result.Name)
	}
	if len(result.CustomFields) != 5 {
		t.Errorf("expected 5 custom fields, got %d", len(result.CustomFields))
	}
	if result.CustomFields[0].FieldFormat != "list" {
		t.Errorf("expected field_format 'list', got %s", result.CustomFields[0].FieldFormat)
	}
	if len(result.CustomFields[0].PossibleValues) != 3 {
		t.Errorf("expected 3 possible values, got %d", len(result.CustomFields[0].PossibleValues))
	}
	if result.CustomFields[0].PossibleValues[0].Value != "high" {
		t.Errorf("expected first value 'high', got %s", result.CustomFields[0].PossibleValues[0].Value)
	}
}

func TestClient_Get_NotFound(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.HandleError("/trackers/999.json", http.StatusNotFound, "Not found")

	c := client.NewClient(mock.URL, "test-key")
	trackerClient := NewClient(c)

	_, err := trackerClient.Get(context.Background(), 999)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestClient_Get_Unauthorized(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.HandleError("/trackers/1.json", http.StatusUnauthorized, "Unauthorized")

	c := client.NewClient(mock.URL, "test-key")
	trackerClient := NewClient(c)

	_, err := trackerClient.Get(context.Background(), 1)
	if err == nil {
		t.Error("expected error, got nil")
	}
}
