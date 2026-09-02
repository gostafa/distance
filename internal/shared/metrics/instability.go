package metrics

// MetricInstability is the metric name for package Instability.
const MetricInstability = "instability"

// DefinitionInstability versions the Instability formula.
const DefinitionInstability = "distance/instability-v1"

// Instability computes package instability:
//
//	Ca = analyzed packages that import this package   (afferent)
//	Ce = packages this package imports, per scope     (efferent)
//	Instability = Ce / (Ca + Ce)
//
// Duplicate and self-dependencies are ignored by the caller. An isolated
// package (Ca + Ce == 0) is defined as maximally stable: instability 0,
// with the convention noted in the result's Reason.
func Instability(afferent, efferent int) MetricResult {
	if afferent+efferent == 0 {
		result := applicable(MetricInstability, ScopePackage, DefinitionInstability, 0)
		result.Reason = "package has no dependencies in scope (isolated); defined as 0"

		return result
	}

	return applicable(MetricInstability, ScopePackage, DefinitionInstability,
		float64(efferent)/float64(afferent+efferent))
}
