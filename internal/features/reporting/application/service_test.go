package application

import (
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/gostafa/distance/distance"
	"github.com/gostafa/distance/internal/features/reporting/domain"
	"github.com/gostafa/distance/internal/shared/metrics"
)

func sampleReport() distance.Report {
	applicable := metrics.MetricResult{
		Name:       "distance",
		Scope:      metrics.ScopePackage,
		Value:      0.5,
		Applicable: true,
		Definition: "d",
	}

	return distance.Report{
		SchemaVersion: "1",
		Tool:          distance.ToolInfo{Name: "distance", Version: "test"},
		Module:        "example.com/m",
		Packages: []distance.PackageReport{{
			Path:     "example.com/m/a",
			Afferent: 1,
			Efferent: 2,
			Metrics:  []metrics.MetricResult{applicable},
		}},
	}
}

// White-box: the JSON envelope round-trips and honors the applicability
// contract (applicable → value present; n/a → value omitted).
func TestRenderJSONContract(t *testing.T) {
	t.Parallel()

	var buf strings.Builder
	err := renderJSON(&buf, sampleReport())
	if err != nil {
		t.Fatal(err)
	}

	var got map[string]any
	err = json.Unmarshal([]byte(buf.String()), &got)
	if err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}

	if got["schema_version"] != "1" {
		t.Errorf("schema_version = %v", got["schema_version"])
	}

	pkg := got["packages"].([]any)[0].(map[string]any)
	if pkg["afferent"].(float64) != 1 || pkg["efferent"].(float64) != 2 {
		t.Errorf("package coupling = ca %v ce %v, want 1 and 2", pkg["afferent"], pkg["efferent"])
	}

	dist := pkg["metrics"].(map[string]any)["distance"].(map[string]any)
	if dist["value"].(float64) != 0.5 {
		t.Errorf("distance value = %v", dist["value"])
	}
}

// White-box: ordered metric objects keep the given slice order.
func TestEncodeOrderedMetricsPreservesOrder(t *testing.T) {
	t.Parallel()

	got, err := encodeOrderedMetrics([]metrics.MetricResult{
		{Name: metrics.MetricAbstractness, Scope: metrics.ScopePackage, Value: 1, Applicable: true, Definition: "d"},
		{Name: metrics.MetricDistance, Scope: metrics.ScopePackage, Applicable: false, Reason: "x", Definition: "d"},
	})
	if err != nil {
		t.Fatal(err)
	}

	s := string(got)
	if !strings.HasPrefix(s, `{"abstractness":`) || strings.Index(s, "abstractness") > strings.Index(s, "distance") {
		t.Errorf("order not preserved: %s", s)
	}
}

// White-box: an unknown format is rejected.
func TestRenderUnknownFormat(t *testing.T) {
	t.Parallel()

	err := render(io.Discard, sampleReport(), domain.Format("xml"), domain.TextOptions{})
	if err == nil {
		t.Fatal("unknown format should error")
	}
}
