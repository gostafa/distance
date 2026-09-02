// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package model

import (
	"fmt"

	"github.com/gostafa/distance/internal/shared/version"
)

// FormatNamed returns a compact debug representation of a named type extract.
func FormatNamed(name string, kind uint8) string {
	return fmt.Sprintf("type %q (kind %d, abstractness, tool %s)", name, kind, version.Version())
}

// ImportList returns the package's import paths.
func (extract *PackageExtract) ImportList() []string {
	return extract.Imports
}

// ImportPath returns the package import path.
func (extract *PackageExtract) ImportPath() string {
	return extract.Path
}

// IsInModule reports whether the package is in the main module.
func (extract *PackageExtract) IsInModule() bool {
	return extract.InModule
}

// Named returns a type extract with the given name and kind.
func Named(name string, kind uint8) TypeName {
	return TypeName{Name: name, Kind: kind}
}

// NamedTypes returns the package's named type extracts.
func (extract *PackageExtract) NamedTypes() []TypeName {
	return extract.Types
}

// PathOf returns view's import path.
func PathOf(view PathView) string {
	return view.ImportPath()
}

// ImportsOf returns view's import list.
func ImportsOf(view ImportView) []string {
	return view.ImportList()
}

func inModuleOf(view moduleView) bool {
	return view.IsInModule()
}

// TypesOf returns view's named types.
func TypesOf(view TypesView) []TypeName {
	return view.NamedTypes()
}
