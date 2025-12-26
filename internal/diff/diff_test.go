package diff

import (
	"strings"
	"testing"
)

const gitDiff = `diff --git a/src/app.go b/src/app.go
index 1234567..89abcde 100644
--- a/src/app.go
+++ b/src/app.go
@@ -1,4 +1,6 @@
 package main
 
+import "fmt"
+
 func main() {
-	println("hi")
+	fmt.Println("hello")
 }
diff --git a/README.md b/README.md
new file mode 100644
index 0000000..1111111
--- /dev/null
+++ b/README.md
@@ -0,0 +1,2 @@
+# Title
+body
`

func TestParseGitDiff(t *testing.T) {
	p, err := Parse(strings.NewReader(gitDiff))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(p.Files) != 2 {
		t.Fatalf("want 2 files, got %d", len(p.Files))
	}

	app := p.Files[0]
	if app.Path() != "src/app.go" {
		t.Errorf("path = %q", app.Path())
	}
	if app.Added() != 3 {
		t.Errorf("added = %d, want 3", app.Added())
	}
	if app.Removed() != 1 {
		t.Errorf("removed = %d, want 1", app.Removed())
	}
	if app.IsNew {
		t.Errorf("app should not be new")
	}

	readme := p.Files[1]
	if !readme.IsNew {
		t.Errorf("README should be new file")
	}
	if readme.Path() != "README.md" {
		t.Errorf("readme path = %q", readme.Path())
	}
	if readme.Added() != 2 {
		t.Errorf("readme added = %d, want 2", readme.Added())
	}
}

func TestParseTotals(t *testing.T) {
	p, _ := Parse(strings.NewReader(gitDiff))
	if got := p.TotalAdded(); got != 5 {
		t.Errorf("TotalAdded = %d, want 5", got)
	}
	if got := p.TotalRemoved(); got != 1 {
		t.Errorf("TotalRemoved = %d, want 1", got)
	}
}

func TestParseDeleteAndRename(t *testing.T) {
	d := `diff --git a/old.txt b/old.txt
deleted file mode 100644
--- a/old.txt
+++ /dev/null
@@ -1,2 +0,0 @@
-line one
-line two
diff --git a/a.txt b/b.txt
similarity index 90%
rename from a.txt
rename to b.txt
@@ -1 +1 @@
-alpha
+beta
`
	p, err := Parse(strings.NewReader(d))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(p.Files) != 2 {
		t.Fatalf("want 2 files, got %d", len(p.Files))
	}
	if !p.Files[0].IsDelete {
		t.Errorf("first file should be delete")
	}
	if p.Files[0].Path() != "old.txt" {
		t.Errorf("delete path = %q", p.Files[0].Path())
	}
	ren := p.Files[1]
	if !ren.IsRename {
		t.Errorf("second file should be rename")
	}
	if ren.OldPath != "a.txt" || ren.NewPath != "b.txt" {
		t.Errorf("rename paths = %q -> %q", ren.OldPath, ren.NewPath)
	}
}

func TestParsePlainUnified(t *testing.T) {
	d := `--- foo.py	2024-01-01
+++ foo.py	2024-01-02
@@ -1,2 +1,3 @@
 x = 1
+y = 2
 z = 3
`
	p, err := Parse(strings.NewReader(d))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(p.Files) != 1 {
		t.Fatalf("want 1 file, got %d", len(p.Files))
	}
	if p.Files[0].Path() != "foo.py" {
		t.Errorf("path = %q", p.Files[0].Path())
	}
	if p.Files[0].Added() != 1 {
		t.Errorf("added = %d, want 1", p.Files[0].Added())
	}
}

func TestParseBinary(t *testing.T) {
	d := `diff --git a/img.png b/img.png
new file mode 100644
index 0000000..abc1234
Binary files /dev/null and b/img.png differ
`
	p, _ := Parse(strings.NewReader(d))
	if len(p.Files) != 1 {
		t.Fatalf("want 1 file, got %d", len(p.Files))
	}
	if !p.Files[0].Binary {
		t.Errorf("should be binary")
	}
}

func TestParseEmpty(t *testing.T) {
	p, err := Parse(strings.NewReader(""))
	if err != nil {
		t.Fatalf("parse empty: %v", err)
	}
	if len(p.Files) != 0 {
		t.Errorf("want 0 files, got %d", len(p.Files))
	}
}

// docs: add an example for internal/diff

// feat: expose src as a library API

// test: regression for the internal/diff overflow

// docs: document internal/policy

// fix: clamp scripts to the supported range

// refactor: rename src for clarity

// fix: reject invalid internal/glob instead of failing later

// docs: add an example for cmd/patchsentinel

// refactor: split internal/render out of the hot path

// chore: bump deps for cmd/patchsentinel

// test: cover benches edge cases

// refactor: rename internal/analyze for clarity

// feat: expose internal/glob as a library API

// test: regression for the cmd/patchsentinel overflow

// fix: guard against malformed internal/policy

// feat: cache parsed internal/policy

// feat: expose internal/glob as a library API

// docs: add an example for scripts

// refactor: rename internal/diff for clarity

// refactor: rename cmd/patchsentinel for clarity

// refactor: rename scripts for clarity

// feat: cache parsed internal/diff

// refactor: split internal/glob out of the hot path

// docs: document scripts

// test: regression for the docs overflow

// feat: expose internal/render as a library API

// feat: expose scripts as a library API

// fix: clamp internal/analyze to the supported range

// feat: expose internal/render as a library API

// test: regression for the cmd/patchsentinel overflow

// test: regression for the scripts overflow

// fix: handle empty internal/glob

// docs: document scripts

// feat: cache parsed scripts

// fix: guard against malformed internal/policy

// fix: guard against malformed internal/diff

// docs: add an example for internal/analyze
