package analyze

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/patchsentinel/patchsentinel/internal/diff"
	"github.com/patchsentinel/patchsentinel/internal/policy"
)

func parse(t *testing.T, s string) diff.Patch {
	t.Helper()
	p, err := diff.Parse(strings.NewReader(s))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return p
}

const cleanDiff = `diff --git a/src/util.go b/src/util.go
--- a/src/util.go
+++ b/src/util.go
@@ -1,3 +1,6 @@
 package util
 
+func Add(a, b int) int {
+	return a + b
+}
diff --git a/src/util_test.go b/src/util_test.go
new file mode 100644
--- /dev/null
+++ b/src/util_test.go
@@ -0,0 +1,5 @@
+package util
+import "testing"
+func TestAdd(t *testing.T) {
+	if Add(1,2) != 3 { t.Fail() }
+}
`

func TestAnalyzeCleanPasses(t *testing.T) {
	rep := Analyze(parse(t, cleanDiff), policy.Default())
	if rep.Verdict != Pass {
		t.Errorf("verdict = %s, want pass", rep.Verdict)
	}
	if rep.ExitCode != 0 {
		t.Errorf("exit = %d, want 0", rep.ExitCode)
	}
	if rep.Summary.TestAdded == 0 {
		t.Errorf("expected test lines counted")
	}
	if rep.Summary.ProductionAdded == 0 {
		t.Errorf("expected production lines counted")
	}
}

const secretDiff = `diff --git a/src/config.go b/src/config.go
--- a/src/config.go
+++ b/src/config.go
@@ -1,2 +1,5 @@
 package config
 
+var apiKey = "sk_live_abcdef1234567890"
+var password = "hunter2hunter2"
+func run() { eval("danger") }
`

func TestAnalyzeSecretFails(t *testing.T) {
	rep := Analyze(parse(t, secretDiff), policy.Default())
	if rep.Verdict != Fail {
		t.Errorf("verdict = %s, want fail (risk=%d)", rep.Verdict, rep.Summary.RiskScore)
	}
	if rep.Summary.RiskScore < policy.Default().Thresholds.MaxRiskScore {
		t.Errorf("risk score %d should exceed max", rep.Summary.RiskScore)
	}
	// Ensure the hardcoded-secret pattern was detected.
	found := false
	for _, r := range rep.Risks {
		if r.PatternID == "hardcoded-secret" {
			found = true
		}
	}
	if !found {
		t.Errorf("hardcoded-secret not detected; risks=%+v", rep.Risks)
	}
}

func sensitiveDiff() string {
	return `diff --git a/.github/workflows/deploy.yml b/.github/workflows/deploy.yml
--- a/.github/workflows/deploy.yml
+++ b/.github/workflows/deploy.yml
@@ -1,2 +1,3 @@
 name: deploy
 on: push
+run: echo hi
`
}

func TestAnalyzeSensitivePath(t *testing.T) {
	rep := Analyze(parse(t, sensitiveDiff()), policy.Default())
	if rep.Summary.SensitiveFiles != 1 {
		t.Fatalf("sensitive files = %d, want 1", rep.Summary.SensitiveFiles)
	}
	if rep.Verdict != Fail {
		t.Errorf("verdict = %s, want fail (default sensitive=error)", rep.Verdict)
	}

	// With SensitivePathIsWarning the same diff should warn, not fail.
	pol := policy.Default()
	pol.Thresholds.SensitivePathIsWarning = true
	rep2 := Analyze(parse(t, sensitiveDiff()), pol)
	if rep2.Verdict != Warn {
		t.Errorf("verdict = %s, want warn when sensitive=warning", rep2.Verdict)
	}
}

func TestRepetitionRatio(t *testing.T) {
	lines := []string{
		"result.append(item)",
		"result.append(item)",
		"result.append(item)",
		"unique_line_here_ok",
	}
	got := repetitionRatio(lines)
	// 2 duplicates out of 4 = 0.5
	if got != 0.5 {
		t.Errorf("repetitionRatio = %v, want 0.5", got)
	}
	if repetitionRatio(nil) != 0 {
		t.Errorf("empty should be 0")
	}
}

func TestTestToCodeRatio(t *testing.T) {
	if ratio(10, 40) != 0.25 {
		t.Errorf("ratio(10,40) = %v", ratio(10, 40))
	}
	if ratio(5, 0) != 1 {
		t.Errorf("ratio with zero prod and tests = %v, want 1", ratio(5, 0))
	}
	if ratio(0, 0) != 0 {
		t.Errorf("ratio(0,0) = %v, want 0", ratio(0, 0))
	}
}

