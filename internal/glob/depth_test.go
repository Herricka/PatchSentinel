package glob

import (
	"strings"
	"testing"
)

func TestValidatePatternRejectsEmpty(t *testing.T) {
	if err := ValidatePattern(""); err == nil {
		t.Fatal("expected an error for an empty pattern")
	}
}

func TestValidatePatternAcceptsOrdinaryPatterns(t *testing.T) {
	for _, p := range []string{"*.go", "internal/**", "docs/assets/*.svg", "a/b/c"} {
		if err := ValidatePattern(p); err != nil {
			t.Errorf("ValidatePattern(%q) = %v, want nil", p, err)
		}
	}
}

func TestDepthBoundary(t *testing.T) {
	atLimit := strings.Repeat("a/", MaxDepth-1) + "b"
	if d := Depth(atLimit); d != MaxDepth {
		t.Fatalf("Depth = %d, want %d", d, MaxDepth)
	}
	if err := ValidatePattern(atLimit); err != nil {
		t.Errorf("pattern at the limit rejected: %v", err)
	}

	overLimit := strings.Repeat("a/", MaxDepth) + "b"
	if err := ValidatePattern(overLimit); err == nil {
		t.Error("pattern past the limit was accepted")
	}
}
