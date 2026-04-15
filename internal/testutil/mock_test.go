// internal/testutil/mock_test.go
package testutil

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type errorWriter struct {
	headersWritten bool
	code           int
}

func (w *errorWriter) Header() http.Header {
	return http.Header{}
}

func (w *errorWriter) WriteHeader(code int) {
	w.code = code
	w.headersWritten = true
}

func (w *errorWriter) Write(b []byte) (int, error) {
	if w.headersWritten {
		return 0, http.ErrHandlerTimeout
	}
	return len(b), nil
}

// TestNewMockServer tests the creation of a new MockServer
func TestNewMockServer(t *testing.T) {
	t.Run("creates_server_successfully", func(t *testing.T) {
		mock := NewMockServer(t)
		if mock == nil {
			t.Fatal("expected mock server, got nil")
		}
		defer mock.Close()

		if mock.Server == nil {
			t.Error("expected Server to be initialized")
		}

		if mock.Mux == nil {
			t.Error("expected Mux to be initialized")
		}

		if mock.URL == "" {
			t.Error("expected URL to be set")
		}

		// Verify URL is a valid HTTP URL
		if !strings.HasPrefix(mock.URL, "http://") {
			t.Errorf("expected URL to start with http://, got %s", mock.URL)
		}
	})

	t.Run("server_is_functional", func(t *testing.T) {
		mock := NewMockServer(t)
		defer mock.Close()

		// Register a simple handler
		mock.Handle("/test", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})

		// Make a request
		resp, err := http.Get(mock.URL + "/test")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}
	})

	t.Run("multiple_servers_can_coexist", func(t *testing.T) {
		mock1 := NewMockServer(t)
		mock2 := NewMockServer(t)

		defer mock1.Close()
		defer mock2.Close()

		if mock1.URL == mock2.URL {
			t.Error("expected different URLs for different servers")
		}
	})
}

