// Package client provides the HTTP client for interacting with Redmine API.
package client

import (
	"context"
	"sync"
)

// DefaultConcurrency is the default maximum concurrency for batch operations.
const DefaultConcurrency = 5

// BatchResult represents a single result from a batch operation.
type BatchResult[T any] struct {
	Index  int   // Original index in the request list
	Result T     // Operation result
	Error  error // Error if operation failed, nil otherwise
}

// BatchGet concurrently executes multiple GET requests and returns results.
// Uses semaphore pattern to control maximum concurrency.
//
//nolint:revive // context-as-argument: client first for API consistency
func BatchGet[T any](c *Client, ctx context.Context, paths []string, concurrency int) []BatchResult[T] {
	return BatchGetFunc(paths, ctx, func(innerCtx context.Context, _ int, path string) (T, error) {
		var result T
		err := c.Get(innerCtx, path, &result)
		return result, err
	}, concurrency)
}

// BatchGetFunc executes batch operations with a custom function.
// Use when more complex request logic is needed.
//
// Parameters:
//   - items: Input items
//   - ctx: Context for cancellation
//   - fn: Processing function that receives index and item, returns result and error
//   - concurrency: Maximum concurrency, defaults to 5 if <= 0
//
// Returns:
//   - []BatchResult[R]: Results in the same order as input items
//
//nolint:revive // context-as-argument: items first for API consistency
func BatchGetFunc[T any, R any](items []T, ctx context.Context, fn func(ctx context.Context, index int, item T) (R, error), concurrency int) []BatchResult[R] {
	if len(items) == 0 {
		return nil
	}

	if concurrency <= 0 {
		concurrency = DefaultConcurrency
	}

	if concurrency > len(items) {
		concurrency = len(items)
	}

	results := make([]BatchResult[R], len(items))
	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrency)

	for i, item := range items {
		select {
		case <-ctx.Done():
			results[i] = BatchResult[R]{
				Index: i,
				Error: ctx.Err(),
			}
			continue
		default:
		}

		wg.Add(1)
		go func(index int, itm T) {
			defer wg.Done()

			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				results[index] = BatchResult[R]{
					Index: index,
					Error: ctx.Err(),
				}
				return
			}

			result, err := fn(ctx, index, itm)
			results[index] = BatchResult[R]{
				Index:  index,
				Result: result,
				Error:  err,
			}
		}(i, item)
	}

	wg.Wait()
	return results
}
