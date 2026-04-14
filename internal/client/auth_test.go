// internal/client/auth_test.go
package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTestAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users/current.json" {
			t.Errorf("expected path /users/current.json, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"user": {"id": 1, "login": "testuser"}}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-key")
	err := c.TestAuth(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTestAuthError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	c := NewClient(server.URL, "invalid-key")
	err := c.TestAuth(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSetAPIKey(t *testing.T) {
	c := NewClient("https://example.com", "old-key")
	c.SetAPIKey("new-key")
	c.mu.RLock()
	got := c.apiKey
	c.mu.RUnlock()
	if got != "new-key" {
		t.Errorf("expected apiKey 'new-key', got %s", got)
	}
}

func TestSetBaseURL(t *testing.T) {
	c := NewClient("https://old.example.com", "test-key")
	c.SetBaseURL("https://new.example.com")
	c.mu.RLock()
	got := c.baseURL
	c.mu.RUnlock()
	if got != "https://new.example.com" {
		t.Errorf("expected baseURL 'https://new.example.com', got %s", got)
	}
}
