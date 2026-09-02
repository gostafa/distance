// Package domain contains the pure package-distance policy model.
//
// A Policy is an ordered list of package-path rules. Each rule is a go list
// pattern plus a maximum distance. Evaluate applies the first matching rule
// and fails when a package's applicable distance exceeds that bound.
// Packages that match no rule are not gated. The package performs no I/O;
// adapters build a Policy from CLI flags or golangci-lint settings.
package domain
