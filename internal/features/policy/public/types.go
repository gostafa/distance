// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package public

type (
	// MetricResult is a computed package metric value.
	MetricResult = struct {
		Name       string
		Scope      string
		Reason     string
		Definition string
		Value      float64
		Applicable bool
	}

	// PackageReport holds metrics and coupling for one analyzed package.
	PackageReport = struct {
		Path     string
		Metrics  []MetricResult
		Afferent int
		Efferent int
	}

	// Report is the complete analysis result for one configuration.
	Report = struct {
		SchemaVersion string
		ToolName      string
		ToolVersion   string
		Module        string
		Packages      []PackageReport
	}

	// Rule is a package glob and maximum allowed distance.
	Rule = struct {
		Pattern string
		Max     float64
	}

	// Violation records a package exceeding its selected policy rule.
	Violation = struct {
		Package   string
		Rule      string
		Value     float64
		Threshold float64
	}
)
