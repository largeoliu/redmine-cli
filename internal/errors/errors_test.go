// internal/errors/errors_test.go
package errors

import (
	"errors"
	"testing"
)

func TestNewValidation(t *testing.T) {
	err := NewValidation("invalid input", WithHint("check your input"))
	if err.Category != CategoryValidation {
		t.Errorf("expected category validation, got %s", err.Category)
	}
	if err.Message != "invalid input" {
		t.Errorf("expected message 'invalid input', got %s", err.Message)
	}
	if err.Hint != "check your input" {
		t.Errorf("expected hint 'check your input', got %s", err.Hint)
	}
}

func TestErrorMethod(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		expected string
	}{
		{
			name:     "simple message",
			err:      &Error{Message: "something went wrong"},
			expected: "something went wrong",
		},
		{
			name:     "empty message",
			err:      &Error{Message: ""},
			expected: "",
		},
		{
			name:     "validation error",
			err:      NewValidation("invalid input"),
			expected: "invalid input",
		},
		{
			name:     "auth error",
			err:      NewAuth("unauthorized"),
			expected: "unauthorized",
		},
		{
			name:     "api error",
			err:      NewAPI("server error"),
			expected: "server error",
		},
		{
			name:     "network error",
			err:      NewNetwork("connection failed"),
			expected: "connection failed",
		},
		{
			name:     "internal error",
			err:      NewInternal("internal failure"),
			expected: "internal failure",
		},
		{
			name:     "timeout error",
			err:      NewTimeout("request timed out"),
			expected: "request timed out",
		},
		{
			name:     "rate limit error",
			err:      NewRateLimit("too many requests"),
			expected: "too many requests",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestWithActions(t *testing.T) {
	actions := []string{"action1", "action2", "action3"}
	err := NewValidation("test error", WithActions(actions...))

	if len(err.Actions) != 3 {
		t.Errorf("expected 3 actions, got %d", len(err.Actions))
	}
	for i, action := range actions {
		if err.Actions[i] != action {
			t.Errorf("expected action[%d] = %s, got %s", i, action, err.Actions[i])
		}
	}
}

func TestWithRetryable(t *testing.T) {
	tests := []struct {
		name      string
		retryable bool
	}{
		{"retryable true", true},
		{"retryable false", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewValidation("test error", WithRetryable(tt.retryable))
			if err.Retryable != tt.retryable {
				t.Errorf("expected Retryable %v, got %v", tt.retryable, err.Retryable)
			}
		})
	}
}

func TestWithRequestID(t *testing.T) {
	requestID := "req-12345"
	err := NewValidation("test error", WithRequestID(requestID))

	if err.RequestID != requestID {
		t.Errorf("expected RequestID %s, got %s", requestID, err.RequestID)
	}
}

func TestMultipleOptions(t *testing.T) {
	err := NewValidation(
		"complex error",
		WithHint("check your input"),
		WithActions("action1", "action2"),
		WithRetryable(true),
		WithRequestID("req-123"),
		WithCause(errors.New("root cause")),
	)

	if err.Message != "complex error" {
		t.Errorf("expected message 'complex error', got %s", err.Message)
	}
	if err.Hint != "check your input" {
		t.Errorf("expected hint 'check your input', got %s", err.Hint)
	}
	if len(err.Actions) != 2 {
		t.Errorf("expected 2 actions, got %d", len(err.Actions))
	}
	if !err.Retryable {
		t.Error("expected Retryable to be true")
	}
	if err.RequestID != "req-123" {
		t.Errorf("expected RequestID 'req-123', got %s", err.RequestID)
	}
	if err.Cause == nil {
		t.Error("expected Cause to be set")
	}
}

func TestNewNetworkIsRetryable(t *testing.T) {
	err := NewNetwork("connection failed")
	if !err.Retryable {
		t.Error("expected Network error to be retryable by default")
	}
}

func TestNewRateLimitIsRetryable(t *testing.T) {
	err := NewRateLimit("too many requests")
	if !err.Retryable {
		t.Error("expected RateLimit error to be retryable by default")
	}
}

func TestExitCode(t *testing.T) {
	tests := []struct {
		err      error
		expected int
	}{
		{nil, 0},
		{NewValidation("test"), 1},
		{NewAuth("test"), 2},
		{NewAPI("test"), 3},
		{NewNetwork("test"), 4},
		{NewInternal("test"), 5},
		{NewTimeout("test"), 6},
		{NewRateLimit("test"), 7},
		{errors.New("unknown"), 1},
	}
	for _, tt := range tests {
		got := ExitCode(tt.err)
		if got != tt.expected {
			t.Errorf("ExitCode(%v) = %d, want %d", tt.err, got, tt.expected)
		}
	}
}

func TestErrorUnwrap(t *testing.T) {
	cause := errors.New("root cause")
	err := NewAPI("api error", WithCause(cause))
	if !errors.Is(err, cause) {
		t.Error("expected error to wrap cause")
	}
}

func TestAs(t *testing.T) {
	t.Run("converts to Error type", func(t *testing.T) {
		err := NewValidation("test error", WithHint("test hint"))
		var appErr *Error
		if !As(err, &appErr) {
			t.Fatal("expected As to return true")
		}
		if appErr.Message != "test error" {
			t.Errorf("expected message 'test error', got %s", appErr.Message)
		}
		if appErr.Hint != "test hint" {
			t.Errorf("expected hint 'test hint', got %s", appErr.Hint)
		}
	})

	t.Run("returns false for non-Error type", func(t *testing.T) {
		err := errors.New("standard error")
		var appErr *Error
		if As(err, &appErr) {
			t.Error("expected As to return false for standard error")
		}
	})
}

func TestUnwrap(t *testing.T) {
	t.Run("returns cause when set", func(t *testing.T) {
		cause := errors.New("root cause")
		err := NewAPI("api error", WithCause(cause))
		unwrapped := err.Unwrap()
		if unwrapped != cause {
			t.Error("expected Unwrap to return cause")
		}
	})

	t.Run("returns nil when no cause", func(t *testing.T) {
		err := NewValidation("test error")
		unwrapped := err.Unwrap()
		if unwrapped != nil {
			t.Error("expected Unwrap to return nil")
		}
	})
}
