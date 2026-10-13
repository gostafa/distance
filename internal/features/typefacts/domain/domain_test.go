// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	"strings"
	"testing"

	"github.com/gostafa/distance/internal/features/typefacts/domain/model"
)

// White-box: the debug Stringers stay informative and panic-free.
func TestStringers(t *testing.T) {
	t.Parallel()

	pf := &ProjectFacts{
		ModulePath: "m",
		Packages:   make([]PackageFacts, 2),
		Types:      make([]TypeFacts, 3),
	}

	if s := ProjectFactsString(
		pf,
	); !strings.Contains(s, "2 packages") ||
		!strings.Contains(s, "3 types") {
		t.Errorf("ProjectFactsString = %q", s)
	}

	tf := &TypeFacts{Name: "W", Kind: model.KindInterface}

	if s := FormatTypeFacts(tf); !strings.Contains(s, `"W"`) {
		t.Errorf("FormatTypeFacts = %q", s)
	}

	if s := DumpFacts(pf); !strings.Contains(s, "2 packages") {
		t.Errorf("DumpFacts = %q", s)
	}

	named := model.FormatNamed("E", model.KindOther)

	if !strings.Contains(named, `"E"`) {
		t.Errorf("FormatNamed = %q", named)
	}
}

func TestAssembleOrderingAndIDs(t *testing.T) {
	extracts := []model.PackageExtract{
		{
			Path: "example.com/m/zeta",
			Types: []model.TypeName{
				{Name: "B", Kind: model.KindOther},
				{Name: "A", Kind: model.KindInterface},
			},
			Imports: []string{"fmt", "example.com/m/alpha", "fmt", "example.com/m/zeta"},
		},
		{
			Path:     "example.com/m/alpha",
			InModule: true,
			Types: []model.TypeName{{
				Name: "A",
				Kind: model.KindStruct,
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

	if facts.Types[0].Kind != model.KindStruct {
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

func TestEmptyDeduplication(t *testing.T) {
	if got := dedupSorted([]string{}); got != nil {
		t.Fatalf("empty imports = %v", got)
	}
}
