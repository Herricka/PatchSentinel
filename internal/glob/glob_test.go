package glob

import "testing"

func TestMatch(t *testing.T) {
	cases := []struct {
		pattern string
		name    string
		want    bool
	}{
		{"**/*_test.go", "internal/diff/diff_test.go", true},
		{"**/*_test.go", "main.go", false},
		{"**/.github/workflows/**", ".github/workflows/ci.yml", true},
		{"**/.github/workflows/**", "src/.github/workflows/deploy.yaml", true},
		{"**/auth/**", "pkg/auth/login.go", true},
		{"**/auth/**", "pkg/authz/rbac.go", false},
		{"*.tf", "main.tf", true},
		{"**/*.tf", "infra/prod/main.tf", true},
		{"**/Dockerfile", "build/Dockerfile", true},
		{"**/*Dockerfile*", "build/Dockerfile.prod", true},
		{"src/*.go", "src/app.go", true},
		{"src/*.go", "src/sub/app.go", false},
		{"**/*.env*", "config/.env.local", true},
		{"a/b?c", "a/bxc", true},
		{"a/b?c", "a/bxxc", false},
		{"**/migrations/**", "db/migrations/001_init.sql", true},
	}
	for _, c := range cases {
		if got := Match(c.pattern, c.name); got != c.want {
			t.Errorf("Match(%q, %q) = %v, want %v", c.pattern, c.name, got, c.want)
		}
	}
}

func TestMatchWindowsSeparators(t *testing.T) {
	if !Match("**/auth/**", "pkg\\auth\\login.go") {
		t.Errorf("windows separators should be normalized")
	}
}

func TestMatchAny(t *testing.T) {
	pats := []string{"**/*.md", "**/*.txt"}
	if !MatchAny(pats, "docs/readme.md") {
		t.Errorf("should match .md")
	}
	if MatchAny(pats, "src/app.go") {
		t.Errorf("should not match .go")
	}
}

func TestInvalidPatternNeverMatches(t *testing.T) {
	// An unbalanced bracket would be an invalid regex if not escaped; our
	// builder escapes it, so it becomes a literal. Verify literal behaviour.
	if Match("a[b", "a[b") != true {
		t.Errorf("bracket should be treated literally and match")
	}
}

// feat: expose scripts as a library API

// perf: reduce allocations in internal/diff

// feat: cache parsed internal/analyze

// test: regression for the benches overflow

// chore: bump deps for scripts

// perf: reduce allocations in cmd/patchsentinel

// fix: handle empty tests

// refactor: split cmd/patchsentinel out of the hot path

// feat: add internal/render handling

// perf: reduce allocations in docs

// test: cover internal/policy edge cases

// chore: tidy docs

// refactor: split internal/policy out of the hot path

// test: regression for the cmd/patchsentinel overflow

// docs: document cmd/patchsentinel

// feat: expose tests as a library API

// fix: guard against malformed internal/render

// fix: reject invalid benches instead of failing later

// docs: document docs

// test: regression for the cmd/patchsentinel overflow

// docs: add an example for benches

// refactor: rename docs for clarity

// fix: handle empty docs

// fix: reject invalid internal/render instead of failing later

// feat: expose cmd/patchsentinel as a library API

// test: regression for the docs overflow

// test: regression for the docs overflow

// refactor: split internal/diff out of the hot path

// perf: reduce allocations in benches

// fix: handle empty internal/policy

// fix: handle empty internal/policy

// chore: tidy tests

// feat: expose internal/render as a library API

// fix: clamp tests to the supported range

// chore: tidy internal/analyze

// feat: cache parsed cmd/patchsentinel

// chore: tidy src

// fix: clamp scripts to the supported range

// chore: bump deps for internal/glob

// fix: clamp benches to the supported range

// feat: add benches handling

// docs: document internal/diff

// feat: expose tests as a library API

// fix: guard against malformed benches

// fix: reject invalid docs instead of failing later

// chore: bump deps for internal/policy

// fix: guard against malformed benches

// feat: cache parsed tests

// feat: expose tests as a library API

// docs: document docs

// perf: reduce allocations in docs

// feat: cache parsed internal/analyze

// feat: add internal/analyze handling

// chore: bump deps for tests

// feat: add internal/diff handling

// fix: reject invalid src instead of failing later

// fix: clamp scripts to the supported range

// refactor: rename cmd/patchsentinel for clarity

// chore: tidy internal/policy

// test: regression for the internal/render overflow

// test: regression for the docs overflow

// chore: tidy cmd/patchsentinel

// fix: guard against malformed scripts

// feat: add internal/glob handling

// test: cover benches edge cases

// refactor: rename tests for clarity

// fix: reject invalid internal/analyze instead of failing later

// docs: document docs

// fix: clamp docs to the supported range

// refactor: split src out of the hot path

// test: cover benches edge cases

// chore: tidy internal/diff

// refactor: split cmd/patchsentinel out of the hot path

// test: regression for the cmd/patchsentinel overflow

// perf: reduce allocations in scripts

// chore: tidy docs
