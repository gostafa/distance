// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	"context"
	"fmt"
	"math"

	coupling "github.com/gostafa/distance/internal/features/packagemetrics/domain"
	typefacts "github.com/gostafa/distance/internal/features/typefacts/application"
	tfoutbound "github.com/gostafa/distance/internal/features/typefacts/ports/outbound"
	"github.com/gostafa/distance/internal/infrastructure/goloader"
)

func defaultPipelineRuntime() pipelineRuntime {
	return pipelineRuntime{runWorkers: goloader.RunWorkers}
}

// NewPipeline returns a Pipeline that collects facts via facts.
func NewPipeline(facts typefacts.Collector) *Pipeline {
	return &Pipeline{facts: facts, runtime: defaultPipelineRuntime()}
}

// NewDefaultPipeline returns a Pipeline wired with the default goloader fact source.
func NewDefaultPipeline() *Pipeline {
	return NewPipeline(typefacts.NewService(goloader.New()))
}

// Analyze collects type facts and computes package metrics for opts.
func (pipe *Pipeline) Analyze(ctx context.Context, opts *Options) (Result, error) {
	facts, err := pipe.facts.Collect(ctx, collectOptions(opts))
	if err != nil {
		return Result{}, fmt.Errorf("application Analyze: %w", err)
	}

	result, err := pipe.assemble(ctx, &assembleIn{facts: &facts, opts: opts})
	if err != nil {
		return Result{}, fmt.Errorf("application assemble: %w", err)
	}

	return result, nil
}

func (pipe *Pipeline) assemble(ctx context.Context, args *assembleIn) (Result, error) {
	err := ctx.Err()
	if err != nil {
		return Result{}, fmt.Errorf("application assembleResult: %w", err)
	}

	result, err := pipe.buildResults(ctx, args)
	if err != nil {
		return Result{}, fmt.Errorf("application build packages: %w", err)
	}

	return result, nil
}

func (pipe *Pipeline) buildResults(ctx context.Context, args *assembleIn) (Result, error) {
	graph := coupling.BuildDependencyGraph(args.facts, args.opts.DependencyScope)
	packageResults := emptyPackageResults(args.facts.PackageCount())

	err := pipe.runtime.runWorkers(ctx, goloader.WorkerRun{
		Workers: goloader.Workers(args.opts.Workers, args.facts.PackageCount()),
		Tasks:   args.facts.PackageCount(),
	}, fillPackageAt(&fillInput{facts: args.facts, graph: &graph, rows: packageResults}))
	if err != nil {
		return Result{}, fmt.Errorf("application buildPackageResults: %w", err)
	}

	return Result{ModulePath: args.facts.ModulePath, Packages: packageResults}, nil
}

func emptyPackageResults(count int) []PackageResult {
	results := make([]PackageResult, zero, count)

	for range count {
		results = append(results, PackageResult{})
	}

	return results
}

func fillPackageAt(input *fillInput) func(int) error {
	pkgResults := computeForPackages(&computeIn{facts: input.facts, graph: input.graph})

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

func collectOptions(opts *Options) *tfoutbound.FactOptions {
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

func analyzePackage(input *analyzePackageInput) PackageResult {
	pkg := &input.facts.Packages[input.pkgID]
	metrics := input.pkgResults[input.pkgID]

	return PackageResult{
		Path:     pkg.Path,
		Afferent: input.afferent,
		Efferent: input.efferent,
		Metrics: []MetricEntry{
			metrics.abstractness,
			metrics.instability,
			metrics.distance,
		},
	}
}

func computeForPackages(in *computeIn) []packageMetrics {
	results := make([]packageMetrics, zero, in.facts.PackageCount())

	for pkgID := range in.facts.Packages {
		results = append(results, computeOne(in, pkgID))
	}

	return results
}

func computeOne(in *computeIn, pkgID int) packageMetrics {
	interfaces, total := coupling.CountTypes(in.facts, pkgID)
	afferent, efferent := in.graph.PackageCoupling(pkgID)
	abstractness := abstractnessMetric(interfaces, total)
	instability := instabilityMetric(afferent, efferent)

	return packageMetrics{
		abstractness: abstractness,
		instability:  instability,
		distance:     distanceMetric(&abstractness, &instability),
	}
}

func abstractnessMetric(namedInterfaceTypes, totalRelevantNamedTypes int) MetricEntry {
	if totalRelevantNamedTypes == zero {
		return notApplicableMetric(
			MetricAbstractness,
			"package declares no relevant named types",
		)
	}

	ratio := float64(namedInterfaceTypes) / float64(totalRelevantNamedTypes)

	return applicableMetric(MetricAbstractness, DefinitionAbstractness, ratio)
}

func distanceMetric(abstractness, instability *MetricEntry) MetricEntry {
	if !abstractness.Applicable {
		return notApplicableMetric(
			MetricDistance,
			"abstractness is not applicable: "+abstractness.Reason,
		)
	}

	if !instability.Applicable {
		return notApplicableMetric(
			MetricDistance,
			"instability is not applicable: "+instability.Reason,
		)
	}

	value := math.Abs(abstractness.Value + instability.Value - mainSequenceBalance)

	return applicableMetric(MetricDistance, DefinitionDistance, value)
}

func instabilityMetric(afferent, efferent int) MetricEntry {
	if afferent+efferent == zero {
		result := applicableMetric(MetricInstability, DefinitionInstability, float64(zero))

		result.Reason = "package has no dependencies in scope (isolated); defined as 0"

		return result
	}

	ratio := float64(efferent) / float64(afferent+efferent)

	return applicableMetric(MetricInstability, DefinitionInstability, ratio)
}

func applicableMetric(name, definition string, value float64) MetricEntry {
	return MetricEntry{
		Name:       name,
		Scope:      ScopePackage,
		Value:      value,
		Applicable: true,
		Definition: definition,
	}
}

func notApplicableMetric(name, reason string) MetricEntry {
	definition := DefinitionAbstractness

	if name == MetricDistance {
		definition = DefinitionDistance
	}

	return MetricEntry{
		Name:       name,
		Scope:      ScopePackage,
		Applicable: false,
		Reason:     reason,
		Definition: definition,
	}
}