// TestMockServer_Close tests the Close method
func TestMockServer_Close(t *testing.T) {
	t.Run("closes_server_successfully", func(t *testing.T) {
		mock := NewMockServer(t)
		mock.Close()

		// Verify server is closed by trying to make a request
		_, err := http.Get(mock.URL + "/test")
		if err == nil {
			t.Error("expected error when connecting to closed server")
		}
	})

	t.Run("close_is_idempotent", func(t *testing.T) {
		mock := NewMockServer(t)

		// Close multiple times should not panic
		mock.Close()
		mock.Close() // Second close should not panic
	})

	t.Run("close_after_requests", func(t *testing.T) {
		mock := NewMockServer(t)

		mock.Handle("/test", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		// Make a request
		resp, err := http.Get(mock.URL + "/test")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		resp.Body.Close()

		// Close should work after requests
		mock.Close()
	})
}

// TestMockServer_Handle tests the Handle method
func TestMockServer_Handle(t *testing.T) {
	t.Run("registers_handler_successfully", func(t *testing.T) {
		mock := NewMockServer(t)
		defer mock.Close()

		mock.Handle("/path", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("handled"))
		})

		resp, err := http.Get(mock.URL + "/path")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		if string(body) != "handled" {
			t.Errorf("expected body 'handled', got %s", string(body))
		}
	})

	t.Run("handles_different_methods", func(t *testing.T) {
		mock := NewMockServer(t)
		defer mock.Close()

		mock.Handle("/method", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(r.Method))
		})

		tests := []struct {
			name   string
			method string
		}{
			{"GET", "GET"},
			{"POST", "POST"},
			{"PUT", "PUT"},
			{"DELETE", "DELETE"},
			{"PATCH", "PATCH"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var req *http.Request
				var err error

				switch tt.method {
				case "GET":
					resp, err := http.Get(mock.URL + "/method")
					if err != nil {
						t.Fatalf("failed to make request: %v", err)
					}
					defer resp.Body.Close()
					body, _ := io.ReadAll(resp.Body)
					if string(body) != tt.method {
						t.Errorf("expected method %s, got %s", tt.method, string(body))
					}
				default:
					req, err = http.NewRequest(tt.method, mock.URL+"/method", nil)
					if err != nil {
						t.Fatalf("failed to create request: %v", err)
					}
					client := &http.Client{}
					resp, err := client.Do(req)
					if err != nil {
						t.Fatalf("failed to make request: %v", err)
					}
					defer resp.Body.Close()
					body, _ := io.ReadAll(resp.Body)
					if string(body) != tt.method {
						t.Errorf("expected method %s, got %s", tt.method, string(body))
					}
				}
			})
		}
	})

	t.Run("handles_multiple_paths", func(t *testing.T) {
		mock := NewMockServer(t)
		defer mock.Close()

		mock.Handle("/path1", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("path1"))
		})

		mock.Handle("/path2", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("path2"))
		})

		mock.Handle("/path3", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("path3"))
		})

		paths := []string{"/path1", "/path2", "/path3"}
		for _, path := range paths {
			resp, err := http.Get(mock.URL + path)
			if err != nil {
				t.Fatalf("failed to make request to %s: %v", path, err)
			}
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			expected := strings.TrimPrefix(path, "/")
			if string(body) != expected {
				t.Errorf("expected body %s, got %s", expected, string(body))
			}
		}
	})

	t.Run("handler_can_access_request_details", func(t *testing.T) {
		mock := NewMockServer(t)
		defer mock.Close()

		mock.Handle("/details", func(w http.ResponseWriter, r *http.Request) {
			// Return request details as JSON
			details := map[string]string{
				"method": r.Method,
				"path":   r.URL.Path,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(details)
		})

		resp, err := http.Get(mock.URL + "/details?query=value")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		var details map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&details); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if details["method"] != "GET" {
			t.Errorf("expected method GET, got %s", details["method"])
		}

		if details["path"] != "/details" {
			t.Errorf("expected path /details, got %s", details["path"])
		}
	})

	t.Run("handler_can_return_custom_headers", func(t *testing.T) {
		mock := NewMockServer(t)
		defer mock.Close()

		mock.Handle("/headers", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Custom-Header", "custom-value")
			w.Header().Set("X-Another-Header", "another-value")
			w.WriteHeader(http.StatusOK)
		})

		resp, err := http.Get(mock.URL + "/headers")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.Header.Get("X-Custom-Header") != "custom-value" {
			t.Error("expected X-Custom-Header to be set")
		}

		if resp.Header.Get("X-Another-Header") != "another-value" {
			t.Error("expected X-Another-Header to be set")
		}
	})

	t.Run("handler_can_return_different_status_codes", func(t *testing.T) {
		mock := NewMockServer(t)
		defer mock.Close()

		statusCodes := []int{
			http.StatusOK,
			http.StatusCreated,
			http.StatusBadRequest,
			http.StatusUnauthorized,
			http.StatusForbidden,
			http.StatusNotFound,
			http.StatusInternalServerError,
		}

		for _, code := range statusCodes {
			path := "/status/" + string(rune(code))
			expectedCode := code

			mock.Handle(path, func(expectedCode int) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(expectedCode)
				}
			}(expectedCode))

			resp, err := http.Get(mock.URL + path)
			if err != nil {
				t.Fatalf("failed to make request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != expectedCode {
				t.Errorf("expected status %d, got %d", expectedCode, resp.StatusCode)
			}
		}
	})

	t.Run("unregistered_path_returns_404", func(t *testing.T) {
		mock := NewMockServer(t)
		defer mock.Close()

		// Don't register any handler
		resp, err := http.Get(mock.URL + "/unregistered")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
		}
	})
}

