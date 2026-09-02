// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	"github.com/gostafa/distance/distance"
	"github.com/gostafa/distance/internal/features/reporting/domain"
	"github.com/gostafa/distance/internal/shared/metrics"
)

type (

	// docsPayload wraps the guide entries with the tool identity for the page
	// header.
	docsPayload struct {
		// Tool identifies the producing tool.
		Tool jsonTool `json:"tool"`
		// Docs are the guide entries in render order.
		Docs []jsonMetricDoc `json:"docs"`
	}

	// jsonMetricDoc mirrors domain.MetricDoc for the embedded payloads.
	jsonMetricDoc struct {
		FormulaLaTeX   string `json:"formula_latex,omitempty"`
		NotApplicable  string `json:"not_applicable,omitempty"`
		FullName       string `json:"full_name"`
		Scope          string `json:"scope"`
		Definition     string `json:"definition,omitempty"`
		FormulaMathML  string `json:"formula_mathml,omitempty"`
		How            string `json:"how"`
		Example        string `json:"example,omitempty"`
		Label          string `json:"label"`
		Interpretation string `json:"interpretation"`
		Name           string `json:"name"`
		Direction      string `json:"direction"`
		Summary        string `json:"summary"`
		Bounded        bool   `json:"bounded"`
	}

	// jsonReport mirrors the versioned report schema (§ output). Metric maps
	// are orderedMetrics so keys always appear in the fixed metric order.
	jsonReport struct {
		// SchemaVersion is the report schema version.
		SchemaVersion string `json:"schema_version"`
		// Tool identifies the producing tool.
		Tool jsonTool `json:"tool"`
		// Packages are the analyzed packages in report order.
		Packages []jsonPackage `json:"packages"`
	}

	jsonTool struct {
		// Name is the tool's canonical name.
		Name string `json:"name"`
		// Version is the tool's build version.
		Version string `json:"version"`
	}

	jsonPackage struct {
		Path     string         `json:"path"`
		Metrics  orderedMetrics `json:"metrics"`
		Afferent int            `json:"afferent"`
		Efferent int            `json:"efferent"`
	}

	// jsonMetric serializes one MetricResult. A non-applicable metric carries
	// its reason and no value — never a fake zero.
	jsonMetric struct {
		Value      *float64 `json:"value,omitempty"`
		Scope      string   `json:"scope"`
		Reason     string   `json:"reason,omitempty"`
		Definition string   `json:"definition"`
		Applicable bool     `json:"applicable"`
	}

	// orderedMetrics marshals as a JSON object keyed by metric name, preserving
	// slice order (the fixed metric order).
	orderedMetrics []metrics.MetricResult

	// webPayload wraps the versioned JSON report with the module path for the
	// page header. The v1 JSON schema itself stays untouched.
	webPayload struct {
		// Module is the analyzed main module's path, when known.
		Module string `json:"module"`
		// Report is the same document the JSON format emits.
		Report jsonReport `json:"report"`
	}

	// WriteOptions configures Write: the output format and text-only options.
	WriteOptions struct {
		Format domain.Format
		Text   domain.TextOptions
	}

	renderOptions struct {
		report *distance.Report
		text   *domain.TextOptions
		format domain.Format
	}

	placeholderSwap struct {
		page        string
		placeholder string
		value       string
		errMsg      string
	}

	reportingRuntime struct {
		jsonMarshal func(any) ([]byte, error)
	}

	unknownFormatError struct {
		format domain.Format
	}

	missingPlaceholderError struct {
		message string
	}

	orderedEntry struct {
		result *metrics.MetricResult
		index  int
	}
)
