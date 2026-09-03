// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package wire

import (
	"context"
	"fmt"

	"github.com/gostafa/distance/distance"
	"github.com/gostafa/distance/internal/features/packagemetrics/domain"
	projapp "github.com/gostafa/distance/internal/features/projectanalysis/application"
	"github.com/gostafa/distance/internal/shared/version"
)

type (
	backend struct {
		analyze func(context.Context, *distance.Config) (distance.Report, error)
	}

	analyzeIn = struct {
		cfg      *distance.Config
		analyzer *projapp.Pipeline
	}
)

// Analyze runs the injected analyze function for cfg.
func (b backend) Analyze(ctx context.Context, cfg *distance.Config) (distance.Report, error) {
	report, err := b.analyze(ctx, cfg)
	if err != nil {
		return distance.Report{}, fmt.Errorf(errWrapBackend, err)
	}

	return report, nil
}

// AnalyzeWithDefault runs Analyze using the built-in analyzer backend.
func AnalyzeWithDefault(ctx context.Context, cfg *distance.Config) (distance.Report, error) {
	report, err := analyzeWithPipeline(ctx, &analyzeIn{
		cfg:      cfg,
		analyzer: projapp.NewDefaultPipeline(),
	})
	if err != nil {
		return distance.Report{}, fmt.Errorf(errWrapAnalyze, err)
	}

	report.ToolName, report.ToolVersion = distance.ToolIdent(
		distance.MetricDistance,
		version.Version(),
	)

	return report, nil
}

func analyzeWithPipeline(ctx context.Context, in *analyzeIn) (distance.Report, error) {
	report, err := distance.Analyze(ctx, in.cfg, backendFor(in.analyzer))
	if err != nil {
		return distance.Report{}, fmt.Errorf(errWrapAnalyze, err)
	}

	return report, nil
}

func backendFor(analyzer *projapp.Pipeline) backend {
	return backend{
		analyze: func(ctx context.Context, cfg *distance.Config) (distance.Report, error) {
			result, analyzeErr := analyzer.Analyze(ctx, toAppOptions(cfg))
			if analyzeErr != nil {
				return distance.Report{}, fmt.Errorf(errWrapBackend, analyzeErr)
			}

			return fromAppResult(&result), nil
		},
	}
}

func toAppOptions(cfg *distance.Config) *projapp.Options {
	return &projapp.Options{
		Directory:        cfg.Directory,
		Patterns:         cfg.Patterns,
		IncludeTests:     cfg.IncludeTests,
		IncludeGenerated: cfg.IncludeGenerated,
		BuildTags:        cfg.BuildTags,
		Workers:          cfg.Workers,
		DependencyScope:  wireScope(cfg.DependencyScope),
		ContinueOnError:  cfg.ContinueOnError,
	}
}

func wireScope(scope string) string {
	scopes := [...]string{domain.ScopeProject, domain.ScopeModule, domain.ScopeAll}

	for i := range scopes {
		if scope == scopes[i] {
			return scopes[i]
		}
	}

	return scope
}

func fromAppResult(result *projapp.Result) distance.Report {
	return distance.Report{
		Module:   result.ModulePath,
		Packages: mapPackages(result.Packages),
	}
}

func mapPackages(pkgs []projapp.PackageResult) []distance.PackageReport {
	packages := make([]distance.PackageReport, zero, len(pkgs))

	for i := range pkgs {
		packages = append(packages, mapPackage(&pkgs[i]))
	}

	return packages
}

func mapPackage(pkg *projapp.PackageResult) distance.PackageReport {
	return distance.PackageReport{
		Path:     pkg.Path,
		Metrics:  mapMetrics(pkg.Metrics),
		Afferent: pkg.Afferent,
		Efferent: pkg.Efferent,
	}
}

func mapMetrics(entries []projapp.MetricEntry) []distance.MetricResult {
	metrics := make([]distance.MetricResult, zero, len(entries))

	for i := range entries {
		metrics = append(metrics, mapMetric(&entries[i]))
	}

	return metrics
}

func mapMetric(entry *projapp.MetricEntry) distance.MetricResult {
	return distance.MetricResult{
		Name:       entry.Name,
		Scope:      entry.Scope,
		Reason:     entry.Reason,
		Definition: entry.Definition,
		Value:      entry.Value,
		Applicable: entry.Applicable,
	}
}
