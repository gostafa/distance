package model

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