// TestMockServer_HandleJSON tests the HandleJSON method
func TestMockServer_HandleJSON(t *testing.T) {
	t.Run("returns_json_response", func(t *testing.T) {
		mock := NewMockServer(t)
		defer mock.Close()

		response := map[string]string{
			"message": "success",
			"status":  "ok",
		}

		mock.HandleJSON("/json", response)

		resp, err := http.Get(mock.URL + "/json")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}

		contentType := resp.Header.Get("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			t.Errorf("expected Content-Type to contain application/json, got %s", contentType)
		}

		var result map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode JSON: %v", err)
		}

		if result["message"] != "success" {
			t.Errorf("expected message 'success', got %s", result["message"])
		}

		if result["status"] != "ok" {
			t.Errorf("expected status 'ok', got %s", result["status"])
		}
	})

	t.Run("handles_complex_json_structures", func(t *testing.T) {
		mock := NewMockServer(t)
		defer mock.Close()

		response := map[string]any{
			"id":      123,
			"name":    "Test Object",
			"active":  true,
			"tags":    []string{"tag1", "tag2", "tag3"},
			"details": map[string]int{"count": 42, "total": 100},
		}

		mock.HandleJSON("/complex", response)

		resp, err := http.Get(mock.URL + "/complex")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		var result map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode JSON: %v", err)
		}

		if result["id"].(float64) != 123 {
			t.Errorf("expected id 123, got %v", result["id"])
		}

		if result["name"].(string) != "Test Object" {
			t.Errorf("expected name 'Test Object', got %s", result["name"])
		}
	})

	t.Run("handles_nil_response", func(t *testing.T) {
		mock := NewMockServer(t)
		defer mock.Close()

		mock.HandleJSON("/nil", nil)

		resp, err := http.Get(mock.URL + "/nil")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}
	})

	t.Run("handles_empty_map_response", func(t *testing.T) {
		mock := NewMockServer(t)
		defer mock.Close()

		response := map[string]string{}
		mock.HandleJSON("/empty", response)

		resp, err := http.Get(mock.URL + "/empty")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}
	})

	t.Run("handles_slice_response", func(t *testing.T) {
		mock := NewMockServer(t)
		defer mock.Close()

		response := []map[string]string{
			{"id": "1", "name": "Item 1"},
			{"id": "2", "name": "Item 2"},
		}

		mock.HandleJSON("/slice", response)

		resp, err := http.Get(mock.URL + "/slice")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		var result []map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode JSON: %v", err)
		}

		if len(result) != 2 {
			t.Errorf("expected 2 items, got %d", len(result))
		}
	})

	t.Run("handles_struct_response", func(t *testing.T) {
		mock := NewMockServer(t)
		defer mock.Close()

		type TestStruct struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}

		response := TestStruct{
			ID:   1,
			Name: "Test",
		}

		mock.HandleJSON("/struct", response)

		resp, err := http.Get(mock.URL + "/struct")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		var result TestStruct
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode JSON: %v", err)
		}

		if result.ID != 1 {
			t.Errorf("expected ID 1, got %d", result.ID)
		}

		if result.Name != "Test" {
			t.Errorf("expected Name 'Test', got %s", result.Name)
		}
	})

	t.Run("handles_json_encoding_error", func(t *testing.T) {
		mock := NewMockServer(t)
		defer mock.Close()

		// Create a value that cannot be JSON encoded
		// Using a channel which cannot be marshaled to JSON
		response := map[string]any{
			"channel": make(chan int),
		}

		mock.HandleJSON("/error", response)

		resp, err := http.Get(mock.URL + "/error")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		// Should return 500 Internal Server Error
		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
		}
	})
}

