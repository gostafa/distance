# 21. Pattern list is the only policy; size counts are gone

Date: 2026-09-02

## Status

Accepted (matching, `--max-distance`, and load-from-rules superseded by ADR 0023)

## Context

ADR 0018 introduced a modularity policy gate over structural facts (types,
funcs, vars, consts, fields, methods) and every metric. ADR 0020 then split
this repository into a single-public-metric distance linter, but the policy
surface and the public report still carried size counts and declaration lists
that no longer feed a reported metric.

Callers configured `package.types`, `funcs.lines`, `--max=key=value`, and
`--min=key=value`. Those knobs implied a size-budget linter. Distance is
`|A + I − 1|`; the only gate that matches the product is a maximum on that
number, scoped by package path.

## Decision

Replace the structural policy surface with a first-match list of package-path
rules. Each rule is a `go list` pattern plus a maximum distance.

* Empty `packages` means `[{pattern: "./...", max-distance: 0.5}]`.
* Load patterns are the union of the rule `pattern` fields.
* The first matching rule in list order wins. Packages that match no rule
  are not gated.
* The bound is a maximum: fail when `distance > threshold`.
* CLI `--max`/`--min` key=value flags are gone. `--check` with optional
  `--max-distance` (default 0.5) applies that bound to every positional
  pattern.

Drop vars/consts/funcs/types **counts and declaration lists** from the public
report, JSON/CSV/HTML, and from extraction that existed only to feed them.
Keep Ca/Ce on the package report. Keep named-type **kind** internally so
abstractness can still count interfaces. Bump the report schema to version 5.

This ADR supersedes the structural keys of ADR 0018 for this linter. ADRs
0018 and 0020 are left as inherited history and are not rewritten.

## Consequences

golangci settings and the CLI share one policy shape. Reports are smaller and
stable. Extraction no longer walks function bodies or field lists. Callers
who gated on size counts must move those checks to another tool.
