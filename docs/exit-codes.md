# Exit-code contract (sketch)

| Code | Meaning | Caller should |
|---|---|---|
| 0 | gate passed | ship it |
| 1 | gate failed on findings | look at the report |
| 2 | policy failed to parse | fix the config, do not retry |
| 3 | input unreadable | fix the path |

## Open questions

- Should 2 and 3 collapse into one, or stay apart for CI dashboards?
- Do we ever need a distinct code for "passed with warnings"?