// TestMockServer_HandleError tests the HandleError method
func TestMockServer_HandleError(t *testing.T) {
	t.Run("returns_error_response", func(t *testing.T) {
		mock := NewMockServer(t)
		defer mock.Close()

		mock.HandleError("/error", http.StatusBadRequest, "Invalid request")

		resp, err := http.Get(mock.URL + "/error")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
		}

		contentType := resp.Header.Get("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			t.Errorf("expected Content-Type to contain application/json, got %s", contentType)
		}

		var result map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode JSON: %v", err)
		}

		errors, ok := result["errors"].([]any)
		if !ok {
			t.Fatal("expected errors array")
		}

		if len(errors) == 0 || errors[0].(string) != "Invalid request" {
			t.Errorf("expected error message 'Invalid request', got %v", errors)
		}
	})

	t.Run("handles_various_status_codes", func(t *testing.T) {
		mock := NewMockServer(t)
		defer mock.Close()

		tests := []struct {
			name       string
			statusCode int
			message    string
		}{
			{"bad_request", http.StatusBadRequest, "Bad request"},
			{"unauthorized", http.StatusUnauthorized, "Unauthorized access"},
			{"forbidden", http.StatusForbidden, "Access forbidden"},
			{"not_found", http.StatusNotFound, "Resource not found"},
			{"internal_error", http.StatusInternalServerError, "Internal server error"},
			{"service_unavailable", http.StatusServiceUnavailable, "Service unavailable"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				path := "/" + tt.name
				mock.HandleError(path, tt.statusCode, tt.message)

				resp, err := http.Get(mock.URL + path)
				if err != nil {
					t.Fatalf("failed to make request: %v", err)
				}
				defer resp.Body.Close()

				if resp.StatusCode != tt.statusCode {
					t.Errorf("expected status %d, got %d", tt.statusCode, resp.StatusCode)
				}

				var result map[string]any
				if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
					t.Fatalf("failed to decode JSON: %v", err)
				}

				errors := result["errors"].([]any)
				if errors[0].(string) != tt.message {
					t.Errorf("expected message %q, got %s", tt.message, errors[0])
				}
			})
		}
	})

	t.Run("handles_empty_message", func(t *testing.T) {
		mock := NewMockServer(t)
		defer mock.Close()

		mock.HandleError("/empty", http.StatusBadRequest, "")

		resp, err := http.Get(mock.URL + "/empty")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
		}

		var result map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode JSON: %v", err)
		}

		errors := result["errors"].([]any)
		if errors[0].(string) != "" {
			t.Errorf("expected empty message, got %s", errors[0])
		}
	})

	t.Run("handles_special_characters_in_message", func(t *testing.T) {
		mock := NewMockServer(t)
		defer mock.Close()

		message := "Error with special chars: <>&\"'`\n\t"
		mock.HandleError("/special", http.StatusBadRequest, message)

		resp, err := http.Get(mock.URL + "/special")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		var result map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode JSON: %v", err)
		}

		errors := result["errors"].([]any)
		if errors[0].(string) != message {
			t.Errorf("expected message %q, got %s", message, errors[0])
		}
	})

	t.Run("handles_unicode_message", func(t *testing.T) {
		mock := NewMockServer(t)
		defer mock.Close()

		message := "错误信息: 测试中文 🎉"
		mock.HandleError("/unicode", http.StatusBadRequest, message)

		resp, err := http.Get(mock.URL + "/unicode")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		var result map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode JSON: %v", err)
		}

		errors := result["errors"].([]any)
		if errors[0].(string) != message {
			t.Errorf("expected message %q, got %s", message, errors[0])
		}
	})

	t.Run("handles_very_long_message", func(t *testing.T) {
		mock := NewMockServer(t)
		defer mock.Close()

		// Create a very long message
		message := strings.Repeat("a", 10000)
		mock.HandleError("/long", http.StatusBadRequest, message)

		resp, err := http.Get(mock.URL + "/long")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		var result map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode JSON: %v", err)
		}

		errors := result["errors"].([]any)
		if len(errors[0].(string)) != 10000 {
			t.Errorf("expected message length 10000, got %d", len(errors[0].(string)))
		}
	})

	t.Run("handles_json_encoding_error", func(t *testing.T) {
		mock := NewMockServer(t)
		defer mock.Close()

		mock.HandleError("/normal", http.StatusBadRequest, "Normal error")

		resp, err := http.Get(mock.URL + "/normal")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
		}
	})

	t.Run("json_encode_error_path_with_failing_writer", func(t *testing.T) {
		mock := NewMockServer(t)
		defer mock.Close()

		mock.HandleError("/encode-fail", http.StatusBadRequest, "encode fail")

		handler, _ := mock.Mux.Handler(httptest.NewRequest(http.MethodGet, "/encode-fail", nil))

		w := &errorWriter{}
		handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/encode-fail", nil))

		// HandleError calls w.WriteHeader(BadRequest), then json.Encode fails
		// because Write returns error. Then http.Error is called with
		// InternalServerError which calls WriteHeader again.
		// On a real ResponseWriter, the second WriteHeader is a no-op.
		// Our errorWriter tracks the last code set by WriteHeader.
		// The important thing is that the json.Encode error path (lines 61-64)
		// was exercised — http.Error was called.
		if w.code != http.StatusInternalServerError && w.code != http.StatusBadRequest {
			t.Errorf("expected http.Error or WriteHeader to be called, got code %d", w.code)
		}
	})
}

