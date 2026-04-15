package issues

import (
	"testing"
)

func TestParseCustomFieldFlags_IDFormat(t *testing.T) {
	tests := []struct {
		name    string
		flags   []string
		want    []CustomField
		wantErr bool
	}{
		{
			name:  "single id format",
			flags: []string{"id:5:高"},
			want:  []CustomField{{ID: 5, Value: "高"}},
		},
		{
			name:  "multiple id format",
			flags: []string{"id:5:高", "id:6:v1.0"},
			want:  []CustomField{{ID: 5, Value: "高"}, {ID: 6, Value: "v1.0"}},
		},
		{
			name:  "id format with comma in value",
			flags: []string{"id:5:v1.0,v2.0"},
			want:  []CustomField{{ID: 5, Value: "v1.0,v2.0"}},
		},
		{
			name:    "invalid no colon",
			flags:   []string{"invalid"},
			wantErr: true,
		},
		{
			name:    "invalid id not number",
			flags:   []string{"id:abc:value"},
			wantErr: true,
		},
		{
			name:  "empty flags",
			flags: []string{},
			want:  nil,
		},
		{
			name:  "value with colon",
			flags: []string{"id:5:value:with:colons"},
			want:  []CustomField{{ID: 5, Value: "value:with:colons"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCustomFieldFlags(tt.flags, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseCustomFieldFlags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !equalCustomFields(got, tt.want) {
				t.Errorf("parseCustomFieldFlags() = %v, want %v", got, tt.want)
			}
		})
	}
}

func equalCustomFields(a, b []CustomField) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].ID != b[i].ID {
			return false
		}
		if a[i].Value != b[i].Value {
			return false
		}
	}
	return true
}
