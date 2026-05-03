// Package output provides utilities for formatting and filtering command output.
package output

import (
	"encoding/json"
	"io"

	"github.com/itchyny/gojq"
)

// ParseJQ parses and compiles a jq expression. The returned *gojq.Query
// is safe to reuse across many calls for the same expression.
func ParseJQ(expr string) (*gojq.Query, error) {
	if expr == "" {
		expr = "."
	}
	return gojq.Parse(expr)
}

// ApplyJQNormalized applies a pre-compiled jq query to normalized data.
// The caller is responsible for ensuring data is already unmarshaled.
func ApplyJQNormalized(w io.Writer, data any, query *gojq.Query) error {
	iter := query.Run(data)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			return err
		}
		output, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return err
		}
		if _, err := w.Write(output); err != nil {
			return err
		}
		if _, err := w.Write([]byte("\n")); err != nil {
			return err
		}
	}
	return nil
}

// ApplyJQ applies a jq expression to filter and transform JSON data.
// It parses the expression on each call; for better performance, use
// ParseJQ + ApplyJQNormalized when applying the same expression repeatedly.
func ApplyJQ(w io.Writer, payload any, expr string) error {
	if expr == "" {
		expr = "."
	}
	query, err := ParseJQ(expr)
	if err != nil {
		return err
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	var input any
	if err := json.Unmarshal(data, &input); err != nil {
		return err
	}
	return ApplyJQNormalized(w, input, query)
}

// SelectFieldsNormalized extracts specified fields from already-normalized data.
//
//nolint:gocyclo // complexity 19 is unavoidable for this filtering logic
func SelectFieldsNormalized(input any, fields []string) (any, error) {
	if len(fields) == 0 {
		return input, nil
	}
	if arr, ok := input.([]any); ok {
		fieldSet := make(map[string]struct{}, len(fields))
		for _, f := range fields {
			fieldSet[f] = struct{}{}
		}
		result := make([]any, 0, len(arr))
		for _, item := range arr {
			if itemMap, ok := item.(map[string]any); ok {
				filteredItem := make(map[string]any)
				for _, field := range fields {
					if v, exists := itemMap[field]; exists {
						filteredItem[field] = v
					}
				}
				if len(filteredItem) > 0 {
					result = append(result, filteredItem)
				}
			} else {
				result = append(result, item)
			}
		}
		return result, nil
	}
	m, ok := input.(map[string]any)
	if !ok {
		return input, nil
	}
	fieldSet := make(map[string]struct{}, len(fields))
	for _, f := range fields {
		fieldSet[f] = struct{}{}
	}
	result := make(map[string]any)
	for key, value := range m {
		if _, include := fieldSet[key]; include {
			result[key] = value
		} else {
			if arr, ok := value.([]any); ok {
				filteredArr := make([]any, 0, len(arr))
				for _, item := range arr {
					if itemMap, ok := item.(map[string]any); ok {
						filteredItem := make(map[string]any)
						for _, field := range fields {
							if v, exists := itemMap[field]; exists {
								filteredItem[field] = v
							}
						}
						if len(filteredItem) > 0 {
							filteredArr = append(filteredArr, filteredItem)
						}
					} else {
						filteredArr = append(filteredArr, item)
					}
				}
				result[key] = filteredArr
			}
		}
	}
	return result, nil
}

// SelectFields extracts only the specified fields from a payload.
// For already-normalized data (map[string]any or []any), it uses
// the more efficient SelectFieldsNormalized path.
func SelectFields(payload any, fields []string) (any, error) {
	if len(fields) == 0 {
		return payload, nil
	}
	if _, ok := payload.(map[string]any); ok {
		return SelectFieldsNormalized(payload, fields)
	}
	if _, ok := payload.([]any); ok {
		return SelectFieldsNormalized(payload, fields)
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	var input any
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, err
	}
	return SelectFieldsNormalized(input, fields)
}
