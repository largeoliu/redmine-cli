// internal/output/sanitize_extended_test.go
package output

import (
	"testing"
	"unicode"
)

// TestSanitizeComprehensive 全面测试 Sanitize 函数
func TestSanitizeComprehensive(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "normal text",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "text with newline",
			input:    "hello\nworld",
			expected: "hello\nworld",
		},
		{
			name:     "text with tab",
			input:    "hello\tworld",
			expected: "hello\tworld",
		},
		{
			name:     "text with carriage return",
			input:    "hello\rworld",
			expected: "hello world",
		},
		{
			name:     "text with null character",
			input:    "hello\x00world",
			expected: "hello world",
		},
		{
			name:     "text with bell character",
			input:    "hello\x07world",
			expected: "hello world",
		},
		{
			name:     "text with backspace",
			input:    "hello\x08world",
			expected: "hello world",
		},
		{
			name:     "text with form feed",
			input:    "hello\x0Cworld",
			expected: "hello world",
		},
		{
			name:     "text with vertical tab",
			input:    "hello\x0Bworld",
			expected: "hello world",
		},
		{
			name:     "text with escape character",
			input:    "hello\x1Bworld",
			expected: "hello world",
		},
		{
			name:     "text with delete character",
			input:    "hello\x7Fworld",
			expected: "hello world",
		},
		{
			name:     "text with multiple control characters",
			input:    "hello\x00\x01\x02world",
			expected: "hello   world",
		},
		{
			name:     "text with mixed control and normal characters",
			input:    "hello\x00world\tgood\nmorning",
			expected: "hello world\tgood\nmorning",
		},
		{
			name:     "only control characters",
			input:    "\x00\x01\x02\x03\x04\x05",
			expected: "      ",
		},
		{
			name:     "only newlines and tabs",
			input:    "\n\t\n\t",
			expected: "\n\t\n\t",
		},
		{
			name:     "unicode text",
			input:    "你好世界",
			expected: "你好世界",
		},
		{
			name:     "unicode with control characters",
			input:    "你好\x00世界",
			expected: "你好 世界",
		},
		{
			name:     "emoji text",
			input:    "hello 😀 world",
			expected: "hello 😀 world",
		},
		{
			name:     "emoji with control characters",
			input:    "hello\x00😀world",
			expected: "hello 😀world",
		},
		{
			name:     "text with CRLF",
			input:    "hello\r\nworld",
			expected: "hello \nworld",
		},
		{
			name:     "text with only spaces",
			input:    "     ",
			expected: "     ",
		},
		{
			name:     "text with leading control characters",
			input:    "\x00\x01hello",
			expected: "  hello",
		},
		{
			name:     "text with trailing control characters",
			input:    "hello\x00\x01",
			expected: "hello  ",
		},
		{
			name:     "long text with control characters",
			input:    "this is a long text with \x00 control \x01 characters \x02 embedded",
			expected: "this is a long text with   control   characters   embedded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Sanitize(tt.input)
			if result != tt.expected {
				t.Errorf("Sanitize() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestSanitizePreservesNewlineAndTab 确保换行符和制表符被保留
func TestSanitizePreservesNewlineAndTab(t *testing.T) {
	input := "line1\nline2\ttab"
	result := Sanitize(input)

	// 检查换行符和制表符是否被保留
	if !containsRune(result, '\n') {
		t.Error("expected newline to be preserved")
	}
	if !containsRune(result, '\t') {
		t.Error("expected tab to be preserved")
	}
}

// TestSanitizeRemovesOtherControlChars 确保其他控制字符被移除
func TestSanitizeRemovesOtherControlChars(t *testing.T) {
	// 测试所有控制字符（除了 \n 和 \t）
	for r := rune(0); r < 32; r++ {
		if r == '\n' || r == '\t' {
			continue
		}
		input := string([]rune{'a', r, 'b'})
		result := Sanitize(input)
		// 控制字符应该被替换为空格
		if containsRune(result, r) {
			t.Errorf("expected control character %d to be removed", r)
		}
	}
}

// TestSanitizeRemovesDeleteChar 测试删除字符（DEL）
func TestSanitizeRemovesDeleteChar(t *testing.T) {
	input := "hello\x7Fworld"
	result := Sanitize(input)
	if containsRune(result, '\x7F') {
		t.Error("expected DEL character to be removed")
	}
}

// TestSanitizeWithAllASCIICControlChars 测试所有 ASCII 控制字符
func TestSanitizeWithAllASCIICControlChars(t *testing.T) {
	var inputBuilder []rune
	var expectedBuilder []rune

	// 添加所有 ASCII 控制字符
	for r := rune(0); r < 32; r++ {
		inputBuilder = append(inputBuilder, r)
		if r == '\n' || r == '\t' {
			expectedBuilder = append(expectedBuilder, r)
		} else {
			expectedBuilder = append(expectedBuilder, ' ')
		}
	}
	// 添加 DEL 字符
	inputBuilder = append(inputBuilder, '\x7F')
	expectedBuilder = append(expectedBuilder, ' ')

	input := string(inputBuilder)
	expected := string(expectedBuilder)
	result := Sanitize(input)

	if result != expected {
		t.Errorf("Sanitize() = %q, want %q", result, expected)
	}
}

// TestSanitizeDoesNotModifyNonControlChars 确保非控制字符不被修改
func TestSanitizeDoesNotModifyNonControlChars(t *testing.T) {
	input := "Hello, World! 123 @#$%^&*()"
	result := Sanitize(input)
	if result != input {
		t.Errorf("expected non-control characters to be unchanged, got %q", result)
	}
}

// TestSanitizeWithUnicodeNonControlChars 测试 Unicode 非控制字符
func TestSanitizeWithUnicodeNonControlChars(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "Chinese characters",
			input: "中文测试",
		},
		{
			name:  "Japanese characters",
			input: "日本語テスト",
		},
		{
			name:  "Korean characters",
			input: "한국어 테스트",
		},
		{
			name:  "Arabic characters",
			input: "مرحبا",
		},
		{
			name:  "Hebrew characters",
			input: "שלום",
		},
		{
			name:  "Greek characters",
			input: "Γειά σου",
		},
		{
			name:  "Russian characters",
			input: "Привет",
		},
		{
			name:  "Mathematical symbols",
			input: "∑ ∫ √ ∞",
		},
		{
			name:  "Currency symbols",
			input: "$ € £ ¥",
		},
		{
			name:  "Special punctuation",
			input: "«»‹›—–",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Sanitize(tt.input)
			if result != tt.input {
				t.Errorf("expected Unicode characters to be unchanged, got %q", result)
			}
		})
	}
}

