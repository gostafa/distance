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
	// ExportedFuncCount is the number of declared functions and methods with
	// an exported name in the package's non-excluded files.
	ExportedFuncCount int
	// UnexportedFuncCount is the number of declared functions and methods with
	// an unexported name in the package's non-excluded files.
	UnexportedFuncCount int
	// VarCount is the number of top-level variable names declared in the
	// package's non-excluded files, excluding blank identifiers.
	VarCount int
	// ConstCount is the number of top-level constant names declared in the
	// package's non-excluded files, excluding blank identifiers.
	ConstCount int
	// Variables are the package's top-level variable names.
	Variables []DeclarationFacts
	// Constants are the package's top-level constant names.
	Constants []DeclarationFacts
	// Functions are the package's top-level functions, excluding methods.
	Functions []FunctionFacts
	// TypeIDs are the package's analyzed types in name order.
	TypeIDs []int
}

// TypeFacts describes one analyzed named type. Aliases are never analyzed.
type TypeFacts struct {
	// ID is the type's dense index into ProjectFacts.Types.
	ID int
	// PackageID is the declaring package's dense index.
	PackageID int
	// Name is the type's declared name.
	Name string
	// Exported reports whether the type name is exported.
	Exported bool
	// Kind classifies the type's underlying type.
	Kind TypeKind
	// Pos locates the type declaration in source.
	Pos Position
	// Fields are the struct's fields in declaration order; empty for
	// non-struct types. An embedded field occupies exactly one slot and
	// promoted members are not represented (they belong to the embedded
	// type).
	Fields []FieldFacts
	// Methods are the explicitly declared methods (functions with a
	// receiver, pointer and value receivers normalized to this type),
	// sorted by name then source position. Promoted methods are excluded.
	Methods []MethodFacts
}

// String summarizes the type facts for debugging.
func (t *TypeFacts) String() string {
	return fmt.Sprintf(
		"type %d %q (package %d, kind %d, exported %v) at %v: %d fields, %d methods",
		t.ID,
		t.Name,
		t.PackageID,
		t.Kind,
		t.Exported,
		t.Pos,
		len(t.Fields),
		len(t.Methods),
	)
}

// FieldFacts describes one struct field slot.
type FieldFacts = factmodel.FieldFacts

// DeclarationFacts describes one top-level named var or const declaration.
type DeclarationFacts = factmodel.DeclarationFacts

// FunctionFacts describes one top-level function declaration.
type FunctionFacts = factmodel.FunctionFacts

// MethodFacts describes one explicitly declared method.
type MethodFacts = factmodel.MethodFacts
