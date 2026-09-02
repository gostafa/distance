package model

import "fmt"

// Position locates a declaration in source, with File relative to the
// analysis directory when possible so output is machine-independent.
type Position struct {
	// File is the source file path, relative when possible.
	File string
	// Line is the 1-based source line.
	Line int
	// Column is the 1-based source column.
	Column int
}

// FieldFacts describes one struct field slot.
type FieldFacts struct {
	// Name is the field name (the type name for embedded fields).
	Name string
	// Exported reports whether the field name is exported.
	Exported bool
	// Embedded marks an embedded (anonymous) field.
	Embedded bool
}

// DeclarationFacts describes one top-level named var or const declaration.
type DeclarationFacts struct {
	// Name is the declared identifier.
	Name string
	// Exported reports whether the declaration name is exported.
	Exported bool
	// Pos locates the declaration in source.
	Pos Position
}

// FunctionFacts describes one top-level function declaration.
type FunctionFacts struct {
	// Name is the declared function name.
	Name string
	// Exported reports whether the function name is exported.
	Exported bool
	// Pos locates the function declaration in source.
	Pos Position
	// Lines is the inclusive source line count from func keyword to end.
	Lines int
}

// MethodFacts describes one explicitly declared method.
type MethodFacts struct {
	// Name is the method name.
	Name string
	// Exported reports whether the method name is exported.
	Exported bool
	// Pos locates the method declaration in source.
	Pos Position
	// Lines is the inclusive source line count from func keyword to end.
	Lines int
}

// String summarizes the method facts for debugging.
func (m *MethodFacts) String() string {
	return fmt.Sprintf("method %q (exported %v) at %v", m.Name, m.Exported, m.Pos)
}
