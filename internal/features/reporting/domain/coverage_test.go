package domain

import (
	"strings"
	"testing"

	"github.com/gostafa/distance/internal/shared/metrics"
	"github.com/gostafa/distance/distance"
)

func TestRelPathEdges(t *testing.T) {
	if got := relPath("example.com/m/p", ""); got != "example.com/m/p" {
		t.Fatalf("empty module: got %q", got)
	}
	if got := relPath("other.com/x", "example.com/m"); got != "other.com/x" {
		t.Fatalf("outside module: got %q", got)
	}
}

func TestTextEmptyPackages(t *testing.T) {
	got := Text(distance.Report{
		SchemaVersion: "1",
		Tool:          distance.ToolInfo{Name: "distance", Version: "test"},
		Module:        "example.com/m",
	}, TextOptions{})
	if !strings.Contains(got, "module example.com/m") || strings.Contains(got, "PATH / TYPE") {
		t.Fatalf("empty packages output unexpected:\n%s", got)
	}
}

func TestTextMultiSectionSpacerAndMissingMetrics(t *testing.T) {
	report := distance.Report{
		SchemaVersion: "1",
		Tool:          distance.ToolInfo{Name: "distance", Version: "test"},
		Module:        "example.com/m",
		Packages: []distance.PackageReport{
			{
				Path: "example.com/m",
				Metrics: []metrics.MetricResult{
					{
						Name:       metrics.MetricAbstractness,
						Scope:      metrics.ScopePackage,
						Value:      0.5,
						Applicable: true,
					},
					{
						Name:       metrics.MetricDistance,
						Scope:      metrics.ScopePackage,
						Value:      0.1,
						Applicable: true,
					},
				},
				// No types → typesTotal 0 on the root package row.
			},
			{
				Path: "example.com/m/leaf",
				Metrics: []metrics.MetricResult{
					// Missing distance (present elsewhere) → blank trailing cell.
					{
						Name:       metrics.MetricAbstractness,
						Scope:      metrics.ScopePackage,
						Applicable: false,
						Reason:     "isolated",
					},
				},
				Types: []distance.TypeReport{{
					Name: "T",
				}},
			},
		},
	}

	got := Text(report, TextOptions{Color: true, Explain: true})
	if !strings.Contains(got, "\n\n") {
		t.Fatalf("expected blank spacer between sections:\n%s", got)
	}
	if !strings.Contains(got, "abstractness: isolated") {
		t.Fatalf("package metric reason missing:\n%s", got)
	}
	// Unknown metric name has no quality color.
	if strings.Contains(got, ansiGreen+"9.00") || strings.Contains(got, ansiRed+"9.00") {
		t.Fatalf("unknown metric was quality-colored:\n%q", got)
	}
}

func TestTextExplainAllTypesAndSkipEmptyNotes(t *testing.T) {
	report := distance.Report{
		SchemaVersion: "1",
		Tool:          distance.ToolInfo{Name: "distance", Version: "test"},
		Module:        "example.com/m",
		Packages: []distance.PackageReport{
			{
				Path: "example.com/m/quiet",
				Metrics: []metrics.MetricResult{
					{
						Name:       metrics.MetricAbstractness,
						Scope:      metrics.ScopePackage,
						Value:      0,
						Applicable: true,
					},
				},
			},
			{
				Path: "example.com/m/noisy",
				Types: []distance.TypeReport{
					{Name: "A"},
					{Name: "B"},
				},
			},
		},
	}

	got := Text(report, TextOptions{Explain: true})
	if strings.Contains(got, "notes") {
		t.Fatalf("no reported metric reasons, notes should be absent:\n%s", got)
	}
}

func TestValueColorUnknownMetric(t *testing.T) {
	if got := valueColor("not-a-metric", 1, &columnStats{min: 0, max: 2, count: 2}); got != "" {
		t.Fatalf("valueColor = %q, want empty", got)
	}
}

func TestMeanCellNilStats(t *testing.T) {
	cell := meanCell(nil, func(float64) string { return "" })
	if cell.text != naCell {
		t.Fatalf("meanCell(nil) = %q, want %q", cell.text, naCell)
	}
}

func TestTextTrailingBlankPackageMetric(t *testing.T) {
	// Package-only columns: one package lacks the last metric so the blank
	// trailing cell is trimmed when the row is written.
	report := distance.Report{
		SchemaVersion: "1",
		Tool:          distance.ToolInfo{Name: "distance", Version: "test"},
		Module:        "example.com/m",
		Packages: []distance.PackageReport{
			{
				Path: "example.com/m/a",
				Metrics: []metrics.MetricResult{
					{
						Name:       metrics.MetricAbstractness,
						Scope:      metrics.ScopePackage,
						Value:      1,
						Applicable: true,
					},
					{
						Name:       metrics.MetricDistance,
						Scope:      metrics.ScopePackage,
						Value:      0,
						Applicable: true,
					},
				},
			},
			{
				Path: "example.com/m/b",
				Metrics: []metrics.MetricResult{
					{
						Name:       metrics.MetricAbstractness,
						Scope:      metrics.ScopePackage,
						Value:      0,
						Applicable: true,
					},
				},
			},
		},
	}
	got := Text(report, TextOptions{})
	mustMatch(t, got, `(?m)^b\s+0\s+0\s+0\s+0$`)
}
