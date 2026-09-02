package distance

import (
	"context"

	"github.com/gostafa/distance/internal/features/projectanalysis/ports/inbound"
	"github.com/gostafa/distance/internal/infrastructure/analyzer"
	"github.com/gostafa/distance/internal/shared/version"
)

// Analyze validates the configuration, runs the analysis pipeline once over
// the configured patterns, and returns a deterministic report. The context
// cancels package loading and metric computation.
func Analyze(ctx context.Context, config Config) (Report, error) {
	cfg := configWithDefaults(config)
	if err := validateConfig(cfg); err != nil {
		return Report{}, err
	}

	result, err := analyzer.NewAnalyzer().Analyze(ctx, inbound.Options{
		Directory:        cfg.Directory,
		Patterns:         cfg.Patterns,
		IncludeTests:     cfg.IncludeTests,
		IncludeGenerated: cfg.IncludeGenerated,
		BuildTags:        cfg.BuildTags,
		Workers:          cfg.Workers,
		DependencyScope:  string(cfg.DependencyScope),
		ContinueOnError:  cfg.ContinueOnError,
	})
	if err != nil {
		return Report{}, err
	}

	report := Report{
		SchemaVersion: SchemaVersion,
		Tool:          ToolInfo{Name: ToolName, Version: version.Version},
		Module:        result.ModulePath,
		Packages:      make([]PackageReport, len(result.Packages)),
	}
	for i, pkg := range result.Packages {
		report.Packages[i] = PackageReport{
			Path:     pkg.Path,
			Afferent: pkg.Afferent,
			Efferent: pkg.Efferent,
			Metrics:  pkg.Metrics,
		}
	}

	return report, nil
}
