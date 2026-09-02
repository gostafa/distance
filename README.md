# distance

`distance` analyzes Go packages and reports each package's **distance from the
main sequence** (`|A + I − 1|`). Abstractness and instability are computed
internally and are not reported, selectable, or gateable on their own.

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

# Fail when policy limits are violated.
# distance --max=types=12 --max=package.distance=0.5 ./...
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
* `--max=key=value` and `--min=key=value`

The reported metric is `distance`. A config naming a hidden input (`abstractness`,
`instability`) is rejected as an unknown policy metric.

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
          patterns:
            - ./...
          tests: false
          generated: false
          dependency-scope: module

          package:
            types:
              max: 12
            funcs:
              max: 30
            metrics:
              distance:
                max: 0.5
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
