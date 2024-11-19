package database

import "testing"

func TestCompareWordsInString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		want     bool
	}{
		{
			name:     "equal",
			input:    "hello world",
			expected: "hello world",
			want:     true,
		},
		{
			name:     "not equal",
			input:    "hello world",
			expected: "hello",
			want:     false,
		},
		{
			name:     "equal with special characters",
			input:    "hello, world",
			expected: "hello world",
			want:     true,
		},
		{
			name:     "not equal with special characters",
			input:    "hello, world",
			expected: "hello",
			want:     false,
		},
		{
			name:     "equal with special characters and spaces",
			input:    "hello, world",
			expected: "hello world",
			want:     true,
		},
		{
			name:     "not equal with special characters and spaces",
			input:    "hello, world",
			expected: "hello",
			want:     false,
		},
		{
			name:     "equal with special characters and spaces",
			input:    "hello, world",
			expected: "hello world",
			want:     true,
		},
		{
			name:     "not equal with special characters and spaces",
			input:    "hello, 'world'",
			expected: "hello",
			want:     false,
		},
		{
			name:     "equal with special characters and spaces",
			input:    `hello, "world"`,
			expected: "hello world",
			want:     true,
		},
		{
			name:     "not equal with special characters and spaces",
			input:    "hello, world",
			expected: "hello",
			want:     false,
		},
		{
			name:     "migration string",
			input:    `pq: column "has_content" of relation "bookmark" already exists`,
			expected: "pq column has_content of relation bookmark already exists",
			want:     true,
		},
		{
			name:     "migration string",
			input:    `pq: column »has_content« of relation »bookmark« already exists`,
			expected: "pq column has_content of relation bookmark already exists",
			want:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := compareWordsInString(tt.input, tt.expected); got != tt.want {
				t.Errorf("compareWordsInString() = %v, want %v", got, tt.want)
			}
		})
	}
}
