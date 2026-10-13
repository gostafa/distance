// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package presentation

import (
	"regexp"
	"strings"
	"testing"

	reportmodel "github.com/gostafa/distance/internal/features/reporting/domain"
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
	got := Text(&reportmodel.Report{
		SchemaVersion: "1",
		ToolName:      "distance", ToolVersion: "test",
		Module: "example.com/m",
	}, &TextOptions{})

	if !strings.Contains(got, "module example.com/m") || strings.Contains(got, "PATH / TYPE") {
		t.Fatalf("empty packages output unexpected:\n%s", got)
	}
}

func TestTextMultiSectionSpacerAndMissingMetrics(t *testing.T) {
	report := reportmodel.Report{
		SchemaVersion: "1",
		ToolName:      "distance", ToolVersion: "test",
		Module: "example.com/m",
		Packages: []reportmodel.PackageReport{
			{
				Path: "example.com/m",
				Metrics: []reportmodel.MetricResult{
					{
						Name:       string(reportmodel.MetricAbstractness),
						Scope:      reportmodel.ScopePackage,
						Value:      0.5,
						Applicable: true,
					},
					{
						Name:       string(reportmodel.MetricDistance),
						Scope:      reportmodel.ScopePackage,
						Value:      0.1,
						Applicable: true,
					},
				},
				// No types → typesTotal 0 on the root package row.
			},
			{
				Path: "example.com/m/leaf",
				Metrics: []reportmodel.MetricResult{
					// Missing distance (present elsewhere) → blank trailing cell.
					{
						Name:       string(reportmodel.MetricAbstractness),
						Scope:      reportmodel.ScopePackage,
						Applicable: false,
						Reason:     "isolated",
					},
				},
			},
		},
	}

	got := Text(&report, &TextOptions{Color: true, Explain: true})

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
	report := reportmodel.Report{
		SchemaVersion: "1",
		ToolName:      "distance", ToolVersion: "test",
		Module: "example.com/m",
		Packages: []reportmodel.PackageReport{
			{
				Path: "example.com/m/quiet",
				Metrics: []reportmodel.MetricResult{
					{
						Name:       string(reportmodel.MetricAbstractness),
						Scope:      reportmodel.ScopePackage,
						Value:      0,
						Applicable: true,
					},
				},
			},
			{
				Path: "example.com/m/noisy",
			},
		},
	}

	got := Text(&report, &TextOptions{Explain: true})

	if strings.Contains(got, "notes") {
		t.Fatalf("no reported metric reasons, notes should be absent:\n%s", got)
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
	report := reportmodel.Report{
		SchemaVersion: "1",
		ToolName:      "distance", ToolVersion: "test",
		Module: "example.com/m",
		Packages: []reportmodel.PackageReport{
			{
				Path: "example.com/m/a",
				Metrics: []reportmodel.MetricResult{
					{
						Name:       string(reportmodel.MetricAbstractness),
						Scope:      reportmodel.ScopePackage,
						Value:      1,
						Applicable: true,
					},
					{
						Name:       string(reportmodel.MetricDistance),
						Scope:      reportmodel.ScopePackage,
						Value:      0,
						Applicable: true,
					},
				},
			},
			{
				Path: "example.com/m/b",
				Metrics: []reportmodel.MetricResult{
					{
						Name:       string(reportmodel.MetricAbstractness),
						Scope:      reportmodel.ScopePackage,
						Value:      0,
						Applicable: true,
					},
				},
			},
		},
	}
	got := Text(&report, &TextOptions{})
	mustMatch(t, got, `(?m)^b\s+0\.00$`)
}

// White-box: the guide's direction, boundedness, and column labels mirror
// the renderers' quality map and abbreviations — the guide may never
// contradict how the report actually colors and titles a column.
func TestMetricDocsDirectionMatchesQuality(t *testing.T) {
	t.Parallel()

	for _, d := range MetricDocs() {
		if d.Scope == DocScopeStructural {
			if d.Direction != DirectionNeutral {
				t.Errorf(
					"%s: structural direction = %q, want %q",
					d.Name,
					d.Direction,
					DirectionNeutral,
				)
			}

			continue
		}

		if d.Label != abbrev(d.Name) {
			t.Errorf("%s: label = %q, want column heading %q", d.Name, d.Label, abbrev(d.Name))
		}

		direction, bounded, colored := qualityFor(d.Name)

		if !colored {
			if d.Direction != DirectionNeutral {
				t.Errorf(
					"%s: uncolored metric direction = %q, want %q",
					d.Name,
					d.Direction,
					DirectionNeutral,
				)
			}

			continue
		}

		if d.Direction != direction {
			t.Errorf("%s: direction = %q, want %q", d.Name, d.Direction, direction)
		}

		if d.Bounded != bounded {
			t.Errorf("%s: bounded = %v, want %v", d.Name, d.Bounded, bounded)
		}
	}
}

func tableReport() reportmodel.Report {
	return reportmodel.Report{
		SchemaVersion: "1",
		ToolName:      "distance", ToolVersion: "test",
		Module: "example.com/mod",
		Packages: []reportmodel.PackageReport{{
			Path: "example.com/mod",
			Metrics: []reportmodel.MetricResult{
				{
					Name:       string(reportmodel.MetricAbstractness),
					Scope:      reportmodel.ScopePackage,
					Value:      0.25,
					Applicable: true,
				},
				{
					Name:       string(reportmodel.MetricDistance),
					Scope:      reportmodel.ScopePackage,
					Value:      0.15,
					Applicable: true,
				},
			},
		}},
	}
}

func mustMatch(t *testing.T, got, pattern string) {
	t.Helper()

	if !regexp.MustCompile(pattern).MatchString(got) {
		t.Errorf("output does not match %q\ngot:\n%s", pattern, got)
	}
}

func TestTextTreeTableLayout(t *testing.T) {
	got := Text(func() *reportmodel.Report {
		r := tableReport()
		return &r
	}(), &TextOptions{})

	mustMatch(t, got, `(?m)^module example\.com/mod$`)
	mustMatch(t, got, `(?m)^PATH\s+A\s+Dist$`)
	mustMatch(t, got, `(?m)^\.\s+0\.25\s+0\.15$`)

	if strings.Contains(got, "mean") {
		t.Errorf("output still contains a separate mean row:\n%s", got)
	}

	if strings.Contains(got, "\x1b[") {
		t.Errorf("uncolored output contains ANSI escapes:\n%q", got)
	}
}

func TestTextTreeGroupsPackagesUnderSharedPath(t *testing.T) {
	report := tableReport()

	report.Packages = []reportmodel.PackageReport{
		{
			Path: "example.com/mod/internal/a",
			Metrics: []reportmodel.MetricResult{
				{
					Name:       string(reportmodel.MetricDistance),
					Scope:      reportmodel.ScopePackage,
					Value:      0.5,
					Applicable: true,
				},
			},
		},
		{
			Path: "example.com/mod/internal/b/deep",
			Metrics: []reportmodel.MetricResult{
				{
					Name:       string(reportmodel.MetricDistance),
					Scope:      reportmodel.ScopePackage,
					Value:      1,
					Applicable: true,
				},
			},
		},
	}

	got := Text(&report, &TextOptions{})

	// Shared "internal" directory heads the section and aggregates all
	// packages beneath it: A mean (1+0)/2, TCC mean (0.5+1)/2. The
	// single-child chain b → deep is compressed into one branch.
	mustMatch(t, got, `(?m)^internal\s+0\.75$`)
	mustMatch(t, got, `(?m)^├── a\s+0\.50$`)
	mustMatch(t, got, `(?m)^│$`)
	mustMatch(t, got, `(?m)^└── b/deep\s+1\.00$`)
}

func TestTextParentPackageGroupSkipsNonApplicableMetrics(t *testing.T) {
	report := tableReport()

	report.Packages = []reportmodel.PackageReport{
		{
			Path: "example.com/mod/group",
			Metrics: []reportmodel.MetricResult{
				{Name: string(reportmodel.MetricAbstractness), Scope: reportmodel.ScopePackage},
				{
					Name:       string(reportmodel.MetricInstability),
					Scope:      reportmodel.ScopePackage,
					Value:      0.2,
					Applicable: true,
				},
				{Name: string(reportmodel.MetricDistance), Scope: reportmodel.ScopePackage},
			},
		},
		{
			Path: "example.com/mod/group/child",
			Metrics: []reportmodel.MetricResult{
				{
					Name:       string(reportmodel.MetricAbstractness),
					Scope:      reportmodel.ScopePackage,
					Value:      0.6,
					Applicable: true,
				},
				{
					Name:       string(reportmodel.MetricInstability),
					Scope:      reportmodel.ScopePackage,
					Value:      0.8,
					Applicable: true,
				},
				{
					Name:       string(reportmodel.MetricDistance),
					Scope:      reportmodel.ScopePackage,
					Value:      0.4,
					Applicable: true,
				},
			},
		},
	}

	got := Text(&report, &TextOptions{})

	// The parent package is also a group. Its n/a abstractness and distance
	// are skipped, while instability averages both applicable package values.
	mustMatch(t, got, `(?m)^group\s+0\.60\s+0\.50\s+0\.40$`)
	mustMatch(t, got, `(?m)^└── child\s+0\.60\s+0\.80\s+0\.40$`)
}

func TestTextModuleRootSummarizesApplicablePackageMetrics(t *testing.T) {
	report := tableReport()

	report.Packages = []reportmodel.PackageReport{
		{
			Path: "example.com/mod",
			Metrics: []reportmodel.MetricResult{
				{Name: string(reportmodel.MetricAbstractness), Scope: reportmodel.ScopePackage},
				{
					Name:       string(reportmodel.MetricInstability),
					Scope:      reportmodel.ScopePackage,
					Value:      1,
					Applicable: true,
				},
				{Name: string(reportmodel.MetricDistance), Scope: reportmodel.ScopePackage},
			},
		},
		{
			Path: "example.com/mod/child",
			Metrics: []reportmodel.MetricResult{
				{
					Name:       string(reportmodel.MetricAbstractness),
					Scope:      reportmodel.ScopePackage,
					Value:      0.6,
					Applicable: true,
				},
				{
					Name:       string(reportmodel.MetricInstability),
					Scope:      reportmodel.ScopePackage,
					Value:      0.5,
					Applicable: true,
				},
				{
					Name:       string(reportmodel.MetricDistance),
					Scope:      reportmodel.ScopePackage,
					Value:      0.4,
					Applicable: true,
				},
			},
		},
	}

	got := Text(&report, &TextOptions{})

	// Root n/a metrics are skipped. Instability averages the applicable root
	// and child values, while abstractness and distance use the child value.
	mustMatch(t, got, `(?m)^\.\s+0\.60\s+0\.75\s+0\.40$`)
	mustMatch(t, got, `(?m)^child\s+0\.60\s+0\.50\s+0\.40$`)
}

func TestTextReasonsOnlyWithExplain(t *testing.T) {
	report := tableReport()

	if got := Text(&report, &TextOptions{}); strings.Contains(got, "fewer than two methods") {
		t.Errorf("reasons shown without Explain:\n%s", got)
	}

	got := Text(&report, &TextOptions{Explain: true})

	if strings.Contains(got, "tcc:") || strings.Contains(got, "amc:") {
		t.Errorf("hidden metrics leaked into notes:\n%s", got)
	}
}

func TestTextMeanSkipsNonApplicable(t *testing.T) {
	got := Text(func() *reportmodel.Report {
		r := tableReport()
		return &r
	}(), &TextOptions{})

	// The TCC mean averages only Cart's 0.75; Order's n/a must not drag
	// it down to 0.375.
	if strings.Contains(got, "0.38") || strings.Contains(got, "0.37") {
		t.Errorf("mean included a non-applicable value:\n%s", got)
	}
}

func TestTextColorAppliesQualityAndBold(t *testing.T) {
	got := Text(
		func() *reportmodel.Report {
			r := tableReport()
			return &r
		}(),
		&TextOptions{Color: true},
	)

	if !strings.Contains(got, ansiGreen+"0.15"+ansiReset) {
		t.Errorf("low distance not green:\n%q", got)
	}

	if strings.Contains(got, ansiGreen+"0.25") || strings.Contains(got, ansiRed+"0.25") ||
		strings.Contains(got, ansiYellow+"0.25") {

		t.Errorf("abstractness was quality-colored:\n%q", got)
	}
}

func TestTextSingleTypeLeavesUnboundedPlain(t *testing.T) {
	report := tableReport()

	got := Text(&report, &TextOptions{Color: true})

	if strings.Contains(got, ansiGreen+"2.00"+ansiReset) ||
		strings.Contains(got, ansiRed+"2.00"+ansiReset) ||
		strings.Contains(got, ansiYellow+"2.00"+ansiReset) {

		t.Errorf("lone AMC value was relatively colored:\n%q", got)
	}
}

func TestFormatCell(t *testing.T) {
	cases := map[float64]string{
		12:        "12.00",
		0:         "0.00",
		4.25:      "4.25",
		2.0 / 3.0: "0.67",
		0.5:       "0.50",
	}

	for value, want := range cases {
		if got := formatCell(value); got != want {
			t.Errorf("formatCell(%v) = %q, want %q", value, got, want)
		}
	}
}

func TestTextFallbacks(t *testing.T) {
	if got := abbrev("custom"); got != "CUSTOM" {
		t.Fatalf("unknown abbreviation: %q", got)
	}
	got := Text(&reportmodel.Report{}, &TextOptions{})
	if strings.Contains(got, "Module:") {
		t.Fatalf("empty module line: %q", got)
	}
}
