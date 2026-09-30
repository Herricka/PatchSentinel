package policy

import (
	"math"
	"testing"
)

func TestValidateName(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"plain", "repetition", false},
		{"with space", "sensitive path", false},
		{"empty", "", true},
		{"whitespace only", "   ", true},
		{"slash", "a/b", true},
		{"backslash", `a\b`, true},
	}
	for _, c := range cases {
		err := ValidateName(c.input)
		if (err != nil) != c.wantErr {
			t.Errorf("ValidateName(%q) error = %v, wantErr %v", c.input, err, c.wantErr)
		}
	}
}

func TestValidateThreshold(t *testing.T) {
	cases := []struct {
		name    string
		value   float64
		wantErr bool
	}{
		{"mid range", 0.5, false},
		{"lower bound", 0, false},
		{"upper bound", 1, false},
		{"NaN", math.NaN(), true},
		{"infinite", math.Inf(1), true},
		{"negative", -0.1, true},
		{"above one", 1.1, true},
	}
	for _, c := range cases {
		if err := ValidateThreshold("score", c.value); (err != nil) != c.wantErr {
			t.Errorf("ValidateThreshold(%v) error = %v, wantErr %v", c.value, err, c.wantErr)
		}
	}
}

func TestClampScore(t *testing.T) {
	cases := []struct{ in, want float64 }{
		{0.4, 0.4}, {-5, 0}, {5, 1}, {math.NaN(), 1},
	}
	for _, c := range cases {
		if got := ClampScore(c.in); got != c.want {
			t.Errorf("ClampScore(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}
