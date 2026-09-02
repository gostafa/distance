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

// analyzePackage assembles one package's structural facts, its reported
// metric, and its types. It writes only into its own result value, so package
// workers never share mutable state.
func analyzePackage(
	facts *tfdomain.ProjectFacts,
	pkgID int,
	pkgCoupling coupling.Coupling,
	pkgResults []pkgmetrics.Result,
) inbound.PackageResult {
	pkg := &facts.Packages[pkgID]

	result := inbound.PackageResult{
		Path:            pkg.Path,
		Afferent:        pkgCoupling.Afferent,
		Efferent:        pkgCoupling.Efferent,
		ExportedFuncs:   pkg.ExportedFuncCount,
		UnexportedFuncs: pkg.UnexportedFuncCount,
		Vars:            pkg.VarCount,
		Consts:          pkg.ConstCount,
		Variables:       declarationResults(pkg.Variables),
		Constants:       declarationResults(pkg.Constants),
		Functions:       functionResults(pkg.Functions),
		Metrics:         packageMetrics(pkgResults[pkgID]),
	}

	result.Types = make([]inbound.TypeResult, 0, len(pkg.TypeIDs))
	for _, typeID := range pkg.TypeIDs {
		result.Types = append(result.Types, typeResult(&facts.Types[typeID]))
	}

	return result
}

// packageMetrics reports distance only; abstractness and instability are its
// internal inputs.
func packageMetrics(result pkgmetrics.Result) []metrics.MetricResult {
	return []metrics.MetricResult{result.Distance}
}

// typeResult carries one type's structural facts. Types have no reported
// metric in this linter.
func typeResult(t *tfdomain.TypeFacts) inbound.TypeResult {
	return inbound.TypeResult{
		Name:          t.Name,
		Exported:      t.Exported,
		Kind:          t.Kind,
		Pos:           t.Pos,
		Fields:        len(t.Fields),
		FieldDetails:  append([]tfdomain.FieldFacts(nil), t.Fields...),
		Methods:       len(t.Methods),
		MethodDetails: methodResults(t),
	}
}

func declarationResults(decls []tfdomain.DeclarationFacts) []inbound.DeclarationResult {
	if len(decls) == 0 {
		return nil
	}

	out := make([]inbound.DeclarationResult, len(decls))
	for i, d := range decls {
		out[i] = inbound.DeclarationResult{
			Name:     d.Name,
			Exported: d.Exported,
			Pos:      d.Pos,
		}
	}

	return out
}

func functionResults(functions []tfdomain.FunctionFacts) []inbound.FunctionResult {
	if len(functions) == 0 {
		return nil
	}

	out := make([]inbound.FunctionResult, len(functions))
	for i, fn := range functions {
		out[i] = inbound.FunctionResult{
			Name:     fn.Name,
			Exported: fn.Exported,
			Pos:      fn.Pos,
			Lines:    fn.Lines,
		}
	}

	return out
}

func methodResults(t *tfdomain.TypeFacts) []inbound.FunctionResult {
	if len(t.Methods) == 0 {
		return nil
	}

	out := make([]inbound.FunctionResult, len(t.Methods))
	for i, method := range t.Methods {
		out[i] = inbound.FunctionResult{
			Name:     method.Name,
			Exported: method.Exported,
			Receiver: t.Name,
			Pos:      method.Pos,
			Lines:    method.Lines,
		}
	}

	return out
}
