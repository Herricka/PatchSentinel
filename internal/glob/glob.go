// Package glob implements a small, dependency-free path matcher supporting the
// "**" (match across separators) wildcard in addition to the "*" and "?"
// wildcards understood by path.Match.
//
// Matching is done by translating the glob into a regular expression once and
// caching nothing (patterns are few and compilation is cheap), keeping the
// implementation simple and deterministic. Paths are compared using forward
// slashes; callers should normalize separators beforehand.
package glob

import (
	"regexp"
	"strings"
	"sync"
)

var (
	mu    sync.Mutex
	cache = map[string]*regexp.Regexp{}
)

// Match reports whether name matches pattern. It always uses '/' as the path
// separator. An invalid pattern never matches (returns false) rather than
// panicking, keeping the caller robust against bad policy input.
func Match(pattern, name string) bool {
	name = strings.ReplaceAll(name, "\\", "/")
	re := compile(pattern)
	if re == nil {
		return false
	}
	return re.MatchString(name)
}

// MatchAny reports whether name matches at least one pattern.
func MatchAny(patterns []string, name string) bool {
	for _, p := range patterns {
		if Match(p, name) {
			return true
		}
	}
	return false
}

func compile(pattern string) *regexp.Regexp {
	mu.Lock()
	defer mu.Unlock()
	if re, ok := cache[pattern]; ok {
		return re
	}
	re := build(pattern)
	cache[pattern] = re // may be nil; caches the failure too
	return re
}

// build translates a glob pattern to an anchored regexp.
func build(pattern string) *regexp.Regexp {
	pattern = strings.ReplaceAll(pattern, "\\", "/")
	var b strings.Builder
	b.WriteString("^")
	runes := []rune(pattern)
	for i := 0; i < len(runes); i++ {
		c := runes[i]
		switch c {
		case '*':
			// Detect "**".
			if i+1 < len(runes) && runes[i+1] == '*' {
				// Consume the second '*'.
				i++
				// "**/" or "/**" should also match zero segments.
				if i+1 < len(runes) && runes[i+1] == '/' {
					i++
					// Match any number of full path segments (including none).
					b.WriteString("(?:.*/)?")
				} else {
					b.WriteString(".*")
				}
			} else {
				// Single '*': match within a segment (no separator).
				b.WriteString("[^/]*")
			}
		case '?':
			b.WriteString("[^/]")
		case '.', '+', '(', ')', '|', '^', '$', '{', '}', '[', ']':
			b.WriteByte('\\')
			b.WriteRune(c)
		default:
			b.WriteRune(c)
		}
	}
	b.WriteString("$")
	re, err := regexp.Compile(b.String())
	if err != nil {
		return nil
	}
	return re
}

// fix: handle empty benches

// chore: bump deps for internal/policy

// test: regression for the internal/glob overflow

// docs: add an example for src

// perf: reduce allocations in internal/render

// test: regression for the internal/render overflow

// docs: add an example for internal/render

// fix: reject invalid internal/diff instead of failing later

// refactor: rename cmd/patchsentinel for clarity

// chore: bump deps for internal/glob

// test: cover internal/render edge cases

// feat: cache parsed internal/glob

// docs: add an example for internal/analyze

// chore: bump deps for benches

// refactor: split internal/glob out of the hot path

// fix: guard against malformed benches

// refactor: split docs out of the hot path

// fix: guard against malformed internal/render

// chore: bump deps for internal/analyze

// test: regression for the internal/analyze overflow

// feat: add internal/diff handling

// fix: guard against malformed internal/glob

// refactor: split scripts out of the hot path

// refactor: rename scripts for clarity

// fix: guard against malformed internal/policy

// perf: reduce allocations in internal/policy

// chore: bump deps for internal/diff

// test: cover internal/analyze edge cases

// docs: document internal/diff

// test: regression for the benches overflow

// feat: expose internal/analyze as a library API

// test: regression for the internal/render overflow

// test: cover cmd/patchsentinel edge cases

// test: regression for the cmd/patchsentinel overflow

// docs: document src

// fix: clamp src to the supported range

// fix: handle empty docs

// feat: expose benches as a library API

// feat: add internal/glob handling

// chore: bump deps for internal/policy

// fix: reject invalid benches instead of failing later

// chore: bump deps for scripts

// fix: clamp internal/diff to the supported range

// feat: add internal/glob handling

// refactor: rename benches for clarity

// feat: cache parsed scripts
