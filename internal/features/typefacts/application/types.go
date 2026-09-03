// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	"context"

	"github.com/gostafa/distance/internal/features/typefacts/domain"
	"github.com/gostafa/distance/internal/features/typefacts/domain/model"
	"github.com/gostafa/distance/internal/features/typefacts/ports/outbound"
)

type (
	// Collector loads and assembles project type facts.
	Collector interface {
		Collect(ctx context.Context, opts *outbound.FactOptions) (domain.ProjectFacts, error)
	}

	// Service is the default Collector implementation.
	Service struct {
		load func(context.Context, *outbound.FactOptions) (string, []model.PackageExtract, error)
	}

	// ProjectFacts is the assembled project snapshot returned by Collect.
	ProjectFacts = domain.ProjectFacts

	packageAssemble = struct {
		extract *model.PackageExtract
		typeID  *int
		facts   *domain.ProjectFacts
		pkgID   int
	}
)
