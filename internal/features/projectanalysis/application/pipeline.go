package application

import (
	"context"

	pkgmetrics "github.com/gostafa/distance/internal/features/packagemetrics/application"
	coupling "github.com/gostafa/distance/internal/features/packagemetrics/domain"
	"github.com/gostafa/distance/internal/features/projectanalysis/ports/inbound"
	typefacts "github.com/gostafa/distance/internal/features/typefacts/application"
	tfdomain "github.com/gostafa/distance/internal/features/typefacts/domain"
	tfoutbound "github.com/gostafa/distance/internal/features/typefacts/ports/outbound"
	"github.com/gostafa/distance/internal/shared/metrics"
	"github.com/gostafa/distance/internal/shared/workerpool"
)

// runWorkers is a seam so tests can force workerpool.Run failures.
var runWorkers = workerpool.Run

// Pipeline implements the inbound Analyzer port.
type Pipeline struct {
	facts typefacts.Collector
}

// NewPipeline returns a pipeline backed by the given fact collector.
func NewPipeline(facts typefacts.Collector) *Pipeline {
	return &Pipeline{facts: facts}
}

var _ inbound.Analyzer = (*Pipeline)(nil)

// Analyze runs the full pipeline for one request.
func (p *Pipeline) Analyze(ctx context.Context, opts inbound.Options) (inbound.Result, error) {
	facts, err := p.facts.Collect(ctx, collectOptions(opts))
	if err != nil {
		return inbound.Result{}, err
	}

	return assembleResult(ctx, &facts, opts)
}

// assembleResult computes the package metrics and every package's results in
// parallel, honouring cancellation, and folds them into the final report.
func assembleResult(
	ctx context.Context,
	facts *tfdomain.ProjectFacts,
	opts inbound.Options,
) (inbound.Result, error) {
	if err := ctx.Err(); err != nil {
		return inbound.Result{}, err
	}

	// The dependency graph is cheap and feeds the structural Ca/Ce facts,
	// so it is built regardless.
	graph := coupling.BuildDependencyGraph(facts, coupling.Scope(opts.DependencyScope))
	pkgResults := pkgmetrics.ComputeForPackages(facts, graph)

	packageResults := make([]inbound.PackageResult, len(facts.Packages))
	workers := workerpool.Workers(opts.Workers, len(facts.Packages))

	err := runWorkers(ctx, workers, len(facts.Packages), func(i int) error {
		packageResults[i] = analyzePackage(facts, i, graph.Coupling(i), pkgResults)

		return nil
	})
	if err != nil {
		return inbound.Result{}, err
	}

	return inbound.Result{ModulePath: facts.ModulePath, Packages: packageResults}, nil
}

// collectOptions maps the analysis request onto the fact-source options.
func collectOptions(opts inbound.Options) tfoutbound.FactOptions {
	return tfoutbound.FactOptions{
		Directory:        opts.Directory,
		Patterns:         opts.Patterns,
		IncludeTests:     opts.IncludeTests,
		IncludeGenerated: opts.IncludeGenerated,
		BuildTags:        opts.BuildTags,
		Workers:          opts.Workers,
		ContinueOnError:  opts.ContinueOnError,
	}
}

// analyzePackage assembles one package's coupling facts and its reported
// metrics. It writes only into its own result value, so package workers never
// share mutable state.
func analyzePackage(
	facts *tfdomain.ProjectFacts,
	pkgID int,
	pkgCoupling coupling.Coupling,
	pkgResults []pkgmetrics.Result,
) inbound.PackageResult {
	pkg := &facts.Packages[pkgID]

	return inbound.PackageResult{
		Path:     pkg.Path,
		Afferent: pkgCoupling.Afferent,
		Efferent: pkgCoupling.Efferent,
		Metrics:  packageMetrics(pkgResults[pkgID]),
	}
}

// packageMetrics reports abstractness, instability, and distance in the
// fixed public order. Only distance is policy-gateable.
func packageMetrics(result pkgmetrics.Result) []metrics.MetricResult {
	return []metrics.MetricResult{
		result.Abstractness,
		result.Instability,
		result.Distance,
	}
}
