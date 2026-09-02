# 20. Split into a single-public-metric distance linter

Date: 2026-09-02

## Status

Superseded in part by [ADR 0022](0022-report-abstractness-and-instability.md):
abstractness and instability are now reported. They remain not selectable or
gateable.

## Context

`github.com/gostafa/modularity` computed eight metrics across two natural
scopes: type-level (`amc`, `lcom`, `tcc`, `cbo`, `reusability`) and
package-level (`abstractness`, `instability`, `distance`). Callers could
select any subset. The metric dependency graph already drew the line:
`distance → {abstractness, instability}`.

The combined tool asked users to choose among internals they should not have
to know. The package-level linter's public contract is one number: how far a
package sits from the main sequence.

This repository is a full copy of that tree, not a shared core. Zero coupling
between the two linters is the point. ADRs 0001–0019 are inherited history
from the combined tool.

## Decision

Ship `github.com/gostafa/distance` as a standalone linter whose only reported,
selectable, and gateable metric is `distance`.

* Abstractness and instability remain in the compute closure and may appear
  in docs prose as inputs to the formula. They are not columns, flags, or
  policy keys. A config naming them is rejected as an unknown policy metric.
* The CLI has no `--metrics` flag. The pipeline display set is hardcoded to
  `{distance}`.
* Default policy gates `distance <= 0.5` plus the structural limits that still
  apply. Type-level metric bounds and cyclomatic extraction are gone.
* The facade package, binary, and golangci plugin are all named `distance`.

## Consequences

Reports, JSON/CSV/HTML, and the golangci plugin all expose one metric column.
Hidden inputs stay testable as internals. The combined `modularity` module is
unchanged by this split.
