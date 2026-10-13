// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package distance

const (
	// DependencyScopeProject counts only imports of other analyzed packages.
	DependencyScopeProject = "project"

	// DependencyScopeModule counts imports of packages in the same module.
	DependencyScopeModule = "module"

	// DependencyScopeAll counts every import, including standard-library and external packages.
	DependencyScopeAll = "all"

	// MetricAbstractness is the package interface ratio A.
	MetricAbstractness = "abstractness"

	// MetricInstability is the package coupling ratio I = Ce/(Ca+Ce).
	MetricInstability = "instability"

	// MetricDistance is a package's distance from the main sequence.
	MetricDistance = "distance"

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

	emptyString = ""

	zero = 0

	errPublicDetails = "%w: %s"

	errContextAnalysis = "distance: %w"
)
