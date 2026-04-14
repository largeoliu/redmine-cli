// internal/resources/users/client_test.go
package users

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/largeoliu/redmine-cli/internal/client"
	"github.com/largeoliu/redmine-cli/internal/testutil"
)

// sampleUser 返回一个示例 User 用于测试
func sampleUser() User {
	createdOn := time.Date(2023, 6, 15, 8, 0, 0, 0, time.UTC)
	lastLoginOn := time.Date(2024, 1, 20, 12, 30, 0, 0, time.UTC)

	return User{
		ID:                 1,
		Login:              "johndoe",
		Firstname:          "John",
		Lastname:           "Doe",
		Mail:               "john.doe@example.com",
		CreatedOn:          &createdOn,
		LastLoginOn:        &lastLoginOn,
		Admin:              false,
		Status:             1,
		MustChangePassword: false,
		AvatarURL:          "https://example.com/avatar/johndoe.png",
	}
}

// newTestClient 创建一个用于测试的 users.Client
func newTestClient(mockServer *testutil.MockServer) *Client {
	c := client.NewClient(mockServer.URL, "test-api-key")
	return NewClient(c)
}

func TestNewClient(t *testing.T) {
	c := client.NewClient("https://example.com", "test-key")
	usersClient := NewClient(c)
	if usersClient == nil {
		t.Fatal("expected non-nil client")
	}
	if usersClient.client == nil {
		t.Fatal("expected non-nil internal client")
	}
}

func TestClient_List_Success(t *testing.T) {
	mock := testutil.NewMockServer(t)
	t.Cleanup(mock.Close)

	// 准备测试数据
	user1 := sampleUser()
	user2 := sampleUser()
	user2.ID = 2
	user2.Login = "janedoe"
	user2.Mail = "jane.doe@example.com"

	response := UserList{
		Users:      []User{user1, user2},
		TotalCount: 2,
		Limit:      25,
		Offset:     0,
	}

	mock.HandleJSON("/users.json", response)

	usersClient := newTestClient(mock)
	result, err := usersClient.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TotalCount != 2 {
		t.Errorf("expected TotalCount 2, got %d", result.TotalCount)
	}
	if len(result.Users) != 2 {
		t.Errorf("expected 2 users, got %d", len(result.Users))
	}
	if result.Users[0].Login != "johndoe" {
		t.Errorf("expected first user login 'johndoe', got %s", result.Users[0].Login)
	}
}

