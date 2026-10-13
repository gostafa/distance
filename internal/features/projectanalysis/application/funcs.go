// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	"context"
	"fmt"

	coupling "github.com/gostafa/distance/internal/features/packagemetrics/public"
	executionport "github.com/gostafa/distance/internal/features/projectanalysis/ports/execution"
	"github.com/gostafa/distance/internal/features/projectanalysis/ports/outbound"
)

// NewPipeline returns a Pipeline that collects facts via facts.
func NewPipeline(facts outbound.FactCollector, workers executionport.WorkerExecutor) *Pipeline {
	return pipelineWithWorkers(facts, workers.Run)
}

// Analyze collects type facts and computes package metrics for opts.
func (pipe *Pipeline) Analyze(ctx context.Context, opts *Options) (Result, error) {
	result, err := pipe.analyze(ctx, opts)
	if err != nil {
		return Result{ModulePath: "", Packages: nil}, fmt.Errorf(errWrapAnalyze, err)
	}

	return result, nil
}

func assemble(ctx context.Context, runtime *pipelineRuntime, args *assembleIn) (Result, error) {
	err := ctx.Err()
	if err != nil {
		return Result{ModulePath: "", Packages: nil}, fmt.Errorf("assemble result: %w", err)
	}

	result, err := buildResults(ctx, runtime, args)
	if err != nil {
		return Result{ModulePath: "", Packages: nil}, fmt.Errorf(errBuildPackages, err)
	}

	return result, nil
}

func buildResults(ctx context.Context, runtime *pipelineRuntime, args *assembleIn) (Result, error) {
	rows := emptyPackageResults(len(args.facts.Packages))
	fill := packageResultFiller(args, rows)

	err := runtime.runWorkers(
		ctx,
		executionport.WorkerRun{Workers: args.opts.Workers, Tasks: len(rows)},
		fill,
	)
	if err != nil {
		return Result{ModulePath: "", Packages: nil}, fmt.Errorf(errBuildPackages, err)
	}

	return Result{ModulePath: args.facts.ModulePath, Packages: rows}, nil
}

func collectOptions(opts *Options) *outbound.FactOptions {
	return &outbound.FactOptions{
		Directory:        opts.Directory,
		Patterns:         opts.Patterns,
		IncludeTests:     opts.IncludeTests,
		IncludeGenerated: opts.IncludeGenerated,
		BuildTags:        opts.BuildTags,
		Workers:          opts.Workers,
		ContinueOnError:  opts.ContinueOnError,
	}
}

func emptyPackageResults(count int) []PackageResult {
	results := make([]PackageResult, zero, count)

	for range count {
		results = append(
			results,
			PackageResult{Path: "", Metrics: nil, Afferent: zero, Efferent: zero},
		)
	}

	return results
}

func packageResultFiller(args *assembleIn, rows []PackageResult) func(int) error {
	compute := coupling.Prepare(args.facts, args.opts.DependencyScope)

	return func(index int) error {
		metrics := compute(index)

		rows[index] = PackageResult{
			Path: args.facts.Packages[index].Path, Metrics: metrics.Metrics,
			Afferent: metrics.Afferent, Efferent: metrics.Efferent,
		}

		return nil
	}
}

func pipelineWithWorkers(facts outbound.FactCollector, runWorkers workerRunFunc) *Pipeline {
	return &Pipeline{
		analyze: func(ctx context.Context, opts *Options) (Result, error) {
			return runPipeline(ctx, &pipelineRuntime{facts: facts, runWorkers: runWorkers}, opts)
		},
	}
}

func runPipeline(ctx context.Context, runtime *pipelineRuntime, opts *Options) (Result, error) {
	facts, err := runtime.facts.Collect(ctx, collectOptions(opts))
	if err != nil {
		return Result{ModulePath: "", Packages: nil}, fmt.Errorf(errWrapAnalyze, err)
	}

	result, err := assemble(ctx, runtime, &assembleIn{facts: &facts, opts: opts})
	if err != nil {
		return Result{ModulePath: "", Packages: nil}, fmt.Errorf("application assemble: %w", err)
	}

	return result, nil
}
