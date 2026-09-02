package inbound

import (
	"context"
	"fmt"

	"github.com/gostafa/distance/internal/shared/metrics"
)

// Options is a fully validated, defaults-applied analysis request.
type Options struct {
	// Directory is the working directory package loading runs from.
	Directory string
	// Patterns are the package patterns to analyze (e.g. "./...").
	Patterns []string
	// IncludeTests also analyzes test files and test packages.
	IncludeTests bool
	// IncludeGenerated also analyzes generated files.
	IncludeGenerated bool
	// BuildTags are extra build tags applied while loading.
	BuildTags []string
	// Workers bounds package-level concurrency; 0 selects a default.
	Workers int
	// DependencyScope is "project", "module", or "all".
	DependencyScope string
	// ContinueOnError skips packages that fail to load or type-check.
	ContinueOnError bool
}

// PackageResult carries one package's coupling facts and display metrics.
type PackageResult struct {
	// Path is the package's import path.
	Path string
	// Afferent counts analyzed packages importing this package (Ca).
	Afferent int
	// Efferent counts this package's in-scope imports (Ce).
	Efferent int
	// Metrics holds the package's display metrics in the fixed order.
	Metrics []metrics.MetricResult
}

// String summarizes the package result for debugging.
func (p PackageResult) String() string {
	return fmt.Sprintf("%s: %d package metrics", p.Path, len(p.Metrics))
}

// Result is a deterministic analysis outcome: packages sorted by import
// path, metrics in the fixed order.
type Result struct {
	// ModulePath is the analyzed main module's path, when known.
	ModulePath string
	// Packages are the analyzed packages, sorted by import path.
	Packages []PackageResult
}

// Analyzer runs the full analysis pipeline. It implements no metric
// formula itself.
type Analyzer interface {
	Analyze(ctx context.Context, opts Options) (Result, error)
}
