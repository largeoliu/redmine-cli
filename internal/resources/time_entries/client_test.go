// internal/resources/time_entries/client_test.go
package time_entries

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/largeoliu/redmine-cli/internal/client"
	"github.com/largeoliu/redmine-cli/internal/testutil"
)

func TestNewClient(t *testing.T) {
	c := client.NewClient("https://example.com", "test-key")
	teClient := NewClient(c)
	if teClient == nil {
		t.Error("expected client to be created")
	}
}

func TestClient_List_Success(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	now := time.Now()
	response := TimeEntryList{
		TimeEntries: []TimeEntry{
			{
				ID:        1,
				Hours:     2.5,
				Comments:  "Working on feature",
				SpentOn:   "2024-01-15",
				Project:   &Reference{ID: 1, Name: "Project A"},
				Issue:     &Reference{ID: 10, Name: "Issue 10"},
				User:      &Reference{ID: 1, Name: "User 1"},
				Activity:  &Reference{ID: 1, Name: "Development"},
				CreatedOn: &now,
			},
		},
		TotalCount: 1,
		Limit:      25,
		Offset:     0,
	}
	mock.HandleJSON("/time_entries.json", response)

	c := client.NewClient(mock.URL, "test-key")
	teClient := NewClient(c)

	result, err := teClient.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.TimeEntries) != 1 {
		t.Errorf("expected 1 time entry, got %d", len(result.TimeEntries))
	}
	if result.TimeEntries[0].Hours != 2.5 {
		t.Errorf("expected hours 2.5, got %f", result.TimeEntries[0].Hours)
	}
	if result.TotalCount != 1 {
		t.Errorf("expected total count 1, got %d", result.TotalCount)
	}
}

