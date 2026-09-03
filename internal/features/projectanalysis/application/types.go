// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	"context"

	coupling "github.com/gostafa/distance/internal/features/packagemetrics/domain"
	typefacts "github.com/gostafa/distance/internal/features/typefacts/application"
	tfdomain "github.com/gostafa/distance/internal/features/typefacts/domain"
	"github.com/gostafa/distance/internal/infrastructure/goloader"
)

type (
	// Options configures one Pipeline.Analyze call.
	Options = struct {
		Directory        string
		DependencyScope  string
		Patterns         []string
		BuildTags        []string
		Workers          int
		IncludeTests     bool
		IncludeGenerated bool
		ContinueOnError  bool
	}

	// MetricEntry is one computed metric value exposed by analysis results.
	MetricEntry = struct {
		Name       string
		Scope      string
		Reason     string
		Definition string
		Value      float64
		Applicable bool
	}

	// PackageResult holds metrics and coupling for one analyzed package.
	PackageResult = struct {
		// Path is the package import path.
		Path string
		// Metrics are the computed package-level metric entries.
		Metrics []MetricEntry
		// Afferent is incoming coupling (Ca).
		Afferent int
		// Efferent is outgoing coupling (Ce).
		Efferent int
	}

	// Result is the complete analysis outcome for one Options value.
	Result = struct {
		// ModulePath is the import path of the main module, when known.
		ModulePath string
		// Packages are the analyzed package results in report order.
		Packages []PackageResult
	}

	analyzer interface {
		Analyze(ctx context.Context, opts *Options) (Result, error)
	}

	// Pipeline orchestrates type-fact collection and package metric computation.
	Pipeline struct {
		analyze func(context.Context, *Options) (Result, error)
	}

	packageMetrics = struct {
		abstractness MetricEntry
		instability  MetricEntry
		distance     MetricEntry
	}

	pipelineRuntime = struct {
		facts      typefacts.Collector
		runWorkers workerRunFunc
	}

	workerRunFunc = func(context.Context, goloader.WorkerRun, func(int) error) error

	analyzePackageInput = struct {
		facts      *tfdomain.ProjectFacts
		pkgResults []packageMetrics
		afferent   int
		efferent   int
		pkgID      int
	}

	assembleIn = struct {
		facts *tfdomain.ProjectFacts
		opts  *Options
	}

	fillInput = struct {
		facts *tfdomain.ProjectFacts
		graph *coupling.DependencyGraph
		rows  []PackageResult
	}

	computeIn = struct {
		facts *tfdomain.ProjectFacts
		graph coupling.CouplingGraph
	}
)
