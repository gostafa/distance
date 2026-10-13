// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

const (
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

	// FormatText selects terminal text.
	FormatText = "text"

	// FormatJSON selects schema version 6 JSON.
	FormatJSON = "json"

	// FormatCSV selects tabular CSV.
	FormatCSV = "csv"

	// FormatWeb selects the standalone HTML report.
	FormatWeb = "web"
)
