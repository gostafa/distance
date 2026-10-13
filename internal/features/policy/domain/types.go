// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

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

	// Violation records one package whose distance exceeded a matching rule.
	Violation = struct {
		Package   string
		Rule      string
		Value     float64
		Threshold float64
	}

	// Rule is a package-path glob paired with an exclusive maximum
	// When more than one rule matches, the most specific pattern wins; exact
	// ties use the later rule.
	Rule = struct {
		// Pattern is a glob against the full import path; * matches one segment,
		// ** matches zero or more segments, using / as the separator.
		Pattern string
		// Max is the exclusive upper bound in [0, 1]: fail when distance > Max.
		Max float64
	}

	matchPos = struct {
		pi, si int
	}

	packageGate = struct {
		pkg       string
		pattern   string
		threshold float64
		value     float64
		ok        bool
	}

	gateInput = struct {
		pkg       *PackageReport
		res       *MetricResult
		pattern   string
		threshold float64
	}
)
