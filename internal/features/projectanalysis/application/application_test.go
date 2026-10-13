// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	"context"
	"errors"
	"testing"

	executionport "github.com/gostafa/distance/internal/features/projectanalysis/ports/execution"
	"github.com/gostafa/distance/internal/features/projectanalysis/ports/outbound"
)

type testWorkers struct{}

func (testWorkers) Run(
	ctx context.Context,
	cfg executionport.WorkerRun,
	task func(int) error,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	for index := range cfg.Tasks {
		if err := task(index); err != nil {
			return err
		}
	}
	return nil
}

func TestAssembleResultWorkerError(t *testing.T) {
	sentinel := errors.New("workers failed")
	pipeline := pipelineWithWorkers(
		coverageCollector{},
		func(context.Context, executionport.WorkerRun, func(int) error) error {
			return sentinel
		},
	)

	_, err := pipeline.Analyze(t.Context(), &Options{
		Patterns: []string{"./..."},
	})

	if !errors.Is(err, sentinel) {
		t.Fatalf("Analyze error = %v, want sentinel", err)
	}
}

func TestReportedMetrics(t *testing.T) {
	pipeline := NewPipeline(coverageCollector{}, testWorkers{})

	result, err := pipeline.Analyze(t.Context(), &Options{
		Patterns:        []string{"./..."},
		DependencyScope: "project",
	})
	if err != nil {
		t.Fatal(err)
	}

	pkg := result.Packages[0]
	want := []string{
		string(MetricAbstractness),
		string(MetricInstability),
		string(MetricDistance),
	}

	if len(pkg.Metrics) != len(want) {
		t.Fatalf("package metrics = %v, want %v", pkg.Metrics, want)
	}

	for i, name := range want {
		if pkg.Metrics[i].Name != name {
			t.Fatalf("package metrics = %v, want %v", pkg.Metrics, want)
		}
	}
}

// White-box: the request→fact-options mapping.
func TestCollectOptionsMapping(t *testing.T) {
	t.Parallel()

	fo := collectOptions(&Options{
		Directory: "d", Patterns: []string{"./..."}, IncludeTests: true,
		IncludeGenerated: true, BuildTags: []string{"tag"}, Workers: 3, ContinueOnError: true,
	})

	if fo.Directory != "d" || !fo.IncludeTests || !fo.IncludeGenerated ||
		fo.Workers != 3 || !fo.ContinueOnError || len(fo.BuildTags) != 1 {

		t.Fatalf("collectOptions = %+v", fo)
	}
}

type coverageCollector struct{}

func (coverageCollector) Collect(
	ctx context.Context,
	opts *outbound.FactOptions,
) (outbound.ProjectFacts, error) {
	return outbound.ProjectFacts{ModulePath: "example.com/m", Packages: []struct {
		Path     string
		Imports  []string
		TypeIDs  []int
		ID       int
		InModule bool
	}{{Path: "example.com/m/a", Imports: nil, TypeIDs: []int{0}, ID: 0, InModule: true}}, Types: []struct {
		Name      string
		ID        int
		PackageID int
		Kind      uint8
	}{{Name: "A", ID: 0, PackageID: 0, Kind: 0}}}, nil
}
