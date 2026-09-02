// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	"math"
	"testing"

	"github.com/gostafa/distance/distance"
)

func TestEvaluateSkipsInapplicablePackageMetric(t *testing.T) {
	report := distance.Report{
		Module: "example.com",
		Packages: []distance.PackageReport{{
			Path: "example.com/p",
			Metrics: []distance.MetricResult{{
				Name:       string(distance.MetricDistance),
				Scope:      distance.ScopePackage,
				Applicable: false,
			}},
		}},
	}

	if got := Evaluate(&report, DefaultRules()); len(got) != 0 {
		t.Fatalf("Evaluate() = %#v, want no violations", got)
	}
}

func TestEvaluateEmptyRulesUsesDefaults(t *testing.T) {
	report := distance.Report{
		Packages: []distance.PackageReport{{
			Path: "example.com/p",
			Metrics: []distance.MetricResult{{
				Name:       string(distance.MetricDistance),
				Scope:      distance.ScopePackage,
				Value:      0.9,
				Applicable: true,
			}},
		}},
	}

	got := Evaluate(&report, nil)
	if len(got) != 1 || got[0].Rule != "**" || got[0].Threshold != DefaultMaxDistance {
		t.Fatalf("empty rules = %#v, want default ** / 0.5", got)
	}
}

func TestMoreSpecific(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		path     string
		rules    []Rule
		wantRule string
		wantMax  float64
	}{
		{
			name:     "literal segments beat wildcard baseline",
			path:     "example.com/m/internal/store",
			rules:    []Rule{{Pattern: "**", Max: 0.6}, {Pattern: "**/internal/**", Max: 0.5}},
			wantRule: "**/internal/**", wantMax: 0.5,
		},
		{
			name: "fewer wildcards break equal literal count",
			path: "example.com/m/store",
			rules: []Rule{
				{Pattern: "example.com/*/store", Max: 0.6},
				{Pattern: "example.com/m/store", Max: 0.5},
			},
			wantRule: "example.com/m/store",
			wantMax:  0.5,
		},
		{
			name:     "fewer wildcards beat longer patterns",
			path:     "example.com/m/internal/store",
			rules:    []Rule{{Pattern: "**/store", Max: 0.6}, {Pattern: "**/**/store", Max: 0.5}},
			wantRule: "**/store", wantMax: 0.6,
		},
		{
			name: "later rule wins exact specificity tie",
			path: "example.com/m/store",
			rules: []Rule{
				{Pattern: "example.com/*/store", Max: 0.6},
				{Pattern: "example.com/*/store", Max: 0.5},
			},
			wantRule: "example.com/*/store",
			wantMax:  0.5,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotMax, gotRule := matchingRule(tc.path, tc.rules)
			if gotRule != tc.wantRule || gotMax != tc.wantMax {
				t.Fatalf(
					"matchingRule() = (%v, %q), want (%v, %q)",
					gotMax,
					gotRule,
					tc.wantMax,
					tc.wantRule,
				)
			}
		})
	}
}

func TestValidateRejectsBadRules(t *testing.T) {
	for _, tc := range []struct {
		name  string
		rules []Rule
	}{
		{"empty pattern", []Rule{{Pattern: "", Max: 0.5}}},
		{"nan max", []Rule{{Pattern: "**", Max: math.NaN()}}},
		{"inf max", []Rule{{Pattern: "**", Max: math.Inf(1)}}},
		{"max below zero", []Rule{{Pattern: "**", Max: -0.1}}},
		{"max above one", []Rule{{Pattern: "**", Max: 1.1}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := Validate(tc.rules)
			if err == nil {
				t.Fatal("Validate succeeded, want error")
			}
		})
	}
}

func TestValidateAcceptsDefault(t *testing.T) {
	if err := Validate(DefaultRules()); err != nil {
		t.Fatal(err)
	}
}

func TestDefaultRules(t *testing.T) {
	t.Parallel()

	rules := DefaultRules()
	if len(rules) != 1 || rules[0].Pattern != "**" || rules[0].Max != DefaultMaxDistance {
		t.Fatalf("DefaultRules() = %+v", rules)
	}
}
