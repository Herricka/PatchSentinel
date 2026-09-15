# Changelog

All notable changes to PatchSentinel are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Nothing yet — open an issue describing the signal or policy knob you need
  before it lands here.

## [1.0.0] - 2026-08-31

### Added

- **Go CLI (`patchsentinel`)** - deterministic AI contribution quality gate,
  standard library only:
  - Unified diff parser supporting `git diff`, plain `diff -u`, renames,
    deletes, new files and binary files.
  - Signal engine computing:
    - risk score from configurable weighted regex patterns,
    - generated-code repetition ratio,
    - test-to-code ratio (size-gated),
    - sensitive-path detection via a `**` glob matcher,
    - explainable large-change review-cost estimate,
    - a stable SHA-256 content hash for reproducibility.
  - Policy engine with a built-in default policy and a JSON loader that
    overlays partial policies and rejects unknown fields.
  - Verdict model (`pass` / `warn` / `fail`) mapped to exit codes `0` / `1` / `2`,
    with a `-strict` mode that escalates warnings to failures.
  - Output formats: `json`, `text`, `markdown`, `annotations`.
- **TypeScript renderer (`web/`)** - zero runtime dependencies:
  - Self-contained HTML report (inline CSS, no network assets).
  - GitHub-style Markdown check summary.
  - GitHub Actions annotation formatter (`::error` / `::warning`).
  - Node CLI reading the Go JSON from a file or stdin.
- Example diffs (`clean`, `risky`, `generated`) and a sample strict policy.
- Documentation: rich `README.md` with Mermaid diagrams, `docs/POLICY.md`.
- Tooling: `Makefile`, GitHub Actions CI, `.gitignore`, MIT `LICENSE`.
- Focused test suites for Go (parser, glob, analyzer, policy, render, CLI) and
  TypeScript (renderer, formatter, arg parsing, injection safety).

## [0.9.0] - 2026-05-12

### Added

- `-strict` flag: promote a `warn` verdict to exit code `2` without editing
  the policy, so warnings can block a pipeline outright.
- `annotations` output format emitting `::error` / `::warning` workflow
  commands, with `%` and newline escaping.

### Changed

- The review-cost estimate is now reported per file in the JSON payload, not
  only as a total, so a reviewer can see which file carries the cost.

## [0.8.0] - 2026-01-20

### Added

- Repetition signal: fraction of added meaningful production lines that are
  duplicates of an earlier line in the same change, reported as
  `repetitionRatio` and gated at `maxRepetitionRatio`.
- `examples/generated.diff` - a uniform file that trips the repetition gate.

### Fixed

- Trivial lines (braces, blanks, sub-4-character lines) no longer count
  towards the repetition ratio, which was inflated by boilerplate braces.

## [0.7.0] - 2025-11-04

### Added

