# JSON report (sketch)

```json
{
  "version": 1,
  "verdict": "pass" | "fail",
  "findings": [
    {"rule": "max-function-length", "path": "internal/diff/diff.go",
     "line": 42, "severity": "warn"}
  ]
}
```

## Open questions

- Do we emit findings with score 0, or omit them?
- Should the text report be generated from this, or stay hand-written?
