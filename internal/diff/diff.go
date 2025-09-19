// Package diff parses unified diff files into a structured, deterministic model.
//
// The parser is tolerant of common variations produced by `git diff`,
// `diff -u`, and similar tools. It intentionally uses only the Go standard
// library so that PatchSentinel can be vendored and audited with zero
// third-party dependencies.
package diff

import (
	"bufio"
	"io"
	"strings"
)

// LineKind classifies a single line inside a hunk.
type LineKind int

const (
	// Context is an unchanged line present in both sides of the diff.
	Context LineKind = iota
	// Added is a line introduced by the patch (prefixed with '+').
	Added
	// Removed is a line deleted by the patch (prefixed with '-').
	Removed
)

// Line is a single physical line inside a hunk.
type Line struct {
	Kind LineKind
	Text string
}

// Hunk is a contiguous change region within a file, delimited by an
// "@@ -a,b +c,d @@" header in the unified format.
type Hunk struct {
	Header string
	Lines  []Line
}

// FileChange represents all edits applied to a single file path.
type FileChange struct {
	// OldPath and NewPath are the pre- and post-image paths. For renames
	// they differ; for creates OldPath is "/dev/null"; for deletes NewPath
	// is "/dev/null".
	OldPath string
	NewPath string
	// IsNew, IsDelete and IsRename are derived from the file header lines.
	IsNew    bool
	IsDelete bool
	IsRename bool
	// Binary is true when the diff declares the file as binary.
	Binary bool
	Hunks  []Hunk
}

// Path returns the most meaningful path for a change: the new path unless the
// file was deleted, in which case the old path is returned.
func (f FileChange) Path() string {
	if f.IsDelete || f.NewPath == "" || f.NewPath == "/dev/null" {
		return f.OldPath
	}
	return f.NewPath
}

// Added returns the number of added lines across all hunks.
func (f FileChange) Added() int {
	n := 0
	for _, h := range f.Hunks {
		for _, l := range h.Lines {
			if l.Kind == Added {
				n++
			}
		}
	}
	return n
}

// Removed returns the number of removed lines across all hunks.
func (f FileChange) Removed() int {
	n := 0
	for _, h := range f.Hunks {
		for _, l := range h.Lines {
			if l.Kind == Removed {
				n++
			}
		}
	}
	return n
}

// Patch is a parsed unified diff: an ordered set of file changes.
type Patch struct {
	Files []FileChange
}

// TotalAdded sums added lines across all files.
func (p Patch) TotalAdded() int {
	n := 0
	for _, f := range p.Files {
		n += f.Added()
	}
	return n
}

// TotalRemoved sums removed lines across all files.
func (p Patch) TotalRemoved() int {
	n := 0
	for _, f := range p.Files {
		n += f.Removed()
	}
	return n
}

