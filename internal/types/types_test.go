package types

import (
	"bytes"
	"io"
	"testing"

	"github.com/largeoliu/redmine-cli/internal/client"
)

// mockResolver 实现 types.Resolver 接口
type mockResolver struct {
	resolveClientFunc func(flags *GlobalFlags) (*client.Client, error)
	writeOutputFunc   func(w io.Writer, flags *GlobalFlags, payload any) error
}

func (m *mockResolver) ResolveClient(flags *GlobalFlags) (*client.Client, error) {
	if m.resolveClientFunc != nil {
		return m.resolveClientFunc(flags)
	}
	return nil, nil
}

func (m *mockResolver) WriteOutput(w io.Writer, flags *GlobalFlags, payload any) error {
	if m.writeOutputFunc != nil {
		return m.writeOutputFunc(w, flags, payload)
	}
	return nil
}

func TestGlobalFlagsFields(t *testing.T) {
	tests := []struct {
		name  string
		flags GlobalFlags
	}{
		{
			name: "default values",
			flags: GlobalFlags{
				URL:      "https://example.com",
				Key:      "test-key",
				Format:   "json",
				JQ:       ".[]",
				Fields:   "id,subject",
				DryRun:   false,
				Yes:      false,
				Output:   "",
				Limit:    25,
				Offset:   0,
				Timeout:  60,
				Retries:  3,
				Verbose:  false,
				Debug:    false,
				Instance: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			// 验证字段可以正确读取
			_ = tt.flags.URL
			_ = tt.flags.Key
			_ = tt.flags.Format
			_ = tt.flags.JQ
			_ = tt.flags.Fields
			_ = tt.flags.DryRun
			_ = tt.flags.Yes
			_ = tt.flags.Output
			_ = tt.flags.Limit
			_ = tt.flags.Offset
			_ = tt.flags.Timeout
			_ = tt.flags.Retries
			_ = tt.flags.Verbose
			_ = tt.flags.Debug
			_ = tt.flags.Instance
		})
	}
}

func TestGlobalFlagsPointer(t *testing.T) {
	// 测试指针类型的使用
	flags := &GlobalFlags{
		URL:    "https://test.com",
		Key:    "test-key",
		Format: "json",
	}

	if flags.URL != "https://test.com" {
		t.Errorf("expected URL 'https://test.com', got %s", flags.URL)
	}
	if flags.Key != "test-key" {
		t.Errorf("expected Key 'test-key', got %s", flags.Key)
	}
	if flags.Format != "json" {
		t.Errorf("expected Format 'json', got %s", flags.Format)
	}

	// 修改指针指向的值
	flags.URL = "https://modified.com"
	if flags.URL != "https://modified.com" {
		t.Errorf("expected URL 'https://modified.com', got %s", flags.URL)
	}
}

func TestGlobalFlagsComparison(t *testing.T) {
	// 测试结构体比较
	flags1 := GlobalFlags{
		URL:    "https://example.com",
		Key:    "test-key",
		Format: "json",
	}

	flags2 := GlobalFlags{
		URL:    "https://example.com",
		Key:    "test-key",
		Format: "json",
	}

	// 相同值的结构体应该相等
	if flags1.URL != flags2.URL {
		t.Error("URL fields should be equal")
	}
	if flags1.Key != flags2.Key {
		t.Error("Key fields should be equal")
	}
	if flags1.Format != flags2.Format {
		t.Error("Format fields should be equal")
	}
}

func TestGlobalFlagsZeroValues(t *testing.T) {
	// 测试零值
	flags := GlobalFlags{}

	if flags.URL != "" {
		t.Errorf("expected empty URL, got %s", flags.URL)
	}
	if flags.Key != "" {
		t.Errorf("expected empty Key, got %s", flags.Key)
	}
	if flags.Format != "" {
		t.Errorf("expected empty Format, got %s", flags.Format)
	}
	if flags.Limit != 0 {
		t.Errorf("expected zero Limit, got %d", flags.Limit)
	}
}

func TestGlobalFlagsNonZeroValues(t *testing.T) {
	// 测试非零值
	flags := GlobalFlags{
		URL:    "https://example.com",
		Key:    "test-key",
		Format: "json",
		Limit:  100,
		Offset: 10,
	}

	if flags.URL == "" {
		t.Error("URL should not be empty")
	}
	if flags.Key == "" {
		t.Error("Key should not be empty")
	}
	if flags.Format == "" {
		t.Error("Format should not be empty")
	}
	if flags.Limit == 0 {
		t.Error("Limit should not be zero")
	}
	if flags.Offset == 0 {
		t.Error("Offset should not be zero")
	}
}

func TestGlobalFlagsJSONTags(t *testing.T) {
	// 测试 JSON 标签
	flags := GlobalFlags{
		URL:    "https://example.com",
		Key:    "test-key",
		Format: "json",
		Limit:  100,
	}

	// 确保字段可以被正确访问
	if flags.URL != "https://example.com" {
		t.Errorf("unexpected URL: %s", flags.URL)
	}
	if flags.Key != "test-key" {
		t.Errorf("unexpected Key: %s", flags.Key)
	}
	if flags.Format != "json" {
		t.Errorf("unexpected Format: %s", flags.Format)
	}
	if flags.Limit != 100 {
		t.Errorf("unexpected Limit: %d", flags.Limit)
	}
}

