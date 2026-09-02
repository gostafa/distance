// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package distance

import (
	"github.com/gostafa/distance/internal/shared/metrics"
)

type (
	// DependencyScope selects which imports count toward efferent coupling.
	DependencyScope string

	// MetricName identifies a reported package metric.
	MetricName string
)

const (

	// DependencyScopeProject counts only imports of other analyzed packages.
	DependencyScopeProject DependencyScope = "project"
	// DependencyScopeModule counts imports of packages in the same module.
	DependencyScopeModule DependencyScope = "module"
	// DependencyScopeAll counts every import, including external modules and
	// the standard library.
	DependencyScopeAll DependencyScope = "all"

	// MetricAbstractness is the package interface ratio A.
	MetricAbstractness MetricName = metrics.MetricAbstractness
	// MetricInstability is the package coupling ratio I = Ce/(Ca+Ce).
	MetricInstability MetricName = metrics.MetricInstability
	// MetricDistance is a package's distance from the main sequence,
	// |A + I - 1|. Abstractness and instability are reported beside it
	// but are not selectable or gateable on their own.
	MetricDistance MetricName = metrics.MetricDistance

	// SchemaVersion identifies the JSON/text report schema this build emits.
	SchemaVersion = "6"

	allPackagesPattern = "./..."

	errWrapAnalyze = "distance Analyze: %w"

	zero = 0
)
