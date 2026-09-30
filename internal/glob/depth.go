package glob

import "strings"

// MaxDepth bounds how far a `**` pattern may walk. Without a limit a pattern
// such as `**/**/**/x` keeps expanding and a hostile path can be used to make
// the matcher do unbounded work.
const MaxDepth = 64

// Depth reports the nesting depth implied by a pattern, counting a leading
// `**` as one level per segment.
func Depth(pattern string) int {
	depth := 0
	for _, seg := range strings.Split(pattern, "/") {
		if seg == "" {
			continue
		}
		depth++
	}
	if strings.Contains(pattern, "**") && depth < 1 {
		return 1
	}
	return depth
}

// ValidatePattern rejects patterns that cannot be evaluated in bounded time.
func ValidatePattern(pattern string) error {
	if pattern == "" {
		return &PatternError{Reason: "pattern must not be empty"}
	}
	if d := Depth(pattern); d > MaxDepth {
		return &PatternError{Reason: "pattern nesting exceeds the supported depth"}
	}
	return nil
}

// PatternError reports a pattern the matcher refuses to compile.
type PatternError struct{ Reason string }

func (e *PatternError) Error() string { return "invalid pattern: " + e.Reason }
