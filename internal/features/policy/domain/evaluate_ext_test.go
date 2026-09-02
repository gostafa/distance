// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain_test

import (
	"math"
	"strings"
	"testing"

	"github.com/gostafa/distance/distance"
	"github.com/gostafa/distance/internal/features/policy/domain"
)

func sampleReport() distance.Report {
	return distance.Report{
		Module: "example.com/m",
		Packages: []distance.PackageReport{
			{
				Path:     "example.com/m/foo",
				Afferent: 2,
				Efferent: 20,
				Metrics: []distance.MetricResult{{
					Name:       string(distance.MetricDistance),
					Scope:      distance.ScopePackage,
					Value:      0.9,
					Applicable: true,
				}},
			},
			{
				Path: "example.com/m/internal/domain",
				Metrics: []distance.MetricResult{{
					Name:       string(distance.MetricDistance),
					Scope:      distance.ScopePackage,
					Value:      0.3,
					Applicable: true,
				}},
			},
			{
				Path: "example.com/other",
				Metrics: []distance.MetricResult{{
					Name:       string(distance.MetricDistance),
					Scope:      distance.ScopePackage,
					Value:      1,
					Applicable: true,
				}},
			},
		},
	}
}

func TestEvaluateMostSpecificRuleWins(t *testing.T) {
	t.Parallel()

	rules := []domain.Rule{
		{Pattern: "**", Max: 0.5},
		{Pattern: "**/internal/**", Max: 0.2},
	}

	got := domain.Evaluate(func() *distance.Report {
		r := sampleReport()

		return &r
	}(), rules)

	if len(got) != 3 {
		t.Fatalf("violations = %#v, want 3", got)
	}

	byPkg := map[string]domain.Violation{}

	for _, item := range got {
		byPkg[item.Package] = item
	}

	if item, ok := byPkg["example.com/m/internal/domain"]; !ok || item.Threshold != 0.2 {
		t.Fatalf("internal/domain = %#v, want specific max 0.2", item)
	}

	if item, ok := byPkg["example.com/m/foo"]; !ok || item.Threshold != 0.5 {
		t.Fatalf("foo = %#v, want ** max 0.5", item)
	}

	if item, ok := byPkg["example.com/other"]; !ok || item.Threshold != 0.5 {
		t.Fatalf("other = %#v, want ** max 0.5", item)
	}
}

func TestEvaluateExactImportPath(t *testing.T) {
	t.Parallel()

	rules := []domain.Rule{{Pattern: "example.com/m/foo", Max: 0.5}}

	got := domain.Evaluate(func() *distance.Report {
		r := sampleReport()

		return &r
	}(), rules)

	if len(got) != 1 || got[0].Package != "example.com/m/foo" {
		t.Fatalf("violations = %#v, want foo only", got)
	}
}

func TestMatchPackage(t *testing.T) {
	t.Parallel()

	cases := []struct {
		pattern string
		path    string
		want    bool
	}{
		{"**", "example.com/m/foo", true},
		{"**", "a", true},
		{"example.com/m/foo", "example.com/m/foo", true},
		{"example.com/m/foo", "example.com/m/bar", false},
		{"**/internal/**", "example.com/m/internal/store", true},
		{"**/internal/**", "example.com/m/store", false},
		{"example.com/*/foo", "example.com/m/foo", true},
		{"example.com/*/foo", "example.com/m/n/foo", false},
		{"**/foo", "example.com/foo", true},
		{"**/foo", "example.com/a/b/foo", true},
	}

	for _, tc := range cases {
		if got := domain.MatchPackage(tc.pattern, tc.path); got != tc.want {
			t.Errorf("MatchPackage(%q, %q) = %v, want %v", tc.pattern, tc.path, got, tc.want)
		}
	}
}

func TestEvaluateBoundaryTolerance(t *testing.T) {
	t.Parallel()

	report := distance.Report{
		Module: "example.com/m",
		Packages: []distance.PackageReport{{
			Path: "example.com/m/p",
			Metrics: []distance.MetricResult{{
				Name:       string(distance.MetricDistance),
				Scope:      distance.ScopePackage,
				Value:      0.5,
				Applicable: true,
			}},
		}},
	}
	rules := []domain.Rule{{Pattern: "**", Max: 0.5}}

	if got := domain.Evaluate(&report, rules); len(got) != 0 {
		t.Fatalf("exact boundary should pass: %#v", got)
	}

	report.Packages[0].Metrics[0].Value = 0.5 + 1e-9

	if got := domain.Evaluate(&report, rules); len(got) != 1 {
		t.Fatalf("over-threshold should fail: %#v", got)
	}
}

func TestEvaluateIgnoresAbstractnessAndInstability(t *testing.T) {
	t.Parallel()

	report := distance.Report{
		Module: "example.com/m",
		Packages: []distance.PackageReport{{
			Path: "example.com/m/p",
			Metrics: []distance.MetricResult{
				{
					Name:       string(distance.MetricAbstractness),
					Scope:      distance.ScopePackage,
					Value:      1,
					Applicable: true,
				},
				{
					Name:       string(distance.MetricInstability),
					Scope:      distance.ScopePackage,
					Value:      1,
					Applicable: true,
				},
				{
					Name:       string(distance.MetricDistance),
					Scope:      distance.ScopePackage,
					Value:      0.2,
					Applicable: true,
				},
			},
		}},
	}
	rules := []domain.Rule{{Pattern: "**", Max: 0.5}}

	if got := domain.Evaluate(&report, rules); len(got) != 0 {
		t.Fatalf("A/I must not be gated: %#v", got)
	}
}

func TestFormatViolations(t *testing.T) {
	t.Parallel()

	if got := domain.FormatViolations(nil); got != "" {
		t.Fatalf("empty = %q", got)
	}

	got := domain.FormatViolations([]domain.Violation{{
		Package:   "example.com/p",
		Value:     0.9,
		Threshold: 0.5,
		Rule:      "**",
	}})

	want := "example.com/p (package): distance 0.90 exceeds max 0.50 (rule **)"
	if !strings.Contains(got, "1 violation") || !strings.Contains(got, want) {
		t.Fatalf("FormatViolations = %q", got)
	}

	got = domain.FormatViolations([]domain.Violation{
		{Package: "a", Value: 1, Threshold: 0, Rule: "**"},
		{Package: "b", Value: math.Pi, Threshold: 0.5, Rule: "**"},
	})

	if !strings.Contains(got, "2 violations") {
		t.Fatalf("plural = %q", got)
	}
}
