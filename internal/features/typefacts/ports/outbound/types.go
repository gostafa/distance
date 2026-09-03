// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package outbound

import (
	"context"

	"github.com/gostafa/distance/internal/features/typefacts/domain/model"
)

type (
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

	// FactSource loads raw package extracts for later assembly.
	FactSource interface {
		// Load returns the main module path (empty when unknown) and one
		// PackageExtract per analyzed package, honoring the ordering contract
		// documented on model.PackageExtract.
		Load(
			ctx context.Context,
			opts *FactOptions,
		) (modulePath string, packages []model.PackageExtract, err error)
	}
)
