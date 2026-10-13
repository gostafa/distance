// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package outbound

import (
	"context"
)

type (
	// PackageFacts describes one package in a collected snapshot.
	PackageFacts = struct {
		Path     string
		Imports  []string
		TypeIDs  []int
		ID       int
		InModule bool
	}

	// TypeFacts describes one named type in a collected snapshot.
	TypeFacts = struct {
		Name      string
		ID        int
		PackageID int
		Kind      uint8
	}

	// ProjectFacts is the assembled type graph for one analysis run.
	ProjectFacts = struct {
		ModulePath string
		Packages   []PackageFacts
		Types      []TypeFacts
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

	// FactCollector collects facts without exposing a loader implementation.
	FactCollector interface {
		// Collect returns a deterministic, read-only snapshot; errors retain cancellation identity.
		Collect(ctx context.Context, opts *FactOptions) (ProjectFacts, error)
	}
)
