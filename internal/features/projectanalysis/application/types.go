// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	"context"

	pkgmetrics "github.com/gostafa/distance/internal/features/packagemetrics/application"
	coupling "github.com/gostafa/distance/internal/features/packagemetrics/domain"
	"github.com/gostafa/distance/internal/features/projectanalysis/ports/inbound"
	typefacts "github.com/gostafa/distance/internal/features/typefacts/application"
	tfdomain "github.com/gostafa/distance/internal/features/typefacts/domain"
	"github.com/gostafa/distance/internal/shared/workerpool"
)

type (
	// Pipeline orchestrates type-fact collection and package metric computation.
	Pipeline struct {
		facts   typefacts.Collector
		runtime pipelineRuntime
	}

	pipelineRuntime struct {
		runWorkers func(context.Context, workerpool.PoolConfig, func(int) error) error
	}

	analyzePackageInput struct {
		facts      *tfdomain.ProjectFacts
		pkgResults []pkgmetrics.Result
		afferent   int
		efferent   int
		pkgID      int
	}

	assembleIn struct {
		facts *tfdomain.ProjectFacts
		opts  *inbound.Options
	}

	fillInput struct {
		facts *tfdomain.ProjectFacts
		graph *coupling.DependencyGraph
		rows  []inbound.PackageResult
	}
)
