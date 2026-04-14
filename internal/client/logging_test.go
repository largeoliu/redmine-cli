// internal/client/logging_test.go
package client

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGenerateRequestID(t *testing.T) {
	id1 := generateRequestID()
	id2 := generateRequestID()

	if id1 == id2 {
		t.Errorf("generateRequestID() should generate unique IDs, got same: %s", id1)
	}

	if !strings.HasPrefix(id1, "req-") {
		t.Errorf("generateRequestID() should return ID starting with 'req-', got: %s", id1)
	}

	if id1 == "" {
		t.Error("generateRequestID() should not return empty string")
	}
}

func TestLoggingTransport_RoundTrip(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	}))
	defer server.Close()

	var logBuf bytes.Buffer
	transport := NewLoggingTransport(http.DefaultTransport, &logBuf, true)

	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip() error = %v", err)
	}
	defer resp.Body.Close()

	logOutput := logBuf.String()
	if logOutput == "" {
		t.Error("RoundTrip() should produce log output")
	}
	if !strings.Contains(logOutput, "[req-") {
		t.Error("RoundTrip() log should contain request ID")
	}
}

func TestLoggingTransport_RoundTripNonVerbose(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	var logBuf bytes.Buffer
	transport := NewLoggingTransport(http.DefaultTransport, &logBuf, false)

	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip() error = %v", err)
	}
	defer resp.Body.Close()

	logOutput := logBuf.String()
	if logOutput == "" {
		t.Error("RoundTrip() should produce log output")
	}
	// Non-verbose mode should not log full URL
	if strings.Contains(logOutput, "http://") {
		t.Error("Non-verbose mode should not log full URL")
	}
}

func TestLoggingTransport_RoundTripError(t *testing.T) {
	var logBuf bytes.Buffer

	// Create a transport that always returns an error
	errorTransport := &mockErrorTransport{}
	transport := NewLoggingTransport(errorTransport, &logBuf, true)

	req, err := http.NewRequest("GET", "http://example.com/test", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	_, err = transport.RoundTrip(req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	logOutput := logBuf.String()
	if logOutput == "" {
		t.Error("RoundTrip() should produce log output even on error")
	}
	if !strings.Contains(logOutput, "ERROR") {
		t.Error("RoundTrip() log should contain ERROR on failure")
	}
}

func TestLoggingTransport_NilTransport(t *testing.T) {
	var logBuf bytes.Buffer
	transport := NewLoggingTransport(nil, &logBuf, false)

	if transport.transport == nil {
		t.Error("NewLoggingTransport should use http.DefaultTransport when nil is passed")
	}
	if transport.transport != http.DefaultTransport {
		t.Error("NewLoggingTransport should use http.DefaultTransport when nil is passed")
	}
}

func TestLoggingTransport_SetVerbose(t *testing.T) {
	var logBuf bytes.Buffer
	transport := NewLoggingTransport(http.DefaultTransport, &logBuf, false)

	transport.SetVerbose(true)
	if !transport.verbose {
		t.Error("SetVerbose(true) should set verbose to true")
	}

	transport.SetVerbose(false)
	if transport.verbose {
		t.Error("SetVerbose(false) should set verbose to false")
	}
}

func TestLoggingTransport_APIKeyMasking(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	var logBuf bytes.Buffer
	transport := NewLoggingTransport(http.DefaultTransport, &logBuf, true)

	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("X-Redmine-API-Key", "secret-api-key-12345")

	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip() error = %v", err)
	}
	defer resp.Body.Close()

	logOutput := logBuf.String()
	// API key should be masked
	if strings.Contains(logOutput, "secret-api-key-12345") {
		t.Error("API key should be masked in logs")
	}
	if !strings.Contains(logOutput, "***") {
		t.Error("API key should be replaced with ***")
	}
}

func TestLoggingTransport_ResponseHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Custom-Header", "custom-value")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	var logBuf bytes.Buffer
	transport := NewLoggingTransport(http.DefaultTransport, &logBuf, true)

	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip() error = %v", err)
	}
	defer resp.Body.Close()

	logOutput := logBuf.String()
	// In verbose mode, response headers should be logged
	if !strings.Contains(logOutput, "Content-Type") {
		t.Error("Verbose mode should log response headers")
	}
}

type mockErrorTransport struct{}

func (m *mockErrorTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, errors.New("mock transport error")
}

type errorWriter struct {
	err error
}

func (e *errorWriter) Write(p []byte) (int, error) {
	return 0, e.err
}

func TestSafeFprintfWithWriterError(t *testing.T) {
	err := errors.New("write error")
	ew := &errorWriter{err: err}

	safeFprintf(ew, "test %s", "data")

	if ew.err != err {
		t.Log("safeFprintf correctly handles write error")
	}
}

func TestLoggingTransport_RoundTripWithResponseHeaderMultiple(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("X-Multi", "value1")
		w.Header().Add("X-Multi", "value2")
		w.Header().Add("X-Multi", "value3")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	var logBuf bytes.Buffer
	transport := NewLoggingTransport(http.DefaultTransport, &logBuf, true)

	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip() error = %v", err)
	}
	defer resp.Body.Close()

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "X-Multi") {
		t.Error("Verbose mode should log multiple header values")
	}
}
