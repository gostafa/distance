// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	"github.com/gostafa/distance/internal/features/packagemetrics/domain"
	typefacts "github.com/gostafa/distance/internal/features/typefacts/domain"
	"github.com/gostafa/distance/internal/shared/metrics"
)

// ComputeForPackages returns one Result per package in facts.
func ComputeForPackages(facts *typefacts.ProjectFacts, graph domain.CouplingGraph) []Result {
	results := make([]Result, zero, facts.PackageCount())

	for pkgID := range facts.Packages {
		results = append(results, computeOne(facts, graph, pkgID))
	}

	return results
}

func computeOne(facts *typefacts.ProjectFacts, graph domain.CouplingGraph, pkgID int) Result {
	interfaces, total := domain.CountTypes(facts, pkgID)
	afferent, efferent := graph.PackageCoupling(pkgID)
	abstractness := metrics.Abstractness(interfaces, total)
	instability := metrics.Instability(afferent, efferent)

	return Result{
		Abstractness: abstractness,
		Instability:  instability,
		Distance:     metrics.Distance(&abstractness, &instability),
	}
}
