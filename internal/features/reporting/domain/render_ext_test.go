package domain_test

import (
	"strings"
	"testing"

	reporting "github.com/gostafa/distance/internal/features/reporting/domain"
	"github.com/gostafa/distance/internal/shared/metrics"
	"github.com/gostafa/distance/distance"
)

// Black-box: format parsing accepts the known encodings and rejects others.
func TestParseFormat(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"text", "json", "csv", "web"} {
		if f, ok := reporting.ParseFormat(name); !ok || string(f) != name {
			t.Errorf("ParseFormat(%q) = %v,%v", name, f, ok)
		}
	}

	if _, ok := reporting.ParseFormat("xml"); ok {
		t.Error("xml must be rejected")
	}
}

// Black-box: the text and CSV renderers emit the module and a per-metric
// header/records.
func TestTextAndCSVRendering(t *testing.T) {
	t.Parallel()

	rep := distance.Report{
		SchemaVersion: "1",
		Tool:          distance.ToolInfo{Name: "distance", Version: "t"},
		Module:        "example.com/m",
		Packages: []distance.PackageReport{
			{
				Path: "example.com/m/a",
				Metrics: []metrics.MetricResult{
					{
						Name:       "abstractness",
						Scope:      metrics.ScopePackage,
						Value:      0.5,
						Applicable: true,
					},
				},
				Types: []distance.TypeReport{{Name: "A"}},
			},
		},
	}

	text := reporting.Text(rep, reporting.TextOptions{})
	if !strings.Contains(text, "example.com/m") || !strings.Contains(text, "A") {
		t.Errorf("text output missing content:\n%s", text)
	}

	if len(reporting.CSVHeader()) == 0 {
		t.Error("empty CSV header")
	}

	if len(reporting.CSVRecords(rep)) == 0 {
		t.Error("no CSV records produced")
	}

	if reporting.FormatValue(0.5) == "" {
		t.Error("FormatValue produced empty string")
	}
}