func TestClient_List_WithParams(t *testing.T) {
	mock := testutil.NewMockServer(t)
	t.Cleanup(mock.Close)

	response := UserList{
		Users:      []User{sampleUser()},
		TotalCount: 1,
		Limit:      10,
		Offset:     0,
	}

	// 使用自定义 handler 来验证请求参数
	mock.Handle("/users.json", func(w http.ResponseWriter, r *http.Request) {
		// 验证查询参数
		query := r.URL.Query()
		if query.Get("status") != "1" {
			t.Errorf("expected status=1, got %s", query.Get("status"))
		}
		if query.Get("limit") != "10" {
			t.Errorf("expected limit=10, got %s", query.Get("limit"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	usersClient := newTestClient(mock)
	params := map[string]string{
		"status": "1",
		"limit":  "10",
	}
	result, err := usersClient.List(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TotalCount != 1 {
		t.Errorf("expected TotalCount 1, got %d", result.TotalCount)
	}
}

func TestClient_List_Empty(t *testing.T) {
	mock := testutil.NewMockServer(t)
	t.Cleanup(mock.Close)

	response := UserList{
		Users:      []User{},
		TotalCount: 0,
		Limit:      25,
		Offset:     0,
	}

	mock.HandleJSON("/users.json", response)

	usersClient := newTestClient(mock)
	result, err := usersClient.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TotalCount != 0 {
		t.Errorf("expected TotalCount 0, got %d", result.TotalCount)
	}
	if len(result.Users) != 0 {
		t.Errorf("expected 0 users, got %d", len(result.Users))
	}
}

func TestClient_List_Error(t *testing.T) {
	mock := testutil.NewMockServer(t)
	t.Cleanup(mock.Close)

	mock.HandleError("/users.json", http.StatusUnauthorized, "Unauthorized")

	usersClient := newTestClient(mock)
	_, err := usersClient.List(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_Get_Success(t *testing.T) {
	mock := testutil.NewMockServer(t)
	t.Cleanup(mock.Close)

	sample := sampleUser()
	response := struct {
		User User `json:"user"`
	}{
		User: sample,
	}

	mock.HandleJSON("/users/1.json", response)

	usersClient := newTestClient(mock)
	result, err := usersClient.Get(context.Background(), 1, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != 1 {
		t.Errorf("expected ID 1, got %d", result.ID)
	}
	if result.Login != "johndoe" {
		t.Errorf("expected login 'johndoe', got %s", result.Login)
	}
	if result.Mail != "john.doe@example.com" {
		t.Errorf("expected mail 'john.doe@example.com', got %s", result.Mail)
	}
}

func TestClient_Get_WithParams(t *testing.T) {
	mock := testutil.NewMockServer(t)
	t.Cleanup(mock.Close)

	sample := sampleUser()
	response := struct {
		User User `json:"user"`
	}{
		User: sample,
	}

	// 使用自定义 handler 来验证请求参数
	mock.Handle("/users/1.json", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if query.Get("include") != "memberships,groups" {
			t.Errorf("expected include=memberships,groups, got %s", query.Get("include"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	usersClient := newTestClient(mock)
	params := map[string]string{
		"include": "memberships,groups",
	}
	result, err := usersClient.Get(context.Background(), 1, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != 1 {
		t.Errorf("expected ID 1, got %d", result.ID)
	}
}

func TestClient_Get_NotFound(t *testing.T) {
	mock := testutil.NewMockServer(t)
	t.Cleanup(mock.Close)

	mock.HandleError("/users/999.json", http.StatusNotFound, "User not found")

	usersClient := newTestClient(mock)
	_, err := usersClient.Get(context.Background(), 999, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_Get_Forbidden(t *testing.T) {
	mock := testutil.NewMockServer(t)
	t.Cleanup(mock.Close)

	mock.HandleError("/users/1.json", http.StatusForbidden, "Forbidden")

	usersClient := newTestClient(mock)
	_, err := usersClient.Get(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_GetCurrent_Success(t *testing.T) {
	mock := testutil.NewMockServer(t)
	t.Cleanup(mock.Close)

	sample := sampleUser()
	response := struct {
		User User `json:"user"`
	}{
		User: sample,
	}

	mock.HandleJSON("/users/current.json", response)

	usersClient := newTestClient(mock)
	result, err := usersClient.GetCurrent(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != 1 {
		t.Errorf("expected ID 1, got %d", result.ID)
	}
	if result.Login != "johndoe" {
		t.Errorf("expected login 'johndoe', got %s", result.Login)
	}
}

func TestClient_GetCurrent_Unauthorized(t *testing.T) {
	mock := testutil.NewMockServer(t)
	t.Cleanup(mock.Close)

	mock.HandleError("/users/current.json", http.StatusUnauthorized, "Unauthorized")

	usersClient := newTestClient(mock)
	_, err := usersClient.GetCurrent(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_Create_Success(t *testing.T) {
	mock := testutil.NewMockServer(t)
	t.Cleanup(mock.Close)

	createdUser := sampleUser()
	createdUser.ID = 10
	createdUser.Login = "newuser"
	createdUser.Mail = "newuser@example.com"

	response := struct {
		User User `json:"user"`
	}{
		User: createdUser,
	}

	// 使用自定义 handler 来验证请求
	mock.Handle("/users.json", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST method, got %s", r.Method)
		}

		// 验证请求体
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}

		var req map[string]any
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("failed to unmarshal request body: %v", err)
		}

		userData, ok := req["user"].(map[string]any)
		if !ok {
			t.Fatal("expected user object in request body")
		}

		if userData["login"] != "newuser" {
			t.Errorf("expected login 'newuser', got %v", userData["login"])
		}
		if userData["mail"] != "newuser@example.com" {
			t.Errorf("expected mail 'newuser@example.com', got %v", userData["mail"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	})

	usersClient := newTestClient(mock)
	req := &UserCreateRequest{
		Login:     "newuser",
		Firstname: "New",
		Lastname:  "User",
		Mail:      "newuser@example.com",
		Password:  "password123",
	}

	result, err := usersClient.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != 10 {
		t.Errorf("expected ID 10, got %d", result.ID)
	}
	if result.Login != "newuser" {
		t.Errorf("expected login 'newuser', got %s", result.Login)
	}
}

func TestClient_Create_ValidationError(t *testing.T) {
	mock := testutil.NewMockServer(t)
	t.Cleanup(mock.Close)

	mock.HandleError("/users.json", http.StatusUnprocessableEntity, "Validation failed: Login can't be blank")

	usersClient := newTestClient(mock)
	req := &UserCreateRequest{
		Firstname: "New",
		Lastname:  "User",
	}

	_, err := usersClient.Create(context.Background(), req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_Create_Forbidden(t *testing.T) {
	mock := testutil.NewMockServer(t)
	t.Cleanup(mock.Close)

	mock.HandleError("/users.json", http.StatusForbidden, "Forbidden")

	usersClient := newTestClient(mock)
	req := &UserCreateRequest{
		Login:     "newuser",
		Firstname: "New",
		Lastname:  "User",
		Mail:      "newuser@example.com",
		Password:  "password123",
	}

	_, err := usersClient.Create(context.Background(), req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_Update_Success(t *testing.T) {
	mock := testutil.NewMockServer(t)
	t.Cleanup(mock.Close)

	// 使用自定义 handler 来验证请求
	mock.Handle("/users/1.json", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT method, got %s", r.Method)
		}

		// 验证请求体
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}

		var req map[string]any
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("failed to unmarshal request body: %v", err)
		}

		userData, ok := req["user"].(map[string]any)
		if !ok {
			t.Fatal("expected user object in request body")
		}

		if userData["firstname"] != "Updated" {
			t.Errorf("expected firstname 'Updated', got %v", userData["firstname"])
		}

		w.WriteHeader(http.StatusNoContent)
	})

	usersClient := newTestClient(mock)
	req := &UserUpdateRequest{
		Firstname: "Updated",
		Lastname:  "Name",
	}

	err := usersClient.Update(context.Background(), 1, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_Update_NotFound(t *testing.T) {
	mock := testutil.NewMockServer(t)
	t.Cleanup(mock.Close)

	mock.HandleError("/users/999.json", http.StatusNotFound, "User not found")

	usersClient := newTestClient(mock)
	req := &UserUpdateRequest{
		Firstname: "Updated",
	}

	err := usersClient.Update(context.Background(), 999, req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_Update_Forbidden(t *testing.T) {
	mock := testutil.NewMockServer(t)
	t.Cleanup(mock.Close)

	mock.HandleError("/users/1.json", http.StatusForbidden, "Forbidden")

	usersClient := newTestClient(mock)
	req := &UserUpdateRequest{
		Firstname: "Updated",
	}

	err := usersClient.Update(context.Background(), 1, req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_Delete_Success(t *testing.T) {
	mock := testutil.NewMockServer(t)
	t.Cleanup(mock.Close)

	// 使用自定义 handler 来验证请求
	mock.Handle("/users/1.json", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE method, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	usersClient := newTestClient(mock)
	err := usersClient.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_Delete_NotFound(t *testing.T) {
	mock := testutil.NewMockServer(t)
	t.Cleanup(mock.Close)

	mock.HandleError("/users/999.json", http.StatusNotFound, "User not found")

	usersClient := newTestClient(mock)
	err := usersClient.Delete(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_Delete_Forbidden(t *testing.T) {
	mock := testutil.NewMockServer(t)
	t.Cleanup(mock.Close)

	mock.HandleError("/users/1.json", http.StatusForbidden, "Forbidden")

	usersClient := newTestClient(mock)
	err := usersClient.Delete(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestBuildListParams_AllFields(t *testing.T) {
	flags := ListFlags{
		Status:  1,
		Name:    "john",
		GroupID: 5,
		Limit:   10,
		Offset:  20,
	}

	params := BuildListParams(flags)

	if params["status"] != "1" {
		t.Errorf("expected status '1', got %s", params["status"])
	}
	if params["name"] != "john" {
		t.Errorf("expected name 'john', got %s", params["name"])
	}
	if params["group_id"] != "5" {
		t.Errorf("expected group_id '5', got %s", params["group_id"])
	}
	if params["limit"] != "10" {
		t.Errorf("expected limit '10', got %s", params["limit"])
	}
	if params["offset"] != "20" {
		t.Errorf("expected offset '20', got %s", params["offset"])
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
	tests := []struct {
		name     string
		flags    ListFlags
		expected map[string]string
	}{
		{
			name: "only status",
			flags: ListFlags{
				Status: 1,
			},
			expected: map[string]string{
				"status": "1",
			},
		},
		{
			name: "only name",
			flags: ListFlags{
				Name: "test",
			},
			expected: map[string]string{
				"name": "test",
			},
		},
		{
			name: "only group_id",
			flags: ListFlags{
				GroupID: 10,
			},
			expected: map[string]string{
				"group_id": "10",
			},
		},
		{
			name: "only limit",
			flags: ListFlags{
				Limit: 50,
			},
			expected: map[string]string{
				"limit": "50",
			},
		},
		{
			name: "only offset",
			flags: ListFlags{
				Offset: 100,
			},
			expected: map[string]string{
				"offset": "100",
			},
		},
		{
			name: "status and limit",
			flags: ListFlags{
				Status: 1,
				Limit:  25,
			},
			expected: map[string]string{
				"status": "1",
				"limit":  "25",
			},
		},
		{
			name: "zero values should be excluded",
			flags: ListFlags{
				Status:  0,
				Name:    "",
				GroupID: 0,
				Limit:   0,
				Offset:  0,
			},
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := BuildListParams(tt.flags)

			if len(params) != len(tt.expected) {
				t.Errorf("expected %d params, got %d", len(tt.expected), len(params))
			}

			for key, expectedValue := range tt.expected {
				if params[key] != expectedValue {
					t.Errorf("expected %s='%s', got '%s'", key, expectedValue, params[key])
				}
			}
		})
	}
}

func TestBuildListParams_ZeroValuesExcluded(t *testing.T) {
	flags := ListFlags{
		Status:  0,
		Name:    "",
		GroupID: 0,
		Limit:   0,
		Offset:  0,
	}

	params := BuildListParams(flags)

	// 所有零值都应该被排除
	for key, value := range params {
		t.Errorf("unexpected param %s=%s (zero values should be excluded)", key, value)
	}
}

func TestUser_Fields(t *testing.T) {
	// 测试 User 结构体的字段
	user := sampleUser()

	if user.ID != 1 {
		t.Errorf("expected ID 1, got %d", user.ID)
	}
	if user.Login != "johndoe" {
		t.Errorf("expected Login 'johndoe', got %s", user.Login)
	}
	if user.Firstname != "John" {
		t.Errorf("expected Firstname 'John', got %s", user.Firstname)
	}
	if user.Lastname != "Doe" {
		t.Errorf("expected Lastname 'Doe', got %s", user.Lastname)
	}
	if user.Mail != "john.doe@example.com" {
		t.Errorf("expected Mail 'john.doe@example.com', got %s", user.Mail)
	}
	if user.Admin != false {
		t.Errorf("expected Admin false, got %v", user.Admin)
	}
	if user.Status != 1 {
		t.Errorf("expected Status 1, got %d", user.Status)
	}
	if user.MustChangePassword != false {
		t.Errorf("expected MustChangePassword false, got %v", user.MustChangePassword)
	}
	if user.AvatarURL != "https://example.com/avatar/johndoe.png" {
		t.Errorf("expected AvatarURL 'https://example.com/avatar/johndoe.png', got %s", user.AvatarURL)
	}
}

func TestUserList_Fields(t *testing.T) {
	// 测试 UserList 结构体的字段
	user := sampleUser()
	userList := UserList{
		Users:      []User{user},
		TotalCount: 1,
		Limit:      25,
		Offset:     0,
	}

	if userList.TotalCount != 1 {
		t.Errorf("expected TotalCount 1, got %d", userList.TotalCount)
	}
	if userList.Limit != 25 {
		t.Errorf("expected Limit 25, got %d", userList.Limit)
	}
	if userList.Offset != 0 {
		t.Errorf("expected Offset 0, got %d", userList.Offset)
	}
	if len(userList.Users) != 1 {
		t.Errorf("expected 1 user, got %d", len(userList.Users))
	}
}

func TestCustomField(t *testing.T) {
	// 测试 CustomField 结构体
	cf := CustomField{
		ID:    1,
		Name:  "Department",
		Value: "Engineering",
	}

	if cf.ID != 1 {
		t.Errorf("expected ID 1, got %d", cf.ID)
	}
	if cf.Name != "Department" {
		t.Errorf("expected Name 'Department', got %s", cf.Name)
	}
	if cf.Value != "Engineering" {
		t.Errorf("expected Value 'Engineering', got %v", cf.Value)
	}
}

func TestUserCreateRequest_Fields(t *testing.T) {
	// 测试 UserCreateRequest 结构体
	req := UserCreateRequest{
		Login:              "testuser",
		Firstname:          "Test",
		Lastname:           "User",
		Mail:               "test@example.com",
		Password:           "password123",
		Admin:              true,
		Status:             1,
		AuthSourceID:       1,
		MustChangePassword: true,
	}

	if req.Login != "testuser" {
		t.Errorf("expected Login 'testuser', got %s", req.Login)
	}
	if req.Firstname != "Test" {
		t.Errorf("expected Firstname 'Test', got %s", req.Firstname)
	}
	if req.Lastname != "User" {
		t.Errorf("expected Lastname 'User', got %s", req.Lastname)
	}
	if req.Mail != "test@example.com" {
		t.Errorf("expected Mail 'test@example.com', got %s", req.Mail)
	}
	if req.Password != "password123" {
		t.Errorf("expected Password 'password123', got %s", req.Password)
	}
	if req.Admin != true {
		t.Errorf("expected Admin true, got %v", req.Admin)
	}
	if req.Status != 1 {
		t.Errorf("expected Status 1, got %d", req.Status)
	}
	if req.AuthSourceID != 1 {
		t.Errorf("expected AuthSourceID 1, got %d", req.AuthSourceID)
	}
	if req.MustChangePassword != true {
		t.Errorf("expected MustChangePassword true, got %v", req.MustChangePassword)
	}
}

func TestUserUpdateRequest_Fields(t *testing.T) {
	// 测试 UserUpdateRequest 结构体
	req := UserUpdateRequest{
		Login:              "updateduser",
		Firstname:          "Updated",
		Lastname:           "Name",
		Mail:               "updated@example.com",
		Password:           "newpassword123",
		Admin:              false,
		Status:             0,
		AuthSourceID:       2,
		MustChangePassword: false,
	}

	if req.Login != "updateduser" {
		t.Errorf("expected Login 'updateduser', got %s", req.Login)
	}
	if req.Firstname != "Updated" {
		t.Errorf("expected Firstname 'Updated', got %s", req.Firstname)
	}
	if req.Lastname != "Name" {
		t.Errorf("expected Lastname 'Name', got %s", req.Lastname)
	}
	if req.Mail != "updated@example.com" {
		t.Errorf("expected Mail 'updated@example.com', got %s", req.Mail)
	}
	if req.Password != "newpassword123" {
		t.Errorf("expected Password 'newpassword123', got %s", req.Password)
	}
	if req.Admin != false {
		t.Errorf("expected Admin false, got %v", req.Admin)
	}
	if req.Status != 0 {
		t.Errorf("expected Status 0, got %d", req.Status)
	}
	if req.AuthSourceID != 2 {
		t.Errorf("expected AuthSourceID 2, got %d", req.AuthSourceID)
	}
	if req.MustChangePassword != false {
		t.Errorf("expected MustChangePassword false, got %v", req.MustChangePassword)
	}
}

func TestClient_List_ContextCancellation(t *testing.T) {
	mock := testutil.NewMockServer(t)
	t.Cleanup(mock.Close)

	response := UserList{
		Users:      []User{sampleUser()},
		TotalCount: 1,
		Limit:      25,
		Offset:     0,
	}

	mock.HandleJSON("/users.json", response)

	usersClient := newTestClient(mock)

	// 创建一个已取消的 context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := usersClient.List(ctx, nil)
	if err == nil {
		t.Fatal("expected error due to cancelled context, got nil")
	}
}

func TestClient_Get_ContextCancellation(t *testing.T) {
	mock := testutil.NewMockServer(t)
	t.Cleanup(mock.Close)

	sample := sampleUser()
	response := struct {
		User User `json:"user"`
	}{
		User: sample,
	}

	mock.HandleJSON("/users/1.json", response)

	usersClient := newTestClient(mock)

	// 创建一个已取消的 context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := usersClient.Get(ctx, 1, nil)
	if err == nil {
		t.Fatal("expected error due to cancelled context, got nil")
	}
}

func TestClient_Create_ContextCancellation(t *testing.T) {
	mock := testutil.NewMockServer(t)
	t.Cleanup(mock.Close)

	usersClient := newTestClient(mock)

	// 创建一个已取消的 context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req := &UserCreateRequest{
		Login:     "test",
		Firstname: "Test",
		Lastname:  "User",
		Mail:      "test@example.com",
		Password:  "password",
	}

	_, err := usersClient.Create(ctx, req)
	if err == nil {
		t.Fatal("expected error due to cancelled context, got nil")
	}
}

func TestClient_Update_ContextCancellation(t *testing.T) {
	mock := testutil.NewMockServer(t)
	t.Cleanup(mock.Close)

	usersClient := newTestClient(mock)

	// 创建一个已取消的 context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req := &UserUpdateRequest{
		Firstname: "Updated",
	}

	err := usersClient.Update(ctx, 1, req)
	if err == nil {
		t.Fatal("expected error due to cancelled context, got nil")
	}
}

func TestClient_Delete_ContextCancellation(t *testing.T) {
	mock := testutil.NewMockServer(t)
	t.Cleanup(mock.Close)

	usersClient := newTestClient(mock)

	// 创建一个已取消的 context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := usersClient.Delete(ctx, 1)
	if err == nil {
		t.Fatal("expected error due to cancelled context, got nil")
	}
}

func TestClient_GetCurrent_ContextCancellation(t *testing.T) {
	mock := testutil.NewMockServer(t)
	t.Cleanup(mock.Close)

	usersClient := newTestClient(mock)

	// 创建一个已取消的 context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := usersClient.GetCurrent(ctx)
	if err == nil {
		t.Fatal("expected error due to cancelled context, got nil")
	}
}

// 边界测试：测试不同 ID 值
func TestClient_Get_DifferentIDs(t *testing.T) {
	tests := []struct {
		name string
		id   int
	}{
		{"ID 1", 1},
		{"ID 100", 100},
		{"ID 999999", 999999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := testutil.NewMockServer(t)
			t.Cleanup(mock.Close)

			sample := sampleUser()
			sample.ID = tt.id
			response := struct {
				User User `json:"user"`
			}{
				User: sample,
			}

			mock.HandleJSON("/users/"+strconv.Itoa(tt.id)+".json", response)

			usersClient := newTestClient(mock)
			result, err := usersClient.Get(context.Background(), tt.id, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ID != tt.id {
				t.Errorf("expected ID %d, got %d", tt.id, result.ID)
			}
		})
	}
}

// 测试 User 结构体包含 CustomFields
func TestUser_WithCustomFields(t *testing.T) {
	mock := testutil.NewMockServer(t)
	t.Cleanup(mock.Close)

	sample := sampleUser()
	sample.CustomFields = []CustomField{
		{ID: 1, Name: "Department", Value: "Engineering"},
		{ID: 2, Name: "Location", Value: "Remote"},
	}

	response := struct {
		User User `json:"user"`
	}{
		User: sample,
	}

	mock.HandleJSON("/users/1.json", response)

	usersClient := newTestClient(mock)
	result, err := usersClient.Get(context.Background(), 1, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.CustomFields) != 2 {
		t.Errorf("expected 2 custom fields, got %d", len(result.CustomFields))
	}
	if result.CustomFields[0].Name != "Department" {
		t.Errorf("expected first custom field name 'Department', got %s", result.CustomFields[0].Name)
	}
}

// 测试服务器错误响应
func TestClient_List_ServerError(t *testing.T) {
	mock := testutil.NewMockServer(t)
	t.Cleanup(mock.Close)

	mock.HandleError("/users.json", http.StatusInternalServerError, "Internal Server Error")

	usersClient := newTestClient(mock)
	_, err := usersClient.List(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_Get_ServerError(t *testing.T) {
	mock := testutil.NewMockServer(t)
	t.Cleanup(mock.Close)

	mock.HandleError("/users/1.json", http.StatusInternalServerError, "Internal Server Error")

	usersClient := newTestClient(mock)
	_, err := usersClient.Get(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_Create_ServerError(t *testing.T) {
	mock := testutil.NewMockServer(t)
	t.Cleanup(mock.Close)

	mock.HandleError("/users.json", http.StatusInternalServerError, "Internal Server Error")

	usersClient := newTestClient(mock)
	req := &UserCreateRequest{
		Login:     "test",
		Firstname: "Test",
		Lastname:  "User",
		Mail:      "test@example.com",
		Password:  "password",
	}

	_, err := usersClient.Create(context.Background(), req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_Update_ServerError(t *testing.T) {
	mock := testutil.NewMockServer(t)
	t.Cleanup(mock.Close)

	mock.HandleError("/users/1.json", http.StatusInternalServerError, "Internal Server Error")

	usersClient := newTestClient(mock)
	req := &UserUpdateRequest{
		Firstname: "Updated",
	}

	err := usersClient.Update(context.Background(), 1, req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_Delete_ServerError(t *testing.T) {
	mock := testutil.NewMockServer(t)
	t.Cleanup(mock.Close)

	mock.HandleError("/users/1.json", http.StatusInternalServerError, "Internal Server Error")

	usersClient := newTestClient(mock)
	err := usersClient.Delete(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_GetCurrent_ServerError(t *testing.T) {
	mock := testutil.NewMockServer(t)
	t.Cleanup(mock.Close)

	mock.HandleError("/users/current.json", http.StatusInternalServerError, "Internal Server Error")

	usersClient := newTestClient(mock)
	_, err := usersClient.GetCurrent(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// 测试速率限制错误
func TestClient_List_RateLimitError(t *testing.T) {
	mock := testutil.NewMockServer(t)
	t.Cleanup(mock.Close)

	mock.HandleError("/users.json", http.StatusTooManyRequests, "Rate limit exceeded")

	usersClient := newTestClient(mock)
	_, err := usersClient.List(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// 测试 ListFlags 结构体
func TestListFlags_Fields(t *testing.T) {
	flags := ListFlags{
		Status:  1,
		Name:    "test",
		GroupID: 5,
		Limit:   10,
		Offset:  20,
	}

	if flags.Status != 1 {
		t.Errorf("expected Status 1, got %d", flags.Status)
	}
	if flags.Name != "test" {
		t.Errorf("expected Name 'test', got %s", flags.Name)
	}
	if flags.GroupID != 5 {
		t.Errorf("expected GroupID 5, got %d", flags.GroupID)
	}
	if flags.Limit != 10 {
		t.Errorf("expected Limit 10, got %d", flags.Limit)
	}
	if flags.Offset != 20 {
		t.Errorf("expected Offset 20, got %d", flags.Offset)
	}
}

// 测试 User 结构体的时间字段
func TestUser_TimeFields(t *testing.T) {
	mock := testutil.NewMockServer(t)
	t.Cleanup(mock.Close)

	createdOn := time.Date(2023, 6, 15, 8, 0, 0, 0, time.UTC)
	lastLoginOn := time.Date(2024, 1, 20, 12, 30, 0, 0, time.UTC)

	sample := sampleUser()
	sample.CreatedOn = &createdOn
	sample.LastLoginOn = &lastLoginOn

	response := struct {
		User User `json:"user"`
	}{
		User: sample,
	}

	mock.HandleJSON("/users/1.json", response)

	usersClient := newTestClient(mock)
	result, err := usersClient.Get(context.Background(), 1, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.CreatedOn == nil {
		t.Fatal("expected CreatedOn to be non-nil")
	}
	if !result.CreatedOn.Equal(createdOn) {
		t.Errorf("expected CreatedOn %v, got %v", createdOn, result.CreatedOn)
	}

	if result.LastLoginOn == nil {
		t.Fatal("expected LastLoginOn to be non-nil")
	}
	if !result.LastLoginOn.Equal(lastLoginOn) {
		t.Errorf("expected LastLoginOn %v, got %v", lastLoginOn, result.LastLoginOn)
	}
}

// 测试 User 结构体的 AuthSourceID 字段
func TestUser_AuthSourceID(t *testing.T) {
	mock := testutil.NewMockServer(t)
	t.Cleanup(mock.Close)

	sample := sampleUser()
	sample.AuthSourceID = 5

	response := struct {
		User User `json:"user"`
	}{
		User: sample,
	}

	mock.HandleJSON("/users/1.json", response)

	usersClient := newTestClient(mock)
	result, err := usersClient.Get(context.Background(), 1, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.AuthSourceID != 5 {
		t.Errorf("expected AuthSourceID 5, got %d", result.AuthSourceID)
	}
}

// 测试 Admin 用户
func TestUser_AdminUser(t *testing.T) {
	mock := testutil.NewMockServer(t)
	t.Cleanup(mock.Close)

	sample := sampleUser()
	sample.Admin = true

	response := struct {
		User User `json:"user"`
	}{
		User: sample,
	}

	mock.HandleJSON("/users/1.json", response)

	usersClient := newTestClient(mock)
	result, err := usersClient.Get(context.Background(), 1, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Admin != true {
		t.Errorf("expected Admin true, got %v", result.Admin)
	}
}

// 测试 MustChangePassword 字段
func TestUser_MustChangePassword(t *testing.T) {
	mock := testutil.NewMockServer(t)
	t.Cleanup(mock.Close)

	sample := sampleUser()
	sample.MustChangePassword = true

	response := struct {
		User User `json:"user"`
	}{
		User: sample,
	}

	mock.HandleJSON("/users/1.json", response)

	usersClient := newTestClient(mock)
	result, err := usersClient.Get(context.Background(), 1, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.MustChangePassword != true {
		t.Errorf("expected MustChangePassword true, got %v", result.MustChangePassword)
	}
}
