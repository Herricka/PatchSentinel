package main

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o644)
}

const passDiff = `diff --git a/src/util.go b/src/util.go
--- a/src/util.go
+++ b/src/util.go
@@ -1,2 +1,4 @@
 package util
 
+func Ok() bool { return true }
`

func TestRunJSONPass(t *testing.T) {
	var out, errBuf bytes.Buffer
	code := run([]string{"-format", "json"}, strings.NewReader(passDiff), &out, &errBuf)
	if code != 0 {
		t.Fatalf("exit = %d, stderr=%s", code, errBuf.String())
	}
	var parsed map[string]any
	if err := json.Unmarshal(out.Bytes(), &parsed); err != nil {
		t.Fatalf("output not valid json: %v\n%s", err, out.String())
	}
	if parsed["verdict"] != "pass" {
		t.Errorf("verdict = %v, want pass", parsed["verdict"])
	}
}

func TestRunVersion(t *testing.T) {
	var out, errBuf bytes.Buffer
	code := run([]string{"-version"}, strings.NewReader(""), &out, &errBuf)
	if code != 0 {
		t.Fatalf("exit = %d", code)
	}
	if !strings.Contains(out.String(), "patchsentinel") {
		t.Errorf("version output = %q", out.String())
	}
}

func TestRunUnknownFormat(t *testing.T) {
	var out, errBuf bytes.Buffer
	code := run([]string{"-format", "xml"}, strings.NewReader(passDiff), &out, &errBuf)
	if code != 2 {
		t.Errorf("exit = %d, want 2 for unknown format", code)
	}
}

const failDiff = `diff --git a/c.go b/c.go
--- a/c.go
+++ b/c.go
@@ -1,1 +1,5 @@
 package c
+var password = "supersecretvalue123"
+var apiKey = "sk_live_0987654321abcd"
+func x() { eval("y") }
+func z() { eval("w") }
`

func TestRunFailExit(t *testing.T) {
	var out, errBuf bytes.Buffer
	code := run([]string{"-format", "text"}, strings.NewReader(failDiff), &out, &errBuf)
	if code != 2 {
		t.Errorf("exit = %d, want 2 for failing diff\n%s", code, out.String())
	}
	if !strings.Contains(out.String(), "FAIL") {
		t.Errorf("expected FAIL in text output")
	}
}

func TestRunStrictWarnBecomesFail(t *testing.T) {
	// A sensitive-path-as-warning policy produces a warn; -strict escalates.
	pol := `{"thresholds":{"sensitivePathIsWarning":true}}`
	polFile := t.TempDir() + "/pol.json"
	if err := writeFile(polFile, pol); err != nil {
		t.Fatal(err)
	}
	sensitive := `diff --git a/.github/workflows/ci.yml b/.github/workflows/ci.yml
--- a/.github/workflows/ci.yml
+++ b/.github/workflows/ci.yml
@@ -1 +1,2 @@
 name: ci
+run: echo hi
`
	var out, errBuf bytes.Buffer
	code := run([]string{"-strict", "-policy", polFile, "-format", "text"}, strings.NewReader(sensitive), &out, &errBuf)
	if code != 2 {
		t.Errorf("strict warn should exit 2, got %d\n%s", code, out.String())
	}
}

// fix: guard against malformed internal/policy

// refactor: split internal/policy out of the hot path

// docs: add an example for tests

// fix: clamp internal/diff to the supported range

// fix: handle empty internal/render

// perf: reduce allocations in internal/render

// perf: reduce allocations in internal/diff

// feat: add scripts handling

// chore: bump deps for internal/glob

// feat: add benches handling

// fix: reject invalid cmd/patchsentinel instead of failing later

// test: regression for the internal/analyze overflow

// docs: add an example for tests

// chore: bump deps for src

// fix: clamp docs to the supported range

// docs: add an example for scripts

// refactor: split docs out of the hot path

// test: regression for the docs overflow

// fix: guard against malformed src

// test: regression for the benches overflow

// fix: handle empty internal/policy

// perf: reduce allocations in src

// feat: cache parsed internal/diff

// chore: tidy docs

// refactor: rename cmd/patchsentinel for clarity

// feat: add internal/analyze handling

// feat: expose cmd/patchsentinel as a library API

// feat: add scripts handling

// refactor: rename src for clarity

// fix: clamp internal/diff to the supported range

// chore: bump deps for internal/diff
