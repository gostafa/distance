// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	"testing"

	typefacts "github.com/gostafa/distance/internal/features/typefacts/domain"
	"github.com/gostafa/distance/internal/shared/metrics"
)

type couplingGraph [][2]int

func (graph couplingGraph) PackageCoupling(packageID int) (int, int) {
	return graph[packageID][0], graph[packageID][1]
}

func assertValue(t *testing.T, result metrics.MetricResult, want float64) {
	t.Helper()

	if !result.Applicable {
		t.Fatalf("%s not applicable (%s), want %v", result.Name, result.Reason, want)
	}

	if result.Value != want {
		t.Fatalf("%s = %v, want %v", result.Name, result.Value, want)
	}
}

// White-box: a stable abstract package and an unstable concrete importer both
// land on the main sequence.
func TestComputeForPackages(t *testing.T) {
	t.Parallel()

	facts := &typefacts.ProjectFacts{
		ModulePath: "example.com/m",
		Packages: []typefacts.PackageFacts{
			{
				ID: 0, Path: "example.com/m/a", InModule: true,
				Imports: []string{"example.com/m/b"}, TypeIDs: []int{0},
			},
			{ID: 1, Path: "example.com/m/b", InModule: true, TypeIDs: []int{1}},
		},
		Types: []typefacts.TypeFacts{
			{ID: 0, PackageID: 0, Name: "A", Kind: typefacts.KindStruct},
			{ID: 1, PackageID: 1, Name: "B", Kind: typefacts.KindInterface},
		},
	}

	got := ComputeForPackages(facts, couplingGraph{
		{0, 1},
		{1, 0},
	})

	if len(got) != 2 {
		t.Fatalf("got %d results", len(got))
	}

	// a: concrete, imports b (Ce=1, Ca=0) → A=0, I=1, D=0.
	assertValue(t, got[0].Abstractness, 0)
	assertValue(t, got[0].Instability, 1)
	assertValue(t, got[0].Distance, 0)
	// b: all-interface, imported by a (Ce=0, Ca=1) → A=1, I=0, D=0.
	assertValue(t, got[1].Abstractness, 1)
	assertValue(t, got[1].Instability, 0)
	assertValue(t, got[1].Distance, 0)
}
