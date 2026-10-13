// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain_test

import (
	"fmt"
	"testing"

	"github.com/gostafa/distance/internal/features/policy/domain"
)

const (
	benchPackages = 500
	benchSpread   = 100.0
)

func benchReport() *domain.Report {
	report := &domain.Report{
		Module:   "example.com/m",
		Packages: make([]domain.PackageReport, benchPackages),
	}

	layers := []string{
		"internal/features/%d/domain",
		"internal/features/%d/adapters",
		"cmd/tool%d",
		"pkg/lib%d",
	}

	for index := range report.Packages {
		layer := fmt.Sprintf(layers[index%len(layers)], index)

		report.Packages[index] = domain.PackageReport{
			Path: "example.com/m/" + layer,
			Metrics: []domain.MetricResult{
				{Name: "abstractness", Scope: "package", Value: 0.2, Applicable: true},
				{Name: "instability", Scope: "package", Value: 0.6, Applicable: true},
				{
					Name:       "distance",
					Scope:      "package",
					Value:      float64(index%100) / benchSpread,
					Applicable: true,
				},
			},
		}
	}

	return report
}

func benchRules() []domain.Rule {
	return []domain.Rule{
		{Pattern: "**", Max: 0.5},
		{Pattern: "**/internal/**", Max: 0.4},
		{Pattern: "example.com/m/internal/features/*/domain", Max: 0.2},
		{Pattern: "example.com/m/cmd/**", Max: 0.9},
	}
}

func BenchmarkEvaluate(b *testing.B) {
	report := benchReport()
	rules := benchRules()

	b.ReportAllocs()

	for b.Loop() {
		_ = domain.Evaluate(report, rules)
	}
}

func BenchmarkMatchPackage(b *testing.B) {
	cases := []struct{ name, pattern, path string }{
		{"exact", "example.com/m/internal/domain", "example.com/m/internal/domain"},
		{
			"single_star",
			"example.com/m/internal/*/domain",
			"example.com/m/internal/features/domain",
		},
		{"double_star", "**/internal/**/domain", "example.com/m/a/b/internal/features/x/y/domain"},
		{"no_match", "**/internal/**/domain", "example.com/m/a/b/c/d/e/f/g/h/adapters"},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			for b.Loop() {
				_ = domain.MatchPackage(tc.pattern, tc.path)
			}
		})
	}
}
