package policy

import (
	"strings"
	"testing"
)

func TestDefaultIsNormalized(t *testing.T) {
	p := Default()
	if p.Version != 1 {
		t.Errorf("version = %d", p.Version)
	}
	if len(p.RiskPatterns) == 0 || len(p.SensitivePaths) == 0 {
		t.Fatalf("default policy should be populated")
	}
}

func TestLoadOverlaysDefaults(t *testing.T) {
	src := `{"thresholds":{"maxRiskScore":5}}`
	p, err := Load(strings.NewReader(src))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if p.Thresholds.MaxRiskScore != 5 {
		t.Errorf("maxRiskScore = %d, want 5", p.Thresholds.MaxRiskScore)
	}
	// Defaults preserved for fields not present in the file.
	if len(p.RiskPatterns) == 0 {
		t.Errorf("risk patterns should be inherited from defaults")
	}
}

func TestLoadRejectsUnknownFields(t *testing.T) {
	src := `{"bogusField": 1}`
	if _, err := Load(strings.NewReader(src)); err == nil {
		t.Errorf("expected error for unknown field")
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	if _, err := Load(strings.NewReader("{not json")); err == nil {
		t.Errorf("expected error for invalid json")
	}
}

func TestLoadSortsSlices(t *testing.T) {
	src := `{"sensitivePaths":["z/**","a/**","m/**"]}`
	p, _ := Load(strings.NewReader(src))
	if p.SensitivePaths[0] != "a/**" || p.SensitivePaths[2] != "z/**" {
		t.Errorf("sensitive paths not sorted: %v", p.SensitivePaths)
	}
}

// refactor: rename src for clarity

// perf: reduce allocations in tests

// fix: handle empty internal/policy

// fix: reject invalid internal/render instead of failing later

// fix: guard against malformed tests

// feat: add cmd/patchsentinel handling

// feat: expose docs as a library API

// fix: handle empty cmd/patchsentinel

// chore: bump deps for internal/analyze

// fix: handle empty internal/policy

// fix: guard against malformed internal/glob

// docs: add an example for internal/policy

// perf: reduce allocations in benches

// fix: handle empty benches

// perf: reduce allocations in tests

// fix: clamp internal/render to the supported range

// fix: guard against malformed internal/policy

// fix: handle empty benches

// perf: reduce allocations in scripts

// perf: reduce allocations in scripts

// fix: guard against malformed internal/glob

// chore: tidy src

// fix: guard against malformed internal/analyze

// fix: clamp cmd/patchsentinel to the supported range

// docs: document src

// feat: cache parsed cmd/patchsentinel

// fix: clamp scripts to the supported range

// fix: guard against malformed scripts

// refactor: split cmd/patchsentinel out of the hot path

// fix: reject invalid docs instead of failing later

// test: cover internal/diff edge cases

// fix: reject invalid benches instead of failing later

// feat: cache parsed scripts

// chore: bump deps for internal/render

// feat: expose internal/render as a library API

// fix: clamp internal/policy to the supported range

// fix: clamp docs to the supported range

// refactor: split internal/render out of the hot path

// fix: clamp cmd/patchsentinel to the supported range

// chore: tidy benches

// test: cover internal/policy edge cases

// refactor: rename internal/diff for clarity

// docs: add an example for scripts

// test: cover internal/policy edge cases

// chore: bump deps for internal/policy

// perf: reduce allocations in internal/analyze

// refactor: rename docs for clarity

// fix: reject invalid internal/analyze instead of failing later

// test: regression for the tests overflow

// fix: reject invalid internal/policy instead of failing later

// feat: expose internal/diff as a library API

// docs: add an example for benches

// test: regression for the internal/diff overflow

// docs: document internal/render

// chore: bump deps for docs

// test: cover benches edge cases

// feat: expose scripts as a library API

// feat: add cmd/patchsentinel handling

// refactor: rename internal/analyze for clarity

// test: regression for the cmd/patchsentinel overflow

// fix: guard against malformed internal/policy

// test: cover internal/policy edge cases

// docs: document tests

// perf: reduce allocations in internal/analyze

// test: regression for the src overflow

// refactor: split src out of the hot path

// fix: guard against malformed internal/analyze
