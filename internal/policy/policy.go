// Package policy defines the configurable rules PatchSentinel evaluates a
// patch against, plus a deterministic loader with sensible built-in defaults.
package policy

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
)

// Policy is the full rule set. All thresholds are inclusive unless noted.
type Policy struct {
	// Version identifies the policy schema for forward compatibility.
	Version int `json:"version"`

	// SensitivePaths are glob patterns (see internal/glob) that mark files
	// requiring extra scrutiny (auth, infra, CI, secrets, etc.).
	SensitivePaths []string `json:"sensitivePaths"`

	// RiskPatterns are regular expressions applied to added lines. Each match
	// contributes its Weight to the risk score.
	RiskPatterns []RiskPattern `json:"riskPatterns"`

	// TestPathPatterns identify files that count as tests for the
	// test-to-code ratio.
	TestPathPatterns []string `json:"testPathPatterns"`

	// Thresholds control verdict computation.
	Thresholds Thresholds `json:"thresholds"`

	// ReviewCost tunes the review-effort model.
	ReviewCost ReviewCost `json:"reviewCost"`
}

// RiskPattern is a named regular expression with a severity weight.
type RiskPattern struct {
	ID      string `json:"id"`
	Regex   string `json:"regex"`
	Message string `json:"message"`
	Weight  int    `json:"weight"`
}

// Thresholds gate the overall verdict.
type Thresholds struct {
	// MaxRiskScore is the highest tolerated cumulative risk score before the
	// verdict fails. Exceeding it is a hard fail.
	MaxRiskScore int `json:"maxRiskScore"`
	// WarnRiskScore raises a warning (not a failure) once reached.
	WarnRiskScore int `json:"warnRiskScore"`
	// MinTestRatio is the minimum acceptable test-to-code line ratio when the
	// change touches non-trivial production code.
	MinTestRatio float64 `json:"minTestRatio"`
	// MinProductionLinesForTestCheck disables the test-ratio gate for tiny
	// changes below this many added production lines.
	MinProductionLinesForTestCheck int `json:"minProductionLinesForTestCheck"`
	// MaxRepetitionRatio flags likely generated/boilerplate code when the
	// fraction of duplicated added lines meets or exceeds this value.
	MaxRepetitionRatio float64 `json:"maxRepetitionRatio"`
	// MaxReviewMinutes fails changes whose estimated review time is too high.
	MaxReviewMinutes int `json:"maxReviewMinutes"`
	// SensitivePathIsWarning downgrades sensitive-path hits from fail to warn.
	SensitivePathIsWarning bool `json:"sensitivePathIsWarning"`
}

// ReviewCost parameters feed a simple, explainable effort model.
type ReviewCost struct {
	// MinutesPerAddedLine is the base cost of reading one new line.
	MinutesPerAddedLine float64 `json:"minutesPerAddedLine"`
	// MinutesPerRemovedLine is the (lower) cost of validating a deletion.
	MinutesPerRemovedLine float64 `json:"minutesPerRemovedLine"`
	// MinutesPerFile accounts for context-switching between files.
	MinutesPerFile float64 `json:"minutesPerFile"`
	// SensitiveMultiplier scales the cost of touching sensitive files.
	SensitiveMultiplier float64 `json:"sensitiveMultiplier"`
}

