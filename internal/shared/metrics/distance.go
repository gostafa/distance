package metrics

import "math"

// MetricDistance is the metric name for Distance from the Main Sequence.
const MetricDistance = "distance"

// DefinitionDistance versions the Distance formula.
const DefinitionDistance = "distance/distance-v1"

// Distance computes a package's distance from the main sequence:
//
//	Distance = abs(Abstractness + Instability − 1)
//
// Not applicable when either input metric is not applicable.
func Distance(abstractness, instability MetricResult) MetricResult {
	if !abstractness.Applicable {
		return notApplicable(MetricDistance, ScopePackage, DefinitionDistance,
			"abstractness is not applicable: "+abstractness.Reason)
	}

	if !instability.Applicable {
		return notApplicable(MetricDistance, ScopePackage, DefinitionDistance,
			"instability is not applicable: "+instability.Reason)
	}

	return applicable(MetricDistance, ScopePackage, DefinitionDistance,
		math.Abs(abstractness.Value+instability.Value-1))
}
