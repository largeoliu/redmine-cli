// internal/resources/projects/client_test.go
package projects

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/largeoliu/redmine-cli/internal/client"
	"github.com/largeoliu/redmine-cli/internal/testutil"
)

// sampleProject 返回一个示例 Project 用于测试
func sampleProject() Project {
	createdOn := time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC)
	updatedOn := time.Date(2024, 1, 10, 16, 30, 0, 0, time.UTC)

	return Project{
		ID:          1,
		Name:        "Sample Project",
		Identifier:  "sample-project",
		Description: "This is a sample project description for testing purposes.",
		Homepage:    "https://example.com/sample-project",
		Status:      1,
		CreatedOn:   &createdOn,
		UpdatedOn:   &updatedOn,
		Trackers: []Reference{
			{ID: 1, Name: "Bug"},
			{ID: 2, Name: "Feature"},
			{ID: 3, Name: "Support"},
		},
		IssueCategories: []Reference{
			{ID: 1, Name: "Development"},
			{ID: 2, Name: "Testing"},
		},
		EnabledModules: []Reference{
			{ID: 1, Name: "issue_tracking"},
			{ID: 2, Name: "time_tracking"},
		},
	}
}

// TestNewClient 测试 NewClient 构造函数
func TestNewClient(t *testing.T) {
	c := client.NewClient("https://example.com", "test-key")
	pc := NewClient(c)

	if pc == nil {
		t.Fatal("expected client, got nil")
	}
}

