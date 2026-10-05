# ADR-0003: Refine the private-URL scan

- Status: Accepted
- Date: 2026-10-05
- Decision-log entries: A-06, D-01

## Context

The inherited CI job scans every file for patterns that suggest private hostnames. It matches the
word `internal` or `intranet` followed by a dot, `corp` surrounded by dots, and `localhost`
followed by a colon and a digit.

The planned Go layout uses `internal/` packages, and documentation for local development uses
`localhost` URLs with ports. Both trigger the scan even though they disclose nothing private.

## Decision

- Keep the Go `internal/` package layout.
- Refine the scan so it targets hostnames in URL or host contexts, not path segments or
  identifiers.
- Allow documented local development URLs that point to `localhost` or `127.0.0.1` where
  appropriate.
- Exceptional matches that are not private hostnames are listed in a reviewed allowlist file,
  `.github/private-url-allowlist.txt`, owned by the maintainers through CODEOWNERS. Each entry is
  a fixed string with a justification comment.
- Implemented in Phase 2 (F-02), before the Go and frontend scaffolding is added.

## Consequences

- The scan keeps catching private hostnames in URLs, while allowing Go packages and local
  development documentation.
- A narrower scan can miss private hostnames written without a URL context. Code review remains
  responsible for those cases.
- Until the refinement is merged, Phase 1 documents avoid the patterns that the current scan
  rejects.

## Alternatives considered

- **Rename `internal/` packages.** Rejected: it removes Go's compiler-enforced encapsulation.
- **Exclude source directories from the scan.** Rejected: source code is where accidental leaks of
  private hostnames are most likely.
