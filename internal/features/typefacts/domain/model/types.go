// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package model

type (

	// PathView exposes a package import path.
	PathView interface {
		ImportPath() string
	}

	// ImportView exposes a package's import list.
	ImportView interface {
		ImportList() []string
	}

	// TypesView exposes the named types declared in a package.
	TypesView interface {
		NamedTypes() []TypeName
	}

	// moduleView reports whether a package belongs to the main module.
	moduleView interface {
		IsInModule() bool
	}

	// TypeName is one named type extract.
	TypeName struct {
		Name string
		Kind uint8
	}

	// PackageExtract is one package's raw facts before project assembly.
	PackageExtract struct {
		Path     string
		Imports  []string
		Types    []TypeName
		InModule bool
	}
)