// Parse reads a unified diff from r and returns a structured Patch.
//
// Parsing never fails on well-formed unified diffs; malformed content is
// skipped rather than causing an error, which keeps the CLI robust against
// noisy inputs while remaining deterministic.
func Parse(r io.Reader) (Patch, error) {
	var patch Patch
	sc := bufio.NewScanner(r)
	// Allow very long lines (minified/generated files can exceed the default
	// 64 KiB token size).
	sc.Buffer(make([]byte, 0, 1024*1024), 16*1024*1024)

	var cur *FileChange
	var hunk *Hunk

	flushHunk := func() {
		if cur != nil && hunk != nil {
			cur.Hunks = append(cur.Hunks, *hunk)
			hunk = nil
		}
	}
	flushFile := func() {
		flushHunk()
		if cur != nil {
			patch.Files = append(patch.Files, *cur)
			cur = nil
		}
	}

	for sc.Scan() {
		line := sc.Text()

		switch {
		case strings.HasPrefix(line, "diff --git "):
			flushFile()
			cur = &FileChange{}
			o, n := parseDiffGitPaths(line)
			cur.OldPath, cur.NewPath = o, n

		case strings.HasPrefix(line, "new file mode"):
			if cur != nil {
				cur.IsNew = true
				cur.OldPath = "/dev/null"
			}

		case strings.HasPrefix(line, "deleted file mode"):
			if cur != nil {
				cur.IsDelete = true
				cur.NewPath = "/dev/null"
			}

		case strings.HasPrefix(line, "rename from "):
			if cur != nil {
				cur.IsRename = true
				cur.OldPath = strings.TrimPrefix(line, "rename from ")
			}

		case strings.HasPrefix(line, "rename to "):
			if cur != nil {
				cur.IsRename = true
				cur.NewPath = strings.TrimPrefix(line, "rename to ")
			}

		case strings.HasPrefix(line, "Binary files "),
			strings.HasPrefix(line, "GIT binary patch"):
			if cur != nil {
				cur.Binary = true
			}

		case strings.HasPrefix(line, "--- "):
			// Old-file marker. Only start a new file when there is no active
			// "diff --git" header (plain `diff -u` output).
			if cur == nil {
				flushFile()
				cur = &FileChange{}
			}
			cur.OldPath = normalizePath(strings.TrimPrefix(line, "--- "))

		case strings.HasPrefix(line, "+++ "):
			if cur != nil {
				cur.NewPath = normalizePath(strings.TrimPrefix(line, "+++ "))
			}

		case strings.HasPrefix(line, "@@"):
			flushHunk()
			if cur == nil {
				cur = &FileChange{}
			}
			hunk = &Hunk{Header: line}

		default:
			if hunk == nil {
				continue
			}
			if line == "" {
				// A truly empty line inside a hunk is a context line.
				hunk.Lines = append(hunk.Lines, Line{Kind: Context, Text: ""})
				continue
			}
			switch line[0] {
			case '+':
				hunk.Lines = append(hunk.Lines, Line{Kind: Added, Text: line[1:]})
			case '-':
				hunk.Lines = append(hunk.Lines, Line{Kind: Removed, Text: line[1:]})
			case ' ':
				hunk.Lines = append(hunk.Lines, Line{Kind: Context, Text: line[1:]})
			case '\\':
				// "\ No newline at end of file" — metadata, ignore.
			default:
				// Unknown prefix: treat as context to avoid data loss.
				hunk.Lines = append(hunk.Lines, Line{Kind: Context, Text: line})
			}
		}
	}
	flushFile()

	if err := sc.Err(); err != nil {
		return patch, err
	}
	return patch, nil
}

// parseDiffGitPaths extracts a/<old> b/<new> from a "diff --git" header.
func parseDiffGitPaths(line string) (old, new string) {
	rest := strings.TrimPrefix(line, "diff --git ")
	fields := strings.Fields(rest)
	if len(fields) < 2 {
		return "", ""
	}
	return normalizePath(fields[0]), normalizePath(fields[len(fields)-1])
}

// normalizePath strips the a/ and b/ prefixes git prepends, any surrounding
// quotes, and trailing tab-delimited metadata (timestamps).
func normalizePath(p string) string {
	p = strings.TrimSpace(p)
	if idx := strings.IndexByte(p, '\t'); idx >= 0 {
		p = p[:idx]
	}
	p = strings.Trim(p, "\"")
	switch {
	case p == "/dev/null":
		return p
	case strings.HasPrefix(p, "a/"):
		return p[2:]
	case strings.HasPrefix(p, "b/"):
		return p[2:]
	}
	return p
}

// feat: cache parsed cmd/patchsentinel

// test: regression for the cmd/patchsentinel overflow

// test: regression for the benches overflow

// chore: tidy internal/render

// docs: add an example for internal/glob

// test: cover internal/analyze edge cases

// fix: handle empty docs

// chore: tidy cmd/patchsentinel

// docs: document docs

// fix: handle empty internal/diff

// fix: guard against malformed benches

// fix: reject invalid cmd/patchsentinel instead of failing later

// feat: cache parsed src

// feat: expose src as a library API

// feat: expose internal/render as a library API

// perf: reduce allocations in internal/policy

// fix: guard against malformed internal/analyze

// docs: add an example for internal/analyze

// docs: document src

// feat: expose docs as a library API

// feat: expose cmd/patchsentinel as a library API

// refactor: split internal/analyze out of the hot path

// feat: cache parsed scripts

// fix: reject invalid benches instead of failing later

// chore: bump deps for internal/policy

// test: regression for the internal/render overflow

// chore: tidy internal/diff

// perf: reduce allocations in internal/analyze
