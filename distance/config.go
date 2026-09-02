package distance

import (
	"errors"
	"fmt"
	"slices"

	"github.com/gostafa/distance/internal/shared/metrics"
)

// DependencyScope selects which import edges count toward package coupling.
type DependencyScope string

const (
	// DependencyScopeProject counts only imports of other analyzed packages.
	DependencyScopeProject DependencyScope = "project"
	// DependencyScopeModule counts imports of packages in the same module.
	DependencyScopeModule DependencyScope = "module"
	// DependencyScopeAll counts every import, including external modules and
	// the standard library.
	DependencyScopeAll DependencyScope = "all"
)

// MetricName identifies a reported metric.
type MetricName string

const (
	// MetricAbstractness is the package interface ratio A.
	MetricAbstractness MetricName = metrics.MetricAbstractness
	// MetricInstability is the package coupling ratio I = Ce/(Ca+Ce).
	MetricInstability MetricName = metrics.MetricInstability
	// MetricDistance is a package's distance from the main sequence,
	// |A + I - 1|. Abstractness and instability are reported beside it
	// but are not selectable or gateable on their own.
	MetricDistance MetricName = metrics.MetricDistance
)

// AllMetrics returns every reported metric name, in column order.
func AllMetrics() []MetricName {
	names := metrics.ReportedMetricOrder()
	out := make([]MetricName, len(names))
	for i, name := range names {
		out[i] = MetricName(name)
	}

	return out
}

// Config controls an analysis run. The zero value is usable: defaults are
// applied by Analyze (pattern "./..." and module dependency scope).
type Config struct {
	// Directory is the working directory for package loading. Empty means the
	// process working directory.
	Directory string
	// Patterns are the package patterns to analyze. Empty means ["./..."].
	Patterns []string
	// IncludeTests also analyzes test files and test packages.
	IncludeTests bool
	// IncludeGenerated also analyzes files carrying the standard
	// "Code generated … DO NOT EDIT." marker.
	IncludeGenerated bool
	// BuildTags are extra build tags for package loading.
	BuildTags []string
	// Workers bounds analysis concurrency. Zero or negative means
	// min(GOMAXPROCS, packageCount).
	Workers int
	// DependencyScope selects the import edges counted by package coupling.
	// Empty means DependencyScopeModule.
	DependencyScope DependencyScope
	// ContinueOnError proceeds past packages that fail to load or type-check.
	ContinueOnError bool
}

// configWithDefaults returns a copy of the config with every empty knob
// replaced by its documented default.
func configWithDefaults(c Config) Config {
	if len(c.Patterns) == 0 {
		c.Patterns = []string{"./..."}
	}

	if c.DependencyScope == "" {
		c.DependencyScope = DependencyScopeModule
	}

	return c
}

// validateConfig checks a defaults-applied config.
func validateConfig(c Config) error {
	switch c.DependencyScope {
	case DependencyScopeProject, DependencyScopeModule, DependencyScopeAll:
	default:
		return fmt.Errorf(
			"invalid dependency scope %q (want project, module, or all)",
			c.DependencyScope,
		)
	}

	if slices.Contains(c.Patterns, "") {
		return errors.New("empty package pattern")
	}

	return nil
}
