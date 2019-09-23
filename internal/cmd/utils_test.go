package cmd

import (
	"reflect"
	"testing"
)

func Test_normalizeSpace(t *testing.T) {
	tests := []struct {
		name string
		args string
		want string
	}{{
		name: "normal sentence",
		args: "What a perfect, beautiful sentence",
		want: "What a perfect, beautiful sentence",
	}, {
		name: "has unnecessary space before and after sentence",
		args: "    I'm surrounded with spaces    ",
		want: "I'm surrounded with spaces",
	}, {
		name: "has unnecessary spaces in middle of sentence",
		args: "I'm hollow         inside",
		want: "I'm hollow inside",
	}, {
		name: "has unnecessary new line in middle of sentence",
		args: "I'm broken \n\n\ninside",
		want: "I'm broken inside",
	}, {
		name: "has unnecessary new line and spaces everywhere",
		args: "    I'm hollow     broken\n\n\n\nand surrounded by spaces    ",
		want: "I'm hollow broken and surrounded by spaces",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeSpace(tt.args); got != tt.want {
				t.Errorf("normalizeSpace() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isURLValid(t *testing.T) {
	tests := []struct {
		name string
		args string
		want bool
	}{{
		name: "valid URL",
		args: "https://www.google.com",
		want: true,
	}, {
		name: "valid localhost URL",
		args: "http://localhost:8080",
		want: true,
	}, {
		name: "valid non-HTTP URL",
		args: "ftp://www.example.com/storage",
		want: true,
	}, {
		name: "invalid URL",
		args: "https:/www.google.com",
		want: false,
	}, {
		name: "hash URL",
		args: "#some-awesome-heading",
		want: false,
	}, {
		name: "relative URL",
		args: "/page/contact",
		want: false,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isURLValid(tt.args); got != tt.want {
				t.Errorf("isURLValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseStrIndices(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    []int
		wantErr bool
	}{{
		name:    "single number",
		args:    []string{"1"},
		want:    []int{1},
		wantErr: false,
	}, {
		name:    "multiple number",
		args:    []string{"1", "2", "3"},
		want:    []int{1, 2, 3},
		wantErr: false,
	}, {
		name:    "single ranged number",
		args:    []string{"1-5"},
		want:    []int{1, 2, 3, 4, 5},
		wantErr: false,
	}, {
		name:    "multiple ranged number",
		args:    []string{"1-5", "8-9"},
		want:    []int{1, 2, 3, 4, 5, 8, 9},
		wantErr: false,
	}, {
		name:    "mixed single and ranged number",
		args:    []string{"1-5", "8-9", "11", "12"},
		want:    []int{1, 2, 3, 4, 5, 8, 9, 11, 12},
		wantErr: false,
	}, {
		name:    "invalid number",
		args:    []string{"AAA"},
		want:    nil,
		wantErr: true,
	}, {
		name:    "mixed number and string",
		args:    []string{"1", "2", "A"},
		want:    nil,
		wantErr: true,
	}, {
		name:    "reversed ranged number",
		args:    []string{"5-1"},
		want:    nil,
		wantErr: true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseStrIndices(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseStrIndices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseStrIndices() = %v, want %v", got, tt.want)
			}
		})
	}
}