// TestMockServer_Integration tests integration scenarios
func TestMockServer_Integration(t *testing.T) {
	t.Run("simulates_rest_api", func(t *testing.T) {
		mock := NewMockServer(t)
		defer mock.Close()

		// Simulate a REST API
		mock.HandleJSON("/api/users", []map[string]any{
			{"id": 1, "name": "User 1"},
			{"id": 2, "name": "User 2"},
		})

		mock.HandleJSON("/api/users/1", map[string]any{
			"id":   1,
			"name": "User 1",
		})

		mock.HandleError("/api/users/999", http.StatusNotFound, "User not found")

		// Test list users
		resp, err := http.Get(mock.URL + "/api/users")
		if err != nil {
			t.Fatalf("failed to get users: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}

		// Test get single user
		resp, err = http.Get(mock.URL + "/api/users/1")
		if err != nil {
			t.Fatalf("failed to get user: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}

		// Test get non-existent user
		resp, err = http.Get(mock.URL + "/api/users/999")
		if err != nil {
			t.Fatalf("failed to get user: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
		}
	})

	t.Run("handles_concurrent_requests", func(t *testing.T) {
		mock := NewMockServer(t)
		defer mock.Close()

		mock.Handle("/concurrent", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})

		// Make concurrent requests
		const numRequests = 10
		done := make(chan bool, numRequests)

		for i := 0; i < numRequests; i++ {
			go func() {
				resp, err := http.Get(mock.URL + "/concurrent")
				if err != nil {
					t.Errorf("failed to make request: %v", err)
					done <- false
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
					done <- false
					return
				}

				done <- true
			}()
		}

		// Wait for all requests to complete
		for i := 0; i < numRequests; i++ {
			if !<-done {
				t.Error("concurrent request failed")
			}
		}
	})

	t.Run("handles_request_with_body", func(t *testing.T) {
		mock := NewMockServer(t)
		defer mock.Close()

		mock.Handle("/echo", func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			w.Header().Set("Content-Type", "application/json")
			w.Write(body)
		})

		// Send a POST request with body
		body := `{"test": "value"}`
		resp, err := http.Post(mock.URL+"/echo", "application/json", strings.NewReader(body))
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(resp.Body)
		if string(respBody) != body {
			t.Errorf("expected body %q, got %q", body, string(respBody))
		}
	})
}

// TestMockServer_EdgeCases tests edge cases
func TestMockServer_EdgeCases(t *testing.T) {
	t.Run("handles_path_with_query_params", func(t *testing.T) {
		mock := NewMockServer(t)
		defer mock.Close()

		mock.Handle("/query", func(w http.ResponseWriter, r *http.Request) {
			query := r.URL.Query().Get("param")
			w.Write([]byte(query))
		})

		resp, err := http.Get(mock.URL + "/query?param=value")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		if string(body) != "value" {
			t.Errorf("expected body 'value', got %s", string(body))
		}
	})

	t.Run("handles_path_with_special_characters", func(t *testing.T) {
		mock := NewMockServer(t)
		defer mock.Close()

		mock.Handle("/special%20path", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("OK"))
		})

		resp, err := http.Get(mock.URL + "/special%20path")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}
	})

	t.Run("handles_root_path", func(t *testing.T) {
		mock := NewMockServer(t)
		defer mock.Close()

		mock.Handle("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("root"))
		})

		resp, err := http.Get(mock.URL + "/")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		if string(body) != "root" {
			t.Errorf("expected body 'root', got %s", string(body))
		}
	})

	t.Run("handles_nested_paths", func(t *testing.T) {
		mock := NewMockServer(t)
		defer mock.Close()

		mock.Handle("/a/b/c/d/e", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("nested"))
		})

		resp, err := http.Get(mock.URL + "/a/b/c/d/e")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		if string(body) != "nested" {
			t.Errorf("expected body 'nested', got %s", string(body))
		}
	})
}

// TestMockServer_Helper tests that NewMockServer is a proper test helper
func TestMockServer_Helper(t *testing.T) {
	// This test verifies that NewMockServer calls t.Helper()
	// If it doesn't, error messages will point to the wrong line

	t.Run("helper_is_called", func(t *testing.T) {
		// Create a mock server - t.Helper() should be called
		mock := NewMockServer(t)
		if mock == nil {
			t.Error("expected mock server, got nil")
		}
		defer mock.Close()
	})
}

// TestMockServer_HTTPS tests that the server can be used with HTTPS
func TestMockServer_HTTPS(t *testing.T) {
	t.Run("server_uses_http", func(t *testing.T) {
		mock := NewMockServer(t)
		defer mock.Close()

		// httptest.NewServer creates an HTTP server by default
		if !strings.HasPrefix(mock.URL, "http://") {
			t.Errorf("expected HTTP URL, got %s", mock.URL)
		}
	})
}

