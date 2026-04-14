// internal/client/batch_test.go
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestBatchGet(t *testing.T) {
	tests := []struct {
		name        string
		paths       []string
		concurrency int
		setupMock   func(*httptest.Server)
		wantCount   int
		wantErrors  int
	}{
		{
			name:        "successful batch requests",
			paths:       []string{"/issues/1.json", "/issues/2.json", "/issues/3.json"},
			concurrency: 2,
			setupMock: func(server *httptest.Server) {
				// Mock handled in handler below
			},
			wantCount:  3,
			wantErrors: 0,
		},
		{
			name:        "empty paths",
			paths:       []string{},
			concurrency: 5,
			setupMock:   func(server *httptest.Server) {},
			wantCount:   0,
			wantErrors:  0,
		},
		{
			name:        "mixed success and failure",
			paths:       []string{"/issues/1.json", "/issues/2.json", "/issues/404.json"},
			concurrency: 2,
			setupMock: func(server *httptest.Server) {
				// Mock handled in handler below
			},
			wantCount:  3,
			wantErrors: 1,
		},
		{
			name:        "default concurrency",
			paths:       []string{"/issues/1.json", "/issues/2.json"},
			concurrency: 0, // Should use default
			setupMock:   func(server *httptest.Server) {},
			wantCount:   2,
			wantErrors:  0,
		},
		{
			name:        "concurrency higher than paths",
			paths:       []string{"/issues/1.json"},
			concurrency: 10,
			setupMock:   func(server *httptest.Server) {},
			wantCount:   1,
			wantErrors:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var requestCount atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestCount.Add(1)

				if r.Header.Get("X-Redmine-API-Key") != "test-key" {
					t.Error("missing API key header")
				}

				// Simulate 404 for specific path
				if r.URL.Path == "/issues/404.json" {
					w.WriteHeader(http.StatusNotFound)
					return
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)

				// Return mock issue data
				response := map[string]interface{}{
					"issue": map[string]interface{}{
						"id":      1,
						"subject": "Test Issue",
					},
				}
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			tt.setupMock(server)

			c := NewClient(server.URL, "test-key")
			results := BatchGet[map[string]interface{}](c, context.Background(), tt.paths, tt.concurrency)

			if len(results) != tt.wantCount {
				t.Errorf("expected %d results, got %d", tt.wantCount, len(results))
			}

			errorCount := 0
			for i, result := range results {
				if result.Index != i {
					t.Errorf("result[%d].Index = %d, want %d", i, result.Index, i)
				}
				if result.Error != nil {
					errorCount++
				}
			}

			if errorCount != tt.wantErrors {
				t.Errorf("expected %d errors, got %d", tt.wantErrors, errorCount)
			}
		})
	}
}

func TestBatchGetWithContextCancellation(t *testing.T) {
	var requestCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		time.Sleep(100 * time.Millisecond) // Slow response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-key")

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	paths := []string{"/issues/1.json", "/issues/2.json", "/issues/3.json"}
	results := BatchGet[map[string]interface{}](c, ctx, paths, 2)

	// All results should have context cancellation error
	for i, result := range results {
		if result.Error != context.Canceled {
			t.Errorf("result[%d].Error = %v, want %v", i, result.Error, context.Canceled)
		}
	}
}

func TestBatchGetConcurrencyLimit(t *testing.T) {
	var maxConcurrent atomic.Int32
	var currentConcurrent atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Track concurrent requests
		current := currentConcurrent.Add(1)
		for {
			max := maxConcurrent.Load()
			if current <= max {
				break
			}
			if maxConcurrent.CompareAndSwap(max, current) {
				break
			}
		}

		time.Sleep(50 * time.Millisecond) // Hold connection
		currentConcurrent.Add(-1)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-key")

	// Make 10 requests with concurrency limit of 3
	paths := make([]string, 10)
	for i := range paths {
		paths[i] = fmt.Sprintf("/issues/%d.json", i+1)
	}

	results := BatchGet[map[string]interface{}](c, context.Background(), paths, 3)

	// Verify all requests completed
	if len(results) != 10 {
		t.Errorf("expected 10 results, got %d", len(results))
	}

	// Verify concurrency limit was respected
	max := maxConcurrent.Load()
	if max > 3 {
		t.Errorf("max concurrent requests was %d, should not exceed 3", max)
	}
}

func TestBatchGetFunc(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}

	results := BatchGetFunc(items, context.Background(), func(ctx context.Context, index int, item int) (string, error) {
		return fmt.Sprintf("item-%d", item), nil
	}, 2)

	if len(results) != len(items) {
		t.Errorf("expected %d results, got %d", len(items), len(results))
	}

	for i, result := range results {
		if result.Index != i {
			t.Errorf("result[%d].Index = %d, want %d", i, result.Index, i)
		}
		if result.Error != nil {
			t.Errorf("result[%d].Error = %v, want nil", i, result.Error)
		}
		expected := fmt.Sprintf("item-%d", items[i])
		if result.Result != expected {
			t.Errorf("result[%d].Result = %s, want %s", i, result.Result, expected)
		}
	}
}

