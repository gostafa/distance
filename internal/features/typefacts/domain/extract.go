package domain

import "fmt"

// PackageExtract is one package's facts as produced by a fact source,
// before dense IDs are assigned.
//
// Contract for producers: Types may be in any order.
type PackageExtract struct {
	// Path is the package's import path.
	Path string
	// InModule reports whether the package belongs to the main module.
	InModule bool
	// Imports are the package's distinct import paths, without self-imports.
	Imports []string
	// Types are the package's extracted named types, in any order.
	Types []TypeExtract
}

// TypeExtract mirrors TypeFacts before dense IDs are assigned. Only name and
// kind are extracted; kind feeds abstractness.
type TypeExtract struct {
	// Name is the type's declared name.
	Name string
	// Kind classifies the type's underlying type.
	Kind TypeKind
}

// String summarizes the extract for debugging.
func (t *TypeExtract) String() string {
	return fmt.Sprintf("type %q (kind %d)", t.Name, t.Kind)
}
