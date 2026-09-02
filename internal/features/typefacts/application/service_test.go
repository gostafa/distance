// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	"context"
	"errors"
	"testing"

	"github.com/gostafa/distance/internal/features/typefacts/domain"
	"github.com/gostafa/distance/internal/features/typefacts/domain/model"
	"github.com/gostafa/distance/internal/features/typefacts/ports/outbound"
)

func TestAssembleOrderingAndIDs(t *testing.T) {
	extracts := []model.PackageExtract{
		{
			Path: "example.com/m/zeta",
			Types: []model.TypeName{
				{Name: "B", Kind: domain.KindOther},
				{Name: "A", Kind: domain.KindInterface},
			},
			Imports: []string{"fmt", "example.com/m/alpha", "fmt", "example.com/m/zeta"},
		},
		{
			Path:     "example.com/m/alpha",
			InModule: true,
			Types: []model.TypeName{{
				Name: "A",
				Kind: domain.KindStruct,
			}},
		},
	}

	facts := Assemble("example.com/m", extracts)

	if facts.ModulePath != "example.com/m" {
		t.Fatalf("module = %q", facts.ModulePath)
	}

	if len(facts.Packages) != 2 || facts.Packages[0].Path != "example.com/m/alpha" {
		t.Fatalf("packages not sorted by path: %+v", facts.Packages)
	}

	for i, pkg := range facts.Packages {
		if pkg.ID != i {
			t.Fatalf("package ID %d at index %d", pkg.ID, i)
		}
	}

	// Types globally sorted by (package path, name) with dense IDs.
	wantNames := []string{"A", "A", "B"}

	for i, typ := range facts.Types {
		if typ.ID != i || typ.Name != wantNames[i] {
			t.Fatalf("types[%d] = {ID:%d Name:%q}, want {ID:%d Name:%q}",
				i, typ.ID, typ.Name, i, wantNames[i])
		}
	}

	if facts.Types[0].Kind != domain.KindStruct {
		t.Fatalf("type kind was not preserved: %+v", facts.Types[0])
	}

	// Imports deduplicated, sorted, self-import removed.
	zeta := facts.Packages[1]

	if len(zeta.Imports) != 2 || zeta.Imports[0] != "example.com/m/alpha" ||
		zeta.Imports[1] != "fmt" {

		t.Fatalf("zeta.Imports = %v", zeta.Imports)
	}

	if len(zeta.TypeIDs) != 2 || facts.Types[zeta.TypeIDs[0]].Name != "A" {
		t.Fatalf("zeta.TypeIDs = %v", zeta.TypeIDs)
	}
}

type errSource struct{ err error }

func (s errSource) Load(
	context.Context,
	*outbound.FactOptions,
) (string, []model.PackageExtract, error) {
	return "", nil, s.err
}

func TestCollectPropagatesLoadError(t *testing.T) {
	sentinel := errors.New("load failed")
	_, err := NewService(
		errSource{err: sentinel},
	).Collect(t.Context(), &outbound.FactOptions{})

	if !errors.Is(err, sentinel) {
		t.Fatalf("Collect error = %v, want sentinel", err)
	}
}

func TestSortedUniqueSelfOnly(t *testing.T) {
	facts := Assemble("example.com/m", []model.PackageExtract{{
		Path:    "example.com/m/p",
		Imports: []string{"example.com/m/p"},
		Types:   []model.TypeName{{Name: "T"}},
	}})

	if imports := facts.Packages[0].Imports; imports != nil {
		t.Fatalf("Imports = %v, want nil", imports)
	}
}
