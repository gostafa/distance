// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package public_test

import (
	"testing"

	policy "github.com/gostafa/distance/internal/features/policy/public"
)

func TestPublishedPolicyContract(t *testing.T) {
	report := &policy.Report{
		Packages: []policy.PackageReport{
			{
				Path: "example.com/a",
				Metrics: []policy.MetricResult{
					{Name: "distance", Scope: "package", Applicable: true, Value: 0.6},
				},
			},
		},
	}
	if _, err := policy.Check(report, []policy.Rule{{Pattern: "**", Max: -1}}); err == nil {
		t.Fatal("accepted invalid rule")
	}
	violations := policy.Evaluate(report, nil)
	if len(violations) != 1 {
		t.Fatalf("default violations: %v", violations)
	}
	if !policy.MatchPackage("example.com/**", "example.com/a") {
		t.Fatal("published glob match")
	}
	if policy.MatchPackage("", "") {
		t.Fatal("empty glob should not match")
	}
	if policy.FormatViolations(violations) == "" {
		t.Fatal("missing summary")
	}
}
