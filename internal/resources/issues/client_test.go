// internal/resources/issues/client_test.go
package issues

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/largeoliu/redmine-cli/internal/client"
)

// sampleIssue 返回一个示例 Issue 用于测试
func sampleIssue() Issue {
	createdOn := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	updatedOn := time.Date(2024, 1, 16, 14, 45, 0, 0, time.UTC)

	return Issue{
		ID:          1,
		Subject:     "Sample Issue Title",
		Description: "This is a sample issue description for testing purposes.",
		Project: &Reference{
			ID:   1,
			Name: "Sample Project",
		},
		Tracker: &Reference{
			ID:   1,
			Name: "Bug",
		},
		Status: &Reference{
			ID:   1,
			Name: "New",
		},
		Priority: &Reference{
			ID:   2,
			Name: "Normal",
		},
		Author: &Reference{
			ID:   1,
			Name: "John Doe",
		},
		AssignedTo: &Reference{
			ID:   2,
			Name: "Jane Smith",
		},
		Category: &Reference{
			ID:   1,
			Name: "Development",
		},
		StartDate:    "2024-01-15",
		DueDate:      "2024-01-31",
		DoneRatio:    50,
		CreatedOn:    &createdOn,
		UpdatedOn:    &updatedOn,
		PrivateNotes: false,
	}
}

// sampleIssueList 返回一个示例 Issue 列表用于测试
func sampleIssueList() IssueList {
	issue1 := sampleIssue()
	issue2 := sampleIssue()
	issue2.ID = 2
	issue2.Subject = "Second Sample Issue"
	issue2.Description = "Another sample issue for testing list operations."
	issue2.DoneRatio = 0

	return IssueList{
		Issues:     []Issue{issue1, issue2},
		TotalCount: 2,
		Limit:      25,
		Offset:     0,
	}
}

// mockServer 创建一个简单的模拟服务器
type mockServer struct {
	*httptest.Server
	mux *http.ServeMux
}

func newMockServer() *mockServer {
	mux := http.NewServeMux()
	return &mockServer{
		Server: httptest.NewServer(mux),
		mux:    mux,
	}
}

// handleJSON 注册一个返回 JSON 响应的处理器
func (m *mockServer) handleJSON(path string, response any) {
	m.mux.HandleFunc(path, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			// Log but don't fail - this is test infrastructure
			_, _ = w.Write([]byte(`{}`))
		}
	})
}

// handleError 注册一个返回错误的处理器
func (m *mockServer) handleError(path string, statusCode int, message string) {
	m.mux.HandleFunc(path, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(statusCode)
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]any{
			"errors": []string{message},
		}); err != nil {
			_, _ = w.Write([]byte(`{}`))
		}
	})
}

// TestNewClient 测试客户端创建
func TestNewClient(t *testing.T) {
	baseClient := client.NewClient("https://example.com", "test-key")
	issueClient := NewClient(baseClient)

	if issueClient == nil {
		t.Fatal("expected client to be created, got nil")
	}
	if issueClient.client == nil {
		t.Error("expected internal client to be set")
	}
}

// TestClient_List 测试 List 方法
func TestClient_List(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func(mock *mockServer)
		params     map[string]string
		wantErr    bool
		wantCount  int
		wantTotal  int
		errMessage string
	}{
		{
			name: "成功获取issue列表",
			setupMock: func(mock *mockServer) {
				mock.handleJSON("/issues.json", sampleIssueList())
			},
			params:    nil,
			wantErr:   false,
			wantCount: 2,
			wantTotal: 2,
		},
		{
			name: "成功获取带参数的issue列表",
			setupMock: func(mock *mockServer) {
				mock.mux.HandleFunc("/issues.json", func(w http.ResponseWriter, r *http.Request) {
					// 验证查询参数
					if r.URL.Query().Get("project_id") != "1" {
						t.Errorf("expected project_id=1, got %s", r.URL.Query().Get("project_id"))
					}
					if r.URL.Query().Get("status_id") != "2" {
						t.Errorf("expected status_id=2, got %s", r.URL.Query().Get("status_id"))
					}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(sampleIssueList())
				})
			},
			params:    map[string]string{"project_id": "1", "status_id": "2"},
			wantErr:   false,
			wantCount: 2,
			wantTotal: 2,
		},
		{
			name: "空列表",
			setupMock: func(mock *mockServer) {
				mock.handleJSON("/issues.json", IssueList{
					Issues:     []Issue{},
					TotalCount: 0,
					Limit:      25,
					Offset:     0,
				})
			},
			params:    nil,
			wantErr:   false,
			wantCount: 0,
			wantTotal: 0,
		},
		{
			name: "认证失败",
			setupMock: func(mock *mockServer) {
				mock.handleError("/issues.json", http.StatusUnauthorized, "Unauthorized")
			},
			params:     nil,
			wantErr:    true,
			errMessage: "authentication failed",
		},
		{
			name: "权限不足",
			setupMock: func(mock *mockServer) {
				mock.handleError("/issues.json", http.StatusForbidden, "Forbidden")
			},
			params:     nil,
			wantErr:    true,
			errMessage: "permission denied",
		},
		{
			name: "服务器错误",
			setupMock: func(mock *mockServer) {
				mock.handleError("/issues.json", http.StatusInternalServerError, "Internal Server Error")
			},
			params:     nil,
			wantErr:    true,
			errMessage: "server error",
		},
		{
			name: "资源未找到",
			setupMock: func(mock *mockServer) {
				mock.handleError("/issues.json", http.StatusNotFound, "Not Found")
			},
			params:     nil,
			wantErr:    true,
			errMessage: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockServer()
			defer mock.Close()

			tt.setupMock(mock)

			baseClient := client.NewClient(mock.URL, "test-key")
			issueClient := NewClient(baseClient)

			result, err := issueClient.List(context.Background(), tt.params)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result == nil {
				t.Fatal("expected result, got nil")
			}

			if len(result.Issues) != tt.wantCount {
				t.Errorf("expected %d issues, got %d", tt.wantCount, len(result.Issues))
			}

			if result.TotalCount != tt.wantTotal {
				t.Errorf("expected total count %d, got %d", tt.wantTotal, result.TotalCount)
			}
		})
	}
}

