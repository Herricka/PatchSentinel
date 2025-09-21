// Package analyze turns a parsed patch and a policy into a deterministic
// Report: a structured set of evidence (risk signals, repetition, test ratio,
// sensitive paths, review cost) plus a policy verdict.
//
// Determinism is a first-class requirement: given the same patch and policy,
// the JSON output is byte-for-byte identical across runs and platforms. All
// iteration happens over sorted collections and floating-point results are
// rounded to a fixed precision.
package analyze

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"

	"github.com/patchsentinel/patchsentinel/internal/diff"
	"github.com/patchsentinel/patchsentinel/internal/glob"
	"github.com/patchsentinel/patchsentinel/internal/policy"
)

// Verdict is the top-level gate decision.
type Verdict string

const (
	// Pass means the change satisfies every gate.
	Pass Verdict = "pass"
	// Warn means the change is acceptable but has notable signals.
	Warn Verdict = "warn"
	// Fail means at least one hard gate was violated.
	Fail Verdict = "fail"
)

// Report is the complete, serializable analysis result.
type Report struct {
	// SchemaVersion identifies the output format.
	SchemaVersion int `json:"schemaVersion"`
	// Tool is the producing tool name.
	Tool string `json:"tool"`
	// PolicyVersion echoes the evaluated policy version.
	PolicyVersion int `json:"policyVersion"`

	Summary  Summary    `json:"summary"`
	Files    []FileStat `json:"files"`
	Risks    []RiskHit  `json:"risks"`
	Findings []Finding  `json:"findings"`
	Verdict  Verdict    `json:"verdict"`
	ExitCode int        `json:"exitCode"`
}

// Summary aggregates headline numbers.
type Summary struct {
	FilesChanged    int     `json:"filesChanged"`
	LinesAdded      int     `json:"linesAdded"`
	LinesRemoved    int     `json:"linesRemoved"`
	ProductionAdded int     `json:"productionAdded"`
	TestAdded       int     `json:"testAdded"`
	TestToCodeRatio float64 `json:"testToCodeRatio"`
	RiskScore       int     `json:"riskScore"`
	RepetitionRatio float64 `json:"repetitionRatio"`
	SensitiveFiles  int     `json:"sensitiveFiles"`
	ReviewMinutes   float64 `json:"reviewMinutes"`
	ContentHash     string  `json:"contentHash"`
}

// FileStat is the per-file breakdown.
type FileStat struct {
	Path        string `json:"path"`
	OldPath     string `json:"oldPath,omitempty"`
	Added       int    `json:"added"`
	Removed     int    `json:"removed"`
	IsTest      bool   `json:"isTest"`
	IsSensitive bool   `json:"isSensitive"`
	IsBinary    bool   `json:"isBinary"`
	Status      string `json:"status"` // added | deleted | renamed | modified
}

// RiskHit is a single risk-pattern match with location context.
type RiskHit struct {
	PatternID string `json:"patternId"`
	Message   string `json:"message"`
	Weight    int    `json:"weight"`
	Path      string `json:"path"`
	Snippet   string `json:"snippet"`
	Count     int    `json:"count"`
}

// Severity classifies a finding.
type Severity string

const (
	SeverityInfo  Severity = "info"
	SeverityWarn  Severity = "warn"
	SeverityError Severity = "error"
)

// Finding is a human-facing gate result explaining why the verdict is what it
// is. Findings are the primary evidence surfaced to reviewers.
type Finding struct {
	ID       string   `json:"id"`
	Severity Severity `json:"severity"`
	Title    string   `json:"title"`
	Detail   string   `json:"detail"`
}

