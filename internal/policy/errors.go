package policy

import "fmt"

// RuleError attaches the name of the rule that produced a failure so a failing
// gate can be traced back to a single line of the policy file.
type RuleError struct {
	Rule string
	Err  error
}

func (e *RuleError) Error() string {
	if e.Rule == "" {
		return e.Err.Error()
	}
	return fmt.Sprintf("policy: rule %q is invalid: %s", e.Rule, e.Err.Error())
}

// Unwrap keeps errors.Is/errors.As working against the underlying cause.
func (e *RuleError) Unwrap() error { return e.Err }

// WithRule annotates err with the rule that produced it.
func WithRule(rule string, err error) error {
	if err == nil {
		return nil
	}
	return &RuleError{Rule: rule, Err: err}
}
