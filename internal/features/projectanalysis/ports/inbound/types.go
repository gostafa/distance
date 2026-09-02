// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package inbound

import (
	"context"

	"github.com/gostafa/distance/internal/shared/metrics"
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

	// PackageResult holds metrics and coupling for one analyzed package.
	PackageResult struct {
		Path     string
		Metrics  []metrics.MetricResult
		Afferent int
		Efferent int
	}

	// Result is the complete analysis outcome for one Options value.
	Result struct {
		// ModulePath is the analyzed main module's path, when known.
		ModulePath string
		// Packages are the analyzed packages, sorted by import path.
		Packages []PackageResult
	}

	// Analyzer runs project-level package-distance analysis.
	Analyzer interface {
		Analyze(ctx context.Context, opts *Options) (Result, error)
	}
)
