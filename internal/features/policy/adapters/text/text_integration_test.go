// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package text_test

import (
	"math"
	"strings"
	"testing"

	domain "github.com/gostafa/distance/internal/features/policy/adapters/text"
	model "github.com/gostafa/distance/internal/features/policy/domain"
)

func TestFormatViolations(t *testing.T) {
	t.Parallel()

	if got := domain.FormatViolations(nil); got != "" {
		t.Fatalf("empty = %q", got)
	}

	got := domain.FormatViolations([]model.Violation{{
		Package:   "example.com/p",
		Value:     0.9,
		Threshold: 0.5,
		Rule:      "**",
	}})

	want := "example.com/p (package): distance 0.90 exceeds max 0.50 (rule **)"
	if !strings.Contains(got, "1 violation") || !strings.Contains(got, want) {
		t.Fatalf("FormatViolations = %q", got)
	}

	got = domain.FormatViolations([]model.Violation{
		{Package: "a", Value: 1, Threshold: 0, Rule: "**"},
		{Package: "b", Value: math.Pi, Threshold: 0.5, Rule: "**"},
	})

	if !strings.Contains(got, "2 violations") {
		t.Fatalf("plural = %q", got)
	}
}
