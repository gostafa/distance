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
	// Service is the default Collector implementation.
	Service struct {
		load func(context.Context, *outbound.FactOptions) (string, []model.PackageExtract, error)
	}

	// ProjectFacts is the assembled project snapshot returned by Collect.
	ProjectFacts = domain.ProjectFacts
)
