// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	"strings"

	typefacts "github.com/gostafa/distance/internal/features/typefacts/domain"
)

// BuildDependencyGraph computes Ca/Ce for each package under scope.
func BuildDependencyGraph(facts *typefacts.ProjectFacts, scope string) DependencyGraph {
	graph := DependencyGraph{
		Afferents: emptyCounts(len(facts.Packages)),
		Efferents: emptyCounts(len(facts.Packages)),
		scope:     normalizeScope(facts.ModulePath, scope),
	}

	fillGraph(&graph, facts)

	return graph
}

// CountTypes tallies named types (and interfaces) for the package at pkgID.
func CountTypes(facts *typefacts.ProjectFacts, pkgID int) (interfaces, total int) {
	pkg := &facts.Packages[pkgID]

	for i := range pkg.TypeIDs {
		total++

		if facts.Types[pkg.TypeIDs[i]].Kind == typefacts.KindInterface {
			interfaces++
		}
	}

	return interfaces, total
}

// PackageCoupling returns the dependency counts for packageID.
func (graph *DependencyGraph) PackageCoupling(packageID int) (afferent, efferent int) {
	return graph.Afferents[packageID], graph.Efferents[packageID]
}

func bumpAfferent(graph *DependencyGraph, facts *typefacts.ProjectFacts, path string) {
	target, ok := packageIndex(facts, path)

	if !ok {
		return
	}

	graph.Afferents[target]++
}

func emptyCounts(count int) []int {
	values := make([]int, zero, count)

	for range count {
		values = append(values, zero)
	}

	return values
}

func fillGraph(graph *DependencyGraph, facts *typefacts.ProjectFacts) {
	for pkgID := range facts.Packages {
		recordPackage(graph, facts, pkgID)
	}
}

func importInScope(facts *typefacts.ProjectFacts, path, scope string) bool {
	if scope == ScopeAll {
		return true
	}

	if scope == ScopeModule {
		return path == facts.ModulePath || strings.HasPrefix(path, facts.ModulePath+"/")
	}

	index, found := packageIndex(facts, path)

	return found && index >= zero
}

func normalizeScope(modulePath, scope string) string {
	if modulePath == "" && scope == ScopeModule {
		return ScopeProject
	}

	return scope
}

func packageIndex(facts *typefacts.ProjectFacts, path string) (int, bool) {
	for i := range facts.Packages {
		if facts.Packages[i].Path == path {
			return i, true
		}
	}

	return zero, false
}

func recordPackage(graph *DependencyGraph, facts *typefacts.ProjectFacts, pkgID int) {
	for i := range facts.Packages[pkgID].Imports {
		path := facts.Packages[pkgID].Imports[i]

		bumpAfferent(graph, facts, path)

		if importInScope(facts, path, graph.scope) {
			graph.Efferents[pkgID]++
		}
	}
}
