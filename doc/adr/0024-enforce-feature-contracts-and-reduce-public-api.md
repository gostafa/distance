# 24. Enforce Feature Contracts and Reduce the Public API

Date: 2026-10-13

## Status

Accepted. Supersedes ADRs 0002, 0003, and 0004, and the public analyzer
construction API described in ADR 0019. The CLI and plugin behavior in
ADRs 0019–0023 continues to apply.

## Context

The existing architecture guards checked only immediate domain directories
and production filenames. They permitted private cross-feature imports,
construction of concrete loaders in application code, and public wiring
packages. Several imports existed solely to affect lint metrics.

## Decision

Each feature owns its domain decisions, application orchestration, ports,
adapters, and published contracts. Cross-feature callers use `public`
contracts, with consumer-owned ports and connecting adapters where runtime
substitution matters. Published DTOs are declared independently of private
models; no public type aliases expose private implementation types.

Typefacts owns package extraction and deterministic fact assembly.
Packagemetrics owns graph coupling, abstractness, instability, distance,
and applicability decisions. Projectanalysis orchestrates collection and
metric computation. Policy validates and evaluates threshold rules.
Reporting owns report values and output orchestration; its adapters encode,
format, write, and close output streams.

Projectanalysis may call the pure packagemetrics published API directly.
This is an accepted stable in-process computation contract, with no I/O,
construction, shared mutable state, or private model dependencies.
Fact collection and worker execution remain injected through consumer ports.
The reusable worker pool is technical infrastructure. Composition code owns
construction; domain and application code never import concrete adapters.

The library facade is `github.com/gostafa/distance/distance`:

```go
func Analyze(ctx context.Context, config *Config) (Report, error)
```

Nil configuration selects defaults. Configuration and its slices are copied
before defaults are applied. Public error identities are `ErrInvalidConfig`
and `ErrAnalysis`; cancellation and deadlines retain their standard context
identities. Error chains expose no private or vendor failures.

The `plugin` package retains its blank-import registration entry point.
Its implementation type and constructor are private; exported methods on
that private type implement the required vendor interface. Analyzer
construction and settings decoding are internal adapter APIs.

Recursive guards enforce file names, declaration placement, package
documentation, feature privacy, core dependencies, public signatures, and
import cycles. Go tooling and nested domain packages are included.
Negative fixtures verify rejection behavior. CI regenerates coverage before
linting and after automatic fixes; thresholds and checks remain unchanged.

## Migration

Replace backend injection and calls to `wire.AnalyzeWithDefault` with:

```go
import "github.com/gostafa/distance/distance"

report, err := distance.Analyze(ctx, &distance.Config{
    Directory: "./project",
})
```

Remove imports of `distance/wire` and the former public `analyzer` package.
Custom golangci-lint builds continue to import
`github.com/gostafa/distance/plugin`; inline settings are unchanged.
`plugin.New`, `distance.Backend`, `distance.ToolIdent`, and
`distance.ReportString` are no longer public library operations. Use strings
in place of the removed `DependencyScope` and `MetricName` aliases; their
constants remain available.

The splitter consolidates implementation and test files and preserves generated
markers, embedded declarations, and test package visibility. It rejects build
constraints and mixed generated/ordinary declarations that cannot share the
required single category file before rewriting sources.

## Consequences

This is a breaking Go library API change. CLI flags, exit codes, metric
formulas, policy selection, report ordering, HTML escaping, and report schema
version 6 remain compatible. Boundary mappings introduce small deliberate
DTO duplication so private feature models can evolve independently.
