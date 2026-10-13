// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	"testing"

	"github.com/gostafa/distance/internal/features/policy/domain"
)

func TestValidatedPolicyWorkflow(t *testing.T) {
	report := domain.Report{
		Packages: []domain.PackageReport{
			{
				Path:    "m/p",
				Metrics: []domain.MetricResult{{Name: "distance", Value: 0.9, Applicable: true}},
			},
		},
	}
	if got, err := Check(&report, nil); err != nil || len(got) != 1 || got[0].Threshold != 0.5 {
		t.Fatalf("defaults: %v %v", got, err)
	}
	if got, err := Check(
		&report,
		[]domain.Rule{{Pattern: "m/p", Max: 1}},
	); err != nil ||
		len(got) != 0 {
		t.Fatalf("explicit rule: %v %v", got, err)
	}
	if got, err := Check(
		&report,
		[]domain.Rule{{Pattern: "", Max: 0.5}},
	); err == nil ||
		got != nil {
		t.Fatalf("invalid rule: %v %v", got, err)
	}
}
