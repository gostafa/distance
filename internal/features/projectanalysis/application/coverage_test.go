// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	"context"
	"errors"
	"testing"

	"github.com/gostafa/distance/distance"
	typefacts "github.com/gostafa/distance/internal/features/typefacts/application"
	tfdomain "github.com/gostafa/distance/internal/features/typefacts/domain"
	tfmodel "github.com/gostafa/distance/internal/features/typefacts/domain/model"
	tfoutbound "github.com/gostafa/distance/internal/features/typefacts/ports/outbound"
	"github.com/gostafa/distance/internal/infrastructure/goloader"
)

type coverageSource struct{}

func (coverageSource) Load(
	context.Context,
	*tfoutbound.FactOptions,
) (string, []tfmodel.PackageExtract, error) {
	return "example.com/m", []tfmodel.PackageExtract{{
		Path: "example.com/m/a", InModule: true,
		Types: []tfmodel.TypeName{
			{
				Name: "A",
				Kind: tfdomain.KindStruct,
			},
		},
	}}, nil
}

func TestAssembleResultWorkerError(t *testing.T) {
	sentinel := errors.New("workers failed")
	pipeline := pipelineWithWorkers(
		typefacts.NewService(coverageSource{}),
		func(context.Context, goloader.WorkerRun, func(int) error) error {
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
	pipeline := NewPipeline(typefacts.NewService(coverageSource{}))

	result, err := pipeline.Analyze(t.Context(), &Options{
		Patterns:        []string{"./..."},
		DependencyScope: "project",
	})
	if err != nil {
		t.Fatal(err)
	}

	pkg := result.Packages[0]
	want := []string{
		string(distance.MetricAbstractness),
		string(distance.MetricInstability),
		string(distance.MetricDistance),
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
