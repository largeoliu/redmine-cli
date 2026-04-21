package agile

import "encoding/json"

// UnmarshalJSON accepts both the legacy agile_sprints payload and the current sprints payload.
func (l *SprintList) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}

	if raw, ok := fields["agile_sprints"]; ok {
		if err := json.Unmarshal(raw, &l.AgileSprints); err != nil {
			return err
		}
	} else if raw, ok := fields["sprints"]; ok {
		if err := json.Unmarshal(raw, &l.AgileSprints); err != nil {
			return err
		}
	}

	if raw, ok := fields["total_count"]; ok {
		if err := json.Unmarshal(raw, &l.TotalCount); err != nil {
			return err
		}
	}
	if raw, ok := fields["limit"]; ok {
		if err := json.Unmarshal(raw, &l.Limit); err != nil {
			return err
		}
	}
	if raw, ok := fields["offset"]; ok {
		if err := json.Unmarshal(raw, &l.Offset); err != nil {
			return err
		}
	}

	return nil
}
