package application

import (
	"github.com/gostafa/distance/internal/features/packagemetrics/domain"
	typefacts "github.com/gostafa/distance/internal/features/typefacts/domain"
	"github.com/gostafa/distance/internal/shared/metrics"
)

// Result is the architecture feature's output for one package.
type Result struct {
	// Abstractness is the package's interface ratio.
	Abstractness metrics.MetricResult
	// Instability is the package's efferent coupling ratio.
	Instability metrics.MetricResult
	// Distance is the package's distance from the main sequence.
	Distance metrics.MetricResult
}

// ComputeForPackages evaluates the package-level metrics for every analyzed
// package, indexed by package ID. The caller supplies the graph so a complete
// analysis can reuse the same coupling traversal for metrics and structure.
func ComputeForPackages(facts *typefacts.ProjectFacts, graph domain.CouplingGraph) []Result {
	results := make([]Result, len(facts.Packages))
	for i := range facts.Packages {
		counts := domain.CountTypes(facts, &facts.Packages[i])
		coupling := graph.Coupling(i)

		abstractness := metrics.Abstractness(counts.Interfaces, counts.Total)
		instability := metrics.Instability(coupling.Afferent, coupling.Efferent)
		results[i] = Result{
			Abstractness: abstractness,
			Instability:  instability,
			Distance:     metrics.Distance(abstractness, instability),
		}
	}

	return results
}
