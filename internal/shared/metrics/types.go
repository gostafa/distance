// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package metrics

type (

	// ResultView is the subset of a metric result needed to compute distance.
	ResultView interface {
		ResultApplicable() bool
		ResultReason() string
		ResultValue() float64
	}

	// MetricResult is one computed metric value (or a not-applicable reason).
	MetricResult struct {
		Name       string
		Scope      string
		Reason     string
		Definition string
		Value      float64
		Applicable bool
	}
)