// Default returns the built-in policy. It is intentionally opinionated but
// conservative, so that PatchSentinel is useful out of the box.
func Default() Policy {
	return Policy{
		Version: 1,
		SensitivePaths: []string{
			"**/.github/workflows/**",
			"**/*Dockerfile*",
			"**/Dockerfile",
			"**/*.tf",
			"**/*.tfvars",
			"**/k8s/**",
			"**/kubernetes/**",
			"**/helm/**",
			"**/auth/**",
			"**/*auth*",
			"**/security/**",
			"**/*secret*",
			"**/*credential*",
			"**/.env*",
			"**/migrations/**",
			"**/*.pem",
			"**/*.key",
		},
		TestPathPatterns: []string{
			"**/*_test.go",
			"**/*.test.ts",
			"**/*.test.js",
			"**/*.spec.ts",
			"**/*.spec.js",
			"**/test/**",
			"**/tests/**",
			"**/__tests__/**",
			"**/*_test.py",
			"**/test_*.py",
		},
		RiskPatterns: []RiskPattern{
			{ID: "eval-usage", Regex: `\beval\s*\(`, Message: "Dynamic code evaluation via eval()", Weight: 8},
			{ID: "shell-exec", Regex: `\b(exec\.Command|child_process|os\.system|subprocess\.)`, Message: "Executes external processes", Weight: 6},
			{ID: "hardcoded-secret", Regex: `(?i)(api[_-]?key|secret|password|token)\s*[:=]\s*["'][^"']{8,}["']`, Message: "Possible hardcoded secret", Weight: 10},
			{ID: "sql-format", Regex: `(?i)(SELECT|INSERT|UPDATE|DELETE)\b.*(\+|%s|\$\{|format\()`, Message: "Possible SQL built via string concatenation", Weight: 7},
			{ID: "disable-tls", Regex: `(?i)(InsecureSkipVerify\s*:\s*true|rejectUnauthorized\s*:\s*false|verify\s*=\s*False)`, Message: "TLS verification disabled", Weight: 9},
			{ID: "broad-except", Regex: `except\s*:\s*$|catch\s*\(\s*\)\s*\{`, Message: "Overly broad exception handling", Weight: 3},
			{ID: "debug-print", Regex: `\b(console\.log|fmt\.Println|print\()`, Message: "Debug print left in code", Weight: 1},
			{ID: "todo-marker", Regex: `(?i)\b(TODO|FIXME|XXX|HACK)\b`, Message: "Unfinished work marker", Weight: 2},
			{ID: "wildcard-cors", Regex: `Access-Control-Allow-Origin["']?\s*[:,]\s*["']\*`, Message: "Wildcard CORS origin", Weight: 5},
			{ID: "unsafe-yaml", Regex: `yaml\.Unmarshal|yaml\.load\s*\(`, Message: "Potentially unsafe deserialization", Weight: 4},
		},
		Thresholds: Thresholds{
			MaxRiskScore:                   25,
			WarnRiskScore:                  10,
			MinTestRatio:                   0.25,
			MinProductionLinesForTestCheck: 40,
			MaxRepetitionRatio:             0.60,
			MaxReviewMinutes:               90,
			SensitivePathIsWarning:         false,
		},
		ReviewCost: ReviewCost{
			MinutesPerAddedLine:   0.30,
			MinutesPerRemovedLine: 0.10,
			MinutesPerFile:        2.0,
			SensitiveMultiplier:   1.5,
		},
	}
}

// Load reads a JSON policy from r, overlaying it on top of the defaults so
// that partial policies are valid. Unknown fields are rejected to catch typos.
func Load(r io.Reader) (Policy, error) {
	p := Default()
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		return Policy{}, fmt.Errorf("decode policy: %w", err)
	}
	if p.Version == 0 {
		p.Version = 1
	}
	normalize(&p)
	return p, nil
}

// normalize sorts slice fields so that policy evaluation and serialization are
// stable regardless of source ordering.
func normalize(p *Policy) {
	sort.Strings(p.SensitivePaths)
	sort.Strings(p.TestPathPatterns)
	sort.Slice(p.RiskPatterns, func(i, j int) bool {
		return p.RiskPatterns[i].ID < p.RiskPatterns[j].ID
	})
}

// test: regression for the internal/diff overflow

// test: regression for the benches overflow

// feat: add internal/diff handling

// feat: expose internal/diff as a library API

// test: cover docs edge cases

// fix: clamp src to the supported range

// test: cover tests edge cases

// docs: add an example for scripts

// test: cover benches edge cases

// test: regression for the internal/analyze overflow

// feat: cache parsed docs

// refactor: split docs out of the hot path

// fix: guard against malformed internal/policy

// feat: expose cmd/patchsentinel as a library API

// fix: handle empty benches

// fix: clamp scripts to the supported range

// test: cover cmd/patchsentinel edge cases

// fix: reject invalid internal/policy instead of failing later

// build(deps): bump github.com/spf13/cobra

// test: cover scripts edge cases

// refactor: split docs out of the hot path

// perf: reduce allocations in cmd/patchsentinel

// refactor: rename benches for clarity

// docs: add an example for cmd/patchsentinel

// fix: reject invalid src instead of failing later

// test: cover internal/diff edge cases

// fix: handle empty internal/glob

// docs: document docs

// refactor: split internal/analyze out of the hot path

// docs: document internal/diff

// feat: expose internal/policy as a library API

// fix: guard against malformed cmd/patchsentinel

// chore: bump deps for scripts

// feat: expose internal/render as a library API

// feat: expose docs as a library API

// test: regression for the internal/analyze overflow

// fix: reject invalid scripts instead of failing later

// fix: guard against malformed internal/diff

// feat: add tests handling

// perf: reduce allocations in internal/policy

// feat: cache parsed cmd/patchsentinel

// feat: expose internal/diff as a library API

// fix: guard against malformed internal/render

// chore: tidy scripts

// docs: document cmd/patchsentinel
