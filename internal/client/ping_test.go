package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientPing(t *testing.T) {
	t.Run("valid URL returns nil", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		c := NewClient(server.URL, "")
		err := c.Ping(context.Background())
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("unreachable URL returns network error", func(t *testing.T) {
		c := NewClient("http://localhost:99999", "")
		err := c.Ping(context.Background())
		if err == nil {
			t.Error("expected error for unreachable URL")
		}
	})

	t.Run("invalid URL format returns error", func(t *testing.T) {
		c := NewClient("://invalid", "")
		err := c.Ping(context.Background())
		if err == nil {
			t.Error("expected error for invalid URL")
		}
	})

	t.Run("server returns 5xx returns API error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		c := NewClient(server.URL, "")
		err := c.Ping(context.Background())
		if err == nil {
			t.Error("expected error for server error")
		}
	})
}
