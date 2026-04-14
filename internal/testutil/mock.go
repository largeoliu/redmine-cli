// internal/testutil/mock.go
package testutil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// MockServer 提供一个用于测试的 HTTP 模拟服务�?
type MockServer struct {
	Server *httptest.Server
	Mux    *http.ServeMux
	URL    string
}

// NewMockServer 创建一个新�?MockServer 实例
func NewMockServer(t *testing.T) *MockServer {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)

	return &MockServer{
		Server: server,
		Mux:    mux,
		URL:    server.URL,
	}
}

// Close 关闭模拟服务�?
func (m *MockServer) Close() {
	m.Server.Close()
}

// Handle 为指定路径注册自定义处理函数
func (m *MockServer) Handle(path string, handler http.HandlerFunc) {
	m.Mux.HandleFunc(path, handler)
}

// HandleJSON 为指定路径注册返�?JSON 响应的处理函�?
func (m *MockServer) HandleJSON(path string, response any) {
	m.Mux.HandleFunc(path, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

// HandleError 为指定路径注册返回错误响应的处理函数
func (m *MockServer) HandleError(path string, statusCode int, message string) {
	m.Mux.HandleFunc(path, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		errorResponse := map[string]any{
			"errors": []string{message},
		}
		if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

// HandlePrefix 为指定路径前缀注册自定义处理函数
// 这个方法会匹配所有以指定路径开头的请求（包括查询参数）
func (m *MockServer) HandlePrefix(path string, handler http.HandlerFunc) {
	// 使用以 / 结尾的路径来启用前缀匹配
	if path[len(path)-1] != '/' {
		path = path + "/"
	}
	m.Mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		// 检查请求路径是否以指定路径开头（忽略查询参数）
		requestPath := r.URL.Path
		prefixPath := path[:len(path)-1] // 移除末尾的 /
		if requestPath == prefixPath || len(requestPath) > len(prefixPath) && requestPath[:len(prefixPath)] == prefixPath {
			handler(w, r)
		} else {
			http.NotFound(w, r)
		}
	})
}
