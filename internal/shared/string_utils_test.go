package shared

import "testing"

func TestStringUtils(t *testing.T) {
	tests := []struct {
		name string
		got  bool
		want bool
	}{
		{
			name: "StringInSlice finds an exact match",
			got:  StringInSlice("mp3", []string{"flac", "mp3"}),
			want: true,
		},
		{
			name: "StringInSlice rejects a missing value",
			got:  StringInSlice("wav", []string{"flac", "mp3"}),
			want: false,
		},
		{
			name: "ContainsAnyOf matches case insensitively",
			got:  ContainsAnyOf("The Beatles", []string{"BEATLES"}),
			want: true,
		},
		{
			name: "ContainsAnyOf rejects an empty input",
			got:  ContainsAnyOf("", []string{"anything"}),
			want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.got != test.want {
				t.Errorf("got %v, want %v", test.got, test.want)
			}
		})
	}
}
