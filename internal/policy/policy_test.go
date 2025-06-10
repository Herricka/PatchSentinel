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
