package application

import (
	"context"
	"errors"
	"testing"

	"github.com/gostafa/distance/internal/features/projectanalysis/ports/inbound"
	typefacts "github.com/gostafa/distance/internal/features/typefacts/application"
	tfdomain "github.com/gostafa/distance/internal/features/typefacts/domain"
	tfoutbound "github.com/gostafa/distance/internal/features/typefacts/ports/outbound"
	"github.com/gostafa/distance/internal/shared/metrics"
)

type coverageSource struct{}

func (coverageSource) Load(
	context.Context,
	tfoutbound.FactOptions,
) (string, []tfdomain.PackageExtract, error) {
	return "example.com/m", []tfdomain.PackageExtract{{
		Path: "example.com/m/a", InModule: true,
		Types: []tfdomain.TypeExtract{
			{
				Name: "A",
				Kind: tfdomain.KindStruct,
			},
		},
	}}, nil
}

func TestAssembleResultWorkerError(t *testing.T) {
	original := runWorkers
	t.Cleanup(func() { runWorkers = original })

	sentinel := errors.New("workers failed")
	runWorkers = func(context.Context, int, int, func(int) error) error {
		return sentinel
	}

	pipeline := NewPipeline(typefacts.NewService(coverageSource{}))
	_, err := pipeline.Analyze(context.Background(), inbound.Options{
		Patterns: []string{"./..."},
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("Analyze error = %v, want sentinel", err)
	}
}

func TestReportedMetrics(t *testing.T) {
	pipeline := NewPipeline(typefacts.NewService(coverageSource{}))
	result, err := pipeline.Analyze(context.Background(), inbound.Options{
		Patterns:        []string{"./..."},
		DependencyScope: "project",
	})
	if err != nil {
		t.Fatal(err)
	}

	pkg := result.Packages[0]
	want := []string{metrics.MetricAbstractness, metrics.MetricInstability, metrics.MetricDistance}
	if len(pkg.Metrics) != len(want) {
		t.Fatalf("package metrics = %v, want %v", pkg.Metrics, want)
	}

	for i, name := range want {
		if pkg.Metrics[i].Name != name {
			t.Fatalf("package metrics = %v, want %v", pkg.Metrics, want)
		}
	}
}
