// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	"context"
	"fmt"

	"github.com/gostafa/distance/internal/features/typefacts/domain"
	"github.com/gostafa/distance/internal/features/typefacts/domain/model"
	"github.com/gostafa/distance/internal/features/typefacts/ports/outbound"
)

// NewService returns a Collector backed by source.
func NewService(source outbound.FactSource) *Service {
	return &Service{
		load: func(ctx context.Context, opts *outbound.FactOptions) (string, []model.PackageExtract, error) {
			return source.Load(ctx, opts)
		},
	}
}

// Collect loads package extracts and assembles project facts.
func (svc *Service) Collect(ctx context.Context, opt *outbound.FactOptions) (ProjectFacts, error) {
	modulePath, extracts, err := svc.load(ctx, opt)
	if err != nil {
		return domain.ProjectFacts{
			ModulePath: "",
			Packages:   nil,
			Types:      nil,
		}, fmt.Errorf(
			"application Collect: %w",
			err,
		)
	}

	return domain.Assemble(modulePath, extracts), nil
}
