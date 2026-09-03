// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

type (
	// ProjectFacts is the assembled type graph for one analysis run.
	ProjectFacts = struct {
		ModulePath string
		Packages   []PackageFacts
		Types      []TypeFacts
	}

	// PackageFacts holds one package's assembled facts and type IDs.
	PackageFacts = struct {
		Path     string
		Imports  []string
		TypeIDs  []int
		ID       int
		InModule bool
	}

	// TypeFacts holds one named type's assembled facts.
	TypeFacts = struct {
		Name      string
		ID        int
		PackageID int
		Kind      uint8
	}
)
