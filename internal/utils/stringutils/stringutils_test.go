package stringutils

import "testing"

func TestIsEmpty(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want bool
	}{
		{
			name: "Empty string",
			arg:  "",
			want: true,
		},
		{
			name: "String with spaces",
			arg:  "   ",
			want: true,
		},
		{
			name: "String with tabs",
			arg:  "			",
			want: true,
		},
		{
			name: "String with text",
			arg:  "Whatever",
			want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := IsEmpty(test.arg); got != test.want {
				t.Errorf("IsEmpty() = %v, want %v", got, test.want)
			}
		})
	}
}
