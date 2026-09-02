// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain_test

import (
	"testing"

	architecture "github.com/gostafa/distance/internal/features/packagemetrics/domain"
	typefacts "github.com/gostafa/distance/internal/features/typefacts/domain"
)

// Black-box: coupling counts and type tallies from the exported entry points.
func TestGraphAndCounts(t *testing.T) {
	t.Parallel()

	facts := &typefacts.ProjectFacts{
		ModulePath: "example.com/m",
		Packages: []typefacts.PackageFacts{
			{
				ID: 0, Path: "example.com/m/a", InModule: true,
				Imports: []string{"example.com/m/b"}, TypeIDs: []int{0},
			},
			{ID: 1, Path: "example.com/m/b", InModule: true, TypeIDs: []int{1, 2}},
		},
		Types: []typefacts.TypeFacts{
			{ID: 0, PackageID: 0, Kind: typefacts.KindStruct},
			{ID: 1, PackageID: 1, Kind: typefacts.KindInterface},
			{ID: 2, PackageID: 1, Kind: typefacts.KindStruct},
		},
	}

	graph := architecture.BuildDependencyGraph(facts, architecture.ScopeProject)
	_, efferent := graph.PackageCoupling(0)

	if efferent != 1 {
		t.Errorf("a efferent = %d, want 1", efferent)
	}

	afferent, _ := graph.PackageCoupling(1)

	if afferent != 1 {
		t.Errorf("b afferent = %d, want 1", afferent)
	}

	interfaces, total := architecture.CountTypes(facts, 1)

	if total != 2 || interfaces != 1 {
		t.Fatalf("b counts = interfaces=%d total=%d, want 1/2", interfaces, total)
	}
}