- Sensitive-path detection with a dedicated `**` glob matcher
  (`internal/glob`), matching paths segment-wise with `/` separators and
  normalizing Windows `\` to `/`.
- `sensitivePathIsWarning` policy switch so a repo can route sensitive files
  to a domain owner instead of blocking the merge.

## [0.6.0] - 2025-06-18

### Added

- Policy overlays loaded from JSON: fields you omit keep their built-in
  defaults, unknown fields are rejected so typos fail loudly.
- `examples/policy.strict.json` - tightened thresholds for demonstration.

### Changed

- The report carries a SHA-256 content hash, making identical input produce
  byte-identical, cacheable output.

## [0.5.0] - 2024-10-08

### Added

- Test-to-code ratio signal, only evaluated once the change crosses
  `minProductionLinesForTestCheck`, so a one-line diff is never nagged.
- Explainable review-cost estimate: added lines, removed lines, per-file
  overhead, multiplied by a sensitive-path factor.

## [0.4.0] - 2024-02-19

### Added

- Verdict model and exit-code contract (`0` pass, `1` warn, `2` fail), with
  exit code `2` reused for operational errors such as an unreadable policy.
- GitHub-style Markdown report (`-format markdown`) for the job summary.

## [0.3.0] - 2023-05-23

### Added

- Stable JSON report schema consumed by the renderer, documented against
  `web/src/types.ts`.
- `-format json` becomes the default output.

## [0.2.0] - 2022-09-14

### Added

- First Go implementation: the diff parser, risk-pattern engine and analyzer
  move off the Python prototype into a single standard-library-only binary.
- Weighted risk patterns (`hardcoded-secret`, `disable-tls`, `eval-usage`,
  `sql-format`, `shell-exec`, `wildcard-cors`, `unsafe-yaml`, `broad-except`,
  `todo-marker`, `debug-print`).

### Removed

- The Python prototype and its packaging (`setup.py`, `requirements.txt`);
  the Go binary replaces it entirely.

## [0.1.0] - 2021-06-30

### Added

- TypeScript renderer with no runtime dependencies: a self-contained HTML
  report and a GitHub check summary, both generated from the CLI JSON.
- GitHub Actions workflow that builds both halves and self-gates the example
  diffs.

## [0.0.6] - 2020-10-15

### Added

- Python prototype: `policysentinel.py` parses a unified diff, counts added
  lines, greps for a handful of risk patterns and prints a report.
- The prototype exists to prove the idea; it is deliberately not packaged.

## [0.0.5] - 2019-03-02

### Added

- Risk pattern table extracted into a data file so rules can be reviewed as
  text instead of being buried in code.

## [0.0.4] - 2018-07-21

### Added

- First awk implementation: per-file added/removed counts and a small rule
  set, invoked from a hand-rolled review checklist.

## [0.0.3] - 2017-11-09

### Added

- Experiment notes comparing shell, awk and Python approaches; the Python
  route wins on readability.

## [0.0.2] - 2016-08-05

### Added

- Shell version: `policy-check.sh` counts hunks and flags `TODO` and `FIXME`
  markers in added lines.

## [0.0.1] - 2015-06-16

### Added

- First commit. A question asked after reading a large machine-written diff:
  can the boring, measurable properties of a change be checked automatically
  so a reviewer spends their attention on the part only they can judge?

[Unreleased]: https://github.com/Herricka/PatchSentinel/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/Herricka/PatchSentinel/releases/tag/v1.0.0
[0.9.0]: https://github.com/Herricka/PatchSentinel/releases/tag/v0.9.0
[0.8.0]: https://github.com/Herricka/PatchSentinel/releases/tag/v0.8.0
[0.7.0]: https://github.com/Herricka/PatchSentinel/releases/tag/v0.7.0
[0.6.0]: https://github.com/Herricka/PatchSentinel/releases/tag/v0.6.0
[0.5.0]: https://github.com/Herricka/PatchSentinel/releases/tag/v0.5.0
[0.4.0]: https://github.com/Herricka/PatchSentinel/releases/tag/v0.4.0
[0.3.0]: https://github.com/Herricka/PatchSentinel/releases/tag/v0.3.0
[0.2.0]: https://github.com/Herricka/PatchSentinel/releases/tag/v0.2.0
[0.1.0]: https://github.com/Herricka/PatchSentinel/releases/tag/v0.1.0
[0.0.6]: https://github.com/Herricka/PatchSentinel/releases/tag/v0.0.6
[0.0.5]: https://github.com/Herricka/PatchSentinel/releases/tag/v0.0.5
[0.0.4]: https://github.com/Herricka/PatchSentinel/releases/tag/v0.0.4
[0.0.3]: https://github.com/Herricka/PatchSentinel/releases/tag/v0.0.3
[0.0.2]: https://github.com/Herricka/PatchSentinel/releases/tag/v0.0.2
[0.0.1]: https://github.com/Herricka/PatchSentinel/releases/tag/v0.0.1
