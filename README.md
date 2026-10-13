# Distance

[![LICENSE](https://img.shields.io/github/license/gostafa/distance)](/LICENSE)
[![`Workflow for distance Action`](https://github.com/gostafa/distance/actions/workflows/main.yml/badge.svg)](https://github.com/gostafa/distance/actions/workflows/main.yml)
[![codecov](https://codecov.io/gh/gostafa/distance/graph/badge.svg)](https://codecov.io/gh/gostafa/distance)
[![CodSpeed](https://img.shields.io/endpoint?url=https://codspeed.io/badge.json)](https://app.codspeed.io/gostafa/distance?utm_source=badge)
[![Go Reference](https://pkg.go.dev/badge/github.com/gostafa/distance/distance)](https://pkg.go.dev/github.com/gostafa/distance/distance)

`distance` analyzes Go packages and reports each package's **distance from the
main sequence** (`|A + I − 1|`), plus the **abstractness** and **instability**
that feed it. Abstractness and instability are shown in reports but are not
selectable or gateable on their own.

It can run as:

* a standalone CLI;
* a plugin inside a custom `golangci-lint` binary.

## Use as a CLI

### Install

```bash
go get github.com/gostafa/distance@latest
go install github.com/gostafa/distance/cmd/distance@latest
```

### Run

```bash
distance

# Check selected packages.
# distance ./internal/...

# Write JSON or CSV.
# distance --format=json --output=report.json ./...
# distance --format=csv --output=report.csv ./...

# Open the HTML report.
# distance --web ./...

# Fail when any loaded package's distance exceeds a rule max.
# distance --check --rule='**':0.5 --rule='**/internal/**':0.2 ./...
```

Flags must come before package patterns:

```bash
distance --format=json ./...
```

Useful flags:

* `--format=text|json|csv|web`
* `--output=path`
* `--tests`
* `--generated`
* `--dependency-scope=project|module|all`
* `--continue-on-error`
* `--check` with `--rule=pattern:max` (repeatable; requires `--check`)

Reports include `abstractness`, `instability`, and `distance`. Policy gates
package distance by import-path glob. When multiple rules match, the most
specific pattern wins: more literal segments, then fewer wildcards, then
longer patterns; exact ties use the later rule. Abstractness and
instability cannot be selected or gated.

### Build from Source

```bash
git clone https://github.com/gostafa/distance.git
cd distance

go build -o ./bin/distance ./cmd/distance
./bin/distance
```

## Use as a Go Library

Import the library facade and call the default analysis entry point:

```go
import "github.com/gostafa/distance/distance"

report, err := distance.Analyze(ctx, &distance.Config{
    Directory: "./project",
    Patterns:  []string{"./..."},
})
```

Passing nil configuration selects the defaults. Analysis copies configuration
and its slices. Use `errors.Is` with `distance.ErrInvalidConfig`,
`distance.ErrAnalysis`, `context.Canceled`, or `context.DeadlineExceeded` to
classify failures. Internal and vendor error types are not exposed.

Existing callers should replace `wire.AnalyzeWithDefault` or backend-injected
`distance.Analyze` calls with the two-argument API above. The public `analyzer`
and `distance/wire` packages, `plugin.New`, `distance.Backend`,
`distance.ToolIdent`, and `distance.ReportString` have been removed.
`DependencyScope` and `MetricName` no longer define named aliases; use strings
and the retained scope and metric constants. The
[architecture decision](doc/adr/0024-enforce-feature-contracts-and-reduce-public-api.md)
documents ownership and migration.

## Use as a Golangci-Lint Plugin

The plugin must be included in a custom `golangci-lint` binary.

Create `.custom-gcl.yml`:

```yaml
version: v2.14.0
name: custom-golangci-lint
destination: ./bin
plugins:
  - module: github.com/gostafa/distance
    import: github.com/gostafa/distance/plugin
    path: .
```

Enable it in `.golangci.yml`:

```yaml
version: "2"

linters:
  default: all
  enable:
    - distance

  settings:
    custom:
      distance:
        type: module
        settings:
          tests: false
          generated: false
          dependency-scope: module
          patterns: [./...]
          rules:
            - pattern: "**"
              max: 0.5
            - pattern: "**/internal/**"
              max: 0.2
```

Build and run the custom linter:

```bash
golangci-lint custom -v
./bin/custom-golangci-lint run ./...
```

Always run the generated `custom-golangci-lint` binary. The standard
`golangci-lint` binary does not include the plugin.

## Exit Codes

* `0`: success
* `1`: analysis or write error
* `2`: command usage error
* `3`: policy violation

## Verify Repository Changes

Run these commands from the repository root. Regenerate reports after fixes
so coverage checks use the final source locations:

```sh
task taskotter:ci:fix
task taskotter:ci
task taskotter:go-junit-report:report
go test -race ./...
go build ./...
```

The test suite includes recursive architecture checks and negative fixtures.
Go packages use the filenames and declaration categories specified in
`AGENTS.md`, including package-named tests and documented `doc.go` files.
