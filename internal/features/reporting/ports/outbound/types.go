// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package outbound

import (
	"io"
)

type (
	// Output renders and closes destinations, preserving stable feature errors.
	Output interface {
		// Write returns ErrOutput for encoding, writing, or closing failures and always closes the destination.
		Write(report *Report, destination io.WriteCloser, opts *struct {
			Format string
			Text   TextOptions
		}) error
		// WriteDocs returns ErrOutput for guide output failures and always closes the destination.
		WriteDocs(destination io.WriteCloser, version string) error
	}

	// TextOptions controls terminal presentation.
	TextOptions = struct {
		Color   bool
		Explain bool
	}

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
)
