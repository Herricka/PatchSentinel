/**
 * Small, dependency-free helpers shared by the HTML renderer and the
 * GitHub-style formatter.
 */

/** Escape a string for safe interpolation into HTML text/attributes. */
export function escapeHtml(input: string): string {
  return input
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#39;");
}

/** Escape a pipe so Markdown table cells are not broken. */
export function escapePipe(input: string): string {
  return input.replace(/\|/g, "\\|");
}

/** Format a ratio (0..1+) to two decimals. */
export function ratio(n: number): string {
  return n.toFixed(2);
}

/** Round minutes to a whole number for display. */
export function minutes(n: number): string {
  return `${Math.round(n)} min`;
}

/** Map a verdict to a display label and colour token. */
export function verdictMeta(verdict: string): { label: string; color: string; emoji: string } {
  switch (verdict) {
    case "pass":
      return { label: "PASS", color: "#1a7f37", emoji: "✅" };
    case "warn":
      return { label: "WARN", color: "#9a6700", emoji: "⚠️" };
    default:
      return { label: "FAIL", color: "#cf222e", emoji: "❌" };
  }
}

/** Map a severity to a display emoji. */
export function severityEmoji(severity: string): string {
  switch (severity) {
    case "error":
      return "❌";
    case "warn":
      return "⚠️";
    default:
      return "ℹ️";
  }
}


