# distance

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

# Fail when any loaded package's distance exceeds 0.3.
# distance --check --max-distance=0.3 ./internal/...
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
* `--check` — enforce `--max-distance` (default 0.5) and exit 3 on violations
* `--max-distance=0.5` — maximum package distance; applies to every positional pattern

Reports include `abstractness`, `instability`, and `distance`. Policy is a
list of package-path patterns, each with a max distance. The first matching
rule wins. Abstractness and instability cannot be selected or gated.

### Build from source

```bash
git clone https://github.com/gostafa/distance.git
cd distance

go build -o ./bin/distance ./cmd/distance
./bin/distance
```

## Use as a golangci-lint plugin

The plugin must be included in a custom `golangci-lint` binary.

Create `.custom-gcl.yml`:

```yaml
version: v2.12.2
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
          packages:
            - pattern: ./...
              max-distance: 0.5
```

Build and run the custom linter:

```bash
golangci-lint custom -v
./bin/custom-golangci-lint run ./...
```

Always run the generated `custom-golangci-lint` binary. The standard
`golangci-lint` binary does not include the plugin.

## Exit codes

* `0`: success
* `1`: analysis or write error
* `2`: command usage error
* `3`: policy violation
