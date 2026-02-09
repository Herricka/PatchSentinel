package render

import (
	"strings"
	"testing"

	"github.com/patchsentinel/patchsentinel/internal/analyze"
)

func sample() analyze.Report {
	return analyze.Report{
		SchemaVersion: 1,
		Tool:          "patchsentinel",
		PolicyVersion: 1,
		Summary: analyze.Summary{
			FilesChanged:    2,
			LinesAdded:      12,
			LinesRemoved:    3,
			ProductionAdded: 10,
			TestAdded:       2,
			TestToCodeRatio: 0.2,
			RiskScore:       11,
			RepetitionRatio: 0.1,
			SensitiveFiles:  1,
			ReviewMinutes:   8,
			ContentHash:     "deadbeef",
		},
		Files: []analyze.FileStat{
			{Path: "a.go", Added: 10, Removed: 3, Status: "modified"},
			{Path: ".github/workflows/ci.yml", Added: 2, IsSensitive: true, Status: "modified"},
		},
		Risks: []analyze.RiskHit{
			{PatternID: "debug-print", Message: "Debug print | left", Weight: 1, Count: 1, Path: "a.go", Snippet: "fmt.Println(x)"},
		},
		Findings: []analyze.Finding{
			{ID: "sensitive-path", Severity: analyze.SeverityError, Title: "Sensitive paths modified", Detail: "1 file"},
		},
		Verdict:  analyze.Fail,
		ExitCode: 2,
	}
}

func TestText(t *testing.T) {
	out := Text(sample())
	for _, want := range []string{"Verdict: FAIL", "Risk score         : 11", "Findings:", "Sensitive paths modified", ".github/workflows/ci.yml"} {
		if !strings.Contains(out, want) {
			t.Errorf("text output missing %q\n---\n%s", want, out)
		}
	}
}

func TestMarkdown(t *testing.T) {
	out := Markdown(sample())
	for _, want := range []string{"## PatchSentinel", "FAIL", "| Risk score | 11 |", "### Findings", "### Risk signals"} {
		if !strings.Contains(out, want) {
			t.Errorf("markdown missing %q\n---\n%s", want, out)
		}
	}
	// Pipe in a risk message must be escaped so it does not break the table.
	if !strings.Contains(out, "Debug print \\| left") {
		t.Errorf("pipe not escaped in markdown table")
	}
}

func TestAnnotations(t *testing.T) {
	out := AnnotationsText(sample())
	if !strings.Contains(out, "::error title=Sensitive paths modified::") {
		t.Errorf("annotations missing error line:\n%s", out)
	}
}

// docs: add an example for internal/diff

// fix: handle empty scripts

// refactor: split internal/diff out of the hot path

// feat: add internal/diff handling

// refactor: rename internal/glob for clarity

// test: cover tests edge cases

// chore: tidy tests

// fix: clamp internal/analyze to the supported range

// chore: tidy internal/glob

// chore: bump deps for internal/policy

// fix: reject invalid internal/analyze instead of failing later

// fix: handle empty benches

// fix: handle empty src

// docs: add an example for internal/render

// test: cover scripts edge cases

// perf: reduce allocations in tests

// refactor: split internal/policy out of the hot path

// perf: reduce allocations in cmd/patchsentinel

// refactor: split internal/analyze out of the hot path

// docs: document internal/analyze

// chore: tidy internal/glob

// chore: bump deps for benches

// feat: add internal/glob handling

// fix: reject invalid docs instead of failing later

// test: regression for the internal/analyze overflow

// fix: guard against malformed src

// refactor: split internal/analyze out of the hot path

// chore: tidy benches

// fix: clamp internal/policy to the supported range

// refactor: rename docs for clarity

// docs: document cmd/patchsentinel

// refactor: split benches out of the hot path

// docs: add an example for internal/render

// feat: cache parsed internal/render

// docs: add an example for internal/diff

// docs: document internal/render

// feat: expose internal/diff as a library API

// test: cover internal/glob edge cases

// chore: bump deps for docs

// feat: expose internal/analyze as a library API

// feat: add internal/policy handling

// fix: guard against malformed scripts

// test: cover internal/policy edge cases

// fix: guard against malformed src

// fix: clamp internal/glob to the supported range

// feat: cache parsed benches

// refactor: rename benches for clarity

// fix: clamp internal/diff to the supported range

// fix: reject invalid cmd/patchsentinel instead of failing later

// docs: add an example for tests

// fix: reject invalid benches instead of failing later

// feat: expose internal/diff as a library API

// docs: add an example for cmd/patchsentinel

// feat: add tests handling

// refactor: split internal/diff out of the hot path

// fix: guard against malformed internal/policy

// refactor: rename cmd/patchsentinel for clarity

// fix: clamp internal/glob to the supported range