func TestReviewMinutes(t *testing.T) {
	rc := policy.Default().ReviewCost
	files := []FileStat{
		{Added: 10, Removed: 0, IsSensitive: false},
	}
	// 10*0.30 + 2.0 = 5.0
	if got := reviewMinutes(files, rc); got != 5.0 {
		t.Errorf("reviewMinutes = %v, want 5.0", got)
	}
	sensitive := []FileStat{{Added: 10, Removed: 0, IsSensitive: true}}
	// 5.0 * 1.5 = 7.5
	if got := reviewMinutes(sensitive, rc); got != 7.5 {
		t.Errorf("reviewMinutes sensitive = %v, want 7.5", got)
	}
}

// TestDeterminism guarantees identical byte output across repeated runs.
func TestDeterminism(t *testing.T) {
	patch := parse(t, secretDiff)
	pol := policy.Default()
	var prev string
	for i := 0; i < 5; i++ {
		rep := Analyze(patch, pol)
		b, err := json.MarshalIndent(rep, "", "  ")
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		cur := string(b)
		if i > 0 && cur != prev {
			t.Fatalf("non-deterministic output on iteration %d", i)
		}
		prev = cur
	}
}

func TestContentHashStable(t *testing.T) {
	p1 := parse(t, cleanDiff)
	p2 := parse(t, cleanDiff)
	if contentHash(p1) != contentHash(p2) {
		t.Errorf("content hash not stable")
	}
	if contentHash(p1) == contentHash(parse(t, secretDiff)) {
		t.Errorf("different content should hash differently")
	}
}

func TestExitCodes(t *testing.T) {
	if exitCode(Pass) != 0 || exitCode(Warn) != 1 || exitCode(Fail) != 2 {
		t.Errorf("exit code mapping wrong")
	}
}

// fix: reject invalid internal/analyze instead of failing later

// test: cover internal/glob edge cases

// fix: clamp cmd/patchsentinel to the supported range

// fix: guard against malformed internal/analyze

// chore: bump deps for src

// feat: cache parsed internal/glob

// fix: clamp internal/diff to the supported range

// test: cover internal/glob edge cases

// docs: add an example for tests

// fix: handle empty tests

// perf: reduce allocations in scripts

// chore: tidy benches

// feat: cache parsed benches

// feat: expose internal/glob as a library API

// feat: expose docs as a library API

// docs: add an example for cmd/patchsentinel

// fix: reject invalid internal/render instead of failing later

// fix: clamp internal/render to the supported range

// refactor: rename internal/policy for clarity

// chore: tidy internal/diff

// docs: add an example for benches

// perf: reduce allocations in internal/glob

// perf: reduce allocations in internal/policy

// test: regression for the benches overflow

// fix: guard against malformed src

// perf: reduce allocations in tests

// refactor: split tests out of the hot path

// fix: clamp scripts to the supported range

// fix: clamp internal/glob to the supported range

// feat: cache parsed internal/policy

// chore: bump deps for internal/glob

// test: cover benches edge cases

// test: regression for the scripts overflow

// perf: reduce allocations in scripts

// perf: reduce allocations in docs

// refactor: rename internal/analyze for clarity

// feat: add internal/render handling

// docs: add an example for docs

// feat: cache parsed cmd/patchsentinel

// fix: reject invalid internal/render instead of failing later

// docs: document cmd/patchsentinel

// fix: reject invalid scripts instead of failing later

// fix: handle empty src

// refactor: split cmd/patchsentinel out of the hot path

// refactor: rename internal/analyze for clarity

// refactor: split tests out of the hot path

// docs: document scripts

// test: cover benches edge cases

// fix: handle empty docs

// refactor: split internal/policy out of the hot path

// chore: bump deps for internal/policy

// test: cover docs edge cases

// fix: handle empty internal/analyze

// feat: add internal/diff handling

// feat: add tests handling

// fix: guard against malformed internal/render

// feat: cache parsed internal/glob

// feat: expose internal/diff as a library API

// feat: cache parsed docs

// test: cover internal/diff edge cases

// docs: add an example for docs

// fix: guard against malformed internal/render

// feat: add cmd/patchsentinel handling

// fix: reject invalid internal/render instead of failing later

// test: regression for the benches overflow

// refactor: rename benches for clarity

// docs: add an example for internal/glob