// TestClient_List 测试 List 方法
func TestClient_List(t *testing.T) {
	t.Run("success without params", func(t *testing.T) {
		mock := testutil.NewMockServer(t)
		defer mock.Close()

		project := sampleProject()
		projectList := ProjectList{
			Projects:   []Project{project},
			TotalCount: 1,
			Limit:      25,
			Offset:     0,
		}
		mock.HandleJSON("/projects.json", projectList)

		c := client.NewClient(mock.URL, "test-key")
		pc := NewClient(c)

		result, err := pc.List(context.Background(), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.TotalCount != 1 {
			t.Errorf("expected TotalCount 1, got %d", result.TotalCount)
		}
		if len(result.Projects) != 1 {
			t.Errorf("expected 1 project, got %d", len(result.Projects))
		}
		if result.Projects[0].ID != project.ID {
			t.Errorf("expected project ID %d, got %d", project.ID, result.Projects[0].ID)
		}
	})

	t.Run("success with params", func(t *testing.T) {
		mock := testutil.NewMockServer(t)
		defer mock.Close()

		project := sampleProject()
		projectList := ProjectList{
			Projects:   []Project{project},
			TotalCount: 1,
			Limit:      10,
			Offset:     0,
		}

		mock.Handle("/projects.json", func(w http.ResponseWriter, r *http.Request) {
			// 验证查询参数
			if r.URL.Query().Get("limit") != "10" {
				t.Errorf("expected limit=10, got %s", r.URL.Query().Get("limit"))
			}
			if r.URL.Query().Get("offset") != "0" {
				t.Errorf("expected offset=0, got %s", r.URL.Query().Get("offset"))
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(projectList)
		})

		c := client.NewClient(mock.URL, "test-key")
		pc := NewClient(c)

		params := map[string]string{
			"limit":  "10",
			"offset": "0",
		}
		result, err := pc.List(context.Background(), params)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Limit != 10 {
			t.Errorf("expected Limit 10, got %d", result.Limit)
		}
	})

	t.Run("error", func(t *testing.T) {
		mock := testutil.NewMockServer(t)
		defer mock.Close()

		mock.HandleError("/projects.json", http.StatusUnauthorized, "Unauthorized")

		c := client.NewClient(mock.URL, "test-key")
		pc := NewClient(c)

		_, err := pc.List(context.Background(), nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

// TestClient_Get 测试 Get 方法
func TestClient_Get(t *testing.T) {
	t.Run("success without params", func(t *testing.T) {
		mock := testutil.NewMockServer(t)
		defer mock.Close()

		project := sampleProject()
		response := map[string]any{
			"project": project,
		}
		mock.HandleJSON("/projects/1.json", response)

		c := client.NewClient(mock.URL, "test-key")
		pc := NewClient(c)

		result, err := pc.Get(context.Background(), 1, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.ID != project.ID {
			t.Errorf("expected ID %d, got %d", project.ID, result.ID)
		}
		if result.Name != project.Name {
			t.Errorf("expected Name %s, got %s", project.Name, result.Name)
		}
		if result.Identifier != project.Identifier {
			t.Errorf("expected Identifier %s, got %s", project.Identifier, result.Identifier)
		}
	})

	t.Run("success with params", func(t *testing.T) {
		mock := testutil.NewMockServer(t)
		defer mock.Close()

		project := sampleProject()
		response := map[string]any{
			"project": project,
		}

		mock.Handle("/projects/1.json", func(w http.ResponseWriter, r *http.Request) {
			// 验证查询参数
			if r.URL.Query().Get("include") != "trackers" {
				t.Errorf("expected include=trackers, got %s", r.URL.Query().Get("include"))
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		})

		c := client.NewClient(mock.URL, "test-key")
		pc := NewClient(c)

		params := map[string]string{
			"include": "trackers",
		}
		result, err := pc.Get(context.Background(), 1, params)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.ID != project.ID {
			t.Errorf("expected ID %d, got %d", project.ID, result.ID)
		}
	})

	t.Run("error not found", func(t *testing.T) {
		mock := testutil.NewMockServer(t)
		defer mock.Close()

		mock.HandleError("/projects/999.json", http.StatusNotFound, "Not found")

		c := client.NewClient(mock.URL, "test-key")
		pc := NewClient(c)

		_, err := pc.Get(context.Background(), 999, nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("error unauthorized", func(t *testing.T) {
		mock := testutil.NewMockServer(t)
		defer mock.Close()

		mock.HandleError("/projects/1.json", http.StatusUnauthorized, "Unauthorized")

		c := client.NewClient(mock.URL, "test-key")
		pc := NewClient(c)

		_, err := pc.Get(context.Background(), 1, nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

// TestClient_GetByIdentifier 测试 GetByIdentifier 方法
func TestClient_GetByIdentifier(t *testing.T) {
	t.Run("success without params", func(t *testing.T) {
		mock := testutil.NewMockServer(t)
		defer mock.Close()

		project := sampleProject()
		response := map[string]any{
			"project": project,
		}
		mock.HandleJSON("/projects/sample-project.json", response)

		c := client.NewClient(mock.URL, "test-key")
		pc := NewClient(c)

		result, err := pc.GetByIdentifier(context.Background(), "sample-project", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.ID != project.ID {
			t.Errorf("expected ID %d, got %d", project.ID, result.ID)
		}
		if result.Identifier != project.Identifier {
			t.Errorf("expected Identifier %s, got %s", project.Identifier, result.Identifier)
		}
	})

	t.Run("success with params", func(t *testing.T) {
		mock := testutil.NewMockServer(t)
		defer mock.Close()

		project := sampleProject()
		response := map[string]any{
			"project": project,
		}

		mock.Handle("/projects/sample-project.json", func(w http.ResponseWriter, r *http.Request) {
			// 验证查询参数
			if r.URL.Query().Get("include") != "issue_categories" {
				t.Errorf("expected include=issue_categories, got %s", r.URL.Query().Get("include"))
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		})

		c := client.NewClient(mock.URL, "test-key")
		pc := NewClient(c)

		params := map[string]string{
			"include": "issue_categories",
		}
		result, err := pc.GetByIdentifier(context.Background(), "sample-project", params)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.ID != project.ID {
			t.Errorf("expected ID %d, got %d", project.ID, result.ID)
		}
	})

	t.Run("error not found", func(t *testing.T) {
		mock := testutil.NewMockServer(t)
		defer mock.Close()

		mock.HandleError("/projects/nonexistent.json", http.StatusNotFound, "Not found")

		c := client.NewClient(mock.URL, "test-key")
		pc := NewClient(c)

		_, err := pc.GetByIdentifier(context.Background(), "nonexistent", nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("error forbidden", func(t *testing.T) {
		mock := testutil.NewMockServer(t)
		defer mock.Close()

		mock.HandleError("/projects/private-project.json", http.StatusForbidden, "Forbidden")

		c := client.NewClient(mock.URL, "test-key")
		pc := NewClient(c)

		_, err := pc.GetByIdentifier(context.Background(), "private-project", nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

// TestClient_Create 测试 Create 方法
func TestClient_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock := testutil.NewMockServer(t)
		defer mock.Close()

		project := sampleProject()
		response := map[string]any{
			"project": project,
		}

		mock.Handle("/projects.json", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("expected POST method, got %s", r.Method)
			}

			// 验证请求体
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("failed to decode request body: %v", err)
			}

			projectData, ok := body["project"].(map[string]any)
			if !ok {
				t.Fatal("expected project field in request body")
			}

			if projectData["name"] != "New Project" {
				t.Errorf("expected name 'New Project', got %v", projectData["name"])
			}
			if projectData["identifier"] != "new-project" {
				t.Errorf("expected identifier 'new-project', got %v", projectData["identifier"])
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(response)
		})

		c := client.NewClient(mock.URL, "test-key")
		pc := NewClient(c)

		req := &ProjectCreateRequest{
			Name:        "New Project",
			Identifier:  "new-project",
			Description: "A new project for testing",
			IsPublic:    true,
		}

		result, err := pc.Create(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.ID != project.ID {
			t.Errorf("expected ID %d, got %d", project.ID, result.ID)
		}
	})

	t.Run("success with all fields", func(t *testing.T) {
		mock := testutil.NewMockServer(t)
		defer mock.Close()

		project := sampleProject()
		response := map[string]any{
			"project": project,
		}

		mock.Handle("/projects.json", func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("failed to decode request body: %v", err)
			}

			projectData := body["project"].(map[string]any)

			// 验证所有字段
			if projectData["name"] != "Full Project" {
				t.Errorf("expected name 'Full Project', got %v", projectData["name"])
			}
			if projectData["identifier"] != "full-project" {
				t.Errorf("expected identifier 'full-project', got %v", projectData["identifier"])
			}
			if projectData["description"] != "Full description" {
				t.Errorf("expected description 'Full description', got %v", projectData["description"])
			}
			if projectData["homepage"] != "https://example.com" {
				t.Errorf("expected homepage 'https://example.com', got %v", projectData["homepage"])
			}
			if projectData["is_public"] != true {
				t.Errorf("expected is_public true, got %v", projectData["is_public"])
			}
			if projectData["parent_id"] != float64(1) {
				t.Errorf("expected parent_id 1, got %v", projectData["parent_id"])
			}
			if projectData["inherit_members"] != true {
				t.Errorf("expected inherit_members true, got %v", projectData["inherit_members"])
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(response)
		})

		c := client.NewClient(mock.URL, "test-key")
		pc := NewClient(c)

		req := &ProjectCreateRequest{
			Name:           "Full Project",
			Identifier:     "full-project",
			Description:    "Full description",
			Homepage:       "https://example.com",
			IsPublic:       true,
			ParentID:       1,
			InheritMembers: true,
		}

		result, err := pc.Create(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result == nil {
			t.Fatal("expected result, got nil")
		}
	})

	t.Run("error validation", func(t *testing.T) {
		mock := testutil.NewMockServer(t)
		defer mock.Close()

		mock.HandleError("/projects.json", http.StatusUnprocessableEntity, "Validation failed")

		c := client.NewClient(mock.URL, "test-key")
		pc := NewClient(c)

		req := &ProjectCreateRequest{
			Name: "", // 缺少必需字段
		}

		_, err := pc.Create(context.Background(), req)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("error unauthorized", func(t *testing.T) {
		mock := testutil.NewMockServer(t)
		defer mock.Close()

		mock.HandleError("/projects.json", http.StatusUnauthorized, "Unauthorized")

		c := client.NewClient(mock.URL, "test-key")
		pc := NewClient(c)

		req := &ProjectCreateRequest{
			Name:       "Test Project",
			Identifier: "test-project",
		}

		_, err := pc.Create(context.Background(), req)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

// TestClient_Update 测试 Update 方法
func TestClient_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock := testutil.NewMockServer(t)
		defer mock.Close()

		mock.Handle("/projects/1.json", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPut {
				t.Errorf("expected PUT method, got %s", r.Method)
			}

			// 验证请求体
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("failed to decode request body: %v", err)
			}

			projectData, ok := body["project"].(map[string]any)
			if !ok {
				t.Fatal("expected project field in request body")
			}

			if projectData["name"] != "Updated Project" {
				t.Errorf("expected name 'Updated Project', got %v", projectData["name"])
			}
			if projectData["description"] != "Updated description" {
				t.Errorf("expected description 'Updated description', got %v", projectData["description"])
			}
			if projectData["status"] != float64(1) {
				t.Errorf("expected status 1, got %v", projectData["status"])
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
		})

		c := client.NewClient(mock.URL, "test-key")
		pc := NewClient(c)

		req := &ProjectUpdateRequest{
			Name:        "Updated Project",
			Description: "Updated description",
			Status:      1,
		}

		err := pc.Update(context.Background(), 1, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("success with all fields", func(t *testing.T) {
		mock := testutil.NewMockServer(t)
		defer mock.Close()

		mock.Handle("/projects/1.json", func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("failed to decode request body: %v", err)
			}

			projectData := body["project"].(map[string]any)

			// 验证所有字段
			if projectData["name"] != "Full Update" {
				t.Errorf("expected name 'Full Update', got %v", projectData["name"])
			}
			if projectData["description"] != "Full update description" {
				t.Errorf("expected description 'Full update description', got %v", projectData["description"])
			}
			if projectData["homepage"] != "https://updated.com" {
				t.Errorf("expected homepage 'https://updated.com', got %v", projectData["homepage"])
			}
			if projectData["status"] != float64(5) {
				t.Errorf("expected status 5, got %v", projectData["status"])
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
		})

		c := client.NewClient(mock.URL, "test-key")
		pc := NewClient(c)

		req := &ProjectUpdateRequest{
			Name:        "Full Update",
			Description: "Full update description",
			Homepage:    "https://updated.com",
			Status:      5,
		}

		err := pc.Update(context.Background(), 1, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("error not found", func(t *testing.T) {
		mock := testutil.NewMockServer(t)
		defer mock.Close()

		mock.HandleError("/projects/999.json", http.StatusNotFound, "Not found")

		c := client.NewClient(mock.URL, "test-key")
		pc := NewClient(c)

		req := &ProjectUpdateRequest{
			Name: "Updated Project",
		}

		err := pc.Update(context.Background(), 999, req)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("error forbidden", func(t *testing.T) {
		mock := testutil.NewMockServer(t)
		defer mock.Close()

		mock.HandleError("/projects/1.json", http.StatusForbidden, "Forbidden")

		c := client.NewClient(mock.URL, "test-key")
		pc := NewClient(c)

		req := &ProjectUpdateRequest{
			Name: "Updated Project",
		}

		err := pc.Update(context.Background(), 1, req)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

// TestClient_Delete 测试 Delete 方法
func TestClient_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock := testutil.NewMockServer(t)
		defer mock.Close()

		mock.Handle("/projects/1.json", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodDelete {
				t.Errorf("expected DELETE method, got %s", r.Method)
			}
			w.WriteHeader(http.StatusNoContent)
		})

		c := client.NewClient(mock.URL, "test-key")
		pc := NewClient(c)

		err := pc.Delete(context.Background(), 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("error not found", func(t *testing.T) {
		mock := testutil.NewMockServer(t)
		defer mock.Close()

		mock.HandleError("/projects/999.json", http.StatusNotFound, "Not found")

		c := client.NewClient(mock.URL, "test-key")
		pc := NewClient(c)

		err := pc.Delete(context.Background(), 999)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("error forbidden", func(t *testing.T) {
		mock := testutil.NewMockServer(t)
		defer mock.Close()

		mock.HandleError("/projects/1.json", http.StatusForbidden, "Forbidden")

		c := client.NewClient(mock.URL, "test-key")
		pc := NewClient(c)

		err := pc.Delete(context.Background(), 1)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("error unauthorized", func(t *testing.T) {
		mock := testutil.NewMockServer(t)
		defer mock.Close()

		mock.HandleError("/projects/1.json", http.StatusUnauthorized, "Unauthorized")

		c := client.NewClient(mock.URL, "test-key")
		pc := NewClient(c)

		err := pc.Delete(context.Background(), 1)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

// TestProjectStruct 测试 Project 结构体的字段
func TestProjectStruct(t *testing.T) {
	project := sampleProject()

	if project.ID != 1 {
		t.Errorf("expected ID 1, got %d", project.ID)
	}
	if project.Name != "Sample Project" {
		t.Errorf("expected Name 'Sample Project', got %s", project.Name)
	}
	if project.Identifier != "sample-project" {
		t.Errorf("expected Identifier 'sample-project', got %s", project.Identifier)
	}
	if project.Description == "" {
		t.Error("expected Description to be non-empty")
	}
	if project.Homepage == "" {
		t.Error("expected Homepage to be non-empty")
	}
	if project.Status != 1 {
		t.Errorf("expected Status 1, got %d", project.Status)
	}
	if project.CreatedOn == nil {
		t.Error("expected CreatedOn to be non-nil")
	}
	if project.UpdatedOn == nil {
		t.Error("expected UpdatedOn to be non-nil")
	}
	if len(project.Trackers) != 3 {
		t.Errorf("expected 3 trackers, got %d", len(project.Trackers))
	}
	if len(project.IssueCategories) != 2 {
		t.Errorf("expected 2 issue categories, got %d", len(project.IssueCategories))
	}
	if len(project.EnabledModules) != 2 {
		t.Errorf("expected 2 enabled modules, got %d", len(project.EnabledModules))
	}
}

// TestReferenceStruct 测试 Reference 结构体
func TestReferenceStruct(t *testing.T) {
	ref := Reference{
		ID:   1,
		Name: "Test Reference",
	}

	if ref.ID != 1 {
		t.Errorf("expected ID 1, got %d", ref.ID)
	}
	if ref.Name != "Test Reference" {
		t.Errorf("expected Name 'Test Reference', got %s", ref.Name)
	}
}

// TestProjectListStruct 测试 ProjectList 结构体
func TestProjectListStruct(t *testing.T) {
	project := sampleProject()
	list := ProjectList{
		Projects:   []Project{project},
		TotalCount: 1,
		Limit:      25,
		Offset:     0,
	}

	if len(list.Projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(list.Projects))
	}
	if list.TotalCount != 1 {
		t.Errorf("expected TotalCount 1, got %d", list.TotalCount)
	}
	if list.Limit != 25 {
		t.Errorf("expected Limit 25, got %d", list.Limit)
	}
	if list.Offset != 0 {
		t.Errorf("expected Offset 0, got %d", list.Offset)
	}
}

// TestProjectCreateRequestStruct 测试 ProjectCreateRequest 结构体
func TestProjectCreateRequestStruct(t *testing.T) {
	req := ProjectCreateRequest{
		Name:           "Test Project",
		Identifier:     "test-project",
		Description:    "Test description",
		Homepage:       "https://test.com",
		IsPublic:       true,
		ParentID:       1,
		InheritMembers: true,
	}

	if req.Name != "Test Project" {
		t.Errorf("expected Name 'Test Project', got %s", req.Name)
	}
	if req.Identifier != "test-project" {
		t.Errorf("expected Identifier 'test-project', got %s", req.Identifier)
	}
	if req.Description != "Test description" {
		t.Errorf("expected Description 'Test description', got %s", req.Description)
	}
	if req.Homepage != "https://test.com" {
		t.Errorf("expected Homepage 'https://test.com', got %s", req.Homepage)
	}
	if !req.IsPublic {
		t.Error("expected IsPublic to be true")
	}
	if req.ParentID != 1 {
		t.Errorf("expected ParentID 1, got %d", req.ParentID)
	}
	if !req.InheritMembers {
		t.Error("expected InheritMembers to be true")
	}
}

// TestProjectUpdateRequestStruct 测试 ProjectUpdateRequest 结构体
func TestProjectUpdateRequestStruct(t *testing.T) {
	req := ProjectUpdateRequest{
		Name:        "Updated Project",
		Description: "Updated description",
		Homepage:    "https://updated.com",
		Status:      5,
	}

	if req.Name != "Updated Project" {
		t.Errorf("expected Name 'Updated Project', got %s", req.Name)
	}
	if req.Description != "Updated description" {
		t.Errorf("expected Description 'Updated description', got %s", req.Description)
	}
	if req.Homepage != "https://updated.com" {
		t.Errorf("expected Homepage 'https://updated.com', got %s", req.Homepage)
	}
	if req.Status != 5 {
		t.Errorf("expected Status 5, got %d", req.Status)
	}
}

// TestProjectWithParent 测试带有 Parent 字段的 Project
func TestProjectWithParent(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	project := sampleProject()
	project.Parent = &Reference{
		ID:   10,
		Name: "Parent Project",
	}
	response := map[string]any{
		"project": project,
	}
	mock.HandleJSON("/projects/1.json", response)

	c := client.NewClient(mock.URL, "test-key")
	pc := NewClient(c)

	result, err := pc.Get(context.Background(), 1, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Parent == nil {
		t.Fatal("expected Parent to be non-nil")
	}
	if result.Parent.ID != 10 {
		t.Errorf("expected Parent ID 10, got %d", result.Parent.ID)
	}
	if result.Parent.Name != "Parent Project" {
		t.Errorf("expected Parent Name 'Parent Project', got %s", result.Parent.Name)
	}
}

// TestClient_ListEmpty 测试空列表返回
func TestClient_ListEmpty(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	projectList := ProjectList{
		Projects:   []Project{},
		TotalCount: 0,
		Limit:      25,
		Offset:     0,
	}
	mock.HandleJSON("/projects.json", projectList)

	c := client.NewClient(mock.URL, "test-key")
	pc := NewClient(c)

	result, err := pc.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCount != 0 {
		t.Errorf("expected TotalCount 0, got %d", result.TotalCount)
	}
	if len(result.Projects) != 0 {
		t.Errorf("expected 0 projects, got %d", len(result.Projects))
	}
}

// TestClient_ListMultipleProjects 测试多项目列表返回
func TestClient_ListMultipleProjects(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	project1 := sampleProject()
	project2 := sampleProject()
	project2.ID = 2
	project2.Name = "Second Project"
	project2.Identifier = "second-project"

	projectList := ProjectList{
		Projects:   []Project{project1, project2},
		TotalCount: 2,
		Limit:      25,
		Offset:     0,
	}
	mock.HandleJSON("/projects.json", projectList)

	c := client.NewClient(mock.URL, "test-key")
	pc := NewClient(c)

	result, err := pc.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCount != 2 {
		t.Errorf("expected TotalCount 2, got %d", result.TotalCount)
	}
	if len(result.Projects) != 2 {
		t.Errorf("expected 2 projects, got %d", len(result.Projects))
	}
	if result.Projects[0].ID != 1 {
		t.Errorf("expected first project ID 1, got %d", result.Projects[0].ID)
	}
	if result.Projects[1].ID != 2 {
		t.Errorf("expected second project ID 2, got %d", result.Projects[1].ID)
	}
}

// TestClient_ServerError 测试服务器错误
func TestClient_ServerError(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.HandleError("/projects.json", http.StatusInternalServerError, "Internal Server Error")

	c := client.NewClient(mock.URL, "test-key")
	pc := NewClient(c)

	_, err := pc.List(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// TestClient_RateLimitError 测试速率限制错误
func TestClient_RateLimitError(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.HandleError("/projects.json", http.StatusTooManyRequests, "Rate limit exceeded")

	c := client.NewClient(mock.URL, "test-key")
	pc := NewClient(c)

	_, err := pc.List(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// TestClient_ContextCancellation 测试上下文取消
func TestClient_ContextCancellation(t *testing.T) {
	mock := testutil.NewMockServer(t)
	defer mock.Close()

	mock.Handle("/projects.json", func(w http.ResponseWriter, r *http.Request) {
		// 模拟延迟响应
		select {
		case <-r.Context().Done():
			// 上下文已取消
			return
		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{}`))
		}
	})

	c := client.NewClient(mock.URL, "test-key")
	pc := NewClient(c)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 立即取消

	_, err := pc.List(ctx, nil)
	if err == nil {
		t.Fatal("expected error due to context cancellation, got nil")
	}
}
