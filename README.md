<!-- PatchSentinel — the review desk between a diff and your main branch. -->

<p align="center">
  <img src="docs/assets/hero-gates.svg" alt="Diff lines riding a conveyor through the risk, tests, repetition and sensitive-path gates, exiting stamped pass, warn or fail." width="100%">
</p>

<h1 align="center">PatchSentinel</h1>

<p align="center">
  <em>An editorial review desk for diffs. Every change lines up, walks the gates, and leaves with a stamp — <strong>pass</strong>, <strong>warn</strong>, or <strong>fail</strong> — that CI can trust because it was earned the same way every time.</em>
</p>

<p align="center">
  <code>stdlib-only Go core</code> · <code>zero-dependency TypeScript renderer</code> · <code>no network</code> · <code>no LLM in the loop</code> · <code>byte-identical output</code>
</p>

---

Maintainers don't drown in bad code. They drown in **plausible** code — large, tidy-looking diffs produced in seconds that take a human an hour to actually vouch for. The secret in the base64 blob, the `InsecureSkipVerify: true` three files down, the forty near-identical getters, the 300 new lines with zero tests: they all *look* fine at a glance. PatchSentinel reads the whole diff, measures the boring-but-load-bearing stuff, and puts the risky material in front of your eyes before it reaches your users.