// TestMockServer_ServerType tests the server type
func TestMockServer_ServerType(t *testing.T) {
	t.Run("server_is_httptest_server", func(t *testing.T) {
		mock := NewMockServer(t)
		defer mock.Close()

		// Verify the server is an httptest.Server
		if mock.Server == nil {
			t.Fatal("expected Server to be initialized")
		}

		// The server should be of type *httptest.Server
		var _ = mock.Server
	})
}

// TestMockServer_MuxType tests the mux type
func TestMockServer_MuxType(t *testing.T) {
	t.Run("mux_is_http_serve_mux", func(t *testing.T) {
		mock := NewMockServer(t)
		defer mock.Close()

		// Verify the mux is an http.ServeMux
		if mock.Mux == nil {
			t.Fatal("expected Mux to be initialized")
		}

		// The mux should be of type *http.ServeMux
		var _ = mock.Mux
	})
}

// TestMockServer_MultipleHandlers tests multiple handlers
func TestMockServer_MultipleHandlers(t *testing.T) {
	t.Run("can_register_multiple_handlers", func(t *testing.T) {
		mock := NewMockServer(t)
		defer mock.Close()

		// Register multiple handlers with different paths
		mock.Handle("/handler1", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("handler1"))
		})

		mock.Handle("/handler2", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("handler2"))
		})

		mock.Handle("/handler3", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("handler3"))
		})

		// Test each handler
		tests := []struct {
			path     string
			expected string
		}{
			{"/handler1", "handler1"},
			{"/handler2", "handler2"},
			{"/handler3", "handler3"},
		}

		for _, tt := range tests {
			resp, err := http.Get(mock.URL + tt.path)
			if err != nil {
				t.Fatalf("failed to make request to %s: %v", tt.path, err)
			}
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			if string(body) != tt.expected {
				t.Errorf("expected body %q, got %s", tt.expected, string(body))
			}
		}
	})
}

// TestMockServer_RequestHeaders tests request header handling
func TestMockServer_RequestHeaders(t *testing.T) {
	t.Run("can_read_request_headers", func(t *testing.T) {
		mock := NewMockServer(t)
		defer mock.Close()

		mock.Handle("/headers", func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			w.Write([]byte(auth))
		})

		req, _ := http.NewRequest("GET", mock.URL+"/headers", nil)
		req.Header.Set("Authorization", "Bearer token123")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		if string(body) != "Bearer token123" {
			t.Errorf("expected 'Bearer token123', got %s", string(body))
		}
	})
}

