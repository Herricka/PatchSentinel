/**
 * PatchSentinel report CLI (TypeScript / Node, zero dependencies).
 *
 * Reads a PatchSentinel JSON report (produced by the Go CLI) from a file or
 * stdin and renders it as either a self-contained HTML page, a GitHub check
 * summary (Markdown), or GitHub Actions annotations.
 *
 * Usage:
 *   node dist/cli.js --format html   [--in report.json] [--out report.html]
 *   node dist/cli.js --format github [--in report.json]
 *   node dist/cli.js --format annotations
 *   patchsentinel -format json <diff> | node dist/cli.js --format html
 */

import { readFileSync, writeFileSync } from "node:fs";
import { pathToFileURL } from "node:url";
import { isReport, Report } from "./types.js";
import { renderHtml } from "./htmlReport.js";
import { renderCheckSummary, renderAnnotations } from "./githubCheck.js";

interface Options {
  format: "html" | "github" | "annotations";
  input?: string;
  output?: string;
}

export function parseArgs(argv: string[]): Options {
  const opts: Options = { format: "html" };
  for (let i = 0; i < argv.length; i++) {
    const arg = argv[i];
    switch (arg) {
      case "--format":
      case "-f": {
        const value = argv[++i];
        if (value !== "html" && value !== "github" && value !== "annotations") {
          throw new Error(`unknown format: ${String(value)} (want html|github|annotations)`);
        }
        opts.format = value;
        break;
      }
      case "--in":
      case "-i":
        opts.input = argv[++i];
        break;
      case "--out":
      case "-o":
        opts.output = argv[++i];
        break;
      case "--help":
      case "-h":
        opts.format = opts.format; // no-op; handled by caller
        throw new HelpRequested();
      default:
        throw new Error(`unexpected argument: ${arg}`);
    }
  }
  return opts;
}

/** Sentinel error used to signal that usage should be printed. */
export class HelpRequested extends Error {}

export function render(report: Report, format: Options["format"]): string {
  switch (format) {
    case "html":
      return renderHtml(report);
    case "github":
      return renderCheckSummary(report);
    case "annotations":
      return renderAnnotations(report);
  }
}

function usage(): string {
  return [
    "patchsentinel-report — render PatchSentinel JSON as HTML or GitHub output",
    "",
    "Usage: node dist/cli.js [--format html|github|annotations] [--in file] [--out file]",
    "",
    "  --format, -f   output format (default: html)",
    "  --in, -i       read JSON from this file (default: stdin)",
    "  --out, -o      write output to this file (default: stdout)",
  ].join("\n");
}

function readStdin(): string {
  try {
    return readFileSync(0, "utf8");
  } catch {
    return "";
  }
}

export function main(argv: string[]): number {
  let opts: Options;
  try {
    opts = parseArgs(argv);
  } catch (err) {
    if (err instanceof HelpRequested) {
      process.stdout.write(usage() + "\n");
      return 0;
    }
    process.stderr.write(`error: ${(err as Error).message}\n`);
    return 2;
  }

  const raw = opts.input ? readFileSync(opts.input, "utf8") : readStdin();
  if (raw.trim() === "") {
    process.stderr.write("error: no input JSON provided\n");
    return 2;
  }

  let parsed: unknown;
  try {
    parsed = JSON.parse(raw);
  } catch (err) {
    process.stderr.write(`error: invalid JSON: ${(err as Error).message}\n`);
    return 2;
  }
  if (!isReport(parsed)) {
    process.stderr.write("error: input is not a valid PatchSentinel report\n");
    return 2;
  }

  const out = render(parsed, opts.format);
  if (opts.output) {
    writeFileSync(opts.output, out);
  } else {
    process.stdout.write(out);
  }
  // Mirror the report's own exit code so the TS stage can also gate CI.
  return parsed.exitCode;
}

// Execute only when run directly (not when imported by tests).
const isDirect =
  process.argv[1] !== undefined &&
  import.meta.url === pathToFileURL(process.argv[1]).href;
if (isDirect) {
  process.exit(main(process.argv.slice(2)));
}


