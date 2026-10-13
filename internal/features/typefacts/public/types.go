// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package public

import (
	"context"
)

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

	// FactOptions configures a FactSource.Load call.
	FactOptions = struct {
		Directory        string
		Patterns         []string
		BuildTags        []string
		Workers          int
		IncludeTests     bool
		IncludeGenerated bool
		ContinueOnError  bool
	}

	// Collector returns deterministic project snapshots without publishing loader details.
	Collector interface {
		// Collect honors inclusion flags and returns packages sorted by import path with stable type IDs.
		Collect(ctx context.Context, opts *FactOptions) (ProjectFacts, error)
	}
)