func TestBatchGetFuncWithError(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}

	results := BatchGetFunc(items, context.Background(), func(ctx context.Context, index int, item int) (string, error) {
		if item == 3 {
			return "", fmt.Errorf("error processing item %d", item)
		}
		return fmt.Sprintf("item-%d", item), nil
	}, 2)

	errorCount := 0
	for i, result := range results {
		if items[i] == 3 {
			if result.Error == nil {
				t.Errorf("result[%d].Error should not be nil", i)
			}
			errorCount++
		} else {
			if result.Error != nil {
				t.Errorf("result[%d].Error = %v, want nil", i, result.Error)
			}
		}
	}

	if errorCount != 1 {
		t.Errorf("expected 1 error, got %d", errorCount)
	}
}

func TestBatchGetFuncWithCancellation(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	results := BatchGetFunc(items, ctx, func(ctx context.Context, index int, item int) (string, error) {
		time.Sleep(10 * time.Millisecond)
		return fmt.Sprintf("item-%d", item), nil
	}, 2)

	// All results should have context cancellation error
	for i, result := range results {
		if result.Error != context.Canceled {
			t.Errorf("result[%d].Error = %v, want %v", i, result.Error, context.Canceled)
		}
	}
}

func TestBatchGetPreservesOrder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Vary response time to test order preservation
		time.Sleep(time.Duration(50) * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"path": "%s"}`, r.URL.Path)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-key")

	paths := []string{"/a.json", "/b.json", "/c.json", "/d.json", "/e.json"}
	results := BatchGet[map[string]interface{}](c, context.Background(), paths, 2)

	// Verify order is preserved
	for i, result := range results {
		if result.Index != i {
			t.Errorf("result[%d].Index = %d, want %d", i, result.Index, i)
		}
		if result.Error != nil {
			t.Errorf("result[%d].Error = %v, want nil", i, result.Error)
		}
		if result.Result != nil {
			resultMap := result.Result
			expectedPath := paths[i]
			if resultMap["path"] != expectedPath {
				t.Errorf("result[%d].path = %v, want %s", i, resultMap["path"], expectedPath)
			}
		}
	}
}

func TestBatchGetFuncHighConcurrency(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}

	results := BatchGetFunc(items, context.Background(), func(ctx context.Context, index int, item int) (int, error) {
		return item * 2, nil
	}, 100)

	if len(results) != len(items) {
		t.Errorf("expected %d results, got %d", len(items), len(results))
	}
	for i, r := range results {
		if r.Index != i {
			t.Errorf("result[%d].Index = %d, want %d", i, r.Index, i)
		}
		if r.Error != nil {
			t.Errorf("result[%d].Error = %v", i, r.Error)
		}
	}
}

func TestBatchGetFuncWithContextAndHighConcurrency(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	items := []int{1, 2, 3}
	results := BatchGetFunc(items, ctx, func(ctx context.Context, index int, item int) (int, error) {
		return item, nil
	}, 10)

	for _, r := range results {
		if r.Error == nil {
			t.Error("expected context cancellation error")
		}
	}
}

func TestBatchGetFuncEmptyItems(t *testing.T) {
	items := []int{}
	results := BatchGetFunc(items, context.Background(), func(ctx context.Context, index int, item int) (int, error) {
		return item, nil
	}, 5)

	if results != nil {
		t.Errorf("expected nil for empty items, got %v", results)
	}
}

func TestBatchGetFuncConcurrencyLessThanOne(t *testing.T) {
	items := []int{1, 2, 3}
	results := BatchGetFunc(items, context.Background(), func(ctx context.Context, index int, item int) (int, error) {
		return item * 2, nil
	}, 0)

	if len(results) != len(items) {
		t.Errorf("expected %d results, got %d", len(items), len(results))
	}
}

func TestBatchGetWithContextCancellationDuringExecution(t *testing.T) {
	var requestCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-key")

	ctx, cancel := context.WithCancel(context.Background())

	paths := []string{"/issues/1.json", "/issues/2.json"}

	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	results := BatchGet[map[string]interface{}](c, ctx, paths, 2)

	errorCount := 0
	for _, result := range results {
		if result.Error != nil {
			errorCount++
		}
	}

	if errorCount == 0 {
		t.Error("expected some errors due to context cancellation during execution")
	}
}

func TestBatchGetFuncWithContextCancellationDuringExecution(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	items := []int{1, 2, 3}

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	results := BatchGetFunc(items, ctx, func(ctx context.Context, index int, item int) (int, error) {
		time.Sleep(100 * time.Millisecond)
		return item, nil
	}, 2)

	errorCount := 0
	for _, r := range results {
		if r.Error != nil {
			errorCount++
		}
	}

	if errorCount == 0 {
		t.Error("expected some errors due to context cancellation during execution")
	}
}
