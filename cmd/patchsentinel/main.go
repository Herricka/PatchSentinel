// Command patchsentinel is a deterministic AI contribution quality gate.
//
// It reads one or more unified diff files (or stdin), analyzes them against a
// policy, and emits structured evidence plus a pass/warn/fail verdict. The
// process exit code mirrors the verdict (0 pass, 1 warn, 2 fail) so it can be
// wired directly into CI pipelines.
//
// Usage:
//
//	patchsentinel [flags] [diff-file ...]
//
// Flags:
//
//	-format   output format: json | text | markdown | annotations (default json)
//	-policy   path to a JSON policy file (defaults to built-in policy)
//	-out      write output to a file instead of stdout
//	-strict   treat warnings as failures (exit 2 on warn)
//	-version  print version and exit
//
// With no diff-file arguments, the diff is read from stdin.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/patchsentinel/patchsentinel/internal/analyze"
	"github.com/patchsentinel/patchsentinel/internal/diff"
	"github.com/patchsentinel/patchsentinel/internal/policy"
	"github.com/patchsentinel/patchsentinel/internal/render"
)

// version is overridable at build time via -ldflags "-X main.version=...".
var version = "1.0.0"

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

// run is the testable core of the CLI. It returns the process exit code.
func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("patchsentinel", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var (
		format      = fs.String("format", "json", "output format: json | text | markdown | annotations")
		policyPath  = fs.String("policy", "", "path to JSON policy file (default: built-in policy)")
		outPath     = fs.String("out", "", "write output to this file instead of stdout")
		strict      = fs.Bool("strict", false, "treat warnings as failures (exit 2 on warn)")
		showVersion = fs.Bool("version", false, "print version and exit")
	)
	fs.Usage = func() {
		fmt.Fprintf(stderr, "patchsentinel — deterministic AI contribution quality gate\n\n")
		fmt.Fprintf(stderr, "Usage: patchsentinel [flags] [diff-file ...]\n\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *showVersion {
		fmt.Fprintf(stdout, "patchsentinel %s\n", version)
		return 0
	}

	// Load policy.
	pol := policy.Default()
	if *policyPath != "" {
		f, err := os.Open(*policyPath)
		if err != nil {
			fmt.Fprintf(stderr, "error: open policy: %v\n", err)
			return 2
		}
		loaded, err := policy.Load(f)
		f.Close()
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 2
		}
		pol = loaded
	}

	// Read diff input from files or stdin.
	raw, err := readInput(fs.Args(), stdin)
	if err != nil {
		fmt.Fprintf(stderr, "error: read input: %v\n", err)
		return 2
	}

	patch, err := diff.Parse(bytes.NewReader(raw))
	if err != nil {
		fmt.Fprintf(stderr, "error: parse diff: %v\n", err)
		return 2
	}

	report := analyze.Analyze(patch, pol)

	out, err := format2bytes(*format, report)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 2
	}

	if err := writeOutput(*outPath, out, stdout); err != nil {
		fmt.Fprintf(stderr, "error: write output: %v\n", err)
		return 2
	}

	exit := report.ExitCode
	if *strict && report.Verdict == analyze.Warn {
		exit = 2
	}
	return exit
}

// readInput concatenates the contents of the named files, or reads stdin when
// no files are given.
func readInput(files []string, stdin io.Reader) ([]byte, error) {
	if len(files) == 0 {
		return io.ReadAll(stdin)
	}
	var buf bytes.Buffer
	for _, name := range files {
		b, err := os.ReadFile(name)
		if err != nil {
			return nil, err
		}
		buf.Write(b)
		if len(b) > 0 && b[len(b)-1] != '\n' {
			buf.WriteByte('\n')
		}
	}
	return buf.Bytes(), nil
}

// format2bytes serializes a report in the requested format.
func format2bytes(format string, r analyze.Report) ([]byte, error) {
	switch strings.ToLower(format) {
	case "json":
		b, err := json.MarshalIndent(r, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("marshal json: %w", err)
		}
		return append(b, '\n'), nil
	case "text":
		return []byte(render.Text(r)), nil
	case "markdown", "md":
		return []byte(render.Markdown(r)), nil
	case "annotations":
		return []byte(render.AnnotationsText(r)), nil
	default:
		return nil, fmt.Errorf("unknown format %q (want json|text|markdown|annotations)", format)
	}
}

func writeOutput(path string, data []byte, stdout io.Writer) error {
	if path == "" {
		_, err := stdout.Write(data)
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// fix: handle empty tests

// chore: bump deps for src

// test: cover internal/glob edge cases

// fix: reject invalid scripts instead of failing later

// refactor: rename internal/analyze for clarity

// test: regression for the cmd/patchsentinel overflow

// docs: add an example for internal/policy

// docs: document internal/diff

// chore: tidy scripts

// test: cover scripts edge cases

// docs: document src
