package application

import (
	"context"
	"sort"

	"github.com/gostafa/distance/internal/features/typefacts/domain"
	"github.com/gostafa/distance/internal/features/typefacts/ports/outbound"
)

// Collector is the type-facts application boundary consumed by analysis
// pipelines. Implementations load and assemble one deterministic project view.
type Collector interface {
	Collect(context.Context, outbound.FactOptions) (domain.ProjectFacts, error)
}

// Service produces ProjectFacts through the outbound fact source.
type Service struct {
	source outbound.FactSource
}

var _ Collector = (*Service)(nil)

// NewService returns a Service backed by the given fact source.
func NewService(source outbound.FactSource) *Service {
	return &Service{source: source}
}

// Collect loads the project once and returns its assembled facts.
func (s *Service) Collect(
	ctx context.Context,
	opts outbound.FactOptions,
) (domain.ProjectFacts, error) {
	modulePath, extracts, err := s.source.Load(ctx, opts)
	if err != nil {
		return domain.ProjectFacts{}, err
	}

	return Assemble(modulePath, extracts), nil
}

// Assemble sorts the extracts, assigns dense numeric IDs, and builds the
// project fact set. Ordering is fully deterministic: packages by import
// path, types by (package path, name).
func Assemble(modulePath string, extracts []domain.PackageExtract) domain.ProjectFacts {
	sort.Slice(extracts, func(i, j int) bool { return extracts[i].Path < extracts[j].Path })

	for i := range extracts {
		types := extracts[i].Types
		sort.Slice(types, func(a, b int) bool { return types[a].Name < types[b].Name })
	}

	totalTypes := 0
	for i := range extracts {
		totalTypes += len(extracts[i].Types)
	}

	facts := domain.ProjectFacts{
		ModulePath: modulePath,
		Packages:   make([]domain.PackageFacts, 0, len(extracts)),
		Types:      make([]domain.TypeFacts, 0, totalTypes),
	}

	typeID := 0

	for pkgID, extract := range extracts {
		pkg := domain.PackageFacts{
			ID:       pkgID,
			Path:     extract.Path,
			InModule: extract.InModule,
			Imports:  sortedUnique(extract.Imports, extract.Path),
			TypeIDs:  make([]int, 0, len(extract.Types)),
		}
		for _, t := range extract.Types {
			id := typeID
			typeID++

			pkg.TypeIDs = append(pkg.TypeIDs, id)
			facts.Types = append(facts.Types, domain.TypeFacts{
				ID:        id,
				PackageID: pkgID,
				Name:      t.Name,
				Kind:      t.Kind,
			})
		}

		facts.Packages = append(facts.Packages, pkg)
	}

	return facts
}

// sortedUnique sorts and deduplicates import paths and removes self-imports.
func sortedUnique(imports []string, self string) []string {
	if len(imports) == 0 {
		return nil
	}

	out := make([]string, 0, len(imports))
	for _, path := range imports {
		if path != self {
			out = append(out, path)
		}
	}

	sort.Strings(out)

	dedup := out[:0]
	for i, path := range out {
		if i == 0 || path != out[i-1] {
			dedup = append(dedup, path)
		}
	}

	if len(dedup) == 0 {
		return nil
	}

	return dedup
}
