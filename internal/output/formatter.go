// Package output provides utilities for formatting and filtering command output.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/largeoliu/redmine-cli/internal/errors"
)

var (
	jsonMarshal     = json.Marshal
	jsonUnmarshalFn = jsonUnmarshal
	typeAssertMap   = typeAssertMapFn
)

func typeAssertMapFn(v any) (map[string]any, bool) {
	m, ok := v.(map[string]any)
	return m, ok
}

// Format represents the output format type.
type Format string

const (
	// FormatJSON outputs data in JSON format.
	FormatJSON Format = "json"
	// FormatTable outputs data in table format.
	FormatTable Format = "table"
	// FormatRaw outputs data in raw format.
	FormatRaw Format = "raw"
)

// Write writes data in the specified format.
func Write(w io.Writer, format Format, payload any) error {
	switch format {
	case FormatJSON:
		return WriteJSON(w, payload)
	case FormatTable:
		if _, ok := payload.([]any); !ok {
			if _, ok := payload.(map[string]any); !ok {
				normalized, err := NormalizePayload(payload)
				if err != nil {
					return err
				}
				return WriteTable(w, normalized)
			}
		}
		return WriteTable(w, payload)
	case FormatRaw:
		return WriteRaw(w, payload)
	default:
		return WriteJSON(w, payload)
	}
}

// WriteJSON writes data in JSON format.
func WriteJSON(w io.Writer, payload any) error {
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return errors.NewInternal("failed to encode JSON")
	}
	_, err = fmt.Fprintln(w, string(data))
	return err
}

// WriteRaw writes data in raw format.
func WriteRaw(w io.Writer, payload any) error {
	if text, ok := payload.(string); ok {
		_, err := fmt.Fprintln(w, Sanitize(text))
		return err
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return errors.NewInternal("failed to encode raw output")
	}
	_, err = fmt.Fprintln(w, Sanitize(string(data)))
	return err
}

// WriteTable writes data in table format.
func WriteTable(w io.Writer, payload any) error {
	normalized, err := NormalizePayload(payload)
	if err != nil {
		return err
	}

	switch typed := normalized.(type) {
	case map[string]any:
		return writeKeyValues(w, typed)
	case []any:
		headers, rows, _ := rowsFromSlice(typed)
		return writeTable(w, headers, rows)
	default:
		return WriteRaw(w, normalized)
	}
}

// NormalizePayload normalizes the payload by converting it to JSON and back.
func NormalizePayload(payload any) (any, error) {
	if payload == nil {
		return nil, nil
	}
	if text, ok := payload.(string); ok {
		return text, nil
	}
	data, err := jsonMarshal(payload)
	if err != nil {
		return nil, errors.NewInternal("failed to normalize payload")
	}
	var normalized any
	if err := jsonUnmarshalFn(data, &normalized); err != nil {
		return nil, errors.NewInternal("failed to decode normalized payload")
	}
	return normalized, nil
}

func normalizePayload(payload any) (any, error) {
	return NormalizePayload(payload)
}

func writeKeyValues(w io.Writer, payload map[string]any) error {
	keys := make([]string, 0, len(payload))
	maxWidth := 0
	for key := range payload {
		keys = append(keys, key)
		if width := len(key); width > maxWidth {
			maxWidth = width
		}
	}
	sort.Strings(keys)
	if maxWidth > 24 {
		maxWidth = 24
	}
	for _, key := range keys {
		if _, err := fmt.Fprintf(w, "%-*s  %s\n", maxWidth, key, formatValue(payload[key])); err != nil {
			return err
		}
	}
	return nil
}

func writeTable(w io.Writer, headers []string, rows [][]string) error {
	widths := calculateColumnWidths(headers, rows)
	if err := writeHeaderLine(w, headers, widths); err != nil {
		return err
	}
	if err := writeSeparatorLine(w, widths); err != nil {
		return err
	}
	return writeDataRows(w, rows, widths)
}

func calculateColumnWidths(headers []string, rows [][]string) []int {
	widths := make([]int, len(headers))
	for i, header := range headers {
		widths[i] = len(header)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i >= len(widths) {
				continue
			}
			if len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}
	for i := range widths {
		if widths[i] > 60 {
			widths[i] = 60
		}
	}
	return widths
}

func writeHeaderLine(w io.Writer, headers []string, widths []int) error {
	for i, header := range headers {
		if i > 0 {
			if _, err := io.WriteString(w, "  "); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(w, "%-*s", widths[i], truncate(header, widths[i])); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintln(w)
	return err
}

func writeSeparatorLine(w io.Writer, widths []int) error {
	for i, width := range widths {
		if i > 0 {
			if _, err := io.WriteString(w, "  "); err != nil {
				return err
			}
		}
		if _, err := io.WriteString(w, strings.Repeat("-", width)); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintln(w)
	return err
}

func writeDataRows(w io.Writer, rows [][]string, widths []int) error {
	for _, row := range rows {
		for i, cell := range row {
			if i >= len(widths) {
				continue
			}
			if i > 0 {
				if _, err := io.WriteString(w, "  "); err != nil {
					return err
				}
			}
			if _, err := fmt.Fprintf(w, "%-*s", widths[i], truncate(cell, widths[i])); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
	}
	return nil
}

func rowsFromSlice(items []any) ([]string, [][]string, bool) {
	if len(items) == 0 {
		return []string{"value"}, [][]string{}, true
	}

	allMaps := true
	keys := make(map[string]struct{})
	for _, item := range items {
		rowMap, ok := typeAssertMap(item)
		if !ok {
			allMaps = false
			break
		}
		for key := range rowMap {
			keys[key] = struct{}{}
		}
	}
	if allMaps {
		headers := sortedKeys(keys)
		rows := make([][]string, 0, len(items))
		for _, item := range items {
			rowMap, ok := typeAssertMap(item)
			if !ok {
				return nil, nil, false
			}
			row := make([]string, 0, len(headers))
			for _, key := range headers {
				row = append(row, formatValue(rowMap[key]))
			}
			rows = append(rows, row)
		}
		return headers, rows, true
	}

	rows := make([][]string, 0, len(items))
	for _, item := range items {
		rows = append(rows, []string{formatValue(item)})
	}
	return []string{"value"}, rows, true
}

func sortedKeys(keys map[string]struct{}) []string {
	out := make([]string, 0, len(keys))
	for key := range keys {
		out = append(out, key)
	}
	sort.Strings(out)
	return out
}

func formatValue(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return Sanitize(typed)
	default:
		data, err := json.Marshal(typed)
		if err != nil {
			return Sanitize(fmt.Sprintf("%v", typed))
		}
		return Sanitize(string(data))
	}
}

func truncate(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= maxWidth {
		return s
	}
	if maxWidth <= 1 {
		return "…"
	}
	return string(runes[:maxWidth-1]) + "…"
}
