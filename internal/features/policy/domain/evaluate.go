package domain

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/gostafa/distance/distance"
	"github.com/gostafa/distance/internal/shared/metrics"
)

// Comparator identifies which bound a violation crossed.
type Comparator string

const (
	// ComparatorMax marks a value that exceeded an upper bound.
	ComparatorMax Comparator = "max"
	// ComparatorMin marks a value that fell below a lower bound.
	ComparatorMin Comparator = "min"
)

// Violation is one broken condition: a package's actual distance against
// the maximum it crossed.
type Violation struct {
	Package    string     // package import path
	Key        string     // condition key: always the distance metric
	Value      float64    // the package's actual distance
	Comparator Comparator // which bound was crossed
	Threshold  float64    // the bound's value
}

// Evaluate checks a report against a policy and returns the violations.
// The first matching rule in list order wins. Packages that match no rule
// are not gated. A metric condition is skipped when distance is not
// applicable, so n/a cells never produce false positives.
func Evaluate(report distance.Report, policy Policy) []Violation {
	var violations []Violation

	for i := range report.Packages {
		pkg := &report.Packages[i]
		rule, ok := matchingRule(policy.Packages, pkg.Path, report.Module)
		if !ok {
			continue
		}

		for _, result := range pkg.Metrics {
			if result.Name != metrics.MetricDistance || !result.Applicable {
				continue
			}

			if result.Value-rule.MaxDistance > comparisonTolerance(result.Value, rule.MaxDistance) {
				violations = append(violations, Violation{
					Package:    pkg.Path,
					Key:        result.Name,
					Value:      result.Value,
					Comparator: ComparatorMax,
					Threshold:  rule.MaxDistance,
				})
			}
		}
	}

	return violations
}

// comparisonTolerance absorbs floating-point representation noise at a policy
// boundary without hiding a meaningful threshold crossing.
func comparisonTolerance(value, threshold float64) float64 {
	return 1e-12 * max(1, math.Abs(value), math.Abs(threshold))
}

// FormatViolations renders violations as a human-readable summary. The empty
// slice yields the empty string, so callers can print unconditionally.
func FormatViolations(violations []Violation) string {
	if len(violations) == 0 {
		return ""
	}

	var b strings.Builder

	noun := "violations"
	if len(violations) == 1 {
		noun = "violation"
	}

	fmt.Fprintf(&b, "policy: %d %s\n", len(violations), noun)

	for _, v := range violations {
		fmt.Fprintf(&b, "  %s (package): %s %s exceeds max %s\n",
			v.Package, v.Key, formatNumber(v.Value), formatNumber(v.Threshold))
	}

	return b.String()
}

// formatNumber prints integers without a fraction and other values with two
// decimals, matching the report's cell formatting.
func formatNumber(value float64) string {
	if value == math.Trunc(value) && !math.IsInf(value, 0) {
		return strconv.FormatFloat(value, 'f', -1, 64)
	}

	return strconv.FormatFloat(value, 'f', 2, 64)
}
