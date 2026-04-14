// internal/client/retry_test.go
package client

import (
	"context"
	stderrors "errors"
	"testing"
	"time"

	"github.com/largeoliu/redmine-cli/internal/errors"
)

func TestWithRetry(t *testing.T) {
	t.Run("success on first attempt", func(t *testing.T) {
		c := NewClient("https://example.com", "test-key", WithRetry(3, 10*time.Millisecond, 100*time.Millisecond))
		attempts := 0

		err := c.WithRetry(context.Background(), func() error {
			attempts++
			return nil
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if attempts != 1 {
			t.Errorf("expected 1 attempt, got %d", attempts)
		}
	})

	t.Run("success after retries", func(t *testing.T) {
		c := NewClient("https://example.com", "test-key", WithRetry(3, 10*time.Millisecond, 100*time.Millisecond))
		attempts := 0

		err := c.WithRetry(context.Background(), func() error {
			attempts++
			if attempts < 3 {
				return errors.NewNetwork("temporary error")
			}
			return nil
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if attempts != 3 {
			t.Errorf("expected 3 attempts, got %d", attempts)
		}
	})

	t.Run("non-retryable error", func(t *testing.T) {
		c := NewClient("https://example.com", "test-key", WithRetry(3, 10*time.Millisecond, 100*time.Millisecond))
		attempts := 0

		err := c.WithRetry(context.Background(), func() error {
			attempts++
			return errors.NewValidation("permanent error")
		})

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if attempts != 1 {
			t.Errorf("expected 1 attempt for non-retryable error, got %d", attempts)
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		c := NewClient("https://example.com", "test-key", WithRetry(3, 10*time.Millisecond, 100*time.Millisecond))
		ctx, cancel := context.WithCancel(context.Background())
		attempts := 0

		// Cancel context after first attempt
		go func() {
			time.Sleep(5 * time.Millisecond)
			cancel()
		}()

		err := c.WithRetry(ctx, func() error {
			attempts++
			return errors.NewNetwork("temporary error")
		})

		if err == nil {
			t.Fatal("expected error due to context cancellation")
		}
		if !stderrors.Is(err, context.Canceled) {
			t.Errorf("expected context.Canceled error, got %v", err)
		}
	})

	t.Run("max retries exceeded", func(t *testing.T) {
		c := NewClient("https://example.com", "test-key", WithRetry(2, 10*time.Millisecond, 100*time.Millisecond))
		attempts := 0

		err := c.WithRetry(context.Background(), func() error {
			attempts++
			return errors.NewNetwork("persistent error")
		})

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		// With MaxRetries=2, we should have 3 attempts (initial + 2 retries)
		if attempts != 3 {
			t.Errorf("expected 3 attempts, got %d", attempts)
		}
	})
}

func TestWithRetryZeroRetries(t *testing.T) {
	c := NewClient("https://example.com", "test-key", WithRetry(0, 10*time.Millisecond, 100*time.Millisecond))
	attempts := 0

	err := c.WithRetry(context.Background(), func() error {
		attempts++
		return errors.NewNetwork("error")
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if attempts != 1 {
		t.Errorf("expected 1 attempt with zero retries, got %d", attempts)
	}
}

func TestWithRetrySuccessOnLastAttempt(t *testing.T) {
	c := NewClient("https://example.com", "test-key", WithRetry(2, 10*time.Millisecond, 100*time.Millisecond))
	attempts := 0

	err := c.WithRetry(context.Background(), func() error {
		attempts++
		if attempts < 3 {
			return errors.NewNetwork("temporary error")
		}
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}
