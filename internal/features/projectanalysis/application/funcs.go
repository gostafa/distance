// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	"context"
	"fmt"

	pkgmetrics "github.com/gostafa/distance/internal/features/packagemetrics/application"
	coupling "github.com/gostafa/distance/internal/features/packagemetrics/domain"
	"github.com/gostafa/distance/internal/features/projectanalysis/ports/inbound"
	typefacts "github.com/gostafa/distance/internal/features/typefacts/application"
	tfoutbound "github.com/gostafa/distance/internal/features/typefacts/ports/outbound"
	"github.com/gostafa/distance/internal/shared/metrics"
	"github.com/gostafa/distance/internal/shared/workerpool"
)

func defaultPipelineRuntime() pipelineRuntime {
	return pipelineRuntime{runWorkers: workerpool.Run}
}

// NewPipeline returns a Pipeline that collects facts via facts.
func NewPipeline(facts typefacts.Collector) *Pipeline {
	return &Pipeline{facts: facts, runtime: defaultPipelineRuntime()}
}

// Analyze collects type facts and computes package metrics for opts.
func (pipe *Pipeline) Analyze(ctx context.Context, opts *inbound.Options) (inbound.Result, error) {
	facts, err := pipe.facts.Collect(ctx, collectOptions(opts))
	if err != nil {
		return inbound.Result{}, fmt.Errorf("application Analyze: %w", err)
	}

	result, err := pipe.assemble(ctx, &assembleIn{facts: &facts, opts: opts})
	if err != nil {
		return inbound.Result{}, fmt.Errorf("application assemble: %w", err)
	}

	return result, nil
}

func (pipe *Pipeline) assemble(ctx context.Context, args *assembleIn) (inbound.Result, error) {
	err := ctx.Err()
	if err != nil {
		return inbound.Result{}, fmt.Errorf("application assembleResult: %w", err)
	}

	result, err := pipe.buildResults(ctx, args)
	if err != nil {
		return inbound.Result{}, fmt.Errorf("application build packages: %w", err)
	}

	return result, nil
}

func (pipe *Pipeline) buildResults(ctx context.Context, args *assembleIn) (inbound.Result, error) {
	graph := coupling.BuildDependencyGraph(args.facts, args.opts.DependencyScope)
	packageResults := emptyPackageResults(args.facts.PackageCount())

	err := pipe.runtime.runWorkers(ctx, &workerpool.Config{
		Workers: workerpool.Workers(args.opts.Workers, args.facts.PackageCount()),
		Tasks:   args.facts.PackageCount(),
	}, fillPackageAt(&fillInput{facts: args.facts, graph: &graph, rows: packageResults}))
	if err != nil {
		return inbound.Result{}, fmt.Errorf("application buildPackageResults: %w", err)
	}

	return inbound.Result{ModulePath: args.facts.ModulePath, Packages: packageResults}, nil
}

func emptyPackageResults(count int) []inbound.PackageResult {
	results := make([]inbound.PackageResult, zero, count)

	for range count {
		results = append(results, inbound.PackageResult{})
	}

	return results
}

func fillPackageAt(input *fillInput) func(int) error {
	pkgResults := pkgmetrics.ComputeForPackages(input.facts, input.graph)

	return func(index int) error {
		afferent, efferent := input.graph.PackageCoupling(index)

		input.rows[index] = analyzePackage(&analyzePackageInput{
			facts:      input.facts,
			pkgID:      index,
			afferent:   afferent,
			efferent:   efferent,
			pkgResults: pkgResults,
		})

		return nil
	}
}

func collectOptions(opts *inbound.Options) *tfoutbound.FactOptions {
	return &tfoutbound.FactOptions{
		Directory:        opts.Directory,
		Patterns:         opts.Patterns,
		IncludeTests:     opts.IncludeTests,
		IncludeGenerated: opts.IncludeGenerated,
		BuildTags:        opts.BuildTags,
		Workers:          opts.Workers,
		ContinueOnError:  opts.ContinueOnError,
	}
}

func analyzePackage(input *analyzePackageInput) inbound.PackageResult {
	pkg := &input.facts.Packages[input.pkgID]

	return inbound.PackageResult{
		Path:     pkg.Path,
		Afferent: input.afferent,
		Efferent: input.efferent,
		Metrics:  packageMetrics(&input.pkgResults[input.pkgID]),
	}
}

func packageMetrics(result *pkgmetrics.Result) []metrics.MetricResult {
	return []metrics.MetricResult{
		result.Abstractness,
		result.Instability,
		result.Distance,
	}
}
