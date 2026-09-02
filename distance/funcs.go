// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package distance

import (
	"context"
	"fmt"
	"slices"

	"github.com/gostafa/distance/internal/features/projectanalysis/ports/inbound"
	"github.com/gostafa/distance/internal/infrastructure/analyzer"
	"github.com/gostafa/distance/internal/shared/metrics"
	"github.com/gostafa/distance/internal/shared/version"
)

// AllMetrics returns the metric names included in every PackageReport.
func AllMetrics() []MetricName {
	names := metrics.ReportedMetricOrder()
	out := make([]MetricName, zero, len(names))

	for i := range names {
		out = append(out, MetricName(names[i]))
	}

	return out
}

// Analyze runs package-distance analysis for config and returns a Report.
func Analyze(ctx context.Context, config *Config) (Report, error) {
	cfg := configWithDefaults(config)

	validateErr := validateConfig(cfg)
	if validateErr != nil {
		return Report{}, fmt.Errorf(errWrapAnalyze, validateErr)
	}

	report, analyzeErr := analyzeValidated(ctx, cfg)
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

// PatternList returns the package patterns to analyze.
func (config *Config) PatternList() []string {
	return config.Patterns
}

// ScopeName returns the dependency scope name.
func (config *Config) ScopeName() string {
	return string(config.DependencyScope)
}

// ToolIdent returns the tool identity embedded in a Report.
func ToolIdent(name, ver string) ToolInfo {
	return ToolInfo{Name: name, Version: ver}
}

func patternListOf(view configured) []string {
	return view.PatternList()
}

func reportPackages(view reporter) []PackageReport {
	return view.PackageList()
}

func scopeOf(view configured) string {
	return view.ScopeName()
}

func analyzeValidated(ctx context.Context, cfg *Config) (Report, error) {
	result, analyzeErr := analyzer.NewAnalyzer().Analyze(ctx, inboundOptions(cfg))
	if analyzeErr != nil {
		return Report{}, fmt.Errorf("distance analyze: %w", analyzeErr)
	}

	return buildReport(&result), nil
}

func buildPackages(result *inbound.Result) []PackageReport {
	packages := make([]PackageReport, zero, len(result.Packages))

	for i := range result.Packages {
		pkg := &result.Packages[i]

		packages = append(packages, PackageReport{
			Path:     pkg.Path,
			Afferent: pkg.Afferent,
			Efferent: pkg.Efferent,
			Metrics:  pkg.Metrics,
		})
	}

	return packages
}

func buildReport(result *inbound.Result) Report {
	return Report{
		SchemaVersion: SchemaVersion,
		Tool:          ToolIdent(string(MetricDistance), version.Version()),
		Module:        result.ModulePath,
		Packages:      reportPackages(&Report{Packages: buildPackages(result)}),
	}
}

func configWithDefaults(config *Config) *Config {
	if len(patternListOf(config)) == zero {
		config.Patterns = []string{allPackagesPattern}
	}

	if scopeOf(config) == "" {
		config.DependencyScope = DependencyScopeModule
	}

	return config
}

func (err configError) Error() string {
	return err.message
}

func inboundOptions(cfg *Config) *inbound.Options {
	return &inbound.Options{
		Directory:        cfg.Directory,
		Patterns:         patternListOf(cfg),
		IncludeTests:     cfg.IncludeTests,
		IncludeGenerated: cfg.IncludeGenerated,
		BuildTags:        cfg.BuildTags,
		Workers:          cfg.Workers,
		DependencyScope:  scopeOf(cfg),
		ContinueOnError:  cfg.ContinueOnError,
	}
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
		return configError{
			message: fmt.Sprintf(
				"invalid dependency scope %q (want project, module, or all)",
				scopeOf(config),
			),
		}
	}
}

func validatePatterns(config *Config) error {
	if slices.Contains(patternListOf(config), "") {
		return configError{message: "empty package pattern"}
	}

	return nil
}
