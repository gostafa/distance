package application_test

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"io"
	"strings"
	"testing"

	reporting "github.com/gostafa/distance/internal/features/reporting/application"
	"github.com/gostafa/distance/internal/features/reporting/domain"
	"github.com/gostafa/distance/internal/shared/metrics"
	"github.com/gostafa/distance/distance"
)

type bufSink struct{ buf *bytes.Buffer }

func (b bufSink) Open() (io.WriteCloser, error) { return nopCloser{b.buf}, nil }

type nopCloser struct{ io.Writer }

func (nopCloser) Close() error { return nil }

func report() distance.Report {
	return distance.Report{
		SchemaVersion: "1",
		Tool:          distance.ToolInfo{Name: "distance", Version: "test"},
		Module:        "example.com/m",
		Packages: []distance.PackageReport{
			{
				Path:   "example.com/m/a",
				Vars:   1,
				Consts: 1,
				Variables: []distance.DeclarationReport{{
					Name:     "Config",
					Exported: true,
					Position: distance.Position{File: "a/a.go", Line: 3, Column: 5},
				}},
				Constants: []distance.DeclarationReport{{
					Name:     "Mode",
					Exported: true,
					Position: distance.Position{File: "a/a.go", Line: 4, Column: 7},
				}},
				Functions: []distance.FunctionReport{{
					Name:     "Run",
					Exported: true,
					Position: distance.Position{File: "a/a.go", Line: 6, Column: 1},
					Lines:    5,
				}},
				Metrics: []metrics.MetricResult{
					{
						Name:       "distance",
						Scope:      metrics.ScopePackage,
						Value:      0.5,
						Applicable: true,
						Definition: "d",
					},
				},
				Types: []distance.TypeReport{{
					Name:     "A",
					Exported: true,
					Kind:     "struct",
					Position: distance.Position{File: "a/a.go", Line: 10, Column: 6},
					Fields:   1,
					FieldDetails: []distance.FieldReport{{
						Name:     "Value",
						Exported: true,
					}},
					Methods: 1,
					MethodDetails: []distance.FunctionReport{{
						Name:     "Do",
						Exported: true,
						Receiver: "A",
						Position: distance.Position{File: "a/a.go", Line: 12, Column: 1},
						Lines:    1,
					}},
				}},
			},
		},
	}
}

// Black-box: the text format includes the module and the type row.
func TestWriteText(t *testing.T) {
	t.Parallel()

	sink := bufSink{&bytes.Buffer{}}
	err := reporting.Write(report(), domain.FormatText, sink, domain.TextOptions{})
	if err != nil {
		t.Fatal(err)
	}

	out := sink.buf.String()
	if !strings.Contains(out, "example.com/m") || !strings.Contains(out, "A") {
		t.Fatalf("text output missing content:\n%s", out)
	}
}

// Black-box: the JSON format is valid and versioned.
func TestWriteJSON(t *testing.T) {
	t.Parallel()

	sink := bufSink{&bytes.Buffer{}}
	err := reporting.Write(report(), domain.FormatJSON, sink, domain.TextOptions{})
	if err != nil {
		t.Fatal(err)
	}

	var got map[string]any
	err = json.Unmarshal(sink.buf.Bytes(), &got)
	if err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if got["schema_version"] != "1" {
		t.Errorf("schema_version = %v", got["schema_version"])
	}
	pkg := got["packages"].([]any)[0].(map[string]any)
	if pkg["vars"].(float64) != 1 || pkg["consts"].(float64) != 1 {
		t.Fatalf("package counts = vars %v consts %v", pkg["vars"], pkg["consts"])
	}
	fn := pkg["functions"].([]any)[0].(map[string]any)
	if fn["name"] != "Run" || fn["lines"].(float64) != 5 {
		t.Fatalf("function details = %+v", fn)
	}
	typ := pkg["types"].([]any)[0].(map[string]any)
	if typ["kind"] != "struct" || typ["exported"] != true {
		t.Fatalf("type details = %+v", typ)
	}
	method := typ["method_details"].([]any)[0].(map[string]any)
	if method["receiver"] != "A" || method["name"] != "Do" {
		t.Fatalf("method details = %+v", method)
	}
}

// Black-box: the CSV format starts with the canonical header and has a row per
// entity/metric.
func TestWriteCSV(t *testing.T) {
	t.Parallel()

	sink := bufSink{&bytes.Buffer{}}
	if err := reporting.Write(report(), domain.FormatCSV, sink, domain.TextOptions{}); err != nil {
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
}