// TestClient_Get 测试 Get 方法
func TestClient_Get(t *testing.T) {
	tests := []struct {
		name       string
		issueID    int
		setupMock  func(mock *mockServer)
		params     map[string]string
		wantErr    bool
		wantID     int
		errMessage string
	}{
		{
			name:    "成功获取单个issue",
			issueID: 1,
			setupMock: func(mock *mockServer) {
				response := map[string]any{
					"issue": sampleIssue(),
				}
				mock.handleJSON("/issues/1.json", response)
			},
			params:  nil,
			wantErr: false,
			wantID:  1,
		},
		{
			name:    "成功获取带参数的单个issue",
			issueID: 1,
			setupMock: func(mock *mockServer) {
				mock.mux.HandleFunc("/issues/1.json", func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Query().Get("include") != "relations" {
						t.Errorf("expected include=relations, got %s", r.URL.Query().Get("include"))
					}
					response := map[string]any{
						"issue": sampleIssue(),
					}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(response)
				})
			},
			params:  map[string]string{"include": "relations"},
			wantErr: false,
			wantID:  1,
		},
		{
			name:    "issue不存在",
			issueID: 999,
			setupMock: func(mock *mockServer) {
				mock.handleError("/issues/999.json", http.StatusNotFound, "Issue not found")
			},
			params:     nil,
			wantErr:    true,
			errMessage: "not found",
		},
		{
			name:    "认证失败",
			issueID: 1,
			setupMock: func(mock *mockServer) {
				mock.handleError("/issues/1.json", http.StatusUnauthorized, "Unauthorized")
			},
			params:     nil,
			wantErr:    true,
			errMessage: "authentication failed",
		},
		{
			name:    "权限不足",
			issueID: 1,
			setupMock: func(mock *mockServer) {
				mock.handleError("/issues/1.json", http.StatusForbidden, "Forbidden")
			},
			params:     nil,
			wantErr:    true,
			errMessage: "permission denied",
		},
		{
			name:    "服务器错误",
			issueID: 1,
			setupMock: func(mock *mockServer) {
				mock.handleError("/issues/1.json", http.StatusInternalServerError, "Internal Server Error")
			},
			params:     nil,
			wantErr:    true,
			errMessage: "server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockServer()
			defer mock.Close()

			tt.setupMock(mock)

			baseClient := client.NewClient(mock.URL, "test-key")
			issueClient := NewClient(baseClient)

			result, err := issueClient.Get(context.Background(), tt.issueID, tt.params)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result == nil {
				t.Fatal("expected result, got nil")
			}

			if result.ID != tt.wantID {
				t.Errorf("expected issue ID %d, got %d", tt.wantID, result.ID)
			}
		})
	}
}

// TestClient_Create 测试 Create 方法
func TestClient_Create(t *testing.T) {
	tests := []struct {
		name       string
		request    *IssueCreateRequest
		setupMock  func(mock *mockServer)
		wantErr    bool
		wantID     int
		errMessage string
	}{
		{
			name: "成功创建issue",
			request: &IssueCreateRequest{
				ProjectID: 1,
				Subject:   "Test Issue",
			},
			setupMock: func(mock *mockServer) {
				mock.mux.HandleFunc("/issues.json", func(w http.ResponseWriter, r *http.Request) {
					if r.Method != http.MethodPost {
						t.Errorf("expected POST method, got %s", r.Method)
					}

					// 验证请求体
					var body map[string]any
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Errorf("failed to decode request body: %v", err)
					}

					issue, ok := body["issue"].(map[string]any)
					if !ok {
						t.Error("expected issue field in request body")
					}

					if issue["subject"] != "Test Issue" {
						t.Errorf("expected subject 'Test Issue', got %v", issue["subject"])
					}

					// 返回创建的 issue
					createdIssue := sampleIssue()
					createdIssue.Subject = "Test Issue"
					response := map[string]any{
						"issue": createdIssue,
					}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(response)
				})
			},
			wantErr: false,
			wantID:  1,
		},
		{
			name: "创建带所有字段的issue",
			request: &IssueCreateRequest{
				ProjectID:      1,
				Subject:        "Full Issue",
				Description:    "Full description",
				TrackerID:      1,
				StatusID:       1,
				PriorityID:     2,
				AssignedToID:   3,
				CategoryID:     1,
				FixedVersionID: 1,
				ParentIssueID:  0,
				StartDate:      "2024-01-15",
				DueDate:        "2024-01-31",
				DoneRatio:      0,
				WatcherUserIDs: []int{1, 2},
			},
			setupMock: func(mock *mockServer) {
				createdIssue := sampleIssue()
				createdIssue.Subject = "Full Issue"
				createdIssue.Description = "Full description"
				response := map[string]any{
					"issue": createdIssue,
				}
				mock.handleJSON("/issues.json", response)
			},
			wantErr: false,
			wantID:  1,
		},
		{
			name: "验证错误 - 缺少必填字段",
			request: &IssueCreateRequest{
				ProjectID: 0,
				Subject:   "",
			},
			setupMock: func(mock *mockServer) {
				mock.handleError("/issues.json", http.StatusUnprocessableEntity, "Subject can't be blank")
			},
			wantErr:    true,
			errMessage: "request failed",
		},
		{
			name: "认证失败",
			request: &IssueCreateRequest{
				ProjectID: 1,
				Subject:   "Test",
			},
			setupMock: func(mock *mockServer) {
				mock.handleError("/issues.json", http.StatusUnauthorized, "Unauthorized")
			},
			wantErr:    true,
			errMessage: "authentication failed",
		},
		{
			name: "权限不足",
			request: &IssueCreateRequest{
				ProjectID: 1,
				Subject:   "Test",
			},
			setupMock: func(mock *mockServer) {
				mock.handleError("/issues.json", http.StatusForbidden, "Forbidden")
			},
			wantErr:    true,
			errMessage: "permission denied",
		},
		{
			name: "服务器错误",
			request: &IssueCreateRequest{
				ProjectID: 1,
				Subject:   "Test",
			},
			setupMock: func(mock *mockServer) {
				mock.handleError("/issues.json", http.StatusInternalServerError, "Internal Server Error")
			},
			wantErr:    true,
			errMessage: "server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockServer()
			defer mock.Close()

			tt.setupMock(mock)

			baseClient := client.NewClient(mock.URL, "test-key")
			issueClient := NewClient(baseClient)

			result, err := issueClient.Create(context.Background(), tt.request)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result == nil {
				t.Fatal("expected result, got nil")
			}

			if result.ID != tt.wantID {
				t.Errorf("expected issue ID %d, got %d", tt.wantID, result.ID)
			}
		})
	}
}