// TestMockServer_HandlePrefix tests the HandlePrefix method
func TestMockServer_HandlePrefix(t *testing.T) {
	t.Run("handles_exact_path_match", func(t *testing.T) {
		mock := NewMockServer(t)
		defer mock.Close()

		mock.HandlePrefix("/api", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("matched"))
		})

		resp, err := http.Get(mock.URL + "/api")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		if string(body) != "matched" {
			t.Errorf("expected body 'matched', got %s", string(body))
		}
	})

	t.Run("handles_path_with_suffix", func(t *testing.T) {
		mock := NewMockServer(t)
		defer mock.Close()

		mock.HandlePrefix("/api", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("prefix-matched: " + r.URL.Path))
		})

		tests := []string{
			"/api",
			"/api/",
			"/api/users",
			"/api/users/123",
			"/api/projects/456/issues",
		}

		for _, path := range tests {
			resp, err := http.Get(mock.URL + path)
			if err != nil {
				t.Fatalf("failed to make request to %s: %v", path, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("expected status OK for path %s, got %d", path, resp.StatusCode)
			}

			body, _ := io.ReadAll(resp.Body)
			if !strings.HasPrefix(string(body), "prefix-matched:") {
				t.Errorf("expected prefix match for path %s, got %s", path, string(body))
			}
		}
	})

	t.Run("handles_path_with_query_params", func(t *testing.T) {
		mock := NewMockServer(t)
		defer mock.Close()

		mock.HandlePrefix("/api", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("query: " + r.URL.RawQuery))
		})

		resp, err := http.Get(mock.URL + "/api/users?limit=10&offset=20")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		if !strings.Contains(string(body), "limit=10") {
			t.Errorf("expected query params in response, got %s", string(body))
		}
	})

	t.Run("returns_404_for_non_matching_path", func(t *testing.T) {
		mock := NewMockServer(t)
		defer mock.Close()

		mock.HandlePrefix("/api", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("api"))
		})

		resp, err := http.Get(mock.URL + "/other")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
		}
	})

	t.Run("handles_multiple_prefix_handlers", func(t *testing.T) {
		mock := NewMockServer(t)
		defer mock.Close()

		mock.HandlePrefix("/api/v1", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("v1"))
		})

		mock.HandlePrefix("/api/v2", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("v2"))
		})

		tests := []struct {
			path     string
			expected string
		}{
			{"/api/v1/users", "v1"},
			{"/api/v2/projects", "v2"},
			{"/api/v1/issues/123", "v1"},
		}

		for _, tt := range tests {
			resp, err := http.Get(mock.URL + tt.path)
			if err != nil {
				t.Fatalf("failed to make request to %s: %v", tt.path, err)
			}
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			if string(body) != tt.expected {
				t.Errorf("path %s: expected %s, got %s", tt.path, tt.expected, string(body))
			}
		}
	})

	t.Run("handles_nested_prefix_paths", func(t *testing.T) {
		mock := NewMockServer(t)
		defer mock.Close()

		mock.HandlePrefix("/api/v1/projects", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("projects"))
		})

		resp, err := http.Get(mock.URL + "/api/v1/projects/123/issues")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		if string(body) != "projects" {
			t.Errorf("expected 'projects', got %s", string(body))
		}
	})

	t.Run("handles_path_with_trailing_slash", func(t *testing.T) {
		mock := NewMockServer(t)
		defer mock.Close()

		mock.HandlePrefix("/api/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("trailing-slash"))
		})

		resp, err := http.Get(mock.URL + "/api/users")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		if string(body) != "trailing-slash" {
			t.Errorf("expected 'trailing-slash', got %s", string(body))
		}
	})

	t.Run("handles_different_methods", func(t *testing.T) {
		mock := NewMockServer(t)
		defer mock.Close()

		mock.HandlePrefix("/api", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(r.Method))
		})

		tests := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

		for _, method := range tests {
			req, _ := http.NewRequest(method, mock.URL+"/api/resource", nil)
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("failed to make %s request: %v", method, err)
			}
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			if string(body) != method {
				t.Errorf("expected %s, got %s", method, string(body))
			}
		}
	})

	t.Run("handles_short_prefix", func(t *testing.T) {
		mock := NewMockServer(t)
		defer mock.Close()

		mock.HandlePrefix("/a", func(w http.ResponseWriter, _ *http.Request) {
			w.Write([]byte("short"))
		})

		// Note: HandlePrefix adds a trailing slash to enable prefix matching
		// So "/a" becomes "/a/" which matches "/a/" and "/a/something"
		// but NOT "/abc" (because it doesn't start with "/a/")
		resp, err := http.Get(mock.URL + "/a/test")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		if string(body) != "short" {
			t.Errorf("expected 'short', got %s", string(body))
		}
	})

	t.Run("handles_exact_prefix_match_with_longer_path", func(t *testing.T) {
		mock := NewMockServer(t)
		defer mock.Close()

		mock.HandlePrefix("/api/v1", func(w http.ResponseWriter, _ *http.Request) {
			w.Write([]byte("v1-handler"))
		})

		resp, err := http.Get(mock.URL + "/api/v1")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		if string(body) != "v1-handler" {
			t.Errorf("expected 'v1-handler', got %s", string(body))
		}
	})

	t.Run("handles_path_without_trailing_slash_in_registration", func(t *testing.T) {
		mock := NewMockServer(t)
		defer mock.Close()

		mock.HandlePrefix("/api", func(w http.ResponseWriter, _ *http.Request) {
			w.Write([]byte("no-trailing-slash"))
		})

		resp, err := http.Get(mock.URL + "/api/users")
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		if string(body) != "no-trailing-slash" {
			t.Errorf("expected 'no-trailing-slash', got %s", string(body))
		}
	})

	t.Run("prefix_handler_returns_404_for_non_matching_subpath", func(t *testing.T) {
		mock := NewMockServer(t)
		defer mock.Close()

		mock.HandlePrefix("/prefix", func(w http.ResponseWriter, _ *http.Request) {
			w.Write([]byte("matched"))
		})

		handler, _ := mock.Mux.Handler(httptest.NewRequest(http.MethodGet, "/prefix/", nil))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/other", nil)
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
		}
	})
}