// TestSanitizePerformance 测试性能（简单基准）
func TestSanitizePerformance(t *testing.T) {
	// 创建一个长字符串
	longStr := ""
	for i := 0; i < 1000; i++ {
		longStr += "hello\x00world\tgood\nmorning"
	}

	// 运行一次以确保没有 panic
	result := Sanitize(longStr)
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

// TestSanitizeWithEmptyInput 测试空输入
func TestSanitizeWithEmptyInput(t *testing.T) {
	result := Sanitize("")
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

// TestSanitizeWithSingleControlChar 测试单个控制字符
func TestSanitizeWithSingleControlChar(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single null",
			input:    "\x00",
			expected: " ",
		},
		{
			name:     "single newline",
			input:    "\n",
			expected: "\n",
		},
		{
			name:     "single tab",
			input:    "\t",
			expected: "\t",
		},
		{
			name:     "single carriage return",
			input:    "\r",
			expected: " ",
		},
		{
			name:     "single bell",
			input:    "\x07",
			expected: " ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Sanitize(tt.input)
			if result != tt.expected {
				t.Errorf("Sanitize() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestSanitizeWithRepeatedControlChars 测试重复的控制字符
func TestSanitizeWithRepeatedControlChars(t *testing.T) {
	input := "hello\x00\x00\x00world"
	result := Sanitize(input)
	expected := "hello   world"
	if result != expected {
		t.Errorf("Sanitize() = %q, want %q", result, expected)
	}
}

// TestSanitizeWithMixedLineEndings 测试混合行结束符
func TestSanitizeWithMixedLineEndings(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "LF only",
			input:    "line1\nline2\nline3",
			expected: "line1\nline2\nline3",
		},
		{
			name:     "CRLF",
			input:    "line1\r\nline2\r\nline3",
			expected: "line1 \nline2 \nline3",
		},
		{
			name:     "CR only",
			input:    "line1\rline2\rline3",
			expected: "line1 line2 line3",
		},
		{
			name:     "mixed",
			input:    "line1\nline2\r\nline3\rline4",
			expected: "line1\nline2 \nline3 line4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Sanitize(tt.input)
			if result != tt.expected {
				t.Errorf("Sanitize() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestSanitizeWithSpecialWhitespace 测试特殊空白字符
func TestSanitizeWithSpecialWhitespace(t *testing.T) {
	// 测试各种空白字符
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "non-breaking space",
			input: "hello\u00A0world",
		},
		{
			name:  "en quad",
			input: "hello\u2000world",
		},
		{
			name:  "em quad",
			input: "hello\u2001world",
		},
		{
			name:  "en space",
			input: "hello\u2002world",
		},
		{
			name:  "em space",
			input: "hello\u2003world",
		},
		{
			name:  "thin space",
			input: "hello\u2009world",
		},
		{
			name:  "hair space",
			input: "hello\u200Aworld",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Sanitize(tt.input)
			// 这些 Unicode 空白字符不是控制字符，应该被保留
			if result != tt.input {
				t.Errorf("expected Unicode whitespace to be preserved, got %q", result)
			}
		})
	}
}

// 辅助函数：检查字符串是否包含指定的 rune
func containsRune(s string, r rune) bool {
	for _, c := range s {
		if c == r {
			return true
		}
	}
	return false
}

// TestSanitizeControlCharRange 测试控制字符范围
func TestSanitizeControlCharRange(t *testing.T) {
	// 测试 C0 控制字符 (0x00-0x1F)
	for i := rune(0x00); i <= 0x1F; i++ {
		input := string(i)
		result := Sanitize(input)

		if i == '\n' || i == '\t' {
			// 换行符和制表符应该被保留
			if result != input {
				t.Errorf("expected rune %d to be preserved, got %q", i, result)
			}
		} else {
			// 其他控制字符应该被替换为空格
			if result != " " {
				t.Errorf("expected control char %d to be replaced with space, got %q", i, result)
			}
		}
	}

	// 测试 DEL 字符 (0x7F)
	input := string(rune(0x7F))
	result := Sanitize(input)
	if result != " " {
		t.Errorf("expected DEL to be replaced with space, got %q", result)
	}
}

// TestSanitizeWithUnicodeControlChars 测试 Unicode 控制字符
func TestSanitizeWithUnicodeControlChars(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "next line (NEL)",
			input: "hello\u0085world",
		},
		{
			name:  "no-break space",
			input: "hello\u00A0world",
		},
		{
			name:  "zero width space",
			input: "hello\u200Bworld",
		},
		{
			name:  "zero width non-joiner",
			input: "hello\u200Cworld",
		},
		{
			name:  "zero width joiner",
			input: "hello\u200Dworld",
		},
		{
			name:  "left-to-right mark",
			input: "hello\u200Eworld",
		},
		{
			name:  "right-to-left mark",
			input: "hello\u200Fworld",
		},
		{
			name:  "byte order mark",
			input: "hello\uFEFFworld",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Sanitize(tt.input)
			// 这些 Unicode 控制字符应该根据 unicode.IsControl 判断
			for i, r := range tt.input {
				if unicode.IsControl(r) && r != '\n' && r != '\t' {
					// 如果是控制字符（除了 \n 和 \t），应该被替换
					if result[i] != ' ' {
						t.Errorf("expected control char %d to be replaced with space", r)
					}
				}
			}
		})
	}
}
