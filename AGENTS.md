## Learned User Preferences
- Fix lint issues by changing code, not by adding `nolint` directives.
- Do not modify `.golangci.yml` when fixing lint failures unless explicitly asked.
- For large lint or fix campaigns, prefer a plan with at least 10 concrete steps.
- Prefer Go packages organized by kind into `consts.go`, `funcs.go`, `types.go`, `vars.go`, and `*_test.go` only.
- Prefer removing unused and duplicated code, logic, docs, and API surface.
- Prefer path-pattern based package policy (patterns with minimum distance) over counting vars/consts/funcs/types.

## Learned Workspace Facts
- Module path is `github.com/gostafa/distance`; it reports distance from the main sequence plus abstractness and instability.
- Abstractness and instability appear in reports but are not independently selectable or gateable.
- The tool runs as a standalone CLI (`cmd/distance`) and as a plugin inside a custom `golangci-lint` binary.
- Lint is driven via `taskotter:golangci-lint`; issue dumps are often captured in `.lint`.
