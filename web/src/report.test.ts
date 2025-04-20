/**
 * Tests for the TypeScript renderer/formatter using Node's built-in test
 * runner (node:test) and assertions (node:assert). No third-party test deps.
 */

import { test } from "node:test";
import assert from "node:assert/strict";

import { Report, isReport } from "./types.js";
import { renderHtml } from "./htmlReport.js";
import { renderCheckSummary, renderAnnotations } from "./githubCheck.js";
import { escapeHtml, escapePipe } from "./util.js";
import { parseArgs, render, HelpRequested } from "./cli.js";

function sampleReport(verdict: Report["verdict"] = "fail"): Report {
  return {
    schemaVersion: 1,
    tool: "patchsentinel",
    policyVersion: 1,
    summary: {
      filesChanged: 2,
      linesAdded: 12,
      linesRemoved: 3,
      productionAdded: 10,
      testAdded: 2,
      testToCodeRatio: 0.2,
      riskScore: 30,
      repetitionRatio: 0.1,
      sensitiveFiles: 1,
      reviewMinutes: 8,
      contentHash: "deadbeef",
    },
    files: [
      { path: "a.go", added: 10, removed: 3, isTest: false, isSensitive: false, isBinary: false, status: "modified" },
      {
        path: ".github/workflows/ci.yml",
        added: 2,
        removed: 0,
        isTest: false,
        isSensitive: true,
        isBinary: false,
        status: "modified",
      },
    ],
    risks: [
      { patternId: "hardcoded-secret", message: "Possible secret | leak", weight: 10, path: "a.go", snippet: "key=...", count: 1 },
    ],
    findings: [
      { id: "risk-score", severity: "error", title: "Risk score exceeds maximum", detail: "score 30 >= 25" },
    ],
    verdict,
    exitCode: verdict === "fail" ? 2 : verdict === "warn" ? 1 : 0,
  };
}

test("isReport accepts valid and rejects invalid", () => {
  assert.equal(isReport(sampleReport()), true);
  assert.equal(isReport({}), false);
  assert.equal(isReport(null), false);
  assert.equal(isReport("nope"), false);
});

test("escapeHtml neutralizes injection", () => {
  assert.equal(escapeHtml(`<script>"x"&'y'</script>`), "&lt;script&gt;&quot;x&quot;&amp;&#39;y&#39;&lt;/script&gt;");
});

test("escapePipe protects markdown tables", () => {
  assert.equal(escapePipe("a|b|c"), "a\\|b\\|c");
});

test("renderHtml produces a self-contained document", () => {
  const html = renderHtml(sampleReport());
  assert.match(html, /<!DOCTYPE html>/);
  assert.match(html, /FAIL/);
  assert.match(html, /Risk score/);
  assert.match(html, /\.github\/workflows\/ci\.yml/);
  // No external resources.
  assert.equal(/src\s*=\s*"http/.test(html), false);
  assert.equal(/href\s*=\s*"http/.test(html), false);
  // Injection safety: the raw risk message pipe should appear escaped only in
  // markdown, but in HTML it is plain text; ensure no unescaped script tags.
  assert.equal(html.includes("<script>"), false);
});

test("renderHtml is deterministic", () => {
  const a = renderHtml(sampleReport());
  const b = renderHtml(sampleReport());
  assert.equal(a, b);
});

test("renderCheckSummary yields GitHub markdown", () => {
  const md = renderCheckSummary(sampleReport());
  assert.match(md, /## PatchSentinel/);
  assert.match(md, /\| Risk score \| 30 \|/);
  assert.match(md, /### Findings/);
  assert.match(md, /### Risk signals/);
  // Pipe inside a risk message must be escaped.
  assert.match(md, /Possible secret \\\| leak/);
});

test("renderAnnotations emits workflow commands", () => {
  const ann = renderAnnotations(sampleReport());
  assert.match(ann, /^::error title=Risk score exceeds maximum::/m);
});

test("renderAnnotations encodes special chars", () => {
  const r = sampleReport();
  r.findings = [{ id: "x", severity: "warn", title: "a:b,c", detail: "line1\nline2" }];
  const ann = renderAnnotations(r);
  assert.match(ann, /title=a%3Ab%2Cc/);
  assert.match(ann, /line1%0Aline2/);
  assert.match(ann, /^::warning /m);
});

test("parseArgs parses format/in/out", () => {
  const opts = parseArgs(["--format", "github", "--in", "r.json", "--out", "o.md"]);
  assert.equal(opts.format, "github");
  assert.equal(opts.input, "r.json");
  assert.equal(opts.output, "o.md");
});

test("parseArgs rejects unknown format", () => {
  assert.throws(() => parseArgs(["--format", "pdf"]), /unknown format/);
});

test("parseArgs signals help", () => {
  assert.throws(() => parseArgs(["--help"]), HelpRequested);
});

test("render dispatches by format", () => {
  const rep = sampleReport("pass");
  assert.match(render(rep, "html"), /<!DOCTYPE html>/);
  assert.match(render(rep, "github"), /## PatchSentinel/);
  assert.equal(typeof render(rep, "annotations"), "string");
});
