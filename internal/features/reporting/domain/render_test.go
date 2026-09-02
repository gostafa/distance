// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	"regexp"
	"strings"
	"testing"

	"github.com/gostafa/distance/distance"
)

func tableReport() distance.Report {
	return distance.Report{
		SchemaVersion: "1",
		ToolName: "distance", ToolVersion: "test",
		Module:        "example.com/mod",
		Packages: []distance.PackageReport{{
			Path: "example.com/mod",
			Metrics: []distance.MetricResult{
				{
					Name:       string(distance.MetricAbstractness),
					Scope:      distance.ScopePackage,
					Value:      0.25,
					Applicable: true,
				},
				{
					Name:       string(distance.MetricDistance),
					Scope:      distance.ScopePackage,
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
	got := Text(func() *distance.Report {
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

	report.Packages = []distance.PackageReport{
		{
			Path: "example.com/mod/internal/a",
			Metrics: []distance.MetricResult{
				{
					Name:       string(distance.MetricDistance),
					Scope:      distance.ScopePackage,
					Value:      0.5,
					Applicable: true,
				},
			},
		},
		{
			Path: "example.com/mod/internal/b/deep",
			Metrics: []distance.MetricResult{
				{
					Name:       string(distance.MetricDistance),
					Scope:      distance.ScopePackage,
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

	report.Packages = []distance.PackageReport{
		{
			Path: "example.com/mod/group",
			Metrics: []distance.MetricResult{
				{Name: string(distance.MetricAbstractness), Scope: distance.ScopePackage},
				{
					Name:       string(distance.MetricInstability),
					Scope:      distance.ScopePackage,
					Value:      0.2,
					Applicable: true,
				},
				{Name: string(distance.MetricDistance), Scope: distance.ScopePackage},
			},
		},
		{
			Path: "example.com/mod/group/child",
			Metrics: []distance.MetricResult{
				{
					Name:       string(distance.MetricAbstractness),
					Scope:      distance.ScopePackage,
					Value:      0.6,
					Applicable: true,
				},
				{
					Name:       string(distance.MetricInstability),
					Scope:      distance.ScopePackage,
					Value:      0.8,
					Applicable: true,
				},
				{
					Name:       string(distance.MetricDistance),
					Scope:      distance.ScopePackage,
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

	report.Packages = []distance.PackageReport{
		{
			Path: "example.com/mod",
			Metrics: []distance.MetricResult{
				{Name: string(distance.MetricAbstractness), Scope: distance.ScopePackage},
				{
					Name:       string(distance.MetricInstability),
					Scope:      distance.ScopePackage,
					Value:      1,
					Applicable: true,
				},
				{Name: string(distance.MetricDistance), Scope: distance.ScopePackage},
			},
		},
		{
			Path: "example.com/mod/child",
			Metrics: []distance.MetricResult{
				{
					Name:       string(distance.MetricAbstractness),
					Scope:      distance.ScopePackage,
					Value:      0.6,
					Applicable: true,
				},
				{
					Name:       string(distance.MetricInstability),
					Scope:      distance.ScopePackage,
					Value:      0.5,
					Applicable: true,
				},
				{
					Name:       string(distance.MetricDistance),
					Scope:      distance.ScopePackage,
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
	got := Text(func() *distance.Report {
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
		func() *distance.Report {
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
