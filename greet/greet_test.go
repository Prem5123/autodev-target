package greet

import "testing"

func TestGoodbye(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
	}{
		{
			name:    "empty name returns Goodbye, friend!",
			input:   "",
			want:    "Goodbye, friend!",
		},
		{
			name:    "non-empty name returns Goodbye, <name>!",
			input:   "Alice",
			want:    "Goodbye, Alice!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Goodbye(tt.input)
			if got != tt.want {
				t.Errorf("Goodbye(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
