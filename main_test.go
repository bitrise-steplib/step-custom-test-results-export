package main

import "testing"

func Test_multipleMatchesWarning(t *testing.T) {
	tests := []struct {
		name    string
		matches []string
		want    string
	}{
		{
			name:    "Five or less matches",
			matches: []string{"match_1", "match_2", "match_3", "match_4", "match_5"},
			want: `Provided search pattern matches 5 files:
- match_1
- match_2
- match_3
- match_4
- match_5
`,
		},
		{
			name:    "More than five matches",
			matches: []string{"match_1", "match_2", "match_3", "match_4", "match_5", "match_6"},
			want: `Provided search pattern matches 6 files:
- match_1
- match_2
- match_3
- match_4
- match_5
...
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := multipleMatchesWarning(tt.matches); got != tt.want {
				t.Errorf("multipleMatchesWarning() = %v, want %v", got, tt.want)
			}
		})
	}
}
