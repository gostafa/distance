// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package distance

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/gostafa/distance/internal/composition"
	projapp "github.com/gostafa/distance/internal/features/projectanalysis/public"
	"github.com/gostafa/distance/internal/shared/version"
)

// AllMetrics returns the metric names included in every PackageReport.
func AllMetrics() []string {
	return []string{MetricAbstractness, MetricInstability, MetricDistance}
}

// Analyze runs package-distance analysis for config and returns a Report.
func Analyze(ctx context.Context, config *Config) (Report, error) {
	cfg := configWithDefaults(config)

	var empty Report

	validateErr := validateConfig(cfg)
	if validateErr != nil {
		return empty, fmt.Errorf(errPublicDetails, ErrInvalidConfig, validateErr)
	}

	report, analyzeErr := analyzeValidated(ctx, cfg)
	if analyzeErr != nil {
		return empty, fmt.Errorf("distance analyze: %w", publicAnalysisError(analyzeErr))
	}

	return report, nil
}

func analyzeValidated(ctx context.Context, cfg *Config) (Report, error) {
	result, err := composition.DefaultAnalyzer().Analyze(ctx, toAppOptions(cfg))
	if err != nil {
		return Report{
			SchemaVersion: "",
			ToolName:      "",
			ToolVersion:   "",
			Module:        "",
			Packages:      nil,
		}, fmt.Errorf(
			"distance pipeline: %w",
			err,
		)
	}

	report := fromAppResult(&result)

	return finalizeReport(&report), nil
}

func cloneConfig(config *Config) *Config {
	if config == nil {
		return new(Config)
	}

	cloned := *config

	cloned.Patterns = slices.Clone(config.Patterns)
	cloned.BuildTags = slices.Clone(config.BuildTags)

	return &cloned
}

func configWithDefaults(config *Config) *Config {
	config = cloneConfig(config)

	if len(config.Patterns) == zero {
		config.Patterns = []string{allPackagesPattern}
	}

	if config.DependencyScope == emptyString {
		config.DependencyScope = DependencyScopeModule
	}

	return config
}

func finalizeReport(report *Report) Report {
	if report.SchemaVersion == emptyString {
		report.SchemaVersion = SchemaVersion
	}

	if report.ToolName == emptyString {
		report.ToolName = MetricDistance
	}

	if report.ToolVersion == emptyString {
		report.ToolVersion = version.Version()
	}

	return *report
}

func fromAppResult(result *projapp.Result) Report {
	return Report{
		Module:   result.ModulePath,
		Packages: mapPackages(result.Packages), SchemaVersion: "", ToolName: "", ToolVersion: "",
	}
}

func mapMetric(entry *projapp.MetricEntry) MetricResult {
	return MetricResult{
		Name:       entry.Name,
		Scope:      entry.Scope,
		Reason:     entry.Reason,
		Definition: entry.Definition,
		Value:      entry.Value,
		Applicable: entry.Applicable,
	}
}

func mapMetrics(entries []projapp.MetricEntry) []MetricResult {
	metrics := make([]MetricResult, zero, len(entries))

	for i := range entries {
		metrics = append(metrics, mapMetric(&entries[i]))
	}

	return metrics
}

func mapPackage(pkg *projapp.PackageResult) PackageReport {
	return PackageReport{
		Path:     pkg.Path,
		Metrics:  mapMetrics(pkg.Metrics),
		Afferent: pkg.Afferent,
		Efferent: pkg.Efferent,
	}
}

func mapPackages(pkgs []projapp.PackageResult) []PackageReport {
	packages := make([]PackageReport, zero, len(pkgs))

	for i := range pkgs {
		packages = append(packages, mapPackage(&pkgs[i]))
	}

	return packages
}

func publicAnalysisError(err error) error {
	if errors.Is(err, context.Canceled) {
		return fmt.Errorf(errContextAnalysis, errors.Join(ErrAnalysis, context.Canceled))
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf(errContextAnalysis, errors.Join(ErrAnalysis, context.DeadlineExceeded))
	}

	return fmt.Errorf(errPublicDetails, ErrAnalysis, err)
}

func toAppOptions(cfg *Config) *projapp.Options {
	return &projapp.Options{
		Directory:        cfg.Directory,
		Patterns:         cfg.Patterns,
		IncludeTests:     cfg.IncludeTests,
		IncludeGenerated: cfg.IncludeGenerated,
		BuildTags:        cfg.BuildTags,
		Workers:          cfg.Workers,
		DependencyScope:  cfg.DependencyScope,
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
		return fmt.Errorf("%w: %q", errInvalidScope, config.DependencyScope)
	}
}

func validatePatterns(config *Config) error {
	if slices.Contains(config.Patterns, emptyString) {
		return errEmptyPattern
	}

	return nil
}
