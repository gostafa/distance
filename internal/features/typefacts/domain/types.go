// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

type (

	// moduleView exposes the analyzed module path.
	moduleView interface {
		Module() string
	}

	// packageCounter exposes how many packages were assembled.
	packageCounter interface {
		PackageCount() int
	}

	// typeCounter exposes how many named types were assembled.
	typeCounter interface {
		TypeCount() int
	}

	// PackagePathView exposes one package's import path.
	PackagePathView interface {
		ImportPath() string
	}

	// FactDumper renders a compact debug representation.
	FactDumper interface {
		String() string
	}

	// ProjectFacts is the assembled type graph for one analysis run.
	ProjectFacts struct {
		// ModulePath is the import path of the main module, when known.
		ModulePath string
		// Packages is sorted by import path; a package's ID is its index.
		Packages []PackageFacts
		// Types is sorted by (package path, type name); a type's ID is its index.
		Types []TypeFacts
	}

	// PackageFacts holds one package's assembled facts and type IDs.
	PackageFacts struct {
		Path     string
		Imports  []string
		TypeIDs  []int
		ID       int
		InModule bool
	}

	// TypeFacts holds one named type's assembled facts.
	TypeFacts struct {
		Name      string
		ID        int
		PackageID int
		Kind      uint8
	}
)