// Analyze evaluates a patch against a policy and returns a deterministic
// Report.
func Analyze(p diff.Patch, pol policy.Policy) Report {
	rep := Report{
		SchemaVersion: 1,
		Tool:          "patchsentinel",
		PolicyVersion: pol.Version,
	}

	// Compile risk patterns once. Invalid regexes are skipped defensively.
	type compiledPattern struct {
		def policy.RiskPattern
		re  *regexp.Regexp
	}
	compiled := make([]compiledPattern, 0, len(pol.RiskPatterns))
	for _, rp := range pol.RiskPatterns {
		re, err := regexp.Compile(rp.Regex)
		if err != nil {
			continue
		}
		compiled = append(compiled, compiledPattern{def: rp, re: re})
	}

	// Per-file stats and added-line collection for repetition analysis.
	var files []FileStat
	var addedProductionLines []string
	riskAgg := map[string]*RiskHit{}

	for _, f := range p.Files {
		path := f.Path()
		isTest := glob.MatchAny(pol.TestPathPatterns, path)
		isSensitive := glob.MatchAny(pol.SensitivePaths, path)

		fs := FileStat{
			Path:        path,
			Added:       f.Added(),
			Removed:     f.Removed(),
			IsTest:      isTest,
			IsSensitive: isSensitive,
			IsBinary:    f.Binary,
			Status:      status(f),
		}
		if f.IsRename {
			fs.OldPath = f.OldPath
		}
		files = append(files, fs)

		// Walk added lines for risk + repetition.
		for _, h := range f.Hunks {
			for _, l := range h.Lines {
				if l.Kind != diff.Added {
					continue
				}
				trimmed := strings.TrimSpace(l.Text)
				if !isTest && isMeaningful(trimmed) {
					addedProductionLines = append(addedProductionLines, trimmed)
				}
				for _, cp := range compiled {
					if cp.re.MatchString(l.Text) {
						key := cp.def.ID + "\x00" + path
						hit := riskAgg[key]
						if hit == nil {
							hit = &RiskHit{
								PatternID: cp.def.ID,
								Message:   cp.def.Message,
								Weight:    cp.def.Weight,
								Path:      path,
								Snippet:   snippet(trimmed),
							}
							riskAgg[key] = hit
						}
						hit.Count++
					}
				}
			}
		}
	}

	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	rep.Files = files

	// Deterministic risk list, sorted by (weight desc, patternId, path).
	risks := make([]RiskHit, 0, len(riskAgg))
	for _, h := range riskAgg {
		risks = append(risks, *h)
	}
	sort.Slice(risks, func(i, j int) bool {
		if risks[i].Weight != risks[j].Weight {
			return risks[i].Weight > risks[j].Weight
		}
		if risks[i].PatternID != risks[j].PatternID {
			return risks[i].PatternID < risks[j].PatternID
		}
		return risks[i].Path < risks[j].Path
	})
	rep.Risks = risks

	// Aggregate summary numbers.
	var sum Summary
	sum.FilesChanged = len(files)
	sum.LinesAdded = p.TotalAdded()
	sum.LinesRemoved = p.TotalRemoved()
	for _, fs := range files {
		if fs.IsTest {
			sum.TestAdded += fs.Added
		} else {
			sum.ProductionAdded += fs.Added
		}
		if fs.IsSensitive {
			sum.SensitiveFiles++
		}
	}
	sum.TestToCodeRatio = ratio(sum.TestAdded, sum.ProductionAdded)

	riskScore := 0
	for _, h := range risks {
		riskScore += h.Weight * h.Count
	}
	sum.RiskScore = riskScore

	sum.RepetitionRatio = repetitionRatio(addedProductionLines)
	sum.ReviewMinutes = reviewMinutes(files, pol.ReviewCost)
	sum.ContentHash = contentHash(p)
	rep.Summary = sum

	// Evaluate gates into findings + verdict.
	findings, verdict := evaluate(sum, risks, pol)
	rep.Findings = findings
	rep.Verdict = verdict
	rep.ExitCode = exitCode(verdict)
	return rep
}

func status(f diff.FileChange) string {
	switch {
	case f.IsNew:
		return "added"
	case f.IsDelete:
		return "deleted"
	case f.IsRename:
		return "renamed"
	default:
		return "modified"
	}
}

