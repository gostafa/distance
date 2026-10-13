// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package composition_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/gostafa/distance/internal/composition"
	application "github.com/gostafa/distance/internal/features/projectanalysis/public"
	"github.com/gostafa/distance/internal/features/typefacts/adapters/goloader"
	tfmodel "github.com/gostafa/distance/internal/features/typefacts/domain/model"
	tfoutbound "github.com/gostafa/distance/internal/features/typefacts/ports/outbound"
	typefacts "github.com/gostafa/distance/internal/features/typefacts/public"
)

func TestAnalyzeFixture(t *testing.T) {
	t.Parallel()

	result, err := composition.NewAnalyzer(goloader.New()).
		Analyze(t.Context(), &application.Options{
			Directory:       filepath.Join("..", "..", "testdata", "fixture"),
			Patterns:        []string{"./..."},
			DependencyScope: "module",
		})
	if err != nil {
		t.Fatal(err)
	}

	if result.ModulePath != "example.com/fixture" {
		t.Fatalf("module = %q", result.ModulePath)
	}

	if len(result.Packages) < 7 {
		t.Fatalf("packages = %d, want >= 7", len(result.Packages))
	}

	for i := 1; i < len(result.Packages); i++ {
		if result.Packages[i-1].Path > result.Packages[i].Path {
			t.Fatalf(
				"packages not sorted: %s before %s",
				result.Packages[i-1].Path,
				result.Packages[i].Path,
			)
		}
	}
}

type errSource struct{ err error }

func (e errSource) Load(
	context.Context,
	*tfoutbound.FactOptions,
) (string, []tfmodel.PackageExtract, error) {
	return "", nil, e.err
}

// Black-box: a fact-source failure propagates out of Analyze.
func TestPipelineLoadError(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("load failed")
	pipeline := composition.NewAnalyzer(errSource{err: sentinel})

	_, err := pipeline.Analyze(t.Context(), &application.Options{
		Patterns: []string{"./..."},
	})

	if !errors.Is(err, sentinel) {
		t.Fatalf("error = %v, want sentinel", err)
	}
}

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
	results []application.MetricEntry,
	name string,
) application.MetricEntry {
	t.Helper()

	for _, r := range results {
		if r.Name == name {
			return r
		}
	}

	t.Fatalf("metric %q not present in %v", name, results)

	return application.MetricEntry{}
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
						Kind: typefacts.KindStruct,
					},
				},
			},
			{
				Path:     "example.com/m/b",
				InModule: true,
				Types: []tfmodel.TypeName{
					{Name: "B", Kind: typefacts.KindInterface},
				},
			},
		},
	}
	pipeline := composition.NewAnalyzer(src)

	result, err := pipeline.Analyze(t.Context(), &application.Options{
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

	if got := findMetric(
		t,
		pkgA.Metrics,
		string(application.MetricAbstractness),
	); !got.Applicable {
		t.Errorf("a abstractness = %+v, want applicable", got)
	}

	if got := findMetric(
		t,
		pkgA.Metrics,
		string(application.MetricInstability),
	); !got.Applicable {
		t.Errorf("a instability = %+v, want applicable", got)
	}

	if got := findMetric(t, pkgA.Metrics, string(application.MetricDistance)); !got.Applicable {
		t.Errorf("a distance = %+v, want applicable", got)
	}

	if got := findMetric(
		t,
		result.Packages[1].Metrics,
		string(application.MetricDistance),
	); !got.Applicable {
		t.Errorf("b distance = %+v, want applicable", got)
	}
}

// Black-box: a canceled context aborts before doing work.
func TestPipelineCancelled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	pipeline := composition.NewAnalyzer(fakeSource{mod: "m"})

	if _, err := pipeline.Analyze(
		ctx,
		&application.Options{Patterns: []string{"./..."}},
	); err == nil {
		t.Fatal("canceled context should abort")
	}
}
