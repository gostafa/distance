package application

import (
	"testing"

	"github.com/gostafa/distance/internal/features/packagemetrics/domain"
	typefacts "github.com/gostafa/distance/internal/features/typefacts/domain"
	"github.com/gostafa/distance/internal/shared/metrics"
)

type couplingGraph []domain.Coupling

func (g couplingGraph) Coupling(packageID int) domain.Coupling {
	return g[packageID]
}

func assertValue(t *testing.T, r metrics.MetricResult, want float64) {
	t.Helper()

	if !r.Applicable {
		t.Fatalf("%s not applicable (%s), want %v", r.Name, r.Reason, want)
	}

	if r.Value != want {
		t.Fatalf("%s = %v, want %v", r.Name, r.Value, want)
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
		{Efferent: 1},
		{Afferent: 1},
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
