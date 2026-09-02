# 23. Glob import-path rules; load patterns are independent

Date: 2026-09-03

## Status

Accepted

## Context

ADR 0021 replaced structural size gates with a first-match list of `go list`
patterns (`./...`, `./internal/...`) plus `--max-distance`. Load patterns were
the union of those rule patterns. That coupled *what to analyze* to *what to
gate*, and first-match order made a broad `./...` rule hide a later specific
one.

go-reusability already treats load and policy as two surfaces and matches
full import paths with `*` / `**`. Distance should use the same shape; the
distance analog of reusability `min` is `max` because lower distance is
better.

ADR 0022 still holds: only distance is gateable. Abstractness and
instability remain reported, not selectable or gated.

## Decision

1. **Policy is `[]Rule{Pattern, Max}`.** Patterns are globs against the full
   import path: `*` matches one segment, `**` matches zero or more. When
   more than one rule matches, the most specific pattern wins (more literal
   segments, then fewer wildcards, then longer); exact ties use the later
   rule. Packages that match no rule are not gated.

2. **Load patterns are independent.** CLI positional args and plugin
   `patterns` (default `./...`) are passed to `go/packages`. They are not
   derived from policy rules.

3. **Exclusive max is unchanged:** fail when `distance > max` with the
   existing comparison epsilon. `Max` must be finite and in `[0, 1]`.

4. **Plugin:** empty/missing `rules` → `DefaultRules()`
   (`[{pattern: "**", max: 0.5}]`). Reject `packages` and `max-distance`.

5. **CLI:** remove `--max-distance`. Add repeatable `--rule=pattern:max`.
   `--check` requires at least one `--rule` (no silent default gate).

This ADR supersedes ADR 0021's go-list matching, first-match order,
`--max-distance`, and load-from-rules decisions. ADR 0021's removal of
size-count gates and ADR 0022's "only distance is gateable" remain.

## Consequences

- CLI and plugin share one glob-rule policy shape with reusability.
- Callers must pass `--rule` with `--check`; `--check` alone is a usage
  error.
- Plugin configs migrate from `packages: [{pattern, max-distance}]` to
  `rules: [{pattern, max}]` plus optional `patterns`.
- A later, equally specific rule overrides an earlier one; a more specific
  glob can tighten or loosen the baseline without list-order tricks.
