// Package output provides utilities for formatting and filtering command output.
package output

import (
	"encoding/json"
	"io"

	"github.com/itchyny/gojq"
)

func ParseJQ(expr string) (*gojq.Query, error) {
	if expr == "" {
		expr = "."
	}
	return gojq.Parse(expr)
}

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