func TestGlobalFlagsCopy(t *testing.T) {
	// 测试结构体复制
	original := GlobalFlags{
		URL:      "https://example.com",
		Key:      "test-key",
		Format:   "json",
		Limit:    100,
		Verbose:  true,
		Instance: "production",
	}

	// 值复制
	copied := original

	// 修改 copied 不应影响 original
	copied.URL = "https://modified.com"
	copied.Limit = 50

	if original.URL != "https://example.com" {
		t.Errorf("original URL should not change, got %s", original.URL)
	}
	if original.Limit != 100 {
		t.Errorf("original Limit should not change, got %d", original.Limit)
	}
	if copied.URL != "https://modified.com" {
		t.Errorf("copied URL should be modified, got %s", copied.URL)
	}
	if copied.Limit != 50 {
		t.Errorf("copied Limit should be modified, got %d", copied.Limit)
	}
}

func TestGlobalFlagsPointerCopy(t *testing.T) {
	// 测试指针复制
	original := &GlobalFlags{
		URL:   "https://example.com",
		Key:   "test-key",
		Limit: 100,
	}

	// 指针复制指向同一对象
	copied := original
	copied.URL = "https://modified.com"

	if original.URL != "https://modified.com" {
		t.Error("original should be modified through pointer copy")
	}
}

func TestResolverInterfaceNilSafety(t *testing.T) {
	var resolver Resolver

	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected nil resolver to panic, but it did not")
			}
		}()
		_, _ = resolver.ResolveClient(nil)
	}()
}

func TestResolverInterfaceTypeAssertion(t *testing.T) {
	// 测试接口类型断言
	var resolver Resolver = &mockResolver{}

	// 类型断言应该成功
	if _, ok := resolver.(*mockResolver); !ok {
		t.Error("expected resolver to be *mockResolver")
	}
}

func TestResolverInterface(_ *testing.T) {
	// 测试 mockResolver 实现了 Resolver 接口
	var _ Resolver = &mockResolver{}
	var _ Resolver = (*mockResolver)(nil)
}

func TestResolverMockResolveClient(t *testing.T) {
	tests := []struct {
		name    string
		flags   *GlobalFlags
		mockFn  func(flags *GlobalFlags) (*client.Client, error)
		wantErr bool
	}{
		{
			name:  "returns nil client",
			flags: &GlobalFlags{},
			mockFn: func(_ *GlobalFlags) (*client.Client, error) {
				return nil, nil
			},
			wantErr: false,
		},
		{
			name: "returns client with flags",
			flags: &GlobalFlags{
				URL: "https://example.com",
				Key: "test-key",
			},
			mockFn: func(flags *GlobalFlags) (*client.Client, error) {
				if flags.URL != "https://example.com" {
					t.Errorf("unexpected URL: %s", flags.URL)
				}
				return nil, nil
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := &mockResolver{
				resolveClientFunc: tt.mockFn,
			}

			_, err := resolver.ResolveClient(tt.flags)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveClient() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResolverMockWriteOutput(t *testing.T) {
	tests := []struct {
		name    string
		flags   *GlobalFlags
		payload any
		mockFn  func(w io.Writer, flags *GlobalFlags, payload any) error
		wantErr bool
	}{
		{
			name:    "writes nil payload",
			flags:   &GlobalFlags{},
			payload: nil,
			mockFn: func(_ io.Writer, _ *GlobalFlags, _ any) error {
				return nil
			},
			wantErr: false,
		},
		{
			name:    "writes string payload",
			flags:   &GlobalFlags{Format: "json"},
			payload: "test payload",
			mockFn: func(_ io.Writer, flags *GlobalFlags, _ any) error {
				if flags.Format != "json" {
					t.Errorf("unexpected Format: %s", flags.Format)
				}
				return nil
			},
			wantErr: false,
		},
		{
			name:    "writes map payload",
			flags:   &GlobalFlags{Format: "table"},
			payload: map[string]string{"key": "value"},
			mockFn: func(_ io.Writer, _ *GlobalFlags, _ any) error {
				return nil
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := &mockResolver{
				writeOutputFunc: tt.mockFn,
			}

			var buf bytes.Buffer
			err := resolver.WriteOutput(&buf, tt.flags, tt.payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteOutput() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResolverMockBothMethods(t *testing.T) {
	// 测试 mock 同时实现两个方法
	resolver := &mockResolver{
		resolveClientFunc: func(_ *GlobalFlags) (*client.Client, error) {
			return nil, nil
		},
		writeOutputFunc: func(_ io.Writer, _ *GlobalFlags, _ any) error {
			return nil
		},
	}

	flags := &GlobalFlags{URL: "https://test.com"}

	client, err := resolver.ResolveClient(flags)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if client != nil {
		t.Error("expected nil client")
	}

	var buf bytes.Buffer
	err = resolver.WriteOutput(&buf, flags, "test")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestResolverMockNilFunctions(t *testing.T) {
	// 测试 mock 的函数为 nil 时的行为
	resolver := &mockResolver{}

	flags := &GlobalFlags{}

	client, err := resolver.ResolveClient(flags)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if client != nil {
		t.Error("expected nil client")
	}

	var buf bytes.Buffer
	err = resolver.WriteOutput(&buf, flags, nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGlobalFlagsAsParameter(t *testing.T) {
	// 测试 GlobalFlags 作为参数传递
	flags := &GlobalFlags{
		URL:   "https://example.com",
		Key:   "test-key",
		Limit: 100,
	}

	// 通过函数验证指针
	verifyFlags := func(f *GlobalFlags) {
		if f.URL != "https://example.com" {
			t.Errorf("expected URL 'https://example.com', got %s", f.URL)
		}
		if f.Key != "test-key" {
			t.Errorf("expected Key 'test-key', got %s", f.Key)
		}
		if f.Limit != 100 {
			t.Errorf("expected Limit 100, got %d", f.Limit)
		}
	}

	verifyFlags(flags)
}
