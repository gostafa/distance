// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package distance

type (
	// DependencyScope selects which imports count toward efferent coupling.
	DependencyScope = string

	// MetricName identifies a reported package metric.
	MetricName = string
)

const (

	// DependencyScopeProject counts only imports of other analyzed packages.
	DependencyScopeProject DependencyScope = "project"
	// DependencyScopeModule counts imports of packages in the same module.
	DependencyScopeModule DependencyScope = "module"
	// DependencyScopeAll counts imports of packages in the same module.
	DependencyScopeAll DependencyScope = "all"

	// MetricAbstractness is the package interface ratio A.
	MetricAbstractness MetricName = "abstractness"
	// MetricInstability is the package coupling ratio I = Ce/(Ca+Ce).
	MetricInstability MetricName = "instability"
	// MetricDistance is a package's distance from the main sequence.
	MetricDistance MetricName = "distance"

	// DefinitionAbstractness identifies the abstractness formula version.
	DefinitionAbstractness = "distance/abstractness-v1"
	// DefinitionDistance identifies the distance formula version.
	DefinitionDistance = "distance/distance-v1"
	// DefinitionInstability identifies the instability formula version.
	DefinitionInstability = "distance/instability-v1"

	// ScopePackage marks a metric computed once per package.
	ScopePackage = "package"

	// SchemaVersion identifies the JSON/text report schema this build emits.
	SchemaVersion = "6"

	allPackagesPattern = "./..."
	emptyString        = ""

	errWrapAnalyze = "distance Analyze: %w"

	zero = 0
)
