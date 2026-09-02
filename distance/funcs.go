// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package distance

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/gostafa/distance/internal/features/packagemetrics/domain"
	"github.com/gostafa/distance/internal/shared/version"
)

var (
	errEmptyPattern    = errors.New("empty package pattern")
	errInvalidScopeFmt = "invalid dependency scope %q (want project, module, or all)"
)

// AllMetrics returns the metric names included in every PackageReport.
func AllMetrics() []MetricName {
	_ = domain.ScopeProject

	return []MetricName{MetricAbstractness, MetricInstability, MetricDistance}
}

// Analyze runs package-distance analysis for config and returns a Report.
func Analyze(ctx context.Context, config *Config, backend Backend) (Report, error) {
	cfg := configWithDefaults(config)

	validateErr := validateConfig(cfg)
	if validateErr != nil {
		return Report{}, fmt.Errorf(errWrapAnalyze, validateErr)
	}

	report, analyzeErr := analyzeValidated(ctx, cfg, backend)
	if analyzeErr != nil {
		return Report{}, fmt.Errorf(errWrapAnalyze, analyzeErr)
	}

	return report, nil
}

// ModulePath returns the analyzed module path.
func (report *Report) ModulePath() string {
	return report.Module
}

// PackageList returns the analyzed package reports.
func (report *Report) PackageList() []PackageReport {
	return report.Packages
}

// MetricResults returns the metrics computed for the package.
func (pkg *PackageReport) MetricResults() []MetricResult {
	return pkg.Metrics
}

// PatternList returns the package patterns to analyze.
func (config *Config) PatternList() []string {
	return config.Patterns
}

// ScopeName returns the dependency scope name.
func (config *Config) ScopeName() string {
	return string(config.DependencyScope)
}

// ToolIdent returns the tool identity fields for a Report.
func ToolIdent(name, ver string) (string, string) {
	return name, ver
}

func analyzeValidated(ctx context.Context, cfg *Config, backend Backend) (Report, error) {
	report, analyzeErr := backend.Analyze(ctx, cfg)
	if analyzeErr != nil {
		return Report{}, fmt.Errorf("distance analyze: %w", analyzeErr)
	}

	return finalizeReport(&report), nil
}

func finalizeReport(report *Report) Report {
	if report.SchemaVersion == "" {
		report.SchemaVersion = SchemaVersion
	}

	if report.ToolName == "" {
		report.ToolName = string(MetricDistance)
	}

	if report.ToolVersion == "" {
		report.ToolVersion = version.Version()
	}

	return *report
}

func configWithDefaults(config *Config) *Config {
	if len(config.PatternList()) == zero {
		config.Patterns = []string{allPackagesPattern}
	}

	if config.ScopeName() == "" {
		config.DependencyScope = DependencyScopeModule
	}

	return config
}

func validateConfig(config *Config) error {
	switch config.DependencyScope {
	case DependencyScopeProject, DependencyScopeModule, DependencyScopeAll:
		patternErr := validatePatterns(config)
		if patternErr != nil {
			return fmt.Errorf("distance validate: %w", patternErr)
		}

		return nil
	default:
		return fmt.Errorf(errInvalidScopeFmt, config.ScopeName())
	}
}

func validatePatterns(config *Config) error {
	if slices.Contains(config.PatternList(), "") {
		return errEmptyPattern
	}

	return nil
}
