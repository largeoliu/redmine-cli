// Package client provides the HTTP client for interacting with Redmine API.
package client

import (
	"context"
	"iter"
)

const (
	// DefaultPageSize 是默认的每页大小
	DefaultPageSize = 100
)

// PageFetcher 是一个泛型函数类型，用于获取指定偏移量和限制的数据
// 返回值：items 是当前页的数据项，total 是总数量，err 是错误信息
type PageFetcher[T any] func(ctx context.Context, offset, limit int) (items []T, total int, err error)

// Paginate 创建一个分页迭代器，使用 Go 1.23 的 iter.Seq2
// 它会自动处理分页逻辑，支持上下文取消
//
// 参数：
//   - ctx: 上下文，用于取消操作
//   - fetcher: 获取分页数据的函数
//   - pageSize: 每页大小，如果为 0 则使用 DefaultPageSize
//
// 返回：
//   - iter.Seq2[T, error]: 一个迭代器，每次迭代返回一个数据项和一个错误
//     当迭代完成时，错误为 nil；当发生错误时，错误不为 nil
//
// 使用示例：
//
//	for item, err := range client.Paginate(ctx, fetcher, 100) {
//	    if err != nil {
//	        // 处理错误
//	        break
//	    }
//	    // 处理 item
//	}
func Paginate[T any](ctx context.Context, fetcher PageFetcher[T], pageSize int) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		// 如果 pageSize 为 0，使用默认值
		if pageSize <= 0 {
			pageSize = DefaultPageSize
		}

		offset := 0
		for {
			// 检查上下文是否取消
			select {
			case <-ctx.Done():
				var zero T
				yield(zero, ctx.Err())
				return
			default:
			}

			// 获取当前页数据
			items, total, err := fetcher(ctx, offset, pageSize)
			if err != nil {
				var zero T
				yield(zero, err)
				return
			}

			// 遍历当前页的所有项
			for _, item := range items {
				// 如果 yield 返回 false，表示调用者停止了迭代
				if !yield(item, nil) {
					return
				}
			}

			// 更新偏移量
			offset += len(items)

			// 如果已经获取了所有数据，或者当前页数据少于 pageSize，说明已经到达最后一页
			if offset >= total || len(items) < pageSize {
				return
			}
		}
	}
}

// CollectAll 收集迭代器中的所有数据，返回切片和错误
// 这是一个便捷函数，用于将迭代器转换为切片
//
// 参数：
//   - seq: 分页迭代器
//
// 返回：
//   - []T: 所有数据项的切片
//   - error: 第一个遇到的错误，如果没有错误则为 nil
//
// 使用示例：
//
//	items, err := client.CollectAll(client.Paginate(ctx, fetcher, 100))
//	if err != nil {
//	    // 处理错误
//	}
//	// 使用 items
func CollectAll[T any](seq iter.Seq2[T, error]) ([]T, error) {
	const initialCapacity = 256
	items := make([]T, 0, initialCapacity)
	for item, err := range seq {
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}
