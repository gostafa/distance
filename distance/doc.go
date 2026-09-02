// Package distance analyzes Go modules and reports each package's distance
// from the main sequence, plus the abstractness and instability that feed it.
//
// Call Analyze with a Config to load packages, compute abstractness,
// instability, and distance, and receive a deterministic Report. Config
// selects patterns and dependency scope. Abstractness and instability are
// reported but are not selectable or gateable on their own.
//
// For policy enforcement via go/analysis, use the sibling analyzer package.
// To register that analyzer as a golangci-lint Module Plugin, blank-import
// the sibling plugin package.
package distance
