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

func TestMergeCustomFields(t *testing.T) {
	tests := []struct {
		name        string
		interactive []CustomField
		flags       []CustomField
		want        []CustomField
	}{
		{
			name:        "both empty",
			interactive: []CustomField{},
			flags:       []CustomField{},
			want:        nil,
		},
		{
			name:        "only interactive",
			interactive: []CustomField{{ID: 1, Value: "v1"}},
			flags:       []CustomField{},
			want:        []CustomField{{ID: 1, Value: "v1"}},
		},
		{
			name:        "only flags",
			interactive: []CustomField{},
			flags:       []CustomField{{ID: 2, Value: "v2"}},
			want:        []CustomField{{ID: 2, Value: "v2"}},
		},
		{
			name:        "flags override interactive",
			interactive: []CustomField{{ID: 1, Value: "interactive"}, {ID: 2, Value: "i2"}},
			flags:       []CustomField{{ID: 1, Value: "flag"}, {ID: 3, Value: "f3"}},
			want:        []CustomField{{ID: 1, Value: "flag"}, {ID: 2, Value: "i2"}, {ID: 3, Value: "f3"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeCustomFields(tt.interactive, tt.flags)
			if !equalCustomFields(got, tt.want) {
				t.Errorf("mergeCustomFields() = %v, want %v", got, tt.want)
			}
		})
	}
}

func equalCustomFields(a, b []CustomField) bool {
	if len(a) != len(b) {
		return false
	}
	m := make(map[int]any)
	for _, cf := range a {
		m[cf.ID] = cf.Value
	}
	for _, cf := range b {
		if m[cf.ID] != cf.Value {
			return false
		}
	}
	return true
}

func FuzzParseCustomFieldFlags(f *testing.F) {
	testcases := []string{
		"id:5:value",
		"id:5:高",
		"name:value",
		"invalid",
		"",
		"id:abc:value",
		"id:5:v1.0,v2.0",
		"id:5:value:with:colons",
		"id:5:",
		"id:5:value,",
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(_ *testing.T, input string) {
		if len(input) > 1000 {
			return
		}

		flags := []string{input}
		_, _ = parseCustomFieldFlags(flags, nil)
	})
}
