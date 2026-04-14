// Package output provides utilities for formatting and filtering command output.
package output

import (
	"strings"
	"unicode"
)

// Sanitize removes control characters from a string.
func Sanitize(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsControl(r) && r != '\n' && r != '\t' {
			b.WriteRune(' ')
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
