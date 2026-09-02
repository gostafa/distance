package domain

import "fmt"

// PackageExtract is one package's facts as produced by a fact source,
// before dense IDs are assigned. Cross-package type references use string
// keys (see TypeKey) because global IDs do not exist yet.
//
// Contract for producers: Types may be in any order, but each type's Fields
// must be in declaration order and its Methods sorted by name then source
// position.
type PackageExtract struct {
	// Path is the package's import path.
	Path string
	// InModule reports whether the package belongs to the main module.
	InModule bool
	// Imports are the package's distinct import paths, without self-imports.
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
	// Types are the package's extracted named types, in any order.
	Types []TypeExtract
}

// TypeExtract mirrors TypeFacts before dense IDs are assigned.
type TypeExtract struct {
	// Name is the type's declared name.
	Name string
	// Exported reports whether the type name is exported.
	Exported bool
	// Kind classifies the type's underlying type.
	Kind TypeKind
	// Pos locates the type declaration in source.
	Pos Position
	// Fields are the struct's fields in declaration order.
	Fields []FieldFacts
	// Methods are the declared methods, sorted by name then position.
	Methods []MethodFacts
}

// String summarizes the extract for debugging.
func (t *TypeExtract) String() string {
	return fmt.Sprintf(
		"type %q (kind %d, exported %v) at %v: %d fields, %d methods",
		t.Name,
		t.Kind,
		t.Exported,
		t.Pos,
		len(t.Fields),
		len(t.Methods),
	)
}

// TypeKey is the canonical cross-package key of a named type.
func TypeKey(pkgPath, typeName string) string {
	return pkgPath + "." + typeName
}
