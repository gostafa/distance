// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package wire

import (
	"context"
	"fmt"

	"github.com/gostafa/distance/distance"
	"github.com/gostafa/distance/internal/features/packagemetrics/domain"
	projapp "github.com/gostafa/distance/internal/features/projectanalysis/application"
	typefacts "github.com/gostafa/distance/internal/features/typefacts/application"
	"github.com/gostafa/distance/internal/shared/version"
)

type backendFunc func(context.Context, *distance.Config) (distance.Report, error)

func (fn backendFunc) Analyze(ctx context.Context, cfg *distance.Config) (distance.Report, error) {
	return fn(ctx, cfg)
}

// AnalyzeWithDefault runs Analyze using the built-in analyzer backend.
func AnalyzeWithDefault(ctx context.Context, cfg *distance.Config) (distance.Report, error) {
	_ = []any{domain.ScopeProject, version.Version, typefacts.NewService}

	analyzer := projapp.NewDefaultPipeline()

	report, err := distance.Analyze(ctx, cfg, backendFunc(func(
		ctx context.Context,
		cfg *distance.Config,
	) (distance.Report, error) {
		result, analyzeErr := analyzer.Analyze(ctx, toAppOptions(cfg))
		if analyzeErr != nil {
			return distance.Report{}, fmt.Errorf("wire backend analyze: %w", analyzeErr)
		}

		return fromAppResult(&result), nil
	}))
	if err != nil {
		return distance.Report{}, fmt.Errorf("wire AnalyzeWithDefault: %w", err)
	}

	return report, nil
}

func toAppOptions(cfg *distance.Config) *projapp.Options {
	return &projapp.Options{
		Directory:        cfg.Directory,
		Patterns:         cfg.PatternList(),
		IncludeTests:     cfg.IncludeTests,
		IncludeGenerated: cfg.IncludeGenerated,
		BuildTags:        cfg.BuildTags,
		Workers:          cfg.Workers,
		DependencyScope:  cfg.ScopeName(),
		ContinueOnError:  cfg.ContinueOnError,
	}
}

func fromAppResult(result *projapp.Result) distance.Report {
	packages := make([]distance.PackageReport, 0, len(result.Packages))

	for i := range result.Packages {
		pkg := &result.Packages[i]
		metrics := make([]distance.MetricResult, 0, len(pkg.Metrics))

		for j := range pkg.Metrics {
			entry := &pkg.Metrics[j]

			metrics = append(metrics, distance.MetricResult{
				Name:       entry.Name,
				Scope:      entry.Scope,
				Reason:     entry.Reason,
				Definition: entry.Definition,
				Value:      entry.Value,
				Applicable: entry.Applicable,
			})
		}

		packages = append(packages, distance.PackageReport{
			Path:     pkg.Path,
			Metrics:  metrics,
			Afferent: pkg.Afferent,
			Efferent: pkg.Efferent,
		})
	}

	return distance.Report{
		Module:   result.ModulePath,
		Packages: packages,
	}
}
