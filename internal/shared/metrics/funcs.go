// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package metrics

import (
	"math"
)

// Abstractness is the ratio of named interfaces to relevant named types.
func Abstractness(namedInterfaceTypes, totalRelevantNamedTypes int) MetricResult {
	if totalRelevantNamedTypes == zero {
		return notApplicable(
			MetricAbstractness,
			"package declares no relevant named types",
		)
	}

	ratio := float64(namedInterfaceTypes) / float64(totalRelevantNamedTypes)

	return applicable(MetricAbstractness, DefinitionAbstractness, ratio)
}

// Distance is |A + I − 1| when both abstractness and instability apply.
func Distance(abstractness, instability ResultView) MetricResult {
	if !abstractness.ResultApplicable() {
		return notApplicable(
			MetricDistance,
			"abstractness is not applicable: "+abstractness.ResultReason(),
		)
	}

	return distanceFromInstability(abstractness, instability)
}

// Instability is Ce / (Ca + Ce), or 0 for packages with no in-scope dependencies.
func Instability(afferent, efferent int) MetricResult {
	if afferent+efferent == zero {
		result := applicable(MetricInstability, DefinitionInstability, float64(zero))

		result.Reason = "package has no dependencies in scope (isolated); defined as 0"

		return result
	}

	ratio := float64(efferent) / float64(afferent+efferent)

	return applicable(MetricInstability, DefinitionInstability, ratio)
}

// ReportedMetricOrder is the stable display order for package metrics.
func ReportedMetricOrder() []string {
	return []string{MetricAbstractness, MetricInstability, MetricDistance}
}

// ResultApplicable reports whether the metric produced a numeric value.
func (result *MetricResult) ResultApplicable() bool {
	return result.Applicable
}

// ResultReason returns the not-applicable or isolated-package explanation.
func (result *MetricResult) ResultReason() string {
	return result.Reason
}

// ResultValue returns the computed metric value.
func (result *MetricResult) ResultValue() float64 {
	return result.Value
}

func applicable(name, definition string, value float64) MetricResult {
	return MetricResult{
		Name:       name,
		Scope:      ScopePackage,
		Value:      value,
		Applicable: true,
		Definition: definition,
	}
}

func distanceFromInstability(abstractness, instability ResultView) MetricResult {
	if !instability.ResultApplicable() {
		return notApplicable(
			MetricDistance,
			"instability is not applicable: "+instability.ResultReason(),
		)
	}

	value := math.Abs(abstractness.ResultValue() + instability.ResultValue() - mainSequenceBalance)

	return applicable(MetricDistance, DefinitionDistance, value)
}

func notApplicable(name, reason string) MetricResult {
	definition := DefinitionAbstractness

	if name == MetricDistance {
		definition = DefinitionDistance
	}

	return MetricResult{
		Name:       name,
		Scope:      ScopePackage,
		Applicable: false,
		Reason:     reason,
		Definition: definition,
	}
}
