// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	"fmt"
	"testing"
)

const (
	benchPackages     = 200
	benchTypesPerPkg  = 20
	benchImportsPerPk = 8
	benchInterfaceMod = 4
)

func benchFacts() *ProjectFacts {
	facts := &ProjectFacts{
		ModulePath: "example.com/m",
		Packages:   make([]PackageFacts, benchPackages),
		Types:      make([]TypeFacts, zero, benchPackages*benchTypesPerPkg),
	}

	for pkgID := range facts.Packages {
		imports := make([]string, zero, benchImportsPerPk+2)
		for offset := 1; offset <= benchImportsPerPk; offset++ {
			imports = append(
				imports,
				fmt.Sprintf("example.com/m/pkg%d", (pkgID+offset*7)%benchPackages),
			)
		}

		imports = append(imports, "fmt", "example.com/other/lib")

		typeIDs := make([]int, zero, benchTypesPerPkg)

		for index := range benchTypesPerPkg {
			typeID := len(facts.Types)

			var kind uint8
			if index%benchInterfaceMod == zero {
				kind = KindInterface
			}

			facts.Types = append(facts.Types, TypeFacts{
				Name:      fmt.Sprintf("Type%02d", index),
				ID:        typeID,
				PackageID: pkgID,
				Kind:      kind,
			})
			typeIDs = append(typeIDs, typeID)
		}

		facts.Packages[pkgID] = PackageFacts{
			Path:     fmt.Sprintf("example.com/m/pkg%d", pkgID),
			Imports:  imports,
			TypeIDs:  typeIDs,
			ID:       pkgID,
			InModule: true,
		}
	}

	return facts
}

func BenchmarkBuildDependencyGraph(b *testing.B) {
	facts := benchFacts()

	for _, scope := range []string{ScopeProject, ScopeModule, ScopeAll} {
		b.Run(scope, func(b *testing.B) {
			b.ReportAllocs()

			for b.Loop() {
				_ = BuildDependencyGraph(facts, scope)
			}
		})
	}
}

func BenchmarkComputeAllPackages(b *testing.B) {
	facts := benchFacts()

	b.ReportAllocs()

	for b.Loop() {
		graph := BuildDependencyGraph(facts, ScopeProject)

		for pkgID := range facts.Packages {
			interfaces, total := CountTypes(facts, pkgID)
			afferent, efferent := PackageCoupling(&graph, pkgID)

			_ = Compute(&struct{ Interfaces, Total, Afferent, Efferent int }{
				interfaces, total, afferent, efferent,
			})
		}
	}
}