// isMeaningful filters out blank lines and pure delimiters so that repetition
// analysis is not dominated by braces and empty lines.
func isMeaningful(s string) bool {
	if len(s) < 4 {
		return false
	}
	switch s {
	case "{", "}", "(", ")", "[", "]", "*/", "/*":
		return false
	}
	return true
}

func snippet(s string) string {
	const max = 120
	s = strings.TrimSpace(s)
	if len(s) > max {
		return s[:max] + "…"
	}
	return s
}

// repetitionRatio estimates how much of the added production code is
// duplicated. It is a strong proxy for boilerplate / machine-generated code:
// the fraction of added meaningful lines that are not unique.
func repetitionRatio(lines []string) float64 {
	if len(lines) == 0 {
		return 0
	}
	seen := map[string]int{}
	for _, l := range lines {
		seen[l]++
	}
	duplicated := 0
	for _, c := range seen {
		if c > 1 {
			duplicated += c - 1 // count every repeat beyond the first
		}
	}
	return round(float64(duplicated) / float64(len(lines)))
}

func ratio(test, prod int) float64 {
	if prod == 0 {
		if test > 0 {
			return 1
		}
		return 0
	}
	return round(float64(test) / float64(prod))
}

// reviewMinutes is a simple, explainable effort model.
func reviewMinutes(files []FileStat, rc policy.ReviewCost) float64 {
	total := 0.0
	for _, f := range files {
		cost := float64(f.Added)*rc.MinutesPerAddedLine +
			float64(f.Removed)*rc.MinutesPerRemovedLine +
			rc.MinutesPerFile
		if f.IsSensitive {
			cost *= rc.SensitiveMultiplier
		}
		total += cost
	}
	return round(total)
}

// contentHash produces a stable fingerprint of the change content, useful for
// caching and for asserting determinism in tests.
func contentHash(p diff.Patch) string {
	h := sha256.New()
	// Iterate over a sorted copy of paths for stability.
	type entry struct {
		path  string
		lines []string
	}
	entries := make([]entry, 0, len(p.Files))
	for _, f := range p.Files {
		var ls []string
		for _, hunk := range f.Hunks {
			for _, l := range hunk.Lines {
				prefix := " "
				switch l.Kind {
				case diff.Added:
					prefix = "+"
				case diff.Removed:
					prefix = "-"
				}
				ls = append(ls, prefix+l.Text)
			}
		}
		entries = append(entries, entry{path: f.Path(), lines: ls})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].path < entries[j].path })
	for _, e := range entries {
		h.Write([]byte(e.path))
		h.Write([]byte{0})
		for _, l := range e.lines {
			h.Write([]byte(l))
			h.Write([]byte{0})
		}
	}
	return hex.EncodeToString(h.Sum(nil))
}

