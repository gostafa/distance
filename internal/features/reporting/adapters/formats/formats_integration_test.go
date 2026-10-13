// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package formats_test

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"io"
	"strings"
	"testing"

	reporting "github.com/gostafa/distance/internal/features/reporting/adapters/formats"
	domain "github.com/gostafa/distance/internal/features/reporting/adapters/presentation"
	reportmodel "github.com/gostafa/distance/internal/features/reporting/domain"
)

type bufWriteCloser struct{ *bytes.Buffer }

func (bufWriteCloser) Close() error { return nil }

var _ io.WriteCloser = bufWriteCloser{}

// Black-box: the metrics guide is a self-contained HTML page carrying the
// tool version, native MathML formulas, and an entry for every metric.
func TestWriteDocs(t *testing.T) {
	t.Parallel()

	sink := &bytes.Buffer{}

	err := reporting.WriteDocs(bufWriteCloser{sink}, "v1.2.3")
	if err != nil {
		t.Fatal(err)
	}

	html := sink.String()

	if !strings.HasPrefix(html, "<!doctype html>") {
		t.Errorf("guide does not start with a doctype: %.40q", html)
	}

	wanted := []string{`id="docs-data"`, `<math`, `"v1.2.3"`}

	for _, name := range []string{reportmodel.MetricAbstractness, reportmodel.MetricInstability, reportmodel.MetricDistance} {
		wanted = append(wanted, `"name":"`+string(name)+`"`)
	}

	for _, want := range wanted {
		if !strings.Contains(html, want) {
			t.Errorf("guide is missing %q", want)
		}
	}

	if strings.Contains(html, "__DOCS_DATA__") {
		t.Error("docs placeholder was not replaced")
	}

	// Self-containment: nothing on the page may fetch an external resource
	// — the MathML must render without any math library.
	for _, ref := range []string{`src="http`, `href="http`, `url(http`, `@import`} {
		if strings.Contains(html, ref) {
			t.Errorf("guide contains external reference %q; it must be self-contained", ref)
		}
	}
}

type nopCloser struct{ io.Writer }

func (nopCloser) Close() error { return nil }

func report() *reportmodel.Report {
	return &reportmodel.Report{
		SchemaVersion: "6",
		ToolName:      "distance", ToolVersion: "test",
		Module: "example.com/m",
		Packages: []reportmodel.PackageReport{
			{
				Path:     "example.com/m/a",
				Afferent: 1,
				Efferent: 2,
				Metrics: []reportmodel.MetricResult{
					{
						Name:       "abstractness",
						Scope:      reportmodel.ScopePackage,
						Value:      0.25,
						Applicable: true,
						Definition: "a",
					},
					{
						Name:       "instability",
						Scope:      reportmodel.ScopePackage,
						Value:      0.75,
						Applicable: true,
						Definition: "i",
					},
					{
						Name:       "distance",
						Scope:      reportmodel.ScopePackage,
						Value:      0.5,
						Applicable: true,
						Definition: "d",
					},
				},
			},
		},
	}
}

// Black-box: the text format includes the module and the package row.
func TestWriteText(t *testing.T) {
	t.Parallel()

	sink := &bytes.Buffer{}

	err := reporting.Write(
		report(),
		nopCloser{sink},
		&reporting.WriteOptions{Format: domain.FormatText},
	)
	if err != nil {
		t.Fatal(err)
	}

	out := sink.String()

	if !strings.Contains(out, "example.com/m") || !strings.Contains(out, "a") {
		t.Fatalf("text output missing content:\n%s", out)
	}
}

// Black-box: the JSON format is valid and versioned.
func TestWriteJSON(t *testing.T) {
	t.Parallel()

	sink := &bytes.Buffer{}

	err := reporting.Write(
		report(),
		nopCloser{sink},
		&reporting.WriteOptions{Format: domain.FormatJSON},
	)
	if err != nil {
		t.Fatal(err)
	}

	var got map[string]any

	err = json.Unmarshal(sink.Bytes(), &got)
	if err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if got["schema_version"] != "6" {
		t.Errorf("schema_version = %v", got["schema_version"])
	}

	pkg := got["packages"].([]any)[0].(map[string]any)

	if pkg["afferent"].(float64) != 1 || pkg["efferent"].(float64) != 2 {
		t.Fatalf("package coupling = ca %v ce %v", pkg["afferent"], pkg["efferent"])
	}

	if _, ok := pkg["funcs"]; ok {
		t.Fatal("JSON still includes funcs")
	}

	if _, ok := pkg["types"]; ok {
		t.Fatal("JSON still includes types")
	}

	metricsMap, ok := pkg["metrics"].(map[string]any)

	if !ok {
		t.Fatal("package metrics missing")
	}

	for _, name := range []string{"abstractness", "instability", "distance"} {
		if _, ok := metricsMap[name]; !ok {
			t.Errorf("JSON metrics missing %q", name)
		}
	}
}

