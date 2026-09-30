package policy

import (
	"fmt"
	"strings"
)

// Gate thresholds below these values make the gate meaningless (it would pass
// everything) or unusable (it would fail everything), so they are rejected at
// load time rather than silently accepted.
const (
	MinScore = 0.0
	MaxScore = 1.0
)

// InvalidRuleError reports a rule that cannot be evaluated as written.
type InvalidRuleError struct {
	Field  string
	Reason string
}

func (e *InvalidRuleError) Error() string {
	return fmt.Sprintf("invalid policy rule: %s: %s", e.Field, e.Reason)
}

// ValidateName rejects names that cannot round-trip through the CLI or the
// renderer: empty, whitespace-only, or containing path separators.
func ValidateName(name string) error {
	if strings.TrimSpace(name) == "" {
		return &InvalidRuleError{Field: "name", Reason: "must not be empty"}
	}
	if strings.ContainsAny(name, "/\\") {
		return &InvalidRuleError{Field: "name", Reason: "must not contain a path separator"}
	}
	return nil
}

// ValidateThreshold rejects NaN, infinities and out-of-range scores. A NaN
// threshold compares false against everything, so the gate would silently pass
// every diff instead of failing loudly.
func ValidateThreshold(field string, value float64) error {
	if value != value { // NaN
		return &InvalidRuleError{Field: field, Reason: "must be a number, got NaN"}
	}
	if value < MinScore || value > MaxScore {
		return &InvalidRuleError{Field: field, Reason: fmt.Sprintf("must be within [%v, %v]", MinScore, MaxScore)}
	}
	return nil
}

// ClampScore maps any input onto the valid score range, treating NaN as the
// maximum so a broken computation surfaces as a failing gate rather than a
// silent pass.
func ClampScore(value float64) float64 {
	if value != value {
		return MaxScore
	}
	if value < MinScore {
		return MinScore
	}
	if value > MaxScore {
		return MaxScore
	}
	return value
}
