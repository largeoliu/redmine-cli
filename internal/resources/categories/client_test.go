// internal/resources/categories/client_test.go
package categories

import (
	"context"
	"net/http"
	"testing"

	"github.com/largeoliu/redmine-cli/internal/client"
	"github.com/largeoliu/redmine-cli/internal/testutil"
)

func TestNewClient(t *testing.T) {
	c := client.NewClient("https://example.com", "test-key")
	catClient := NewClient(c)
	if catClient == nil {
		t.Error("expected client to be created")
	}
}

func TestClient_List_Success(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	response := CategoryList{
		IssueCategories: []Category{
			{ID: 1, Name: "Bug"},
			{ID: 2, Name: "Feature"},
		},
	}
	mock.HandleJSON("/projects/1/issue_categories.json", response)

	c := client.NewClient(mock.URL, "test-key")
	catClient := NewClient(c)

	result, err := catClient.List(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.IssueCategories) != 2 {
		t.Errorf("expected 2 categories, got %d", len(result.IssueCategories))
	}
	if result.IssueCategories[0].Name != "Bug" {
		t.Errorf("expected first category name 'Bug', got %s", result.IssueCategories[0].Name)
	}
}

func TestClient_List_Error(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.HandleError("/projects/1/issue_categories.json", http.StatusNotFound, "Project not found")

	c := client.NewClient(mock.URL, "test-key")
	catClient := NewClient(c)

	_, err := catClient.List(context.Background(), 1)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestClient_Get_Success(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	response := struct {
		IssueCategory Category `json:"issue_category"`
	}{
		IssueCategory: Category{
			ID:   1,
			Name: "Bug",
		},
	}
	mock.HandleJSON("/issue_categories/1.json", response)

	c := client.NewClient(mock.URL, "test-key")
	catClient := NewClient(c)

	result, err := catClient.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != 1 {
		t.Errorf("expected ID 1, got %d", result.ID)
	}
	if result.Name != "Bug" {
		t.Errorf("expected name 'Bug', got %s", result.Name)
	}
}

func TestClient_Get_Error(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.HandleError("/issue_categories/999.json", http.StatusNotFound, "Category not found")

	c := client.NewClient(mock.URL, "test-key")
	catClient := NewClient(c)

	_, err := catClient.Get(context.Background(), 999)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestClient_Create_Success(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	response := struct {
		IssueCategory Category `json:"issue_category"`
	}{
		IssueCategory: Category{
			ID:   1,
			Name: "New Category",
		},
	}
	mock.HandleJSON("/projects/1/issue_categories.json", response)

	c := client.NewClient(mock.URL, "test-key")
	catClient := NewClient(c)

	req := &CategoryCreateRequest{
		Name: "New Category",
	}
	result, err := catClient.Create(context.Background(), 1, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != 1 {
		t.Errorf("expected ID 1, got %d", result.ID)
	}
	if result.Name != "New Category" {
		t.Errorf("expected name 'New Category', got %s", result.Name)
	}
}

func TestClient_Create_Error(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.HandleError("/projects/1/issue_categories.json", http.StatusForbidden, "Permission denied")

	c := client.NewClient(mock.URL, "test-key")
	catClient := NewClient(c)

	req := &CategoryCreateRequest{
		Name: "New Category",
	}
	_, err := catClient.Create(context.Background(), 1, req)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestClient_Update_Success(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.Handle("/issue_categories/1.json", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	c := client.NewClient(mock.URL, "test-key")
	catClient := NewClient(c)

	req := &CategoryUpdateRequest{
		Name: "Updated Category",
	}
	err := catClient.Update(context.Background(), 1, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_Update_Error(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.HandleError("/issue_categories/1.json", http.StatusNotFound, "Category not found")

	c := client.NewClient(mock.URL, "test-key")
	catClient := NewClient(c)

	req := &CategoryUpdateRequest{
		Name: "Updated Category",
	}
	err := catClient.Update(context.Background(), 1, req)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestClient_Delete_Success(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.Handle("/issue_categories/1.json", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	c := client.NewClient(mock.URL, "test-key")
	catClient := NewClient(c)

	err := catClient.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_Delete_Error(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.HandleError("/issue_categories/1.json", http.StatusNotFound, "Category not found")

	c := client.NewClient(mock.URL, "test-key")
	catClient := NewClient(c)

	err := catClient.Delete(context.Background(), 1)
	if err == nil {
		t.Error("expected error, got nil")
	}
}
