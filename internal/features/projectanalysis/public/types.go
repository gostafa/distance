// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package public

import (
	"context"
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

	// Pipeline publishes analysis without exposing construction or private models.
	Pipeline interface {
		// Analyze computes all three metrics in fixed order, preserving cancellation and deterministic output.
		Analyze(ctx context.Context, opts *Options) (Result, error)
	}
)
