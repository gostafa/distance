// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain_test

import (
	"math"
	"strings"
	"testing"

	"github.com/gostafa/distance/distance"
	"github.com/gostafa/distance/internal/features/policy/domain"
	"github.com/gostafa/distance/internal/shared/metrics"
)

func sampleReport() distance.Report {
	return distance.Report{
		Module: "example.com/m",
		Packages: []distance.PackageReport{
			{
				Path:     "example.com/m/foo",
				Afferent: 2,
				Efferent: 20,
				Metrics: []metrics.MetricResult{{
					Name:       metrics.MetricDistance,
					Scope:      metrics.ScopePackage,
					Value:      0.9,
					Applicable: true,
				}},
			},
			{
				Path: "example.com/m/internal/domain",
				Metrics: []metrics.MetricResult{{
					Name:       metrics.MetricDistance,
					Scope:      metrics.ScopePackage,
					Value:      0.3,
					Applicable: true,
				}},
			},
			{
				Path: "example.com/other",
				Metrics: []metrics.MetricResult{{
					Name:       metrics.MetricDistance,
					Scope:      metrics.ScopePackage,
					Value:      1,
					Applicable: true,
				}},
			},
		},
	}
}

func TestEvaluateFirstMatchWins(t *testing.T) {
	t.Parallel()

	policy := domain.Policy{Packages: []domain.PackageRule{
		{Pattern: "./internal/domain/...", MaxDistance: 0.2},
		{Pattern: "./...", MaxDistance: 0.5},
	}}

	got := domain.Evaluate(func() *distance.Report {
		r := sampleReport()
		return &r
	}(), policy)

	if len(got) != 2 {
		t.Fatalf("violations = %#v, want 2", got)
	}

	byPkg := map[string]domain.Violation{}

	for _, v := range got {
		byPkg[v.Package] = v
	}

	if v, ok := byPkg["example.com/m/internal/domain"]; !ok || v.Threshold != 0.2 {
		t.Fatalf("internal/domain = %#v, want first-match max 0.2", v)
	}

	if v, ok := byPkg["example.com/m/foo"]; !ok || v.Threshold != 0.5 {
		t.Fatalf("foo = %#v, want ./... max 0.5", v)
	}

	if _, ok := byPkg["example.com/other"]; ok {
		t.Fatal("unrelated package should not be gated")
	}
}

func TestEvaluateOverlapPrefersListOrder(t *testing.T) {
	t.Parallel()

	// Broad rule first: the specific prefix never wins.
	policy := domain.Policy{Packages: []domain.PackageRule{
		{Pattern: "./...", MaxDistance: 0.8},
		{Pattern: "./foo", MaxDistance: 0.1},
	}}

	got := domain.Evaluate(func() *distance.Report {
		r := sampleReport()
		return &r
	}(), policy)

	if len(got) != 1 || got[0].Package != "example.com/m/foo" || got[0].Threshold != 0.8 {
		t.Fatalf("violations = %#v, want foo against the first ./... rule", got)
	}
}

func TestEvaluateExactImportPath(t *testing.T) {
	t.Parallel()

	policy := domain.Policy{Packages: []domain.PackageRule{
		{Pattern: "example.com/m/foo", MaxDistance: 0.5},
	}}

	got := domain.Evaluate(func() *distance.Report {
		r := sampleReport()
		return &r
	}(), policy)

	if len(got) != 1 || got[0].Package != "example.com/m/foo" {
		t.Fatalf("violations = %#v, want foo only", got)
	}
}

func TestMatchPatternResolvesRelativeAndPrefix(t *testing.T) {
	t.Parallel()

	const module = "example.com/m"

	cases := []struct {
		pattern, path string
		want          bool
	}{
		{"./", "example.com/m", true},
		{"./...", "example.com/m", true},
		{"./...", "example.com/m/foo", true},
		{"./foo", "example.com/m/foo", true},
		{"./foo", "example.com/m/foo/bar", false},
		{"./foo/...", "example.com/m/foo", true},
		{"./foo/...", "example.com/m/foo/bar", true},
		{"./foo/...", "example.com/m/other", false},
		{"example.com/m/foo/...", "example.com/m/foo/bar", true},
		{"example.com/other", "example.com/other", true},
		{"./...", "example.com/other", false},
	}

	for _, tc := range cases {
		if got := domain.MatchPattern(tc.pattern, tc.path, module); got != tc.want {
			t.Errorf("MatchPattern(%q, %q) = %v, want %v", tc.pattern, tc.path, got, tc.want)
		}
	}

	if !domain.MatchPattern("./...", "any/path", "") {
		t.Error("./... with empty module should match all loaded packages")
	}
}

func TestEvaluateBoundaryTolerance(t *testing.T) {
	t.Parallel()

	report := distance.Report{
		Module: "example.com/m",
		Packages: []distance.PackageReport{{
			Path: "example.com/m/p",
			Metrics: []metrics.MetricResult{{
				Name:       metrics.MetricDistance,
				Scope:      metrics.ScopePackage,
				Value:      0.5,
				Applicable: true,
			}},
		}},
	}
	policy := domain.Policy{Packages: []domain.PackageRule{{
		Pattern:     "./...",
		MaxDistance: 0.5,
	}}}

	if got := domain.Evaluate(&report, policy); len(got) != 0 {
		t.Fatalf("exact boundary should pass: %#v", got)
	}

	report.Packages[0].Metrics[0].Value = 0.5 + 1e-9

	if got := domain.Evaluate(&report, policy); len(got) != 1 {
		t.Fatalf("over-threshold should fail: %#v", got)
	}
}

func TestEvaluateIgnoresAbstractnessAndInstability(t *testing.T) {
	t.Parallel()

	report := distance.Report{
		Module: "example.com/m",
		Packages: []distance.PackageReport{{
			Path: "example.com/m/p",
			Metrics: []metrics.MetricResult{
				{
					Name:       metrics.MetricAbstractness,
					Scope:      metrics.ScopePackage,
					Value:      1,
					Applicable: true,
				},
				{
					Name:       metrics.MetricInstability,
					Scope:      metrics.ScopePackage,
					Value:      1,
					Applicable: true,
				},
				{
					Name:       metrics.MetricDistance,
					Scope:      metrics.ScopePackage,
					Value:      0.2,
					Applicable: true,
				},
			},
		}},
	}
	policy := domain.Policy{Packages: []domain.PackageRule{{
		Pattern:     "./...",
		MaxDistance: 0.5,
	}}}

	if got := domain.Evaluate(&report, policy); len(got) != 0 {
		t.Fatalf("A/I must not be gated: %#v", got)
	}
}

func TestFormatViolations(t *testing.T) {
	t.Parallel()

	if got := domain.FormatViolations(nil); got != "" {
		t.Fatalf("empty = %q", got)
	}

	got := domain.FormatViolations([]domain.Violation{{
		Package:    "example.com/p",
		Key:        metrics.MetricDistance,
		Value:      0.9,
		Comparator: domain.ComparatorMax,
		Threshold:  0.5,
	}})

	if !strings.Contains(got, "1 violation") || !strings.Contains(got, "exceeds max 0.50") {
		t.Fatalf("FormatViolations = %q", got)
	}

	got = domain.FormatViolations([]domain.Violation{
		{Package: "a", Key: "distance", Value: 1, Comparator: domain.ComparatorMax, Threshold: 0},
		{
			Package:    "b",
			Key:        "distance",
			Value:      math.Pi,
			Comparator: domain.ComparatorMax,
			Threshold:  0.5,
		},
	})

	if !strings.Contains(got, "2 violations") {
		t.Fatalf("plural = %q", got)
	}
}
