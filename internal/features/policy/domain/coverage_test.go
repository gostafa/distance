// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	"math"
	"testing"

	"github.com/gostafa/distance/distance"
	"github.com/gostafa/distance/internal/shared/metrics"
)

func TestEvaluateSkipsInapplicablePackageMetric(t *testing.T) {
	report := distance.Report{
		Module: "example.com",
		Packages: []distance.PackageReport{{
			Path: "example.com/p",
			Metrics: []metrics.MetricResult{{
				Name:       metrics.MetricDistance,
				Scope:      metrics.ScopePackage,
				Applicable: false,
			}},
		}},
	}
	policy := Policy{Packages: []PackageRule{{
		Pattern:     "./...",
		MaxDistance: 0,
	}}}

	if got := Evaluate(&report, policy); len(got) != 0 {
		t.Fatalf("Evaluate() = %#v, want no violations", got)
	}
}

func TestValidateRejectsBadRules(t *testing.T) {
	for _, tc := range []struct {
		name   string
		policy Policy
	}{
		{"empty pattern", Policy{Packages: []PackageRule{{Pattern: "", MaxDistance: 0.5}}}},
		{"nan max", Policy{Packages: []PackageRule{{Pattern: "./...", MaxDistance: math.NaN()}}}},
		{"inf max", Policy{Packages: []PackageRule{{Pattern: "./...", MaxDistance: math.Inf(1)}}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := Validate(tc.policy)
			if err == nil {
				t.Fatal("Validate succeeded, want error")
			}
		})
	}
}

func TestValidateAcceptsDefault(t *testing.T) {
	policy, err := PolicyFromPatterns(nil, DefaultMaxDistance)
	if err != nil {
		t.Fatal(err)
	}

	if err := Validate(policy); err != nil {
		t.Fatal(err)
	}
}