It does **not** claim to detect "AI-written code" — that claim isn't honest and this tool won't fake it. It measures transparent, explainable properties of a unified diff — repetition, risk-pattern hits, test-to-code balance, sensitive-path touches, review cost — and turns them into a reproducible verdict. See [Heuristics, honestly](#heuristics-honestly) for exactly where the lines are.

## Navigation

[60-second tour](#the-60-second-tour) · [Install](#install--build) · [Three diffs, three verdicts](#three-diffs-three-verdicts) · [Anatomy of a score](#anatomy-of-a-score) · [Policy knobs](#policy-knobs) · [Exit codes](#exit-codes) · [Report formats](#report-formats) · [GitHub Actions](#github-actions) · [Triage playbooks](#triage-playbooks) · [Heuristics, honestly](#heuristics-honestly) · [Edge cases](#edge-cases) · [Limitations](#limitations) · [Customization](#customization-recipes) · [Architecture](#architecture) · [Development](#development) · [Roadmap](#roadmap) · [License](#license)

## The 60-second tour

```bash
make build                                   # -> ./bin/patchsentinel
git diff origin/main... | ./bin/patchsentinel -format text
```

The core is a **Go CLI** built from the standard library alone: parse a unified diff (files or stdin), evaluate it against a policy, print evidence, exit with a code that mirrors the verdict. A separate dependency-free **TypeScript renderer** turns that JSON into a self-contained HTML page or GitHub check output. Two languages, one schema, zero runtime dependencies.

<p align="center">
  <img src="docs/assets/cockpit-report.svg" alt="A terminal window typing out the risky-example report: verdict FAIL, risk score 42, two sensitive files, exit 2, with live gauges." width="100%">
</p>

## Install & build

Requires **Go ≥ 1.24** and **Node ≥ 20** (Node only for HTML/GitHub rendering; the CLI stands alone).

```bash
make build      # compile the Go CLI into ./bin/patchsentinel
make test       # run Go + TypeScript test suites
make demo       # regenerate the example HTML + Markdown reports
make help       # list every target

# or raw:
go build -o bin/patchsentinel ./cmd/patchsentinel
cd web && npm install && npm run build      # dev-only deps; runtime stays zero-dep
```

## Three diffs, three verdicts

The `examples/` directory ships three diffs spanning the full spectrum. Output below is the **actual** text from the built CLI — not a mock-up.

### 🟢 `clean.diff` → PASS (exit 0)

A small helper plus its test. Balanced, mergeable.

```text
$ patchsentinel -format text examples/clean.diff
Verdict: PASS
------------------------------------------------------------
Files changed      : 2
Lines +/-          : +23 / -0
Production added   : 11
Test added         : 12
Test-to-code ratio : 1.09
Risk score         : 0
Repetition ratio   : 0.00
Sensitive files    : 0
Est. review time   : 11 min

Findings:
  [INFO] No policy violations detected
    All configured gates passed.

Files:
  modified  +11   -0    pkg/mathutil/mathutil.go
  added     +12   -0    pkg/mathutil/mathutil_test.go  (test)
```

### 🔴 `risky.diff` → FAIL (exit 2)

Hardcoded secrets, `eval`, string-built SQL, wildcard CORS, a TODO, plus edits to an auth service and a CI workflow. Two hard gates trip.

```text
$ patchsentinel -format text examples/risky.diff
Verdict: FAIL
------------------------------------------------------------
Files changed      : 2
Lines +/-          : +27 / -0
Production added   : 27
Test added         : 0
Test-to-code ratio : 0.00
Risk score         : 42
Repetition ratio   : 0.00
Sensitive files    : 2
Est. review time   : 18 min

Findings:
  [ERROR] Risk score exceeds maximum
    Cumulative risk score 42 ≥ max 25.
  [ERROR] Sensitive paths modified
    2 sensitive file(s) changed; require domain-owner review.

Risk signals:
  (w10 x2) hardcoded-secret [services/auth/login.go] Possible hardcoded secret
    var apiKey = "sk_live_51H8xYzABCDEFghijklmnop"
  (w8 x1) eval-usage [services/auth/login.go] Dynamic code evaluation via eval()
    eval("dangerousDynamicCode()")
  (w7 x1) sql-format [services/auth/login.go] Possible SQL built via string concatenation
    return fmt.Sprintf("SELECT * FROM users WHERE name = '" + name + "'")
  (w5 x1) wildcard-cors [services/auth/login.go] Wildcard CORS origin
  (w2 x1) todo-marker [services/auth/login.go] Unfinished work marker
```

Each signal line also prints the offending snippet (two shown above). The risk score **42** = `hardcoded-secret 10×2 + eval 8 + sql-format 7 + wildcard-cors 5 + todo 2`. Nothing hidden.

### 🟡 `generated.diff` → WARN (exit 1)

A brand-new file of 33 added lines: getters/setters plus the *same* `validate()` stub repeated over and over. No dangerous patterns, no tests — just suspicious uniformity.

```text
$ patchsentinel -format text examples/generated.diff
Verdict: WARN
------------------------------------------------------------
Files changed      : 1
Lines +/-          : +33 / -0
Risk score         : 0
Repetition ratio   : 0.75
Sensitive files    : 0
Est. review time   : 12 min

Findings:
  [WARN] High repetition in added code
    Repetition ratio 0.75 ≥ 0.60 suggests boilerplate or generated code that needs closer review.
```

That `0.75` is the fraction of added *meaningful* production lines that are duplicates — the closest thing to a "smells machine-made" signal, phrased as a nudge, not an accusation.

## Anatomy of a score

Five independent signals feed one verdict; the verdict is the worst finding severity.

| # | Signal | What it measures | Default gate |
|---|--------|------------------|--------------|
| 1 | **Risk score** | Σ `weight × matchCount` over risk patterns matching *added* lines | `error` ≥ 25 · `warn` ≥ 10 |
| 2 | **Test-to-code ratio** | `testAdded ÷ productionAdded`, only once `productionAdded ≥ 40` | `warn` < 0.25 |
| 3 | **Repetition ratio** | fraction of added meaningful production lines that are duplicates | `warn` ≥ 0.60 |
| 4 | **Sensitive paths** | changed files matching sensitive globs | `error` if any (→`warn`) |
| 5 | **Review cost** | explainable minutes estimate | `warn` if `ceil` > 90 |

The ten built-in risk patterns, `id (weight)` → what they flag:

- `hardcoded-secret (10)` — `api_key`/`secret`/`password`/`token = "…8+ chars…"`
- `disable-tls (9)` — `InsecureSkipVerify: true`, `rejectUnauthorized: false`, `verify=False`
- `eval-usage (8)` — dynamic code eval via `eval(` · `sql-format (7)` — `SELECT/INSERT/UPDATE/DELETE` built with `+`, `%s`, `${`, `format(`
- `shell-exec (6)` — `exec.Command`, `child_process`, `os.system`, `subprocess.` · `wildcard-cors (5)` — `Access-Control-Allow-Origin: *`
- `unsafe-yaml (4)` — `yaml.Unmarshal`/`yaml.load(` · `broad-except (3)` — bare `except:` / empty `catch () {`
- `todo-marker (2)` — `TODO`/`FIXME`/`XXX`/`HACK` · `debug-print (1)` — `console.log`/`fmt.Println`/`print(`

**Review cost** is deliberately auditable: per file, `added × minutesPerAddedLine + removed × minutesPerRemovedLine + minutesPerFile`, ×`sensitiveMultiplier` for sensitive files, summed. With defaults the risky example is ~17.9 min → rounded to **18 min**. Every report also carries a SHA-256 **content hash** so identical input yields byte-identical, cacheable output.

## Policy knobs

Pass `-policy path.json` and its fields overlay the [built-in default](internal/policy/policy.go). Anything you omit keeps its default; unknown fields are **rejected** so typos fail loudly instead of silently.

```jsonc
{
  "sensitivePaths":   ["**/.github/workflows/**", "**/auth/**", "**/*secret*"],
  "thresholds": {
    "maxRiskScore": 25, "warnRiskScore": 10,
    "minTestRatio": 0.25, "minProductionLinesForTestCheck": 40,
    "maxRepetitionRatio": 0.60, "maxReviewMinutes": 90,
    "sensitivePathIsWarning": false
  },
  "reviewCost": { "minutesPerAddedLine": 0.30, "minutesPerFile": 2.0, "sensitiveMultiplier": 1.5 }
}
```

Also configurable: `testPathPatterns` and the `riskPatterns` array (id/regex/message/weight) — full schema in [docs/POLICY.md](docs/POLICY.md).

Glob tokens: `*` = within a segment, `?` = one non-separator char, `**` = any number of segments. Paths compare with `/`; Windows `\` is normalized.

A shipped `examples/policy.strict.json` tightens thresholds and flips `sensitivePathIsWarning` to `true`. Run it on the generated example and the lower `minProductionLinesForTestCheck` surfaces a *second* finding (review time rises to 16 min):

```text
$ patchsentinel -policy examples/policy.strict.json -format text examples/generated.diff
Verdict: WARN
Findings:
  [WARN] Insufficient test coverage for change size
    Test-to-code ratio 0.00 < required 0.30 for 33 added production lines.
  [WARN] High repetition in added code
    Repetition ratio 0.75 ≥ 0.50 suggests boilerplate or generated code that needs closer review.
```

## Exit codes

The exit code *is* the verdict — no parsing required to gate a pipeline.

| Exit | Verdict | Meaning |
|------|---------|---------|
| `0` | pass | every gate satisfied |
| `1` | warn | acceptable, but notable signals present |
| `2` | fail | at least one hard (`error`) gate violated |
| `2` | — | also used for operational errors (bad flag, unreadable policy, unparsable diff) |

`-strict` promotes a `warn` to exit `2` — warnings block without editing the policy. The TypeScript renderer mirrors the report's `exitCode`, so a second pipeline stage stays consistent.

## Report formats

**Go CLI** `-format`: `json` (default — canonical schema, [web/src/types.ts](web/src/types.ts)) · `text` (stable CI logs) · `markdown` (GitHub summary / PR comment) · `annotations` (`::error`/`::warning` workflow commands).

**TypeScript renderer** `node web/dist/cli.js --format`: `html` (self-contained page) · `github` (the Markdown summary from JSON) · `annotations` (workflow commands with `%`/newline escaping).

Annotations for the risky example:

```text
::error title=Risk score exceeds maximum::Cumulative risk score 42 ≥ max 25.
::error title=Sensitive paths modified::2 sensitive file(s) changed; require domain-owner review.
```

## GitHub Actions

The gate is stdin-friendly. Fail on `fail`, post a summary regardless, drop inline annotations:

```yaml
name: patchsentinel
on: pull_request
permissions: { contents: read, pull-requests: write }
jobs:
  gate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with: { fetch-depth: 0 }
      - uses: actions/setup-go@v5
        with: { go-version: "1.24" }
      - uses: actions/setup-node@v4
        with: { node-version: "20" }
      - name: Build tools
        run: |
          go build -o bin/patchsentinel ./cmd/patchsentinel
          cd web && npm install --no-audit --no-fund && npm run build
      - name: Analyze + report
        run: |
          git diff origin/${{ github.base_ref }}... > pr.diff
          ./bin/patchsentinel -format json pr.diff > report.json || true
          node web/dist/cli.js --format github      --in report.json >> "$GITHUB_STEP_SUMMARY"
          node web/dist/cli.js --format annotations --in report.json
      - name: Enforce verdict           # non-zero exit fails the job
        run: ./bin/patchsentinel -format text pr.diff
```

The `|| true` always renders the summary; `Enforce verdict` blocks the merge (add `-strict` to block on warnings too). This repo's [`.github/workflows/ci.yml`](.github/workflows/ci.yml) self-gates the examples the same way.

## Triage playbooks

What a maintainer should actually *do* with each stamp.

- **🟢 PASS** — review as normal. Not a correctness proof; it means the load-bearing checks are green.
- **🟡 WARN · repetition** — skim for copy-paste that should be a loop or generator. Legitimately generated (protobuf, mocks, sqlc)? Raise `maxRepetitionRatio` for that repo. A human fanning out near-identical code? That's a refactor conversation.
- **🟡 WARN · thin tests** — the change crossed `minProductionLinesForTestCheck` with too few tests. Ask for tests, or justify and tune the threshold.
- **🟡 WARN · review cost** — the diff is expensive to review carefully. Push back on scope: split mechanical from behavioral.
- **🔴 FAIL · risk score** — read every `Risk signals` line first; `hardcoded-secret`, `disable-tls` and `sql-format` end incidents. Confirm real vs. false positive before overriding.
- **🔴 FAIL · sensitive paths** — auth/infra/CI/secrets changed; route to that path's owner. Flip `sensitivePathIsWarning` if you'd rather it advise than block.

## Heuristics, honestly

The part marketing usually skips.

- **PatchSentinel does not detect AI.** No model, no classifier, no "confidence this was written by an LLM." The `repetition` signal correlates with boilerplate and *some* generated output, but human boilerplate scores identically and hand-crafted AI output can score zero. Treat it as *"worth a closer look,"* never as proof of origin.
- **Risk patterns are regexes over added lines** — they find *shapes*, not *meaning*. `eval(` in a comment counts; an obfuscated secret may not.
- **The verdict is evidence for a human, not a replacement.** Nothing here understands your architecture or threat model; a person still owns the merge.
- **Determinism is the actual promise.** Same diff + same policy ⇒ same bytes, every run, every OS — cacheable and trustworthy, with no claim of intelligence.

## Edge cases

- **Binary files** are tracked (flagged `binary`) but not line-scanned — they add file-level review cost only.
- **Renames** are detected and reported with `oldPath`; **deletions** never trip risk patterns (only *added* lines are scanned) and cost less.
- **Tiny diffs** skip the test-ratio gate entirely — below `minProductionLinesForTestCheck` (default 40) a one-liner is never nagged for missing tests.
- **Trivial lines** (`{`, `}`, blanks, sub-4-char lines) are excluded from repetition analysis so braces don't inflate the ratio.
- **Zero production, some tests** → ratio `1` by definition; **zero of both** → `0`.
- **Invalid regex** in a custom policy is skipped defensively instead of crashing.
- **Multiple diff files** on the command line are concatenated (newline-safe) and analyzed as one change.

## Limitations

- **Line-oriented, not semantic** — no AST, taint tracking or dataflow. It reads diffs, not programs.
- **Regex risk patterns have false positives and negatives** — the accepted trade for zero dependencies and total determinism.
- **Language-agnostic**, so it can't apply language-specific rules (`errcheck`, ESLint). Pair it with real linters; it isn't one.
- **Context-free** — it sees the diff, not the surrounding file or history, so an "unused" new function may actually be wired up elsewhere.
- **Review-cost is a heuristic** calibrated by your numbers — minutes are relative, not a stopwatch.

## Customization recipes

A supplied `riskPatterns` array *replaces* the defaults (arrays don't merge), so add a rule like `{ "id": "banned-md5", "regex": "crypto/md5|hashlib\\.md5", "weight": 9 }` alongside the built-ins you keep. Common threshold tweaks:

```json
{ "thresholds": { "sensitivePathIsWarning": true } }   // sensitive paths advise, not block
{ "thresholds": { "maxRepetitionRatio": 0.85 } }       // relax for a mocks-heavy service
{ "thresholds": { "maxRiskScore": 15, "warnRiskScore": 5, "minTestRatio": 0.5 } }  // tighten security repo
{ "reviewCost": { "minutesPerAddedLine": 0.5, "maxReviewMinutes": 60 } }           // recalibrate effort
```

Pair a tightened policy with `-strict` to make warnings block in CI.

## Architecture

```mermaid
flowchart LR
    A["diff (file/stdin)"] --> B[diff.Parse] --> C[analyze.Analyze]
    P["policy JSON / default"] --> C
    C --> E["Report JSON (schema v1)"]
    E --> F["Go renderer: text · md · annotations"]
    E --> G["TS renderer: html · github · annotations"]
    C -->|"exit 0 / 1 / 2"| K[CI gate]
```

Layout: `cmd/patchsentinel` (CLI entrypoint) · `internal/{diff,policy,glob,analyze,render}` (stdlib-only engine) · `web/src` (zero-dep TypeScript renderer) · `examples/` (the three diffs + strict policy) · `docs/POLICY.md` + `docs/assets/` (the two local SVGs above) · `.github/workflows/` (Go + TS + self-gate CI).

## Development

`make vet` · `make cover` (Go tests + coverage) · `make test-ts` (type-check + build + test the renderer) · `make demo` (regenerate example reports) · `make ci` (what CI runs).

## Roadmap

Direction, not commitments:

- **Baseline mode** — gate only on findings *new* since a stored baseline.
- **SARIF output** so findings land in GitHub's "Code scanning" tab.
- **Per-directory policy overlays** for monorepos.
- **Pattern packs** — versioned, composable risk-pattern bundles.
- **Suppression comments** with an audit trail for accepted false positives.
- **Published JSON schema** for the report format.

## License

[MIT](LICENSE). Bring it to your review desk, tune the gates, ship with a clearer conscience.