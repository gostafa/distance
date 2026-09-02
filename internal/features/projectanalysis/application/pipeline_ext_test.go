package application_test

import (
	"context"
	"testing"

	projectanalysis "github.com/gostafa/distance/internal/features/projectanalysis/application"
	"github.com/gostafa/distance/internal/features/projectanalysis/ports/inbound"
	typefacts "github.com/gostafa/distance/internal/features/typefacts/application"
	tfdomain "github.com/gostafa/distance/internal/features/typefacts/domain"
	tfoutbound "github.com/gostafa/distance/internal/features/typefacts/ports/outbound"
	"github.com/gostafa/distance/internal/shared/metrics"
)

// fakeSource feeds canned extracts so the whole pipeline runs without loading
// real packages through go/packages.
type fakeSource struct {
	mod  string
	pkgs []tfdomain.PackageExtract
}

func (f fakeSource) Load(
	context.Context,
	tfoutbound.FactOptions,
) (string, []tfdomain.PackageExtract, error) {
	return f.mod, f.pkgs, nil
}

func findMetric(t *testing.T, results []metrics.MetricResult, name string) metrics.MetricResult {
	t.Helper()

	for _, r := range results {
		if r.Name == name {
			return r
		}
	}

	t.Fatalf("metric %q not present in %v", name, results)

	return metrics.MetricResult{}
}

// Black-box: the pipeline turns extracts into a deterministic report with
// distance as the only displayed metric.
func TestPipelineAnalyzeEndToEnd(t *testing.T) {
	t.Parallel()

	src := fakeSource{
		mod: "example.com/m",
		pkgs: []tfdomain.PackageExtract{
			{
				Path: "example.com/m/a", InModule: true, Imports: []string{"example.com/m/b"},
				VarCount: 2, ConstCount: 3,
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
			},
			{
				Path:     "example.com/m/b",
				InModule: true,
				Types: []tfdomain.TypeExtract{
					{Name: "B", Exported: true, Kind: tfdomain.KindInterface},
				},
			},
		},
	}
	pipeline := projectanalysis.NewPipeline(typefacts.NewService(src))

	result, err := pipeline.Analyze(context.Background(), inbound.Options{
		Patterns:        []string{"./..."},
		DependencyScope: "project",
	})
	if err != nil {
		t.Fatal(err)
	}

	if result.ModulePath != "example.com/m" {
		t.Fatalf("module = %q", result.ModulePath)
	}

	if len(result.Packages) != 2 || result.Packages[0].Path != "example.com/m/a" ||
		result.Packages[1].Path != "example.com/m/b" {
		t.Fatalf("packages not sorted by path: %+v", result.Packages)
	}

	pkgA := result.Packages[0]
	if pkgA.Vars != 2 || pkgA.Consts != 3 {
		t.Fatalf("a counts = vars %d consts %d, want 2 and 3", pkgA.Vars, pkgA.Consts)
	}

	if got := findMetric(t, pkgA.Metrics, metrics.MetricDistance); !got.Applicable {
		t.Errorf("a distance = %+v, want applicable", got)
	}

	if len(pkgA.Types) != 1 || pkgA.Types[0].Name != "A" {
		t.Fatalf("a types = %+v", pkgA.Types)
	}

	if got := findMetric(t, result.Packages[1].Metrics, metrics.MetricDistance); !got.Applicable {
		t.Errorf("b distance = %+v, want applicable", got)
	}
}

// Black-box: a cancelled context aborts before doing work.
func TestPipelineCancelled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	pipeline := projectanalysis.NewPipeline(typefacts.NewService(fakeSource{mod: "m"}))
	if _, err := pipeline.Analyze(
		ctx,
		inbound.Options{Patterns: []string{"./..."}},
	); err == nil {
		t.Fatal("cancelled context should abort")
	}
}
