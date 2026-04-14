// Package client provides the HTTP client for interacting with Redmine API.
package client

import (
	"context"
	"time"
)

// RetryableFunc is a function that can be retried.
type RetryableFunc func() error

// WithRetry executes the given function with retry logic.
func (c *Client) WithRetry(ctx context.Context, fn RetryableFunc) error {
	var lastErr error
	for attempt := 0; attempt <= c.retry.MaxRetries; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}
		lastErr = err
		if !c.shouldRetry(err) {
			return err
		}
		if attempt < c.retry.MaxRetries {
			delay := c.retryDelay(attempt)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}
	}
	return lastErr
}
