# PatchSentinel Policy Reference

A **policy** is a JSON document describing the rules PatchSentinel evaluates a
patch against. If you pass `-policy path.json`, its fields are overlaid on top
of the built-in [default policy](#default-policy): any field you omit keeps its
default value. Unknown fields are rejected so typos fail loudly.

```bash
patchsentinel -policy examples/policy.strict.json examples/risky.diff
```

## Schema

```jsonc
{
  "version": 1,
  "sensitivePaths":   ["**/.github/workflows/**", "..."],
  "testPathPatterns": ["**/*_test.go", "..."],
  "riskPatterns": [
    { "id": "eval-usage", "regex": "\\beval\\s*\\(", "message": "…", "weight": 8 }
  ],
  "thresholds": {
    "maxRiskScore": 25,
    "warnRiskScore": 10,
    "minTestRatio": 0.25,
    "minProductionLinesForTestCheck": 40,
    "maxRepetitionRatio": 0.60,
    "maxReviewMinutes": 90,
    "sensitivePathIsWarning": false
  },
  "reviewCost": {
    "minutesPerAddedLine": 0.30,
    "minutesPerRemovedLine": 0.10,
    "minutesPerFile": 2.0,
    "sensitiveMultiplier": 1.5
  }
}
```

## Glob syntax

Path patterns (`sensitivePaths`, `testPathPatterns`) support:

| Token | Meaning                                            |
|-------|----------------------------------------------------|
| `*`   | Any run of characters **within** a path segment    |
| `?`   | Exactly one non-separator character                 |
| `**`  | Any number of path segments (including zero)        |

Paths are always compared using `/` separators; Windows `\` separators are
normalized automatically.

## Signals & gates

Each gate produces a **finding** with a severity. The overall verdict is the
worst severity across all findings:

- any `error` → **fail** (exit code `2`)
- otherwise any `warn` → **warn** (exit code `1`)
- otherwise → **pass** (exit code `0`)

### 1. Risk score

Every `riskPattern` whose `regex` matches an **added** line contributes
`weight × matchCount` to the cumulative risk score.

| Condition                              | Severity |
|----------------------------------------|----------|
| `riskScore >= thresholds.maxRiskScore` | error    |
| `riskScore >= thresholds.warnRiskScore`| warn     |

The default patterns cover hardcoded secrets, `eval`, process execution,
string-built SQL, disabled TLS verification, wildcard CORS, unsafe
deserialization, debug prints and unfinished-work markers.

### 2. Test-to-code ratio

`testToCodeRatio = testAddedLines / productionAddedLines`. The gate only
applies once `productionAdded >= minProductionLinesForTestCheck`, so small
changes are never penalized. Below `minTestRatio` yields a **warn**.

Files are classified as tests via `testPathPatterns`.

### 3. Generated-code repetition

`repetitionRatio` is the fraction of added, meaningful production lines that are
duplicates (every occurrence beyond the first). High values indicate
boilerplate or machine-generated code that deserves closer review. At or above
`maxRepetitionRatio` yields a **warn**.

### 4. Sensitive paths

Any changed file matching `sensitivePaths` is counted. When one or more are
present the gate fires — an **error** by default, or a **warn** when
`sensitivePathIsWarning` is `true`.

### 5. Review cost

An explainable effort estimate in minutes:

```
perFile   = added × minutesPerAddedLine
          + removed × minutesPerRemovedLine
          + minutesPerFile
          × (sensitiveMultiplier if the file is sensitive)
reviewMinutes = Σ perFile over all files
```

When `ceil(reviewMinutes) > maxReviewMinutes` the gate yields a **warn**,
nudging authors to split large changes.

## Determinism

Given the same patch and policy, the JSON output is byte-for-byte identical
across runs and platforms:

- collections are sorted before serialization,
- floating-point results are rounded to four decimal places,
- a SHA-256 `contentHash` fingerprints the change for cache keys and tests.

## Default policy

The built-in default is defined in `internal/policy/policy.go` (`Default()`).
It is intentionally conservative so the tool is useful with no configuration.
Override only the fields you need.
