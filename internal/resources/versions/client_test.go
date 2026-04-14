// internal/resources/versions/client_test.go
package versions

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
	versionClient := NewClient(c)
	if versionClient == nil {
		t.Error("expected client to be created")
	}
}

func TestClient_List_Success(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	now := time.Now()
	response := VersionList{
		Versions: []Version{
			{
				ID:          1,
				Name:        "v1.0",
				Description: "First release",
				Status:      "open",
				CreatedOn:   &now,
			},
			{
				ID:          2,
				Name:        "v2.0",
				Description: "Second release",
				Status:      "closed",
				CreatedOn:   &now,
			},
		},
		TotalCount: 2,
		Limit:      25,
		Offset:     0,
	}
	mock.HandleJSON("/projects/1/versions.json", response)

	c := client.NewClient(mock.URL, "test-key")
	versionClient := NewClient(c)

	result, err := versionClient.List(context.Background(), 1, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Versions) != 2 {
		t.Errorf("expected 2 versions, got %d", len(result.Versions))
	}
	if result.Versions[0].Name != "v1.0" {
		t.Errorf("expected first version name 'v1.0', got %s", result.Versions[0].Name)
	}
	if result.TotalCount != 2 {
		t.Errorf("expected total count 2, got %d", result.TotalCount)
	}
}

func TestClient_List_WithParams(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	response := VersionList{
		Versions:   []Version{},
		TotalCount: 0,
		Limit:      25,
		Offset:     0,
	}
	mock.Handle("/projects/1/versions.json", func(w http.ResponseWriter, r *http.Request) {
		// Verify query parameters
		query := r.URL.Query()
		if query.Get("status") != "open" {
			t.Errorf("expected status=open, got %s", query.Get("status"))
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	c := client.NewClient(mock.URL, "test-key")
	versionClient := NewClient(c)

	params := map[string]string{
		"status": "open",
	}
	_, err := versionClient.List(context.Background(), 1, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_List_Empty(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	response := VersionList{
		Versions:   []Version{},
		TotalCount: 0,
		Limit:      25,
		Offset:     0,
	}
	mock.HandleJSON("/projects/1/versions.json", response)

	c := client.NewClient(mock.URL, "test-key")
	versionClient := NewClient(c)

	result, err := versionClient.List(context.Background(), 1, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Versions) != 0 {
		t.Errorf("expected 0 versions, got %d", len(result.Versions))
	}
}

func TestClient_List_Error(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.HandleError("/projects/1/versions.json", http.StatusNotFound, "Project not found")

	c := client.NewClient(mock.URL, "test-key")
	versionClient := NewClient(c)

	_, err := versionClient.List(context.Background(), 1, nil)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestClient_Get_Success(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	now := time.Now()
	response := struct {
		Version Version `json:"version"`
	}{
		Version: Version{
			ID:          1,
			Name:        "v1.0",
			Description: "First release",
			Status:      "open",
			CreatedOn:   &now,
		},
	}
	mock.HandleJSON("/versions/1.json", response)

	c := client.NewClient(mock.URL, "test-key")
	versionClient := NewClient(c)

	result, err := versionClient.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != 1 {
		t.Errorf("expected ID 1, got %d", result.ID)
	}
	if result.Name != "v1.0" {
		t.Errorf("expected name 'v1.0', got %s", result.Name)
	}
	if result.Status != "open" {
		t.Errorf("expected status 'open', got %s", result.Status)
	}
}

func TestClient_Get_Error(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.HandleError("/versions/999.json", http.StatusNotFound, "Version not found")

	c := client.NewClient(mock.URL, "test-key")
	versionClient := NewClient(c)

	_, err := versionClient.Get(context.Background(), 999)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestClient_Create_Success(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	now := time.Now()
	response := struct {
		Version Version `json:"version"`
	}{
		Version: Version{
			ID:          1,
			Name:        "v1.0",
			Description: "First release",
			Status:      "open",
			CreatedOn:   &now,
		},
	}
	mock.HandleJSON("/projects/1/versions.json", response)

	c := client.NewClient(mock.URL, "test-key")
	versionClient := NewClient(c)

	req := &VersionCreateRequest{
		Name:        "v1.0",
		Description: "First release",
		Status:      "open",
	}
	result, err := versionClient.Create(context.Background(), 1, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != 1 {
		t.Errorf("expected ID 1, got %d", result.ID)
	}
	if result.Name != "v1.0" {
		t.Errorf("expected name 'v1.0', got %s", result.Name)
	}
}

func TestClient_Create_Error(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.HandleError("/projects/1/versions.json", http.StatusForbidden, "Permission denied")

	c := client.NewClient(mock.URL, "test-key")
	versionClient := NewClient(c)

	req := &VersionCreateRequest{
		Name: "v1.0",
	}
	_, err := versionClient.Create(context.Background(), 1, req)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestClient_Update_Success(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.Handle("/versions/1.json", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	c := client.NewClient(mock.URL, "test-key")
	versionClient := NewClient(c)

	req := &VersionUpdateRequest{
		Name:        "v1.1",
		Description: "Updated description",
		Status:      "closed",
	}
	err := versionClient.Update(context.Background(), 1, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_Update_Error(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.HandleError("/versions/1.json", http.StatusNotFound, "Version not found")

	c := client.NewClient(mock.URL, "test-key")
	versionClient := NewClient(c)

	req := &VersionUpdateRequest{
		Name: "v1.1",
	}
	err := versionClient.Update(context.Background(), 1, req)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestClient_Delete_Success(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.Handle("/versions/1.json", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	c := client.NewClient(mock.URL, "test-key")
	versionClient := NewClient(c)

	err := versionClient.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_Delete_Error(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.HandleError("/versions/1.json", http.StatusNotFound, "Version not found")

	c := client.NewClient(mock.URL, "test-key")
	versionClient := NewClient(c)

	err := versionClient.Delete(context.Background(), 1)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestClient_Delete_Forbidden(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.HandleError("/versions/1.json", http.StatusForbidden, "Permission denied")

	c := client.NewClient(mock.URL, "test-key")
	versionClient := NewClient(c)

	err := versionClient.Delete(context.Background(), 1)
	if err == nil {
		t.Error("expected error, got nil")
	}
}
