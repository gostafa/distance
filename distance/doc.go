// Package distance analyzes Go modules and reports each package's distance
// from the main sequence.
//
// Call Analyze with a Config to load packages, compute distance (and the
// abstractness and instability inputs it needs internally), and receive a
// deterministic Report. Config selects patterns and dependency scope. The
// supporting metrics are never reported, selectable, or gateable on their
// own.
//
// For policy enforcement via go/analysis, use the sibling analyzer package.
// To register that analyzer as a golangci-lint Module Plugin, blank-import
// the sibling plugin package.
package distance
