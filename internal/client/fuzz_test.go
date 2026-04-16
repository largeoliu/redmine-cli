// internal/client/fuzz_test.go
package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func FuzzBuildPath(f *testing.F) {
	testcases := []struct {
		path   string
		params string
	}{
		{"/issues.json", "limit=10&offset=0"},
		{"/projects.json", ""},
		{"/users.json", "status=active"},
		{"/issues/1.json", "include=children"},
		{"/search.json", "q=test&limit=100"},
	}

	for _, tc := range testcases {
		f.Add(tc.path, tc.params)
	}

	f.Fuzz(func(t *testing.T, path, paramsStr string) {
		if len(path) > 1000 || len(paramsStr) > 1000 {
			return
		}

		client := NewClient("https://example.com", "test-key")

		params := parseParams(paramsStr)
		result := client.BuildPath(path, params)

		if len(result) > 10000 {
			t.Errorf("BuildPath result too long: %d", len(result))
		}
	})
}

func FuzzNewClient(f *testing.F) {
	testcases := []struct {
		baseURL string
		apiKey  string
	}{
		{"https://example.com", "test-key"},
		{"https://redmine.example.org/", "abc123"},
		{"http://localhost:3000", "secret"},
		{"", ""},
		{"https://example.com/", ""},
	}

	for _, tc := range testcases {
		f.Add(tc.baseURL, tc.apiKey)
	}

	f.Fuzz(func(t *testing.T, baseURL, apiKey string) {
		if len(baseURL) > 1000 || len(apiKey) > 1000 {
			return
		}

		client := NewClient(baseURL, apiKey)
		if client == nil {
			t.Error("NewClient returned nil")
		}
	})
}

func FuzzClientGet(f *testing.F) {
	testcases := []struct {
		path       string
		response   string
		statusCode int
	}{
		{"/issues.json", `{"issues": []}`, 200},
		{"/projects.json", `{"projects": []}`, 200},
		{"/users.json", `{"users": []}`, 200},
		{"/issues/1.json", `{"issue": {"id": 1}}`, 200},
		{"/notfound.json", `{"error": "not found"}`, 404},
	}

	for _, tc := range testcases {
		f.Add(tc.path, tc.response, tc.statusCode)
	}

	f.Fuzz(func(t *testing.T, path, response string, statusCode int) {
		if len(path) > 500 || len(response) > 10000 {
			return
		}
		if statusCode < 100 || statusCode > 599 {
			statusCode = 200
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(statusCode)
			w.Write([]byte(response))
		}))
		defer server.Close()

		client := NewClient(server.URL, "test-key",
			WithRetry(0, 0, 0),
			WithTimeout(1*time.Second),
		)

		var result map[string]any
		_ = client.Get(context.Background(), path, &result)
	})
}

func FuzzClientPost(f *testing.F) {
	testcases := []struct {
		path       string
		body       string
		statusCode int
	}{
		{"/issues.json", `{"issue": {"subject": "test"}}`, 201},
		{"/projects.json", `{"project": {"name": "test"}}`, 201},
		{"/time_entries.json", `{"time_entry": {"hours": 1}}`, 201},
		{"/invalid.json", `{}`, 400},
	}

	for _, tc := range testcases {
		f.Add(tc.path, tc.body, tc.statusCode)
	}

	f.Fuzz(func(t *testing.T, path, body string, statusCode int) {
		if len(path) > 500 || len(body) > 10000 {
			return
		}
		if statusCode < 100 || statusCode > 599 {
			statusCode = 201
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(statusCode)
			w.Write([]byte(`{"id": 1}`))
		}))
		defer server.Close()

		client := NewClient(server.URL, "test-key",
			WithRetry(0, 0, 0),
			WithTimeout(1*time.Second),
		)

		var bodyData map[string]any
		if body != "" {
			if err := json.Unmarshal([]byte(body), &bodyData); err != nil {
				return
			}
		}

		var result map[string]any
		_ = client.Post(context.Background(), path, bodyData, &result)
	})
}

func FuzzRetryDelay(f *testing.F) {
	testcases := []int{0, 1, 2, 3, 5, 10}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, attempt int) {
		if attempt < 0 || attempt > 100 {
			return
		}

		client := NewClient("https://example.com", "test-key",
			WithRetry(3, 100*time.Millisecond, 5*time.Second),
		)

		delay := client.retryDelay(attempt)
		if delay < 0 {
			t.Errorf("retryDelay returned negative: %v", delay)
		}
	})
}

func FuzzHandleErrorResponse(f *testing.F) {
	testcases := []struct {
		statusCode int
		body       string
	}{
		{400, `{"errors": ["Bad request"]}`},
		{401, `{"errors": ["Unauthorized"]}`},
		{403, `{"errors": ["Forbidden"]}`},
		{404, `{"errors": ["Not found"]}`},
		{429, `{"errors": ["Rate limit"]}`},
		{500, `{"errors": ["Internal error"]}`},
		{502, `Bad gateway`},
		{503, `Service unavailable`},
	}

	for _, tc := range testcases {
		f.Add(tc.statusCode, tc.body)
	}

	f.Fuzz(func(t *testing.T, statusCode int, body string) {
		if statusCode < 400 || statusCode > 599 {
			return
		}
		if len(body) > 10000 {
			return
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(statusCode)
			w.Write([]byte(body))
		}))
		defer server.Close()

		client := NewClient(server.URL, "test-key",
			WithRetry(0, 0, 0),
			WithTimeout(1*time.Second),
		)

		_ = client.Get(context.Background(), "/test.json", nil)
	})
}

func parseParams(s string) map[string]string {
	if s == "" {
		return nil
	}
	params := make(map[string]string)
	var key, value string
	inValue := false
	for _, r := range s {
		if r == '=' {
			inValue = true
		} else if r == '&' {
			if key != "" {
				params[key] = value
			}
			key, value = "", ""
			inValue = false
		} else if inValue {
			value += string(r)
		} else {
			key += string(r)
		}
	}
	if key != "" {
		params[key] = value
	}
	return params
}
