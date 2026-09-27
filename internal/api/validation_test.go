package api

import "testing"

func TestValidTerm(t *testing.T) {
	tests := []struct {
		name  string
		term  string
		valid bool
	}{
		{
			name:  "valid fall term",
			term:  "2267",
			valid: true,
		},
		{
			name:  "valid numeric term",
			term:  "2255",
			valid: true,
		},
		{
			name:  "empty",
			term:  "",
			valid: false,
		},
		{
			name:  "too short",
			term:  "267",
			valid: false,
		},
		{
			name:  "too long",
			term:  "20267",
			valid: false,
		},
		{
			name:  "contains letters",
			term:  "22A7",
			valid: false,
		},
		{
			name:  "contains punctuation",
			term:  "226-",
			valid: false,
		},
		{
			name:  "contains whitespace",
			term:  "226 ",
			valid: false,
		},
		{
			name:  "term not allowed",
			term:  "2262",
			valid: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := validTerm(test.term)

			if actual != test.valid {
				t.Fatalf(
					"validTerm(%q) = %v, expected %v",
					test.term,
					actual,
					test.valid,
				)
			}
		})
	}
}
