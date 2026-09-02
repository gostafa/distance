// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	"context"

	coupling "github.com/gostafa/distance/internal/features/packagemetrics/domain"
	typefacts "github.com/gostafa/distance/internal/features/typefacts/application"
	tfdomain "github.com/gostafa/distance/internal/features/typefacts/domain"
)

type (
	// Options configures one Analyzer.Analyze call.
	Options struct {
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
	MetricEntry struct {
		Name       string
		Scope      string
		Reason     string
		Definition string
		Value      float64
		Applicable bool
	}

	// PackageResult holds metrics and coupling for one analyzed package.
	PackageResult struct {
		Path     string
		Metrics  []MetricEntry
		Afferent int
		Efferent int
	}

	// Result is the complete analysis outcome for one Options value.
	Result struct {
		ModulePath string
		Packages   []PackageResult
	}

	// Analyzer runs project-level package-distance analysis.
	Analyzer interface {
		Analyze(ctx context.Context, opts *Options) (Result, error)
	}

	// Pipeline orchestrates type-fact collection and package metric computation.
	Pipeline struct {
		facts   typefacts.Collector
		runtime pipelineRuntime
	}

	packageMetrics struct {
		abstractness MetricEntry
		instability  MetricEntry
		distance     MetricEntry
	}

	pipelineRuntime struct {
		runWorkers func(context.Context, int, int, func(int) error) error
	}

	analyzePackageInput struct {
		facts      *tfdomain.ProjectFacts
		pkgResults []packageMetrics
		afferent   int
		efferent   int
		pkgID      int
	}

	assembleIn struct {
		facts *tfdomain.ProjectFacts
		opts  *Options
	}

	fillInput struct {
		facts *tfdomain.ProjectFacts
		graph *coupling.DependencyGraph
		rows  []PackageResult
	}
)
