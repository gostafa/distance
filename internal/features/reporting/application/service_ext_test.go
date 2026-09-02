// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application_test

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/gostafa/distance/distance"
	reporting "github.com/gostafa/distance/internal/features/reporting/application"
	"github.com/gostafa/distance/internal/features/reporting/domain"
	"github.com/gostafa/distance/internal/features/reporting/ports/outbound"
	"github.com/gostafa/distance/internal/shared/metrics"
)

type bufSink struct{ buf *bytes.Buffer }

func (b bufSink) Open() (outbound.Stream, error) {
	return outbound.NewStream(nopCloser{b.buf}), nil
}

type nopCloser struct{ io.Writer }

func (nopCloser) Close() error { return nil }

func report() *distance.Report {
	return &distance.Report{
		SchemaVersion: "6",
		Tool:          distance.ToolIdent("distance", "test"),
		Module:        "example.com/m",
		Packages: []distance.PackageReport{
			{
				Path:     "example.com/m/a",
				Afferent: 1,
				Efferent: 2,
				Metrics: []metrics.MetricResult{
					{
						Name:       "abstractness",
						Scope:      metrics.ScopePackage,
						Value:      0.25,
						Applicable: true,
						Definition: "a",
					},
					{
						Name:       "instability",
						Scope:      metrics.ScopePackage,
						Value:      0.75,
						Applicable: true,
						Definition: "i",
					},
					{
						Name:       "distance",
						Scope:      metrics.ScopePackage,
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

	sink := bufSink{&bytes.Buffer{}}

	err := reporting.Write(report(), sink, &reporting.WriteOptions{Format: domain.FormatText})
	if err != nil {
		t.Fatal(err)
	}

	out := sink.buf.String()

	if !strings.Contains(out, "example.com/m") || !strings.Contains(out, "a") {
		t.Fatalf("text output missing content:\n%s", out)
	}
}

// Black-box: the JSON format is valid and versioned.
func TestWriteJSON(t *testing.T) {
	t.Parallel()

	sink := bufSink{&bytes.Buffer{}}

	err := reporting.Write(report(), sink, &reporting.WriteOptions{Format: domain.FormatJSON})
	if err != nil {
		t.Fatal(err)
	}

	var got map[string]any

	err = json.Unmarshal(sink.buf.Bytes(), &got)
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

	sink := bufSink{&bytes.Buffer{}}

	if err := reporting.Write(
		report(),
		sink,
		&reporting.WriteOptions{Format: domain.FormatCSV},
	); err != nil {
		t.Fatal(err)
	}

	records, err := csv.NewReader(sink.buf).ReadAll()
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
