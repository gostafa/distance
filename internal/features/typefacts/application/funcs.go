// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	"cmp"
	"context"
	"fmt"
	"slices"

	"github.com/gostafa/distance/internal/features/typefacts/domain"
	"github.com/gostafa/distance/internal/features/typefacts/domain/model"
	"github.com/gostafa/distance/internal/features/typefacts/ports/outbound"
)

// NewService returns a Collector backed by source.
func NewService(source outbound.FactSource) *Service {
	return &Service{source: source}
}

// Collect loads package extracts and assembles project facts.
func (svc *Service) Collect(ctx context.Context, opt *outbound.FactOptions) (ProjectFacts, error) {
	modulePath, extracts, err := svc.source.Load(ctx, opt)
	if err != nil {
		return domain.ProjectFacts{}, fmt.Errorf("application Collect: %w", err)
	}

	return Assemble(modulePath, extracts), nil
}

// Assemble builds ProjectFacts from modulePath and package extracts.
func Assemble(modulePath string, extracts []model.PackageExtract) domain.ProjectFacts {
	sortExtracts(extracts)

	return projectFromExtracts(modulePath, extracts)
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

func projectFromExtracts(modulePath string, extracts []model.PackageExtract) domain.ProjectFacts {
	facts := domain.ProjectFacts{
		ModulePath: modulePath,
		Packages:   make([]domain.PackageFacts, zeroIndex, len(extracts)),
		Types:      make([]domain.TypeFacts, zeroIndex, countExtractTypes(extracts)),
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

func countExtractTypes(extracts []model.PackageExtract) int {
	total := zeroIndex

	for i := range extracts {
		total += len(extracts[i].Types)
	}

	return total
}

func packageFromExtract(input *packageAssemble) domain.PackageFacts {
	pkg := domain.PackageFacts{
		ID:       input.pkgID,
		Path:     model.PathOf(input.extract),
		InModule: input.extract.InModule,
		Imports:  sortedUnique(model.ImportsOf(input.extract), model.PathOf(input.extract)),
		TypeIDs:  make([]int, zeroIndex, len(input.extract.Types)),
	}

	appendExtractTypes(input, &pkg)

	return pkg
}

func appendExtractTypes(input *packageAssemble, pkg *domain.PackageFacts) {
	for i := range input.extract.Types {
		named := &input.extract.Types[i]
		id := *input.typeID
		*input.typeID++

		pkg.TypeIDs = append(pkg.TypeIDs, id)
		input.facts.Types = append(input.facts.Types, domain.TypeFacts{
			ID:        id,
			PackageID: input.pkgID,
			Name:      named.Name,
			Kind:      named.Kind,
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

func dedupSorted(sorted []string) []string {
	dedup := sorted[:zeroIndex]

	for i := range sorted {
		if i == zeroIndex || sorted[i] != sorted[i-nextIndex] {
			dedup = append(dedup, sorted[i])
		}
	}

	if len(dedup) == zeroIndex {
		return nil
	}

	return dedup
}