// evaluate applies the policy thresholds, producing sorted findings and the
// final verdict. Verdict escalation: any error → Fail; else any warn → Warn.
func evaluate(sum Summary, risks []RiskHit, pol policy.Policy) ([]Finding, Verdict) {
	var findings []Finding
	worst := Pass

	escalate := func(sev Severity) {
		switch sev {
		case SeverityError:
			worst = Fail
		case SeverityWarn:
			if worst != Fail {
				worst = Warn
			}
		}
	}

	add := func(id string, sev Severity, title, detail string) {
		findings = append(findings, Finding{ID: id, Severity: sev, Title: title, Detail: detail})
		escalate(sev)
	}

	// Risk score gate.
	switch {
	case sum.RiskScore >= pol.Thresholds.MaxRiskScore:
		add("risk-score", SeverityError, "Risk score exceeds maximum",
			fmt.Sprintf("Cumulative risk score %d ≥ max %d.", sum.RiskScore, pol.Thresholds.MaxRiskScore))
	case sum.RiskScore >= pol.Thresholds.WarnRiskScore:
		add("risk-score", SeverityWarn, "Elevated risk score",
			fmt.Sprintf("Cumulative risk score %d ≥ warn threshold %d.", sum.RiskScore, pol.Thresholds.WarnRiskScore))
	}

	// Test-to-code ratio gate (only for non-trivial production changes).
	if sum.ProductionAdded >= pol.Thresholds.MinProductionLinesForTestCheck {
		if sum.TestToCodeRatio < pol.Thresholds.MinTestRatio {
			add("test-ratio", SeverityWarn, "Insufficient test coverage for change size",
				fmt.Sprintf("Test-to-code ratio %.2f < required %.2f for %d added production lines.",
					sum.TestToCodeRatio, pol.Thresholds.MinTestRatio, sum.ProductionAdded))
		}
	}

	// Repetition / generated-code gate.
	if sum.RepetitionRatio >= pol.Thresholds.MaxRepetitionRatio {
		add("repetition", SeverityWarn, "High repetition in added code",
			fmt.Sprintf("Repetition ratio %.2f ≥ %.2f suggests boilerplate or generated code that needs closer review.",
				sum.RepetitionRatio, pol.Thresholds.MaxRepetitionRatio))
	}

	// Sensitive path gate.
	if sum.SensitiveFiles > 0 {
		sev := SeverityError
		if pol.Thresholds.SensitivePathIsWarning {
			sev = SeverityWarn
		}
		add("sensitive-path", sev, "Sensitive paths modified",
			fmt.Sprintf("%d sensitive file(s) changed; require domain-owner review.", sum.SensitiveFiles))
	}

	// Review cost gate.
	if int(math.Ceil(sum.ReviewMinutes)) > pol.Thresholds.MaxReviewMinutes {
		add("review-cost", SeverityWarn, "Large change increases review cost",
			fmt.Sprintf("Estimated review time %.0f min > budget %d min; consider splitting the change.",
				sum.ReviewMinutes, pol.Thresholds.MaxReviewMinutes))
	}

	if len(findings) == 0 {
		add("clean", SeverityInfo, "No policy violations detected",
			"All configured gates passed.")
	}

	sort.SliceStable(findings, func(i, j int) bool {
		return severityRank(findings[i].Severity) < severityRank(findings[j].Severity)
	})
	return findings, worst
}

func severityRank(s Severity) int {
	switch s {
	case SeverityError:
		return 0
	case SeverityWarn:
		return 1
	default:
		return 2
	}
}

func exitCode(v Verdict) int {
	switch v {
	case Fail:
		return 2
	case Warn:
		return 1
	default:
		return 0
	}
}

// round normalizes floats to 4 decimal places for deterministic JSON.
func round(f float64) float64 {
	return math.Round(f*10000) / 10000
}

// refactor: rename internal/diff for clarity

// test: regression for the scripts overflow

// feat: add internal/analyze handling

// fix: clamp internal/policy to the supported range

// docs: document internal/analyze

// test: cover internal/glob edge cases

// fix: guard against malformed internal/policy

// perf: reduce allocations in scripts

// feat: cache parsed internal/render

// docs: document src

// refactor: rename benches for clarity

// chore: bump deps for benches

// fix: clamp internal/policy to the supported range

// refactor: split internal/analyze out of the hot path

// refactor: split scripts out of the hot path

// feat: expose internal/render as a library API

// chore: tidy docs

// feat: cache parsed internal/render

// docs: add an example for docs

// refactor: rename internal/glob for clarity

// refactor: split src out of the hot path

// test: cover benches edge cases

// refactor: split scripts out of the hot path

// feat: add internal/render handling

// fix: handle empty benches

// refactor: split benches out of the hot path

// refactor: rename internal/glob for clarity

// chore: tidy docs

// refactor: split benches out of the hot path

// perf: reduce allocations in cmd/patchsentinel
