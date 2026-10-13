// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package public

import (
	"github.com/gostafa/distance/internal/features/packagemetrics/domain"
)

// Prepare captures a read-only graph and returns an independent computation for each package index.
// Indices must refer to facts.Packages; callers may execute different indices concurrently.
func Prepare(facts *ProjectFacts, scope string) func(int) PackageMetrics {
	graph := domain.BuildDependencyGraph(facts, scope)

	return func(index int) PackageMetrics { return computePackage(facts, &graph, index) }
}

func computePackage(facts *ProjectFacts, graph *domain.DependencyGraph, index int) PackageMetrics {
	interfaces, total := domain.CountTypes(facts, index)
	afferent, efferent := domain.PackageCoupling(graph, index)

	return PackageMetrics{
		Metrics: domain.Compute(
			&struct{ Interfaces, Total, Afferent, Efferent int }{
				interfaces,
				total,
				afferent,
				efferent,
			},
		),
		Afferent: afferent,
		Efferent: efferent,
	}
}
