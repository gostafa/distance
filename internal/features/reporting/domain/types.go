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

	// TextOptions controls terminal presentation.
	TextOptions = struct {
		Color   bool
		Explain bool
	}

	// WriteOptions selects a representation and text presentation.
	WriteOptions = struct {
		Format string
		Text   TextOptions
	}
)
