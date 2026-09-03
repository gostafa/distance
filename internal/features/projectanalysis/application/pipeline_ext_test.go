// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application_test

import (
	"context"
	"testing"

	"github.com/gostafa/distance/distance"
	projectanalysis "github.com/gostafa/distance/internal/features/projectanalysis/application"
	typefacts "github.com/gostafa/distance/internal/features/typefacts/application"
	tfdomain "github.com/gostafa/distance/internal/features/typefacts/domain"
	tfmodel "github.com/gostafa/distance/internal/features/typefacts/domain/model"
	tfoutbound "github.com/gostafa/distance/internal/features/typefacts/ports/outbound"
)

// fakeSource feeds canned extracts so the whole pipeline runs without loading
// real packages through go/packages.
type fakeSource struct {
	mod  string
	pkgs []tfmodel.PackageExtract
}

func (f fakeSource) Load(
	context.Context,
	*tfoutbound.FactOptions,
) (string, []tfmodel.PackageExtract, error) {
	return f.mod, f.pkgs, nil
}

func findMetric(
	t *testing.T,
	results []projectanalysis.MetricEntry,
	name string,
) projectanalysis.MetricEntry {
	t.Helper()

	for _, r := range results {
		if r.Name == name {
			return r
		}
	}

	t.Fatalf("metric %q not present in %v", name, results)

	return projectanalysis.MetricEntry{}
}

// Black-box: the pipeline turns extracts into a deterministic report with
// abstractness, instability, and distance as the displayed domain.
func TestPipelineAnalyzeEndToEnd(t *testing.T) {
	t.Parallel()

	src := fakeSource{
		mod: "example.com/m",
		pkgs: []tfmodel.PackageExtract{
			{
				Path: "example.com/m/a", InModule: true, Imports: []string{"example.com/m/b"},
				Types: []tfmodel.TypeName{
					{
						Name: "A",
						Kind: tfdomain.KindStruct,
					},
				},
			},
			{
				Path:     "example.com/m/b",
				InModule: true,
				Types: []tfmodel.TypeName{
					{Name: "B", Kind: tfdomain.KindInterface},
				},
			},
		},
	}
	pipeline := projectanalysis.NewPipeline(typefacts.NewService(src))

	result, err := pipeline.Analyze(t.Context(), &projectanalysis.Options{
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

	if got := findMetric(t, pkgA.Metrics, string(distance.MetricAbstractness)); !got.Applicable {
		t.Errorf("a abstractness = %+v, want applicable", got)
	}

	if got := findMetric(t, pkgA.Metrics, string(distance.MetricInstability)); !got.Applicable {
		t.Errorf("a instability = %+v, want applicable", got)
	}

	if got := findMetric(t, pkgA.Metrics, string(distance.MetricDistance)); !got.Applicable {
		t.Errorf("a distance = %+v, want applicable", got)
	}

	if got := findMetric(
		t,
		result.Packages[1].Metrics,
		string(distance.MetricDistance),
	); !got.Applicable {
		t.Errorf("b distance = %+v, want applicable", got)
	}
}

// Black-box: a canceled context aborts before doing work.
func TestPipelineCancelled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	pipeline := projectanalysis.NewPipeline(typefacts.NewService(fakeSource{mod: "m"}))

	if _, err := pipeline.Analyze(
		ctx,
		&projectanalysis.Options{Patterns: []string{"./..."}},
	); err == nil {
		t.Fatal("canceled context should abort")
	}
}
