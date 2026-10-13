// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	"math"
	"strings"
)

// BuildDependencyGraph computes Ca/Ce for each package under scope.
func BuildDependencyGraph(facts *ProjectFacts, scope string) DependencyGraph {
	graph := DependencyGraph{
		Afferents: emptyCounts(len(facts.Packages)),
		Efferents: emptyCounts(len(facts.Packages)),
		scope:     normalizeScope(facts.ModulePath, scope),
	}

	fillGraph(&graph, facts)

	return graph
}

// Compute returns abstractness, instability, and distance in report order.
func Compute(counts *struct{ Interfaces, Total, Afferent, Efferent int }) []MetricEntry {
	a := abstractnessMetric(counts.Interfaces, counts.Total)
	i := instabilityMetric(counts.Afferent, counts.Efferent)

	return []MetricEntry{a, i, distanceMetric(&a, &i)}
}

// CountTypes tallies named types (and interfaces) for the package at pkgID.
func CountTypes(facts *ProjectFacts, pkgID int) (interfaces, total int) {
	pkg := &facts.Packages[pkgID]

	for i := range pkg.TypeIDs {
		total++

		if facts.Types[pkg.TypeIDs[i]].Kind == KindInterface {
			interfaces++
		}
	}

	return interfaces, total
}

// PackageCoupling returns the dependency counts for packageID.
func PackageCoupling(graph *DependencyGraph, packageID int) (afferent, efferent int) {
	return graph.Afferents[packageID], graph.Efferents[packageID]
}

func abstractnessMetric(namedInterfaceTypes, totalRelevantNamedTypes int) MetricEntry {
	if totalRelevantNamedTypes == zero {
		return notApplicableMetric(
			MetricAbstractness,
			"package declares no relevant named types",
		)
	}

	ratio := float64(namedInterfaceTypes) / float64(totalRelevantNamedTypes)

	return applicableMetric(MetricAbstractness, DefinitionAbstractness, ratio)
}

func applicableMetric(name, definition string, value float64) MetricEntry {
	return MetricEntry{
		Name:       name,
		Scope:      ScopePackage,
		Value:      value,
		Applicable: true,
		Definition: definition, Reason: "",
	}
}

func bumpAfferent(graph *DependencyGraph, facts *ProjectFacts, path string) {
	target, ok := packageIndex(facts, path)

	if !ok {
		return
	}

	graph.Afferents[target]++
}

func distanceMetric(abstractness, instability *MetricEntry) MetricEntry {
	if !abstractness.Applicable {
		return notApplicableMetric(
			MetricDistance,
			"abstractness is not applicable: "+abstractness.Reason,
		)
	}

	if !instability.Applicable {
		return notApplicableMetric(
			MetricDistance,
			"instability is not applicable: "+instability.Reason,
		)
	}

	value := math.Abs(abstractness.Value + instability.Value - mainSequenceBalance)

	return applicableMetric(MetricDistance, DefinitionDistance, value)
}

func emptyCounts(count int) []int {
	values := make([]int, zero, count)

	for range count {
		values = append(values, zero)
	}

	return values
}

func fillGraph(graph *DependencyGraph, facts *ProjectFacts) {
	for pkgID := range facts.Packages {
		recordPackage(graph, facts, pkgID)
	}
}

func importInScope(facts *ProjectFacts, path, scope string) bool {
	if scope == ScopeAll {
		return true
	}

	if scope == ScopeModule {
		return path == facts.ModulePath || strings.HasPrefix(path, facts.ModulePath+"/")
	}

	index, found := packageIndex(facts, path)

	return found && index >= zero
}

func instabilityMetric(afferent, efferent int) MetricEntry {
	if afferent+efferent == zero {
		result := applicableMetric(MetricInstability, DefinitionInstability, float64(zero))

		result.Reason = "package has no dependencies in scope (isolated); defined as 0"

		return result
	}

	ratio := float64(efferent) / float64(afferent+efferent)

	return applicableMetric(MetricInstability, DefinitionInstability, ratio)
}

func normalizeScope(modulePath, scope string) string {
	if modulePath == "" && scope == ScopeModule {
		return ScopeProject
	}

	return scope
}

func notApplicableMetric(name, reason string) MetricEntry {
	definition := DefinitionAbstractness

	if name == MetricDistance {
		definition = DefinitionDistance
	}

	return MetricEntry{
		Name:       name,
		Scope:      ScopePackage,
		Applicable: false,
		Reason:     reason,
		Definition: definition, Value: zero,
	}
}

func packageIndex(facts *ProjectFacts, path string) (int, bool) {
	for i := range facts.Packages {
		if facts.Packages[i].Path == path {
			return i, true
		}
	}

	return zero, false
}

func recordPackage(graph *DependencyGraph, facts *ProjectFacts, pkgID int) {
	for i := range facts.Packages[pkgID].Imports {
		path := facts.Packages[pkgID].Imports[i]

		bumpAfferent(graph, facts, path)

		if importInScope(facts, path, graph.scope) {
			graph.Efferents[pkgID]++
		}
	}
}
