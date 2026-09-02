// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	"testing"

	typefacts "github.com/gostafa/distance/internal/features/typefacts/domain"
)

func scopedFacts() *typefacts.ProjectFacts {
	return &typefacts.ProjectFacts{
		ModulePath: "example.com/m",
		Packages: []typefacts.PackageFacts{
			{
				ID: 0, Path: "example.com/m/a", InModule: true,
				Imports: []string{"example.com/m/b", "example.com/other/lib", "fmt"},
			},
			{ID: 1, Path: "example.com/m/b", InModule: true},
		},
	}
}

func TestBuildDependencyGraphScopes(t *testing.T) {
	facts := scopedFacts()

	project := BuildDependencyGraph(facts, ScopeProject)
	afferent, efferent := project.PackageCoupling(0)

	if efferent != 1 || afferent != 0 {
		t.Fatalf("project scope a = Ca=%d Ce=%d", afferent, efferent)
	}

	afferent, efferent = project.PackageCoupling(1)

	if afferent != 1 || efferent != 0 {
		t.Fatalf("project scope b = Ca=%d Ce=%d", afferent, efferent)
	}

	module := BuildDependencyGraph(facts, ScopeModule)

	_, efferent = module.PackageCoupling(0)

	if efferent != 1 {
		t.Fatalf("module scope Ce(a) = %d, want 1 (fmt and external excluded)", efferent)
	}

	all := BuildDependencyGraph(facts, ScopeAll)

	_, efferent = all.PackageCoupling(0)

	if efferent != 3 {
		t.Fatalf("all scope Ce(a) = %d, want 3", efferent)
	}

	// Afferent coupling is always measured within the analyzed set.
	afferent, _ = all.PackageCoupling(1)

	if afferent != 1 {
		t.Fatalf("all scope Ca(b) = %d, want 1", afferent)
	}
}

func TestModuleScopeWithoutModuleInfo(t *testing.T) {
	facts := scopedFacts()

	facts.ModulePath = ""

	graph := BuildDependencyGraph(facts, ScopeModule)
	_, efferent := graph.PackageCoupling(0)

	if efferent != 1 {
		t.Fatalf("Ce(a) = %d, want 1 (degrades to project scope)", efferent)
	}
}

func TestCountTypes(t *testing.T) {
	facts := &typefacts.ProjectFacts{
		Packages: []typefacts.PackageFacts{{ID: 0, Path: "p", TypeIDs: []int{0, 1, 2}}},
		Types: []typefacts.TypeFacts{
			{ID: 0, Kind: typefacts.KindStruct},
			{ID: 1, Kind: typefacts.KindInterface},
			{ID: 2, Kind: typefacts.KindOther},
		},
	}

	interfaces, total := CountTypes(facts, 0)

	if interfaces != 1 || total != 3 {
		t.Fatalf("counts = interfaces=%d total=%d", interfaces, total)
	}
}
