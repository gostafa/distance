// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package distance

import (
	"context"
)

type (
	// MetricResult is a computed package metric value.
	MetricResult struct {
		Name       string
		Scope      string
		Reason     string
		Definition string
		Value      float64
		Applicable bool
	}

	// Backend runs package-distance analysis for config.
	Backend interface {
		Analyze(ctx context.Context, cfg *Config) (Report, error)
	}

	reportReader interface {
		ModulePath() string
		PackageList() []PackageReport
	}

	configReader interface {
		PatternList() []string
		ScopeName() string
	}

	metricCarrier interface {
		MetricResults() []MetricResult
	}

	// Config controls a single Analyze run.
	Config struct {
		Directory        string
		DependencyScope  DependencyScope
		Patterns         []string
		BuildTags        []string
		Workers          int
		IncludeTests     bool
		IncludeGenerated bool
		ContinueOnError  bool
	}

	// PackageReport holds metrics and coupling for one analyzed package.
	PackageReport struct {
		Path     string
		Metrics  []MetricResult
		Afferent int
		Efferent int
	}

	// Report is the complete analysis result for one configuration.
	Report struct {
		SchemaVersion string
		ToolName      string
		ToolVersion   string
		Module        string
		Packages      []PackageReport
	}
)

var (
	_ reportReader  = (*Report)(nil)
	_ configReader  = (*Config)(nil)
	_ metricCarrier = (*PackageReport)(nil)
)
