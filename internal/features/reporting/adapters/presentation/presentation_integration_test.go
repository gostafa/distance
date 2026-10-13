// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package presentation_test

import (
	"strings"
	"testing"

	domain "github.com/gostafa/distance/internal/features/reporting/adapters/presentation"
	reportmodel "github.com/gostafa/distance/internal/features/reporting/domain"
)

// Black-box: every reported metric and structural column has a complete
// guide entry, with no duplicate names.
func TestMetricDocsCoverEveryMetric(t *testing.T) {
	t.Parallel()

	docs := domain.MetricDocs()

	byName := make(map[string]domain.MetricDoc, len(docs))

	for _, d := range docs {
		if _, dup := byName[d.Name]; dup {
			t.Errorf("duplicate docs entry %q", d.Name)
		}

		byName[d.Name] = d
	}

	for _, name := range []string{reportmodel.MetricAbstractness, reportmodel.MetricInstability, reportmodel.MetricDistance} {
		assertMetricDoc(t, byName, string(name), domain.DocScopePackage)
	}

	for _, name := range []string{"ca", "ce"} {
		d, ok := byName[name]

		if !ok {
			t.Errorf("structural column %q has no docs entry", name)

			continue
		}

		if d.Scope != domain.DocScopeStructural {
			t.Errorf("%s scope = %q, want %q", name, d.Scope, domain.DocScopeStructural)
		}

		if d.Label == "" || d.FullName == "" || d.Summary == "" ||
			d.HowCalculated == "" || d.Interpretation == "" || d.Example == "" {

			t.Errorf("structural entry %q has empty prose fields", name)
		}

		// Structural columns are counted, not computed: no formula, no id.
		if d.FormulaMathML != "" || d.Definition != "" {
			t.Errorf("structural entry %q must not carry a formula or definition", name)
		}
	}
}

// assertMetricDoc checks one computed metric's entry for completeness.
func assertMetricDoc(
	t *testing.T,
	byName map[string]domain.MetricDoc,
	name string,
	scope domain.DocScope,
) {
	t.Helper()

	d, ok := byName[name]

	if !ok {
		t.Errorf("metric %q has no docs entry", name)

		return
	}

	if d.Scope != scope {
		t.Errorf("%s scope = %q, want %q", name, d.Scope, scope)
	}

	for field, value := range map[string]string{
		"Label":          d.Label,
		"FullName":       d.FullName,
		"FormulaLaTeX":   d.FormulaLaTeX,
		"Summary":        d.Summary,
		"HowCalculated":  d.HowCalculated,
		"Interpretation": d.Interpretation,
		"Example":        d.Example,
	} {
		if value == "" {
			t.Errorf("%s: %s is empty", name, field)
		}
	}

	if !strings.Contains(d.FormulaMathML, "<math") {
		t.Errorf("%s: FormulaMathML carries no <math> markup", name)
	}

	if !strings.HasPrefix(d.Definition, "distance/") {
		t.Errorf("%s: Definition = %q, want a versioned distance id", name, d.Definition)
	}
}

// Black-box: format parsing accepts the known encodings and rejects others.
func TestParseFormat(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"text", "json", "csv", "web"} {
		if f, ok := domain.ParseFormat(name); !ok || string(f) != name {
			t.Errorf("ParseFormat(%q) = %v,%v", name, f, ok)
		}
	}

	if _, ok := domain.ParseFormat("xml"); ok {
		t.Error("xml must be rejected")
	}
}

// Black-box: the text and CSV renderers emit the module and a per-metric
// header/records.
func TestTextAndCSVRendering(t *testing.T) {
	t.Parallel()

	rep := reportmodel.Report{
		SchemaVersion: "1",
		ToolName:      "distance", ToolVersion: "t",
		Module: "example.com/m",
		Packages: []reportmodel.PackageReport{
			{
				Path: "example.com/m/a",
				Metrics: []reportmodel.MetricResult{
					{
						Name:       "abstractness",
						Scope:      reportmodel.ScopePackage,
						Value:      0.5,
						Applicable: true,
					},
				},
			},
		},
	}

	text := domain.Text(&rep, &domain.TextOptions{})

	if !strings.Contains(text, "example.com/m") || !strings.Contains(text, "a") {
		t.Errorf("text output missing content:\n%s", text)
	}

	if len(domain.CSVHeader()) == 0 {
		t.Error("empty CSV header")
	}

	if len(domain.CSVRecords(&rep)) == 0 {
		t.Error("no CSV records produced")
	}

	if domain.FormatValue(0.5) == "" {
		t.Error("FormatValue produced empty string")
	}
}
