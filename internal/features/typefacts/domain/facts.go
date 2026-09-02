package domain

import (
	"fmt"

	factmodel "github.com/gostafa/distance/internal/features/typefacts/domain/model"
)

// TypeKind classifies a named type's underlying type.
type TypeKind uint8

const (
	// KindStruct marks a named type whose underlying type is a struct.
	KindStruct TypeKind = iota
	// KindInterface marks a named type whose underlying type is an interface.
	KindInterface
	// KindOther marks any other named type (basic, slice, func, …).
	KindOther
)

// Position locates a declaration in source, with File relative to the
// analysis directory when possible so output is machine-independent.
type Position = factmodel.Position

// ProjectFacts is everything the metric features need to know about the
// analyzed packages. All slices are deterministically ordered and all
// cross-references use dense numeric IDs (indices into the slices).
type ProjectFacts struct {
	// ModulePath is the import path of the main module, when known.
	ModulePath string
	// Packages is sorted by import path; a package's ID is its index.
	Packages []PackageFacts
	// Types is sorted by (package path, type name); a type's ID is its index.
	Types []TypeFacts
}

// String summarizes the fact set for debugging.
func (f *ProjectFacts) String() string {
	return fmt.Sprintf(
		"module %q: %d packages, %d types",
		f.ModulePath,
		len(f.Packages),
		len(f.Types),
	)
}

// PackageFacts describes one analyzed package.
type PackageFacts struct {
	// ID is the package's dense index into ProjectFacts.Packages.
	ID int
	// Path is the package's import path.
	Path string
	// InModule reports whether the package belongs to the main module.
	InModule bool
	// Imports are the package's distinct import paths, sorted, without
	// self-imports. Scope filtering happens in the architecture feature.
	Imports []string
	// TypeIDs are the package's analyzed types in name order.
	TypeIDs []int
}

// TypeFacts describes one analyzed named type. Aliases are never analyzed.
// Kind is kept so abstractness can count interface vs non-interface types.
type TypeFacts struct {
	// ID is the type's dense index into ProjectFacts.Types.
	ID int
	// PackageID is the declaring package's dense index.
	PackageID int
	// Name is the type's declared name.
	Name string
	// Kind classifies the type's underlying type.
	Kind TypeKind
}

// String summarizes the type facts for debugging.
func (t *TypeFacts) String() string {
	return fmt.Sprintf(
		"type %d %q (package %d, kind %d)",
		t.ID,
		t.Name,
		t.PackageID,
		t.Kind,
	)
}
