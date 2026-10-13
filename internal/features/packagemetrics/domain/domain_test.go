// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	"testing"
)

func scopedFacts() *ProjectFacts {
	return &ProjectFacts{
		ModulePath: "example.com/m",
		Packages: []PackageFacts{
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
	afferent, efferent := PackageCoupling(&project, 0)

	if efferent != 1 || afferent != 0 {
		t.Fatalf("project scope a = Ca=%d Ce=%d", afferent, efferent)
	}

	afferent, efferent = PackageCoupling(&project, 1)

	if afferent != 1 || efferent != 0 {
		t.Fatalf("project scope b = Ca=%d Ce=%d", afferent, efferent)
	}

	module := BuildDependencyGraph(facts, ScopeModule)

	_, efferent = PackageCoupling(&module, 0)

	if efferent != 1 {
		t.Fatalf("module scope Ce(a) = %d, want 1 (fmt and external excluded)", efferent)
	}

	all := BuildDependencyGraph(facts, ScopeAll)

	_, efferent = PackageCoupling(&all, 0)

	if efferent != 3 {
		t.Fatalf("all scope Ce(a) = %d, want 3", efferent)
	}

	// Afferent coupling is always measured within the analyzed set.
	afferent, _ = PackageCoupling(&all, 1)

	if afferent != 1 {
		t.Fatalf("all scope Ca(b) = %d, want 1", afferent)
	}
}

func TestModuleScopeWithoutModuleInfo(t *testing.T) {
	facts := scopedFacts()

	facts.ModulePath = ""

	graph := BuildDependencyGraph(facts, ScopeModule)
	_, efferent := PackageCoupling(&graph, 0)

	if efferent != 1 {
		t.Fatalf("Ce(a) = %d, want 1 (degrades to project scope)", efferent)
	}
}

func TestCountTypes(t *testing.T) {
	facts := &ProjectFacts{
		Packages: []PackageFacts{{ID: 0, Path: "p", TypeIDs: []int{0, 1, 2}}},
		Types: []TypeFacts{
			{ID: 0, Kind: 0},
			{ID: 1, Kind: KindInterface},
			{ID: 2, Kind: 2},
		},
	}

	interfaces, total := CountTypes(facts, 0)

	if interfaces != 1 || total != 3 {
		t.Fatalf("counts = interfaces=%d total=%d", interfaces, total)
	}
}

func TestMetricDecisions(t *testing.T) {
	cases := []struct {
		interfaces, total, afferent, efferent int
		abstractness, instability, distance   float64
		applicable                            bool
	}{
		{
			1,
			2,
			1,
			1,
			0.5,
			0.5,
			0,
			true,
		},
		{0, 1, 0, 0, 0, 0, 1, true},
		{1, 1, 0, 0, 1, 0, 0, true},
		{0, 0, 0, 0, 0, 0, 0, false},
	}
	for _, test := range cases {
		metrics := Compute(
			&struct{ Interfaces, Total, Afferent, Efferent int }{
				test.interfaces,
				test.total,
				test.afferent,
				test.efferent,
			},
		)
		if len(metrics) != 3 || metrics[0].Applicable != test.applicable ||
			metrics[2].Applicable != test.applicable {
			t.Fatalf("applicability: %+v", metrics)
		}
		if metrics[0].Value != test.abstractness || metrics[1].Value != test.instability ||
			metrics[2].Value != test.distance {
			t.Fatalf("values: %+v", metrics)
		}
	}
	abstractness := applicableMetric(MetricAbstractness, DefinitionAbstractness, 0.5)
	instability := notApplicableMetric(MetricInstability, "missing coupling")
	if got := distanceMetric(&abstractness, &instability); got.Applicable || got.Reason == "" {
		t.Fatalf("missing instability: %+v", got)
	}
}