func TestClient_List_WithParams(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	response := TimeEntryList{
		TimeEntries: []TimeEntry{},
		TotalCount:  0,
		Limit:       25,
		Offset:      0,
	}
	mock.Handle("/time_entries.json", func(w http.ResponseWriter, r *http.Request) {
		// Verify query parameters
		query := r.URL.Query()
		if query.Get("project_id") != "1" {
			t.Errorf("expected project_id=1, got %s", query.Get("project_id"))
		}
		if query.Get("issue_id") != "10" {
			t.Errorf("expected issue_id=10, got %s", query.Get("issue_id"))
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	c := client.NewClient(mock.URL, "test-key")
	teClient := NewClient(c)

	params := map[string]string{
		"project_id": "1",
		"issue_id":   "10",
	}
	_, err := teClient.List(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_List_Empty(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	response := TimeEntryList{
		TimeEntries: []TimeEntry{},
		TotalCount:  0,
		Limit:       25,
		Offset:      0,
	}
	mock.HandleJSON("/time_entries.json", response)

	c := client.NewClient(mock.URL, "test-key")
	teClient := NewClient(c)

	result, err := teClient.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.TimeEntries) != 0 {
		t.Errorf("expected 0 time entries, got %d", len(result.TimeEntries))
	}
}

func TestClient_List_Error(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.HandleError("/time_entries.json", http.StatusUnauthorized, "Unauthorized")

	c := client.NewClient(mock.URL, "test-key")
	teClient := NewClient(c)

	_, err := teClient.List(context.Background(), nil)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestClient_Get_Success(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	response := struct {
		TimeEntry TimeEntry `json:"time_entry"`
	}{
		TimeEntry: TimeEntry{
			ID:       1,
			Hours:    3.0,
			Comments: "Code review",
			SpentOn:  "2024-01-15",
			Project:  &Reference{ID: 1, Name: "Project A"},
			User:     &Reference{ID: 1, Name: "User 1"},
		},
	}
	mock.HandleJSON("/time_entries/1.json", response)

	c := client.NewClient(mock.URL, "test-key")
	teClient := NewClient(c)

	result, err := teClient.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != 1 {
		t.Errorf("expected ID 1, got %d", result.ID)
	}
	if result.Hours != 3.0 {
		t.Errorf("expected hours 3.0, got %f", result.Hours)
	}
}

func TestClient_Get_Error(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.HandleError("/time_entries/999.json", http.StatusNotFound, "Time entry not found")

	c := client.NewClient(mock.URL, "test-key")
	teClient := NewClient(c)

	_, err := teClient.Get(context.Background(), 999)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestClient_Create_Success(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	response := struct {
		TimeEntry TimeEntry `json:"time_entry"`
	}{
		TimeEntry: TimeEntry{
			ID:       1,
			Issue:    &Reference{ID: 10, Name: "Issue 10"},
			Hours:    2.0,
			Comments: "Development work",
			SpentOn:  "2024-01-15",
		},
	}
	mock.HandleJSON("/time_entries.json", response)

	c := client.NewClient(mock.URL, "test-key")
	teClient := NewClient(c)

	req := &TimeEntryCreateRequest{
		IssueID:    10,
		Hours:      2.0,
		Comments:   "Development work",
		SpentOn:    "2024-01-15",
		ActivityID: 1,
	}
	result, err := teClient.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != 1 {
		t.Errorf("expected ID 1, got %d", result.ID)
	}
	if result.Hours != 2.0 {
		t.Errorf("expected hours 2.0, got %f", result.Hours)
	}
}

func TestClient_Create_Error(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.HandleError("/time_entries.json", http.StatusBadRequest, "Invalid request")

	c := client.NewClient(mock.URL, "test-key")
	teClient := NewClient(c)

	req := &TimeEntryCreateRequest{
		Hours: 2.0,
	}
	_, err := teClient.Create(context.Background(), req)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestClient_Update_Success(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.Handle("/time_entries/1.json", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	c := client.NewClient(mock.URL, "test-key")
	teClient := NewClient(c)

	req := &TimeEntryUpdateRequest{
		Hours:    3.0,
		Comments: "Updated comment",
	}
	err := teClient.Update(context.Background(), 1, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_Update_Error(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.HandleError("/time_entries/1.json", http.StatusNotFound, "Time entry not found")

	c := client.NewClient(mock.URL, "test-key")
	teClient := NewClient(c)

	req := &TimeEntryUpdateRequest{
		Hours: 3.0,
	}
	err := teClient.Update(context.Background(), 1, req)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestClient_Delete_Success(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.Handle("/time_entries/1.json", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	c := client.NewClient(mock.URL, "test-key")
	teClient := NewClient(c)

	err := teClient.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_Delete_Error(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.HandleError("/time_entries/1.json", http.StatusNotFound, "Time entry not found")

	c := client.NewClient(mock.URL, "test-key")
	teClient := NewClient(c)

	err := teClient.Delete(context.Background(), 1)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestBuildListParams_AllFields(t *testing.T) {
	flags := ListFlags{
		ProjectID:  1,
		IssueID:    10,
		UserID:     5,
		ActivityID: 2,
		From:       "2024-01-01",
		To:         "2024-01-31",
		Limit:      100,
		Offset:     50,
	}

	params := BuildListParams(flags)

	if params["project_id"] != "1" {
		t.Errorf("expected project_id=1, got %s", params["project_id"])
	}
	if params["issue_id"] != "10" {
		t.Errorf("expected issue_id=10, got %s", params["issue_id"])
	}
	if params["user_id"] != "5" {
		t.Errorf("expected user_id=5, got %s", params["user_id"])
	}
	if params["activity_id"] != "2" {
		t.Errorf("expected activity_id=2, got %s", params["activity_id"])
	}
	if params["from"] != "2024-01-01" {
		t.Errorf("expected from=2024-01-01, got %s", params["from"])
	}
	if params["to"] != "2024-01-31" {
		t.Errorf("expected to=2024-01-31, got %s", params["to"])
	}
	if params["limit"] != "100" {
		t.Errorf("expected limit=100, got %s", params["limit"])
	}
	if params["offset"] != "50" {
		t.Errorf("expected offset=50, got %s", params["offset"])
	}
}

func TestBuildListParams_Empty(t *testing.T) {
	flags := ListFlags{}

	params := BuildListParams(flags)

	if len(params) != 0 {
		t.Errorf("expected 0 params, got %d", len(params))
	}
}

func TestBuildListParams_PartialFields(t *testing.T) {
	flags := ListFlags{
		ProjectID: 1,
		IssueID:   10,
		Limit:     25,
	}

	params := BuildListParams(flags)

	if len(params) != 3 {
		t.Errorf("expected 3 params, got %d", len(params))
	}
	if params["project_id"] != "1" {
		t.Errorf("expected project_id=1, got %s", params["project_id"])
	}
	if params["issue_id"] != "10" {
		t.Errorf("expected issue_id=10, got %s", params["issue_id"])
	}
	if params["limit"] != "25" {
		t.Errorf("expected limit=25, got %s", params["limit"])
	}
}

func TestBuildListParams_ZeroValuesIgnored(t *testing.T) {
	flags := ListFlags{
		ProjectID:  0,
		IssueID:    0,
		UserID:     0,
		ActivityID: 0,
		From:       "",
		To:         "",
		Limit:      0,
		Offset:     0,
	}

	params := BuildListParams(flags)

	if len(params) != 0 {
		t.Errorf("expected 0 params for zero values, got %d", len(params))
	}
}

func TestClient_ListActivities_Success(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	response := struct {
		TimeEntryActivities []TimeEntryActivity `json:"time_entry_activities"`
	}{
		TimeEntryActivities: []TimeEntryActivity{
			{ID: 1, Name: "Development", IsDefault: true},
			{ID: 2, Name: "Design", IsDefault: false},
			{ID: 3, Name: "Testing", IsDefault: false},
		},
	}
	mock.HandleJSON("/enumerations/time_entry_activities.json", response)

	c := client.NewClient(mock.URL, "test-key")
	teClient := NewClient(c)

	result, err := teClient.ListActivities(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("expected 3 activities, got %d", len(result))
	}
	if result[0].Name != "Development" {
		t.Errorf("expected first activity name 'Development', got %s", result[0].Name)
	}
	if !result[0].IsDefault {
		t.Error("expected first activity to be default")
	}
}

func TestClient_ListActivities_Empty(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	response := struct {
		TimeEntryActivities []TimeEntryActivity `json:"time_entry_activities"`
	}{
		TimeEntryActivities: []TimeEntryActivity{},
	}
	mock.HandleJSON("/enumerations/time_entry_activities.json", response)

	c := client.NewClient(mock.URL, "test-key")
	teClient := NewClient(c)

	result, err := teClient.ListActivities(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("expected 0 activities, got %d", len(result))
	}
}

func TestClient_ListActivities_Error(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.HandleError("/enumerations/time_entry_activities.json", http.StatusUnauthorized, "Unauthorized")

	c := client.NewClient(mock.URL, "test-key")
	teClient := NewClient(c)

	_, err := teClient.ListActivities(context.Background())
	if err == nil {
		t.Error("expected error, got nil")
	}
}
