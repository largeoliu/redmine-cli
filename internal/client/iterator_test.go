// internal/client/iterator_test.go
package client

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestPaginate(t *testing.T) {
	t.Run("normal pagination", func(t *testing.T) {
		callCount := 0
		fetcher := func(ctx context.Context, offset, limit int) ([]int, int, error) {
			callCount++
			total := 250

			var items []int
			start := offset
			end := offset + limit
			if end > total {
				end = total
			}
			for i := start; i < end; i++ {
				items = append(items, i)
			}

			return items, total, nil
		}

		ctx := context.Background()
		var results []int
		for item, err := range Paginate(ctx, fetcher, 100) {
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			results = append(results, item)
		}

		if callCount != 3 {
			t.Errorf("expected 3 calls, got %d", callCount)
		}

		if len(results) != 250 {
			t.Errorf("expected 250 items, got %d", len(results))
		}

		for i := 0; i < 250; i++ {
			if results[i] != i {
				t.Errorf("expected results[%d] = %d, got %d", i, i, results[i])
			}
		}
	})

	t.Run("empty result", func(t *testing.T) {
		fetcher := func(ctx context.Context, offset, limit int) ([]int, int, error) {
			return []int{}, 0, nil
		}

		ctx := context.Background()
		var results []int
		for item, err := range Paginate(ctx, fetcher, 100) {
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			results = append(results, item)
		}

		if len(results) != 0 {
			t.Errorf("expected 0 items, got %d", len(results))
		}
	})

	t.Run("error handling", func(t *testing.T) {
		expectedErr := errors.New("fetch error")
		fetcher := func(ctx context.Context, offset, limit int) ([]int, int, error) {
			return nil, 0, expectedErr
		}

		ctx := context.Background()
		var results []int
		var gotErr error
		for item, err := range Paginate(ctx, fetcher, 100) {
			if err != nil {
				gotErr = err
				break
			}
			results = append(results, item)
		}

		if gotErr == nil {
			t.Fatal("expected error, got nil")
		}
		if gotErr.Error() != expectedErr.Error() {
			t.Errorf("expected error %v, got %v", expectedErr, gotErr)
		}
		if len(results) != 0 {
			t.Errorf("expected 0 items, got %d", len(results))
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		fetcher := func(ctx context.Context, offset, limit int) ([]int, int, error) {
			select {
			case <-ctx.Done():
				return nil, 0, ctx.Err()
			default:
			}

			time.Sleep(50 * time.Millisecond)

			var items []int
			for i := offset; i < offset+limit && i < 1000; i++ {
				items = append(items, i)
			}
			return items, 1000, nil
		}

		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			time.Sleep(10 * time.Millisecond)
			cancel()
		}()

		var results []int
		var gotErr error
		for item, err := range Paginate(ctx, fetcher, 100) {
			if err != nil {
				gotErr = err
				break
			}
			results = append(results, item)
		}

		if gotErr == nil {
			t.Fatal("expected context canceled error, got nil")
		}
		if !errors.Is(gotErr, context.Canceled) {
			t.Errorf("expected context.Canceled error, got %v", gotErr)
		}
	})

	t.Run("early stop iteration", func(t *testing.T) {
		callCount := 0
		fetcher := func(ctx context.Context, offset, limit int) ([]int, int, error) {
			callCount++
			var items []int
			for i := offset; i < offset+limit && i < 1000; i++ {
				items = append(items, i)
			}
			return items, 1000, nil
		}

		ctx := context.Background()
		var results []int
		for item, err := range Paginate(ctx, fetcher, 100) {
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			results = append(results, item)
			if len(results) >= 150 {
				break
			}
		}

		if len(results) != 150 {
			t.Errorf("expected 150 items, got %d", len(results))
		}

		if callCount != 2 {
			t.Errorf("expected 2 calls, got %d", callCount)
		}
	})

	t.Run("default page size", func(t *testing.T) {
		fetcher := func(ctx context.Context, offset, limit int) ([]int, int, error) {
			if limit != DefaultPageSize {
				t.Errorf("expected limit %d, got %d", DefaultPageSize, limit)
			}
			return []int{1, 2, 3}, 3, nil
		}

		ctx := context.Background()
		var results []int
		for item, err := range Paginate(ctx, fetcher, 0) {
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			results = append(results, item)
		}

		if len(results) != 3 {
			t.Errorf("expected 3 items, got %d", len(results))
		}
	})
}

func TestCollectAll(t *testing.T) {
	t.Run("normal collection", func(t *testing.T) {
		fetcher := func(ctx context.Context, offset, limit int) ([]int, int, error) {
			total := 250
			var items []int
			start := offset
			end := offset + limit
			if end > total {
				end = total
			}
			for i := start; i < end; i++ {
				items = append(items, i)
			}
			return items, total, nil
		}

		ctx := context.Background()
		items, err := CollectAll(Paginate(ctx, fetcher, 100))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(items) != 250 {
			t.Errorf("expected 250 items, got %d", len(items))
		}

		for i := 0; i < 250; i++ {
			if items[i] != i {
				t.Errorf("expected items[%d] = %d, got %d", i, i, items[i])
			}
		}
	})

	t.Run("error handling", func(t *testing.T) {
		expectedErr := errors.New("fetch error")
		fetcher := func(ctx context.Context, offset, limit int) ([]int, int, error) {
			return nil, 0, expectedErr
		}

		ctx := context.Background()
		items, err := CollectAll(Paginate(ctx, fetcher, 100))
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != expectedErr.Error() {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
		if items != nil {
			t.Errorf("expected nil items, got %v", items)
		}
	})
}
