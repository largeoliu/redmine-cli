// internal/client/benchmark_test.go
package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// BenchmarkClientGet 测试 HTTP GET 请求性能
func BenchmarkClientGet(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"name":"test","description":"benchmark test data"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result map[string]any
		_ = client.Get(context.Background(), "/test.json", &result)
	}
}

// BenchmarkClientGetParallel 测试并行 HTTP GET 请求性能
func BenchmarkClientGetParallel(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"name":"test","description":"benchmark test data"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var result map[string]any
			_ = client.Get(context.Background(), "/test.json", &result)
		}
	})
}

// BenchmarkClientGetLargeResponse 测试大响应体�?GET 请求性能
func BenchmarkClientGetLargeResponse(b *testing.B) {
	// 生成较大的响应数�?
	largeData := make([]map[string]any, 100)
	for i := 0; i < 100; i++ {
		largeData[i] = map[string]any{
			"id":          i,
			"name":        "test-item-" + string(rune('A'+i%26)),
			"description": "This is a test description for benchmark purposes",
			"status":      "open",
			"priority":    i % 5,
		}
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// 返回大型 JSON 响应
		w.Write([]byte(`{"items":[` + largeDataToJSON(largeData) + `],"total_count":100}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result map[string]any
		_ = client.Get(context.Background(), "/test.json", &result)
	}
}

// BenchmarkClientWithRetry 测试重试机制性能
func BenchmarkClientWithRetry(b *testing.B) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		// 前两次失败，第三次成�?
		if attempts%3 != 0 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	client := NewClient(
		server.URL,
		"test-api-key",
		WithRetry(3, 1*time.Millisecond, 10*time.Millisecond),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		attempts = 0
		var result map[string]any
		_ = client.Get(context.Background(), "/test.json", &result)
	}
}

// BenchmarkClientWithRetryNoFail 测试无失败时的重试机制开销
func BenchmarkClientWithRetryNoFail(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	client := NewClient(
		server.URL,
		"test-api-key",
		WithRetry(3, 1*time.Millisecond, 10*time.Millisecond),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result map[string]any
		_ = client.Get(context.Background(), "/test.json", &result)
	}
}

// BenchmarkClientPost 测试 HTTP POST 请求性能
func BenchmarkClientPost(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":1,"status":"created"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	body := map[string]any{
		"name":        "test issue",
		"description": "test description",
		"priority":    3,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result map[string]any
		_ = client.Post(context.Background(), "/issues.json", body, &result)
	}
}

// BenchmarkClientBuildPath 测试 URL 路径构建性能
func BenchmarkClientBuildPath(b *testing.B) {
	client := NewClient("https://example.com", "test-api-key")
	params := map[string]string{
		"limit":  "100",
		"offset": "0",
		"status": "open",
		"sort":   "updated_on:desc",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.BuildPath("/issues.json", params)
	}
}

// BenchmarkClientBuildPathNoParams 测试无参数时的 URL 路径构建性能
func BenchmarkClientBuildPathNoParams(b *testing.B) {
	client := NewClient("https://example.com", "test-api-key")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.BuildPath("/issues.json", nil)
	}
}

// BenchmarkNewClient 测试客户端创建性能
func BenchmarkNewClient(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewClient("https://example.com", "test-api-key")
	}
}

// BenchmarkNewClientWithOptions 测试带选项的客户端创建性能
func BenchmarkNewClientWithOptions(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewClient(
			"https://example.com",
			"test-api-key",
			WithTimeout(60*time.Second),
			WithRetry(5, 100*time.Millisecond, 10*time.Second),
		)
	}
}

// 辅助函数：将数据转换�?JSON 字符�?
func largeDataToJSON(data []map[string]any) string {
	var result string
	for i, item := range data {
		if i > 0 {
			result += ","
		}
		result += `{"id":` + itoa(int(item["id"].(int))) +
			`,"name":"` + item["name"].(string) +
			`","description":"` + item["description"].(string) +
			`","status":"` + item["status"].(string) +
			`","priority":` + itoa(item["priority"].(int)) + `}`
	}
	return result
}

// 简单的 itoa 实现
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var result []byte
	negative := n < 0
	if negative {
		n = -n
	}
	for n > 0 {
		result = append([]byte{byte('0' + n%10)}, result...)
		n /= 10
	}
	if negative {
		result = append([]byte{'-'}, result...)
	}
	return string(result)
}
