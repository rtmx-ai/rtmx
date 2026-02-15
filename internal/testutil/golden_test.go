package testutil

import (
	"testing"
)

func TestStripANSIString(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		want   string
	}{
		{
			name:   "no_ansi",
			input:  "plain text",
			want:   "plain text",
		},
		{
			name:   "simple_color",
			input:  "\033[32mgreen\033[0m",
			want:   "green",
		},
		{
			name:   "multiple_colors",
			input:  "\033[31mred\033[0m and \033[34mblue\033[0m",
			want:   "red and blue",
		},
		{
			name:   "bold",
			input:  "\033[1mbold\033[0m",
			want:   "bold",
		},
		{
			name:   "nested_codes",
			input:  "\033[1;31;40mbold red on black\033[0m",
			want:   "bold red on black",
		},
		{
			name:   "empty",
			input:  "",
			want:   "",
		},
		{
			name:   "only_escape",
			input:  "\033[0m",
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StripANSIString(tt.input)
			if got != tt.want {
				t.Errorf("StripANSIString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStripANSI(t *testing.T) {
	input := []byte("\033[32mgreen\033[0m")
	got := StripANSI(input)
	want := []byte("green")

	if string(got) != string(want) {
		t.Errorf("StripANSI() = %q, want %q", got, want)
	}
}
