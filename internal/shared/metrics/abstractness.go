package metrics

// MetricAbstractness is the metric name for package Abstractness.
const MetricAbstractness = "abstractness"

// DefinitionAbstractness versions the Abstractness formula.
const DefinitionAbstractness = "distance/abstractness-v1"

// Abstractness computes package abstractness:
//
//	Abstractness = namedInterfaceTypes / totalRelevantNamedTypes
//
// Type aliases are excluded from both counts. Not applicable when the
// package declares no relevant named types.
func Abstractness(namedInterfaceTypes, totalRelevantNamedTypes int) MetricResult {
	if totalRelevantNamedTypes == 0 {
		return notApplicable(MetricAbstractness, ScopePackage, DefinitionAbstractness,
			"package declares no relevant named types")
	}

	return applicable(MetricAbstractness, ScopePackage, DefinitionAbstractness,
		float64(namedInterfaceTypes)/float64(totalRelevantNamedTypes))
}
