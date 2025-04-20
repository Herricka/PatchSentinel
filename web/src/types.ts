/**
 * Type definitions mirroring the deterministic JSON emitted by the
 * `patchsentinel` Go CLI (schema version 1). Keeping these in one place lets
 * the renderer and formatter share a single source of truth.
 *
 * This module has zero runtime dependencies.
 */

export type Verdict = "pass" | "warn" | "fail";
export type Severity = "info" | "warn" | "error";

export interface Summary {
  filesChanged: number;
  linesAdded: number;
  linesRemoved: number;
  productionAdded: number;
  testAdded: number;
  testToCodeRatio: number;
  riskScore: number;
  repetitionRatio: number;
  sensitiveFiles: number;
  reviewMinutes: number;
  contentHash: string;
}

export interface FileStat {
  path: string;
  oldPath?: string;
  added: number;
  removed: number;
  isTest: boolean;
  isSensitive: boolean;
  isBinary: boolean;
  status: "added" | "deleted" | "renamed" | "modified";
}

export interface RiskHit {
  patternId: string;
  message: string;
  weight: number;
  path: string;
  snippet: string;
  count: number;
}

export interface Finding {
  id: string;
  severity: Severity;
  title: string;
  detail: string;
}

export interface Report {
  schemaVersion: number;
  tool: string;
  policyVersion: number;
  summary: Summary;
  files: FileStat[];
  risks: RiskHit[];
  findings: Finding[];
  verdict: Verdict;
  exitCode: number;
}

/**
 * Narrowing runtime guard used by the CLI entrypoints so that malformed JSON
 * produces a clear error instead of an opaque `undefined` crash later.
 */
export function isReport(value: unknown): value is Report {
  if (typeof value !== "object" || value === null) {
    return false;
  }
  const v = value as Record<string, unknown>;
  return (
    typeof v.schemaVersion === "number" &&
    typeof v.verdict === "string" &&
    typeof v.summary === "object" &&
    Array.isArray(v.files) &&
    Array.isArray(v.risks) &&
    Array.isArray(v.findings)
  );
}
