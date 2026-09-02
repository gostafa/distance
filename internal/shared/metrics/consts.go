// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package metrics

const (
	// MetricAbstractness is the reported name for package abstractness.
	MetricAbstractness = "abstractness"
	// DefinitionAbstractness identifies the abstractness formula version.
	DefinitionAbstractness = "distance/abstractness-v1"
	// MetricDistance is the reported name for package distance.
	MetricDistance = "distance"
	// DefinitionDistance identifies the distance formula version.
	DefinitionDistance = "distance/distance-v1"
	// MetricInstability is the reported name for package instability.
	MetricInstability = "instability"
	// DefinitionInstability identifies the instability formula version.
	DefinitionInstability = "distance/instability-v1"

	// ScopeType marks a metric computed per named type.
	ScopeType = "type"
	// ScopePackage marks a metric computed once per package.
	ScopePackage = "package"

	zero                = 0
	mainSequenceBalance = 1
)
