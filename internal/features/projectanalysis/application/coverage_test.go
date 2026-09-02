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
				Name:     "A",
				Exported: true,
				Kind:     tfdomain.KindStruct,
				Methods: []tfdomain.MethodFacts{
					{Name: "Do", Exported: true},
				},
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

func TestReportedMetricIsDistanceOnly(t *testing.T) {
	pipeline := NewPipeline(typefacts.NewService(coverageSource{}))
	result, err := pipeline.Analyze(context.Background(), inbound.Options{
		Patterns:        []string{"./..."},
		DependencyScope: "project",
	})
	if err != nil {
		t.Fatal(err)
	}

	pkg := result.Packages[0]
	if len(pkg.Metrics) != 1 || pkg.Metrics[0].Name != metrics.MetricDistance {
		t.Fatalf("package metrics = %v, want distance only", pkg.Metrics)
	}
}
