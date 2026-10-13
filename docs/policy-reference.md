# Policy file reference (sketch)

A proposal for discussion, not a finished page.

## Shape

    version: 1
    thresholds:
      coverage: 0.80
      repetition: 0.35
    rules:
      - name: max-function-length
        limit: 60
        severity: warn
      - name: error-returns-context
        severity: error

## Open questions

- Should `severity: warn` be able to fail a gate, or only `error`?
- Do we allow per-repository overrides, or is a policy file always repo-local?
- What is the exit code for a policy that fails to parse, as opposed to one that runs?
