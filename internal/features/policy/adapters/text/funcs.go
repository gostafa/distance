// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package text

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/gostafa/distance/internal/features/policy/domain"
)

// FormatViolations renders violations as a human-readable summary. The empty
// slice yields the empty string, so callers can print unconditionally.
func FormatViolations(violations []domain.Violation) string {
	if len(violations) == zero {
		return emptyString
	}

	lines := make([]string, zero, one+len(violations))

	lines = append(lines, violationHeader(len(violations)))

	for index := range violations {
		lines = append(lines, violationLine(&violations[index]))
	}

	return strings.Join(lines, emptyString)
}

func formatNumber(value float64) string {
	if value == math.Trunc(value) && !math.IsInf(value, zero) {
		return strconv.FormatFloat(value, 'f', -one, floatBits)
	}

	return strconv.FormatFloat(value, 'f', two, floatBits)
}

func violationHeader(count int) string {
	noun := "violations"

	if count == one {
		noun = "violation"
	}

	return fmt.Sprintf("policy: %d %s\n", count, noun)
}

func violationLine(violation *domain.Violation) string {
	return fmt.Sprintf(
		"  %s (package): %s %s exceeds max %s (rule %s)\n",
		violation.Package,
		domain.MetricDistance,
		formatNumber(violation.Value),
		formatNumber(violation.Threshold),
		violation.Rule,
	)
}