// Black-box: the CSV format starts with the canonical header and has a row per
// entity/metric.
func TestWriteCSV(t *testing.T) {
	t.Parallel()

	sink := &bytes.Buffer{}

	if err := reporting.Write(
		report(),
		nopCloser{sink},
		&reporting.WriteOptions{Format: domain.FormatCSV},
	); err != nil {
		t.Fatal(err)
	}

	records, err := csv.NewReader(sink).ReadAll()
	if err != nil {
		t.Fatalf("invalid CSV: %v", err)
	}

	if len(records) < 2 {
		t.Fatalf("csv has %d rows, want header + data", len(records))
	}

	header := strings.Join(records[0], ",")

	if header != strings.Join(domain.CSVHeader(), ",") {
		t.Errorf("csv header = %q", header)
	}

	names := map[string]bool{}

	for _, rec := range records[1:] {
		if len(rec) > 2 {
			names[rec[2]] = true
		}
	}

	for _, name := range []string{"abstractness", "instability", "distance"} {
		if !names[name] {
			t.Errorf("CSV missing metric %q", name)
		}
	}
}

// Black-box: the web format is a self-contained HTML page embedding the
// module and the versioned report payload, with the placeholder replaced.
func TestWriteWeb(t *testing.T) {
	t.Parallel()

	sink := &bytes.Buffer{}

	err := reporting.Write(
		report(),
		bufWriteCloser{sink},
		&reporting.WriteOptions{Format: domain.FormatWeb},
	)
	if err != nil {
		t.Fatal(err)
	}

	html := sink.String()

	if !strings.HasPrefix(html, "<!doctype html>") {
		t.Errorf("web report does not start with a doctype: %.40q", html)
	}

	for _, want := range []string{
		`id="report-data"`,
		`"module":"example.com/m"`,
		`"schema_version":"6"`,
		`"afferent":1`,
		`"efferent":2`,
		`"abstractness"`,
		`"instability"`,
		`"distance"`,
		`id="docs-data"`,
		`"formula_mathml"`,
		`id="report-table"`,
		`id="report-head"`,
		`id="report-body"`,
		`data-sort`,
		`aria-sort`,
		`class: 'help'`,
		`data-help`,
		`summarizeNode`,
		`aggregateMetricCell`,
		`meanStat`,
		`Package`,
		`Ca`,
		`Ce`,
		`label: 'A'`,
		`label: 'I'`,
		`label: 'Dist'`,
		`All`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("web report is missing %q", want)
		}
	}

	for _, placeholder := range []string{"__REPORT_DATA__", "__DOCS_DATA__"} {
		if strings.Contains(html, placeholder) {
			t.Errorf("placeholder %s was not replaced", placeholder)
		}
	}

	for _, gone := range []string{
		`id="kind-pills"`,
		`id="export-pills"`,
		`"functions":`,
		`"variables":`,
		`"types":`,
		`Funcs`,
		`Vars`,
		`Consts`,
		`Types`,
	} {
		if strings.Contains(html, gone) {
			t.Errorf("web report still contains removed surface %q", gone)
		}
	}

	// Self-containment: nothing on the page may fetch an external resource.
	for _, ref := range []string{`src="http`, `href="http`, `url(http`, `@import`} {
		if strings.Contains(html, ref) {
			t.Errorf("web report contains external reference %q; it must be self-contained", ref)
		}
	}
}

// Black-box: hostile identifiers cannot terminate the payload's script
// element early — json.Marshal HTML-escapes every angle bracket.
func TestWriteWebEscapesScriptTerminator(t *testing.T) {
	t.Parallel()

	rep := report()

	rep.Packages[0].Path = "</script><script>alert(1)</script>"

	sink := &bytes.Buffer{}

	err := reporting.Write(
		rep,
		bufWriteCloser{sink},
		&reporting.WriteOptions{Format: domain.FormatWeb},
	)
	if err != nil {
		t.Fatal(err)
	}

	html := sink.String()

	if strings.Contains(html, "</script><script>alert(1)") {
		t.Error("payload contains an unescaped script terminator")
	}

	if !strings.Contains(html, `\u003c/script\u003e`) &&
		!strings.Contains(html, `\u003c`) {

		t.Error("angle brackets in the payload are not escaped")
	}
}

// Black-box: a hostile identifier spelling the docs placeholder cannot
// hijack the docs script element — the trusted docs payload is injected
// before the untrusted report payload.
func TestWriteWebPayloadCannotSpoofDocsPlaceholder(t *testing.T) {
	t.Parallel()

	rep := report()

	rep.Packages[0].Path = "__DOCS_DATA__"

	sink := &bytes.Buffer{}

	err := reporting.Write(
		rep,
		bufWriteCloser{sink},
		&reporting.WriteOptions{Format: domain.FormatWeb},
	)
	if err != nil {
		t.Fatal(err)
	}

	html := sink.String()

	if got := strings.Count(html, `id="docs-data"`); got != 1 {
		t.Errorf("docs-data script elements = %d, want 1", got)
	}

	if !strings.Contains(html, `"path":"__DOCS_DATA__"`) {
		t.Error("hostile package path is missing from the report payload")
	}

	if !strings.Contains(html, `"formula_mathml"`) {
		t.Error("docs payload was not injected")
	}
}
