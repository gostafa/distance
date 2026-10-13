// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/gostafa/distance/internal/features/typefacts/domain/model"
)

// Assemble builds ProjectFacts from modulePath and package extracts.
func Assemble(modulePath string, extracts []model.PackageExtract) ProjectFacts {
	sortExtracts(extracts)

	return projectFromExtracts(modulePath, extracts)
}

// DumpFacts returns a compact debug representation of facts.
func DumpFacts(facts *ProjectFacts) string {
	return ProjectFactsString(facts)
}

// FormatTypeFacts returns a compact debug representation of type facts.
func FormatTypeFacts(facts *TypeFacts) string {
	return fmt.Sprintf(
		"type %d %q (package %d, kind %d)",
		facts.ID,
		facts.Name,
		facts.PackageID,
		facts.Kind,
	)
}

// ProjectFactsString returns a compact debug representation of project facts.
func ProjectFactsString(facts *ProjectFacts) string {
	return fmt.Sprintf(
		"module %q: %d packages, %d types",
		facts.ModulePath,
		len(facts.Packages),
		len(facts.Types),
	)
}

func appendExtractTypes(input *packageAssemble, pkg *PackageFacts) {
	for i := range input.extract.Types {
		named := &input.extract.Types[i]
		id := *input.typeID
		*input.typeID++

		pkg.TypeIDs = append(pkg.TypeIDs, id)
		input.facts.Types = append(input.facts.Types, TypeFacts{
			ID:        id,
			PackageID: input.pkgID,
			Name:      named.Name,
			Kind:      named.Kind,
		})
	}
}

func countExtractTypes(extracts []model.PackageExtract) int {
	total := zeroIndex

	for i := range extracts {
		total += len(extracts[i].Types)
	}

	return total
}

func dedupSorted(sorted []string) []string {
	if len(sorted) == zeroIndex {
		return nil
	}

	return slices.Compact(sorted)
}

func filterSelfImports(imports []string, self string) []string {
	if len(imports) == zeroIndex {
		return nil
	}

	out := make([]string, zeroIndex, len(imports))

	for i := range imports {
		if imports[i] != self {
			out = append(out, imports[i])
		}
	}

	return out
}

func packageFromExtract(input *packageAssemble) PackageFacts {
	pkg := PackageFacts{
		ID:       input.pkgID,
		Path:     input.extract.Path,
		InModule: input.extract.InModule,
		Imports:  sortedUnique(input.extract.Imports, input.extract.Path),
		TypeIDs:  make([]int, zeroIndex, len(input.extract.Types)),
	}

	appendExtractTypes(input, &pkg)

	return pkg
}

func projectFromExtracts(modulePath string, extracts []model.PackageExtract) ProjectFacts {
	facts := ProjectFacts{
		ModulePath: modulePath,
		Packages:   make([]PackageFacts, zeroIndex, len(extracts)),
		Types:      make([]TypeFacts, zeroIndex, countExtractTypes(extracts)),
	}

	typeID := zeroIndex

	for pkgID := range extracts {
		facts.Packages = append(facts.Packages, packageFromExtract(&packageAssemble{
			pkgID:   pkgID,
			extract: &extracts[pkgID],
			typeID:  &typeID,
			facts:   &facts,
		}))
	}

	return facts
}

func sortExtracts(extracts []model.PackageExtract) {
	slices.SortFunc(extracts, func(a, b model.PackageExtract) int {
		return cmp.Compare(a.Path, b.Path)
	})

	for i := range extracts {
		slices.SortFunc(extracts[i].Types, func(left, right model.TypeName) int {
			return cmp.Compare(left.Name, right.Name)
		})
	}
}

func sortedUnique(imports []string, self string) []string {
	filtered := filterSelfImports(imports, self)

	if len(filtered) == zeroIndex {
		return nil
	}

	slices.Sort(filtered)

	return dedupSorted(filtered)
}
