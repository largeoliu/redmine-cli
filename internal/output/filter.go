// Package output provides utilities for formatting and filtering command output.
package output

import (
	"encoding/json"
	"io"

	"github.com/itchyny/gojq"
)

// ApplyJQ applies a jq expression to filter and transform JSON data.
func ApplyJQ(w io.Writer, payload any, expr string) error {
	query, err := gojq.Parse(expr)
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

	iter := query.Run(input)
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

// SelectFields extracts only the specified fields from a payload.
func SelectFields(payload any, fields []string) (any, error) {
	if len(fields) == 0 {
		return payload, nil
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	var input any
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, err
	}

	m, ok := input.(map[string]any)
	if !ok {
		return payload, nil
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