// TestClient_Update 测试 Update 方法
func TestClient_Update(t *testing.T) {
	tests := []struct {
		name       string
		issueID    int
		request    *IssueUpdateRequest
		setupMock  func(mock *mockServer)
		wantErr    bool
		errMessage string
	}{
		{
			name:    "成功更新issue",
			issueID: 1,
			request: &IssueUpdateRequest{
				Subject: "Updated Subject",
				Notes:   "Updated via API",
			},
			setupMock: func(mock *mockServer) {
				mock.mux.HandleFunc("/issues/1.json", func(w http.ResponseWriter, r *http.Request) {
					if r.Method != http.MethodPut {
						t.Errorf("expected PUT method, got %s", r.Method)
					}

					// 验证请求体
					var body map[string]any
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Errorf("failed to decode request body: %v", err)
					}

					issue, ok := body["issue"].(map[string]any)
					if !ok {
						t.Error("expected issue field in request body")
					}

					if issue["subject"] != "Updated Subject" {
						t.Errorf("expected subject 'Updated Subject', got %v", issue["subject"])
					}

					w.WriteHeader(http.StatusNoContent)
				})
			},
			wantErr: false,
		},
		{
			name:    "更新所有字段",
			issueID: 1,
			request: &IssueUpdateRequest{
				Subject:        "Updated Subject",
				Description:    "Updated description",
				StatusID:       3,
				PriorityID:     4,
				AssignedToID:   5,
				CategoryID:     2,
				FixedVersionID: 2,
				StartDate:      "2024-02-01",
				DueDate:        "2024-02-28",
				DoneRatio:      75,
				Notes:          "Bulk update",
				PrivateNotes:   true,
			},
			setupMock: func(mock *mockServer) {
				mock.mux.HandleFunc("/issues/1.json", func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusNoContent)
				})
			},
			wantErr: false,
		},
		{
			name:    "issue不存在",
			issueID: 999,
			request: &IssueUpdateRequest{
				Subject: "Updated",
			},
			setupMock: func(mock *mockServer) {
				mock.handleError("/issues/999.json", http.StatusNotFound, "Issue not found")
			},
			wantErr:    true,
			errMessage: "not found",
		},
		{
			name:    "认证失败",
			issueID: 1,
			request: &IssueUpdateRequest{
				Subject: "Updated",
			},
			setupMock: func(mock *mockServer) {
				mock.handleError("/issues/1.json", http.StatusUnauthorized, "Unauthorized")
			},
			wantErr:    true,
			errMessage: "authentication failed",
		},
		{
			name:    "权限不足",
			issueID: 1,
			request: &IssueUpdateRequest{
				Subject: "Updated",
			},
			setupMock: func(mock *mockServer) {
				mock.handleError("/issues/1.json", http.StatusForbidden, "Forbidden")
			},
			wantErr:    true,
			errMessage: "permission denied",
		},
		{
			name:    "服务器错误",
			issueID: 1,
			request: &IssueUpdateRequest{
				Subject: "Updated",
			},
			setupMock: func(mock *mockServer) {
				mock.handleError("/issues/1.json", http.StatusInternalServerError, "Internal Server Error")
			},
			wantErr:    true,
			errMessage: "server error",
		},
		{
			name:    "验证错误",
			issueID: 1,
			request: &IssueUpdateRequest{
				DoneRatio: 150, // 无效值
			},
			setupMock: func(mock *mockServer) {
				mock.handleError("/issues/1.json", http.StatusUnprocessableEntity, "Done ratio must be between 0 and 100")
			},
			wantErr:    true,
			errMessage: "request failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockServer()
			defer mock.Close()

			tt.setupMock(mock)

			baseClient := client.NewClient(mock.URL, "test-key")
			issueClient := NewClient(baseClient)

			err := issueClient.Update(context.Background(), tt.issueID, tt.request)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// TestClient_Delete 测试 Delete 方法
func TestClient_Delete(t *testing.T) {
	tests := []struct {
		name       string
		issueID    int
		setupMock  func(mock *mockServer)
		wantErr    bool
		errMessage string
	}{
		{
			name:    "成功删除issue",
			issueID: 1,
			setupMock: func(mock *mockServer) {
				mock.mux.HandleFunc("/issues/1.json", func(w http.ResponseWriter, r *http.Request) {
					if r.Method != http.MethodDelete {
						t.Errorf("expected DELETE method, got %s", r.Method)
					}
					w.WriteHeader(http.StatusNoContent)
				})
			},
			wantErr: false,
		},
		{
			name:    "issue不存在",
			issueID: 999,
			setupMock: func(mock *mockServer) {
				mock.handleError("/issues/999.json", http.StatusNotFound, "Issue not found")
			},
			wantErr:    true,
			errMessage: "not found",
		},
		{
			name:    "认证失败",
			issueID: 1,
			setupMock: func(mock *mockServer) {
				mock.handleError("/issues/1.json", http.StatusUnauthorized, "Unauthorized")
			},
			wantErr:    true,
			errMessage: "authentication failed",
		},
		{
			name:    "权限不足",
			issueID: 1,
			setupMock: func(mock *mockServer) {
				mock.handleError("/issues/1.json", http.StatusForbidden, "Forbidden")
			},
			wantErr:    true,
			errMessage: "permission denied",
		},
		{
			name:    "服务器错误",
			issueID: 1,
			setupMock: func(mock *mockServer) {
				mock.handleError("/issues/1.json", http.StatusInternalServerError, "Internal Server Error")
			},
			wantErr:    true,
			errMessage: "server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockServer()
			defer mock.Close()

			tt.setupMock(mock)

			baseClient := client.NewClient(mock.URL, "test-key")
			issueClient := NewClient(baseClient)

			err := issueClient.Delete(context.Background(), tt.issueID)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// TestBuildListParams 测试 BuildListParams 函数
func TestBuildListParams(t *testing.T) {
	tests := []struct {
		name     string
		flags    ListFlags
		expected map[string]string
	}{
		{
			name:     "空参数",
			flags:    ListFlags{},
			expected: map[string]string{},
		},
		{
			name: "单个参数 - ProjectID",
			flags: ListFlags{
				ProjectID: 1,
			},
			expected: map[string]string{
				"project_id": "1",
			},
		},
		{
			name: "单个参数 - TrackerID",
			flags: ListFlags{
				TrackerID: 2,
			},
			expected: map[string]string{
				"tracker_id": "2",
			},
		},
		{
			name: "单个参数 - StatusID",
			flags: ListFlags{
				StatusID: 3,
			},
			expected: map[string]string{
				"status_id": "3",
			},
		},
		{
			name: "单个参数 - AssignedToID",
			flags: ListFlags{
				AssignedToID: 4,
			},
			expected: map[string]string{
				"assigned_to_id": "4",
			},
		},
		{
			name: "单个参数 - Limit",
			flags: ListFlags{
				Limit: 25,
			},
			expected: map[string]string{
				"limit": "25",
			},
		},
		{
			name: "单个参数 - Offset",
			flags: ListFlags{
				Offset: 10,
			},
			expected: map[string]string{
				"offset": "10",
			},
		},
		{
			name: "单个参数 - Query",
			flags: ListFlags{
				Query: "123",
			},
			expected: map[string]string{
				"query_id": "123",
			},
		},
		{
			name: "单个参数 - Sort",
			flags: ListFlags{
				Sort: "updated_on:desc",
			},
			expected: map[string]string{
				"sort": "updated_on:desc",
			},
		},
		{
			name: "多个参数组合",
			flags: ListFlags{
				ProjectID:    1,
				TrackerID:    2,
				StatusID:     3,
				AssignedToID: 4,
				Limit:        25,
				Offset:       10,
				Query:        "123",
				Sort:         "updated_on:desc",
			},
			expected: map[string]string{
				"project_id":     "1",
				"tracker_id":     "2",
				"status_id":      "3",
				"assigned_to_id": "4",
				"limit":          "25",
				"offset":         "10",
				"query_id":       "123",
				"sort":           "updated_on:desc",
			},
		},
		{
			name: "零值参数应被忽略",
			flags: ListFlags{
				ProjectID:    0,
				TrackerID:    0,
				StatusID:     0,
				AssignedToID: 0,
				Limit:        0,
				Offset:       0,
				Query:        "",
				Sort:         "",
			},
			expected: map[string]string{},
		},
		{
			name: "部分参数",
			flags: ListFlags{
				ProjectID: 1,
				StatusID:  2,
				Limit:     10,
			},
			expected: map[string]string{
				"project_id": "1",
				"status_id":  "2",
				"limit":      "10",
			},
		},
		{
			name: "字符串参数 - 空字符串应被忽略",
			flags: ListFlags{
				Query: "",
				Sort:  "",
			},
			expected: map[string]string{},
		},
		{
			name: "字符串参数 - 非空字符串应被包含",
			flags: ListFlags{
				Query: "456",
				Sort:  "created_on:asc",
			},
			expected: map[string]string{
				"query_id": "456",
				"sort":     "created_on:asc",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildListParams(tt.flags)

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d params, got %d", len(tt.expected), len(result))
			}

			for key, expectedValue := range tt.expected {
				if result[key] != expectedValue {
					t.Errorf("expected %s=%s, got %s=%s", key, expectedValue, key, result[key])
				}
			}

			for key, value := range result {
				if _, exists := tt.expected[key]; !exists {
					t.Errorf("unexpected param %s=%s", key, value)
				}
			}
		})
	}
}

// TestIssue_Fields 测试 Issue 结构体的字段
func TestIssue_Fields(t *testing.T) {
	mock := newMockServer()
	defer mock.Close()

	// 创建包含所有字段的 issue
	issue := sampleIssue()
	issue.Parent = &Reference{ID: 10, Name: "Parent Issue"}
	issue.FixedVersion = &Reference{ID: 1, Name: "Version 1.0"}
	issue.ClosedOn = issue.UpdatedOn
	issue.Notes = "Test notes"
	issue.PrivateNotes = true
	issue.Watchers = []Reference{
		{ID: 1, Name: "Watcher 1"},
		{ID: 2, Name: "Watcher 2"},
	}
	issue.CustomFields = []CustomField{
		{ID: 1, Name: "Custom Field 1", Value: "Value 1"},
		{ID: 2, Name: "Custom Field 2", Value: 123},
	}

	response := map[string]any{
		"issue": issue,
	}
	mock.handleJSON("/issues/1.json", response)

	baseClient := client.NewClient(mock.URL, "test-key")
	issueClient := NewClient(baseClient)

	result, err := issueClient.Get(context.Background(), 1, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 验证所有字段
	if result.ID != issue.ID {
		t.Errorf("expected ID %d, got %d", issue.ID, result.ID)
	}
	if result.Subject != issue.Subject {
		t.Errorf("expected Subject %s, got %s", issue.Subject, result.Subject)
	}
	if result.Description != issue.Description {
		t.Errorf("expected Description %s, got %s", issue.Description, result.Description)
	}
	if result.Project == nil || result.Project.ID != issue.Project.ID {
		t.Errorf("expected Project.ID %d, got %v", issue.Project.ID, result.Project)
	}
	if result.Tracker == nil || result.Tracker.ID != issue.Tracker.ID {
		t.Errorf("expected Tracker.ID %d, got %v", issue.Tracker.ID, result.Tracker)
	}
	if result.Status == nil || result.Status.ID != issue.Status.ID {
		t.Errorf("expected Status.ID %d, got %v", issue.Status.ID, result.Status)
	}
	if result.Priority == nil || result.Priority.ID != issue.Priority.ID {
		t.Errorf("expected Priority.ID %d, got %v", issue.Priority.ID, result.Priority)
	}
	if result.Author == nil || result.Author.ID != issue.Author.ID {
		t.Errorf("expected Author.ID %d, got %v", issue.Author.ID, result.Author)
	}
	if result.AssignedTo == nil || result.AssignedTo.ID != issue.AssignedTo.ID {
		t.Errorf("expected AssignedTo.ID %d, got %v", issue.AssignedTo.ID, result.AssignedTo)
	}
	if result.Category == nil || result.Category.ID != issue.Category.ID {
		t.Errorf("expected Category.ID %d, got %v", issue.Category.ID, result.Category)
	}
	if result.Parent == nil || result.Parent.ID != issue.Parent.ID {
		t.Errorf("expected Parent.ID %d, got %v", issue.Parent.ID, result.Parent)
	}
	if result.FixedVersion == nil || result.FixedVersion.ID != issue.FixedVersion.ID {
		t.Errorf("expected FixedVersion.ID %d, got %v", issue.FixedVersion.ID, result.FixedVersion)
	}
	if result.DoneRatio != issue.DoneRatio {
		t.Errorf("expected DoneRatio %d, got %d", issue.DoneRatio, result.DoneRatio)
	}
	if result.Notes != issue.Notes {
		t.Errorf("expected Notes %s, got %s", issue.Notes, result.Notes)
	}
	if result.PrivateNotes != issue.PrivateNotes {
		t.Errorf("expected PrivateNotes %v, got %v", issue.PrivateNotes, result.PrivateNotes)
	}
	if len(result.Watchers) != len(issue.Watchers) {
		t.Errorf("expected %d Watchers, got %d", len(issue.Watchers), len(result.Watchers))
	}
	if len(result.CustomFields) != len(issue.CustomFields) {
		t.Errorf("expected %d CustomFields, got %d", len(issue.CustomFields), len(result.CustomFields))
	}
}

// TestIssueList_Fields 测试 IssueList 结构体的字段
func TestIssueList_Fields(t *testing.T) {
	mock := newMockServer()
	defer mock.Close()

	issueList := sampleIssueList()
	mock.handleJSON("/issues.json", issueList)

	baseClient := client.NewClient(mock.URL, "test-key")
	issueClient := NewClient(baseClient)

	result, err := issueClient.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Issues) != len(issueList.Issues) {
		t.Errorf("expected %d issues, got %d", len(issueList.Issues), len(result.Issues))
	}
	if result.TotalCount != issueList.TotalCount {
		t.Errorf("expected TotalCount %d, got %d", issueList.TotalCount, result.TotalCount)
	}
	if result.Limit != issueList.Limit {
		t.Errorf("expected Limit %d, got %d", issueList.Limit, result.Limit)
	}
	if result.Offset != issueList.Offset {
		t.Errorf("expected Offset %d, got %d", issueList.Offset, result.Offset)
	}
}

// TestClient_Context_Cancellation 测试上下文取消
func TestClient_Context_Cancellation(t *testing.T) {
	mock := newMockServer()
	defer mock.Close()

	// 创建一个会延迟响应的处理器
	mock.mux.HandleFunc("/issues.json", func(_ http.ResponseWriter, r *http.Request) {
		// 检查上下文是否已取消
		select {
		case <-r.Context().Done():
			// 客户端已取消请求
			return
		default:
		}
		// 模拟延迟
		<-r.Context().Done()
	})

	baseClient := client.NewClient(mock.URL, "test-key")
	issueClient := NewClient(baseClient)

	// 创建一个已取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := issueClient.List(ctx, nil)
	if err == nil {
		t.Error("expected error with cancelled context, got nil")
	}
}

// TestClient_List_EmptyParams 测试空参数的 List 调用
func TestClient_List_EmptyParams(t *testing.T) {
	mock := newMockServer()
	defer mock.Close()

	mock.mux.HandleFunc("/issues.json", func(w http.ResponseWriter, r *http.Request) {
		// 验证没有查询参数
		if len(r.URL.Query()) != 0 {
			t.Errorf("expected no query params, got %v", r.URL.Query())
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sampleIssueList())
	})

	baseClient := client.NewClient(mock.URL, "test-key")
	issueClient := NewClient(baseClient)

	result, err := issueClient.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}
}

// TestClient_Get_EmptyParams 测试空参数的 Get 调用
func TestClient_Get_EmptyParams(t *testing.T) {
	mock := newMockServer()
	defer mock.Close()

	mock.mux.HandleFunc("/issues/1.json", func(w http.ResponseWriter, r *http.Request) {
		// 验证没有查询参数
		if len(r.URL.Query()) != 0 {
			t.Errorf("expected no query params, got %v", r.URL.Query())
		}
		response := map[string]any{
			"issue": sampleIssue(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	baseClient := client.NewClient(mock.URL, "test-key")
	issueClient := NewClient(baseClient)

	result, err := issueClient.Get(context.Background(), 1, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}
}

// TestReference 测试 Reference 结构体
func TestReference(t *testing.T) {
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

// TestCustomField 测试 CustomField 结构体
func TestCustomField(t *testing.T) {
	tests := []struct {
		name  string
		field CustomField
	}{
		{
			name: "字符串值",
			field: CustomField{
				ID:    1,
				Name:  "String Field",
				Value: "test value",
			},
		},
		{
			name: "整数值",
			field: CustomField{
				ID:    2,
				Name:  "Integer Field",
				Value: 123,
			},
		},
		{
			name: "布尔值",
			field: CustomField{
				ID:    3,
				Name:  "Boolean Field",
				Value: true,
			},
		},
		{
			name: "数组值",
			field: CustomField{
				ID:    4,
				Name:  "Array Field",
				Value: []string{"value1", "value2"},
			},
		},
		{
			name: "nil值",
			field: CustomField{
				ID:    5,
				Name:  "Nil Field",
				Value: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.field.ID == 0 {
				t.Error("expected non-zero ID")
			}
			if tt.field.Name == "" {
				t.Error("expected non-empty Name")
			}
		})
	}
}

// TestIssueCreateRequest 测试 IssueCreateRequest 结构体
func TestIssueCreateRequest(t *testing.T) {
	req := IssueCreateRequest{
		ProjectID:      1,
		Subject:        "Test Subject",
		Description:    "Test Description",
		TrackerID:      2,
		StatusID:       3,
		PriorityID:     4,
		AssignedToID:   5,
		CategoryID:     6,
		FixedVersionID: 7,
		ParentIssueID:  8,
		StartDate:      "2024-01-01",
		DueDate:        "2024-12-31",
		DoneRatio:      50,
		WatcherUserIDs: []int{1, 2, 3},
	}

	if req.ProjectID != 1 {
		t.Errorf("expected ProjectID 1, got %d", req.ProjectID)
	}
	if len(req.WatcherUserIDs) != 3 {
		t.Errorf("expected 3 watchers, got %d", len(req.WatcherUserIDs))
	}
}

// TestIssueUpdateRequest 测试 IssueUpdateRequest 结构体
func TestIssueUpdateRequest(t *testing.T) {
	req := IssueUpdateRequest{
		Subject:        "Updated Subject",
		Description:    "Updated Description",
		StatusID:       3,
		PriorityID:     4,
		AssignedToID:   5,
		CategoryID:     6,
		FixedVersionID: 7,
		StartDate:      "2024-01-01",
		DueDate:        "2024-12-31",
		DoneRatio:      75,
		Notes:          "Update notes",
		PrivateNotes:   true,
	}

	if req.Subject != "Updated Subject" {
		t.Errorf("expected Subject 'Updated Subject', got %s", req.Subject)
	}
	if req.PrivateNotes != true {
		t.Errorf("expected PrivateNotes true, got %v", req.PrivateNotes)
	}
}

// TestClient_List_WithAllQueryParams 测试 List 方法带所有查询参数
func TestClient_List_WithAllQueryParams(t *testing.T) {
	mock := newMockServer()
	defer mock.Close()

	mock.mux.HandleFunc("/issues.json", func(w http.ResponseWriter, r *http.Request) {
		// 验证所有查询参数
		expected := map[string]string{
			"project_id":     "1",
			"tracker_id":     "2",
			"status_id":      "3",
			"assigned_to_id": "4",
			"limit":          "10",
			"offset":         "20",
			"query_id":       "5",
			"sort":           "updated_on:desc",
		}

		for key, expectedValue := range expected {
			if r.URL.Query().Get(key) != expectedValue {
				t.Errorf("expected %s=%s, got %s", key, expectedValue, r.URL.Query().Get(key))
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sampleIssueList())
	})

	baseClient := client.NewClient(mock.URL, "test-key")
	issueClient := NewClient(baseClient)

	params := map[string]string{
		"project_id":     "1",
		"tracker_id":     "2",
		"status_id":      "3",
		"assigned_to_id": "4",
		"limit":          "10",
		"offset":         "20",
		"query_id":       "5",
		"sort":           "updated_on:desc",
	}

	result, err := issueClient.List(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}
}

// TestClient_Get_WithIncludeParam 测试 Get 方法带 include 参数
func TestClient_Get_WithIncludeParam(t *testing.T) {
	mock := newMockServer()
	defer mock.Close()

	mock.mux.HandleFunc("/issues/1.json", func(w http.ResponseWriter, r *http.Request) {
		include := r.URL.Query().Get("include")
		if include != "children,attachments,relations,changesets,journals,watchers,allowed_statuses" {
			t.Errorf("unexpected include param: %s", include)
		}

		response := map[string]any{
			"issue": sampleIssue(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	baseClient := client.NewClient(mock.URL, "test-key")
	issueClient := NewClient(baseClient)

	params := map[string]string{
		"include": "children,attachments,relations,changesets,journals,watchers,allowed_statuses",
	}

	result, err := issueClient.Get(context.Background(), 1, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}
}

// TestClient_Create_WithAllFields 测试 Create 方法带所有字段
func TestClient_Create_WithAllFields(t *testing.T) {
	mock := newMockServer()
	defer mock.Close()

	mock.mux.HandleFunc("/issues.json", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST method, got %s", r.Method)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}

		issue, ok := body["issue"].(map[string]any)
		if !ok {
			t.Fatal("expected issue field in request body")
		}

		if issue["project_id"] != float64(1) {
			t.Errorf("expected project_id 1, got %v", issue["project_id"])
		}
		if issue["subject"] != "Test with All Fields" {
			t.Errorf("expected subject 'Test with All Fields', got %v", issue["subject"])
		}
		if issue["description"] != "Testing issue creation with all fields" {
			t.Errorf("expected description, got %v", issue["description"])
		}
		if issue["tracker_id"] != float64(1) {
			t.Errorf("expected tracker_id 1, got %v", issue["tracker_id"])
		}
		if issue["status_id"] != float64(1) {
			t.Errorf("expected status_id 1, got %v", issue["status_id"])
		}
		if issue["priority_id"] != float64(2) {
			t.Errorf("expected priority_id 2, got %v", issue["priority_id"])
		}
		if issue["assigned_to_id"] != float64(1) {
			t.Errorf("expected assigned_to_id 1, got %v", issue["assigned_to_id"])
		}

		createdIssue := sampleIssue()
		response := map[string]any{
			"issue": createdIssue,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	baseClient := client.NewClient(mock.URL, "test-key")
	issueClient := NewClient(baseClient)

	request := &IssueCreateRequest{
		ProjectID:    1,
		Subject:      "Test with All Fields",
		Description:  "Testing issue creation with all fields",
		TrackerID:    1,
		StatusID:     1,
		PriorityID:   2,
		AssignedToID: 1,
	}

	result, err := issueClient.Create(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}
}

// TestClient_Update_WithNotes 测试 Update 方法带备注
func TestClient_Update_WithNotes(t *testing.T) {
	mock := newMockServer()
	defer mock.Close()

	mock.mux.HandleFunc("/issues/1.json", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT method, got %s", r.Method)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}

		issue, ok := body["issue"].(map[string]any)
		if !ok {
			t.Fatal("expected issue field in request body")
		}

		if issue["notes"] != "This is a note" {
			t.Errorf("expected notes 'This is a note', got %v", issue["notes"])
		}

		if issue["private_notes"] != true {
			t.Errorf("expected private_notes true, got %v", issue["private_notes"])
		}

		w.WriteHeader(http.StatusNoContent)
	})

	baseClient := client.NewClient(mock.URL, "test-key")
	issueClient := NewClient(baseClient)

	request := &IssueUpdateRequest{
		Notes:        "This is a note",
		PrivateNotes: true,
	}

	err := issueClient.Update(context.Background(), 1, request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestClient_Delete_Confirmation 测试 Delete 方法确认
func TestClient_Delete_Confirmation(t *testing.T) {
	mock := newMockServer()
	defer mock.Close()

	deleted := false
	mock.mux.HandleFunc("/issues/1.json", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE method, got %s", r.Method)
		}
		deleted = true
		w.WriteHeader(http.StatusNoContent)
	})

	baseClient := client.NewClient(mock.URL, "test-key")
	issueClient := NewClient(baseClient)

	err := issueClient.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !deleted {
		t.Error("expected delete to be called")
	}
}

// TestIssue_JSONSerialization 测试 Issue JSON 序列化
func TestIssue_JSONSerialization(t *testing.T) {
	issue := sampleIssue()

	data, err := json.Marshal(issue)
	if err != nil {
		t.Fatalf("failed to marshal issue: %v", err)
	}

	var unmarshaled Issue
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal issue: %v", err)
	}

	if unmarshaled.ID != issue.ID {
		t.Errorf("expected ID %d, got %d", issue.ID, unmarshaled.ID)
	}
	if unmarshaled.Subject != issue.Subject {
		t.Errorf("expected Subject %s, got %s", issue.Subject, unmarshaled.Subject)
	}
}

// TestIssueList_JSONSerialization 测试 IssueList JSON 序列化
func TestIssueList_JSONSerialization(t *testing.T) {
	issueList := sampleIssueList()

	data, err := json.Marshal(issueList)
	if err != nil {
		t.Fatalf("failed to marshal issue list: %v", err)
	}

	var unmarshaled IssueList
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal issue list: %v", err)
	}

	if unmarshaled.TotalCount != issueList.TotalCount {
		t.Errorf("expected TotalCount %d, got %d", issueList.TotalCount, unmarshaled.TotalCount)
	}
	if len(unmarshaled.Issues) != len(issueList.Issues) {
		t.Errorf("expected %d issues, got %d", len(issueList.Issues), len(unmarshaled.Issues))
	}
}

// TestIssueCreateRequest_JSONSerialization 测试 IssueCreateRequest JSON 序列化
func TestIssueCreateRequest_JSONSerialization(t *testing.T) {
	req := IssueCreateRequest{
		ProjectID:      1,
		Subject:        "Test Subject",
		Description:    "Test Description",
		TrackerID:      2,
		StatusID:       3,
		PriorityID:     4,
		AssignedToID:   5,
		CategoryID:     6,
		FixedVersionID: 7,
		ParentIssueID:  8,
		StartDate:      "2024-01-01",
		DueDate:        "2024-12-31",
		DoneRatio:      50,
		WatcherUserIDs: []int{1, 2, 3},
	}

	// 包装为 issue 字段中
	wrapped := map[string]any{
		"issue": req,
	}

	data, err := json.Marshal(wrapped)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	var unmarshaled map[string]any
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal request: %v", err)
	}

	issue, ok := unmarshaled["issue"].(map[string]any)
	if !ok {
		t.Fatal("expected issue field in unmarshaled data")
	}

	if issue["subject"] != "Test Subject" {
		t.Errorf("expected subject 'Test Subject', got %v", issue["subject"])
	}
}

// TestIssueUpdateRequest_JSONSerialization 测试 IssueUpdateRequest JSON 序列化
func TestIssueUpdateRequest_JSONSerialization(t *testing.T) {
	req := IssueUpdateRequest{
		Subject:        "Updated Subject",
		Description:    "Updated Description",
		StatusID:       3,
		PriorityID:     4,
		AssignedToID:   5,
		CategoryID:     6,
		FixedVersionID: 7,
		StartDate:      "2024-01-01",
		DueDate:        "2024-12-31",
		DoneRatio:      75,
		Notes:          "Update notes",
		PrivateNotes:   true,
	}

	// 包装为 issue 字段中
	wrapped := map[string]any{
		"issue": req,
	}

	data, err := json.Marshal(wrapped)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	var unmarshaled map[string]any
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal request: %v", err)
	}

	issue, ok := unmarshaled["issue"].(map[string]any)
	if !ok {
		t.Fatal("expected issue field in unmarshaled data")
	}

	if issue["subject"] != "Updated Subject" {
		t.Errorf("expected subject 'Updated Subject', got %v", issue["subject"])
	}
}

// TestClient_List_BadRequest 测试 List 方法错误请求
func TestClient_List_BadRequest(t *testing.T) {
	mock := newMockServer()
	defer mock.Close()

	mock.handleError("/issues.json", http.StatusBadRequest, "Bad Request")

	baseClient := client.NewClient(mock.URL, "test-key")
	issueClient := NewClient(baseClient)

	_, err := issueClient.List(context.Background(), nil)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// TestClient_Get_BadRequest 测试 Get 方法错误请求
func TestClient_Get_BadRequest(t *testing.T) {
	mock := newMockServer()
	defer mock.Close()

	mock.handleError("/issues/1.json", http.StatusBadRequest, "Bad Request")

	baseClient := client.NewClient(mock.URL, "test-key")
	issueClient := NewClient(baseClient)

	_, err := issueClient.Get(context.Background(), 1, nil)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// TestClient_Create_Conflict 测试 Create 方法冲突
func TestClient_Create_Conflict(t *testing.T) {
	mock := newMockServer()
	defer mock.Close()

	mock.handleError("/issues.json", http.StatusConflict, "Conflict")

	baseClient := client.NewClient(mock.URL, "test-key")
	issueClient := NewClient(baseClient)

	request := &IssueCreateRequest{
		ProjectID: 1,
		Subject:   "Test",
	}

	_, err := issueClient.Create(context.Background(), request)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// TestClient_Update_Conflict 测试 Update 方法冲突
func TestClient_Update_Conflict(t *testing.T) {
	mock := newMockServer()
	defer mock.Close()

	mock.handleError("/issues/1.json", http.StatusConflict, "Conflict")

	baseClient := client.NewClient(mock.URL, "test-key")
	issueClient := NewClient(baseClient)

	request := &IssueUpdateRequest{
		Subject: "Updated",
	}

	err := issueClient.Update(context.Background(), 1, request)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// TestClient_Delete_Conflict 测试 Delete 方法冲突
func TestClient_Delete_Conflict(t *testing.T) {
	mock := newMockServer()
	defer mock.Close()

	mock.handleError("/issues/1.json", http.StatusConflict, "Conflict")

	baseClient := client.NewClient(mock.URL, "test-key")
	issueClient := NewClient(baseClient)

	err := issueClient.Delete(context.Background(), 1)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// TestListFlags 测试 ListFlags 结构体
func TestListFlags(t *testing.T) {
	flags := ListFlags{
		ProjectID:    1,
		TrackerID:    2,
		StatusID:     3,
		AssignedToID: 4,
		Limit:        25,
		Offset:       10,
		Query:        "123",
		Sort:         "updated_on:desc",
	}

	if flags.ProjectID != 1 {
		t.Errorf("expected ProjectID 1, got %d", flags.ProjectID)
	}
	if flags.TrackerID != 2 {
		t.Errorf("expected TrackerID 2, got %d", flags.TrackerID)
	}
	if flags.StatusID != 3 {
		t.Errorf("expected StatusID 3, got %d", flags.StatusID)
	}
	if flags.AssignedToID != 4 {
		t.Errorf("expected AssignedToID 4, got %d", flags.AssignedToID)
	}
	if flags.Limit != 25 {
		t.Errorf("expected Limit 25, got %d", flags.Limit)
	}
	if flags.Offset != 10 {
		t.Errorf("expected Offset 10, got %d", flags.Offset)
	}
	if flags.Query != "123" {
		t.Errorf("expected Query '123', got %s", flags.Query)
	}
	if flags.Sort != "updated_on:desc" {
		t.Errorf("expected Sort 'updated_on:desc', got %s", flags.Sort)
	}
}

// TestIssue_EmptyFields 测试 Issue 空字段
func TestIssue_EmptyFields(t *testing.T) {
	issue := Issue{}

	if issue.ID != 0 {
		t.Errorf("expected ID 0, got %d", issue.ID)
	}
	if issue.Subject != "" {
		t.Errorf("expected empty Subject, got %s", issue.Subject)
	}
	if issue.Project != nil {
		t.Errorf("expected nil Project, got %v", issue.Project)
	}
	if issue.DoneRatio != 0 {
		t.Errorf("expected DoneRatio 0, got %d", issue.DoneRatio)
	}
}

// TestIssueList_EmptyFields 测试 IssueList 空字段
func TestIssueList_EmptyFields(t *testing.T) {
	issueList := IssueList{}

	if issueList.Issues != nil {
		t.Errorf("expected nil Issues, got %v", issueList.Issues)
	}
	if issueList.TotalCount != 0 {
		t.Errorf("expected TotalCount 0, got %d", issueList.TotalCount)
	}
	if issueList.Limit != 0 {
		t.Errorf("expected Limit 0, got %d", issueList.Limit)
	}
	if issueList.Offset != 0 {
		t.Errorf("expected Offset 0, got %d", issueList.Offset)
	}
}

// TestIssueCreateRequest_EmptyFields 测试 IssueCreateRequest 空字段
func TestIssueCreateRequest_EmptyFields(t *testing.T) {
	req := IssueCreateRequest{}

	if req.ProjectID != 0 {
		t.Errorf("expected ProjectID 0, got %d", req.ProjectID)
	}
	if req.Subject != "" {
		t.Errorf("expected empty Subject, got %s", req.Subject)
	}
	if req.WatcherUserIDs != nil {
		t.Errorf("expected nil WatcherUserIDs, got %v", req.WatcherUserIDs)
	}
}

// TestIssueUpdateRequest_EmptyFields 测试 IssueUpdateRequest 空字段
func TestIssueUpdateRequest_EmptyFields(t *testing.T) {
	req := IssueUpdateRequest{}

	if req.Subject != "" {
		t.Errorf("expected empty Subject, got %s", req.Subject)
	}
	if req.DoneRatio != 0 {
		t.Errorf("expected DoneRatio 0, got %d", req.DoneRatio)
	}
	if req.PrivateNotes != false {
		t.Errorf("expected PrivateNotes false, got %v", req.PrivateNotes)
	}
}

// TestCustomField_EmptyFields 测试 CustomField 空字段
func TestCustomField_EmptyFields(t *testing.T) {
	field := CustomField{}

	if field.ID != 0 {
		t.Errorf("expected ID 0, got %d", field.ID)
	}
	if field.Name != "" {
		t.Errorf("expected empty Name, got %s", field.Name)
	}
	if field.Value != nil {
		t.Errorf("expected nil Value, got %v", field.Value)
	}
}

// TestReference_EmptyFields 测试 Reference 空字段
func TestReference_EmptyFields(t *testing.T) {
	ref := Reference{}

	if ref.ID != 0 {
		t.Errorf("expected ID 0, got %d", ref.ID)
	}
	if ref.Name != "" {
		t.Errorf("expected empty Name, got %s", ref.Name)
	}
}
