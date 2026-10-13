// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package formats

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"
	"testing"

	domain "github.com/gostafa/distance/internal/features/reporting/adapters/presentation"
	reportmodel "github.com/gostafa/distance/internal/features/reporting/domain"
	"github.com/gostafa/distance/internal/features/reporting/ports/outbound"
)

type trackingWriteCloser struct {
	err    error
	closed bool
}

func (w *trackingWriteCloser) Write([]byte) (int, error) {
	return 0, w.err
}

func (w *trackingWriteCloser) Close() error {
	w.closed = true

	return nil
}

func TestWriteRenderErrors(t *testing.T) {
	sentinel := errors.New("write failed")
	w := &trackingWriteCloser{err: sentinel}

	err := Write(
		sampleReport(),
		w,
		&WriteOptions{Format: domain.FormatText},
	)
	if !errors.Is(err, sentinel) {
		t.Fatalf("render error = %v, want sentinel", err)
	}

	if !w.closed {
		t.Fatal("writer was not closed after a render error")
	}
}

func TestJSONDebugStringsAndMarshalError(t *testing.T) {
	reportSummary := jsonReportString(&jsonReport{
		SchemaVersion: "3",
		Tool:          jsonTool{Name: "distance", Version: "test"},
		Packages:      []jsonPackage{{Path: "example.com/p"}},
	})

	if !strings.Contains(reportSummary, "schema 3") ||
		!strings.Contains(reportSummary, "1 packages") {

		t.Fatalf("jsonReportString() = %q", reportSummary)
	}

	_, err := encodeOrderedMetrics([]reportmodel.MetricResult{{
		Name:       string(reportmodel.MetricDistance),
		Scope:      reportmodel.ScopePackage,
		Value:      math.NaN(),
		Applicable: true,
	}})
	if err == nil {
		t.Fatal("expected JSON encoding to reject NaN")
	}
}

func TestWriteDocsErrors(t *testing.T) {
	original := docsTemplate

	docsTemplate = "missing placeholder"

	t.Cleanup(func() { docsTemplate = original })

	err := renderDocs(io.Discard, "test")
	if err == nil {
		t.Fatal("expected a missing docs placeholder error")
	}

	w := &trackingWriteCloser{}

	err = WriteDocs(w, "test")
	if err == nil {
		t.Fatal("expected WriteDocs to propagate the render error")
	}

	if !w.closed {
		t.Fatal("writer was not closed after the docs render error")
	}
}

func TestRenderWebMissingPlaceholders(t *testing.T) {
	original := webTemplate

	t.Cleanup(func() { webTemplate = original })

	webTemplate = webDataPlaceholder

	err := renderWeb(io.Discard, sampleReport())
	if err == nil {
		t.Fatal("expected a missing docs placeholder error")
	}

	webTemplate = docsDataPlaceholder

	err = renderWeb(io.Discard, sampleReport())
	if err == nil {
		t.Fatal("expected a missing report placeholder error")
	}
}

type failWriter struct {
	err   error
	allow int
	n     int
}

func (w *failWriter) Write(p []byte) (int, error) {
	w.n++

	if w.n > w.allow {
		return 0, w.err
	}

	return len(p), nil
}

func (failWriter) Close() error { return nil }

func TestRenderCSVWriteErrors(t *testing.T) {
	sentinel := errors.New("csv write failed")

	// csv.Writer buffers through bufio; enough rows force a flush to the
	// underlying writer during WriteAll.
	big := sampleReport()

	for i := range 200 {
		big.Packages = append(big.Packages, reportmodel.PackageReport{
			Path: fmt.Sprintf("example.com/m/p%d", i),
			Metrics: []reportmodel.MetricResult{
				{
					Name: string(
						reportmodel.MetricDistance,
					),
					Scope:      reportmodel.ScopePackage,
					Value:      float64(i),
					Applicable: true,
					Definition: "d",
					Reason:     strings.Repeat("x", 64),
				},
			},
		})
	}

	err := render(
		&failWriter{allow: 0, err: sentinel},
		&renderOptions{
			report: big,
			format: domain.FormatCSV,
		},
	)
	if !errors.Is(err, sentinel) {
		t.Fatalf("csv write error = %v, want sentinel", err)
	}
}

func TestJSONMarshalSeamErrors(t *testing.T) {
	sentinel := errors.New("marshal failed")
	runtime := reportingRuntime{jsonMarshal: func(any) ([]byte, error) { return nil, sentinel }}

	err := renderDocsWith(runtime, io.Discard, "test")
	if !errors.Is(err, sentinel) {
		t.Fatalf("renderDocs = %v, want sentinel", err)
	}

	err = renderWebWith(runtime, io.Discard, sampleReport())
	if !errors.Is(err, sentinel) {
		t.Fatalf("renderWeb = %v, want sentinel", err)
	}

	if _, err := encodeOrderedMetricsWith(runtime, []reportmodel.MetricResult{
		{
			Name: string(
				reportmodel.MetricDistance,
			),
			Scope:      reportmodel.ScopePackage,
			Value:      1,
			Applicable: true,
		},
	}); !errors.Is(err, sentinel) {
		t.Fatalf("encodeOrderedMetrics = %v, want sentinel", err)
	}
}

func TestMarshalDocsErrorViaRenderWeb(t *testing.T) {
	sentinel := errors.New("docs marshal failed")
	runtime := reportingRuntime{jsonMarshal: func(value any) ([]byte, error) {
		if _, ok := value.(docsPayload); ok {
			return nil, sentinel
		}

		return json.Marshal(value)
	}}

	err := renderWebWith(runtime, io.Discard, sampleReport())
	if !errors.Is(err, sentinel) {
		t.Fatalf("renderWeb docs error = %v, want sentinel", err)
	}
}

func sampleReport() *reportmodel.Report {
	applicable := reportmodel.MetricResult{
		Name:       "distance",
		Scope:      reportmodel.ScopePackage,
		Value:      0.5,
		Applicable: true,
		Definition: "d",
	}

	return &reportmodel.Report{
		SchemaVersion: "1",
		ToolName:      "distance", ToolVersion: "test",
		Module: "example.com/m",
		Packages: []reportmodel.PackageReport{{
			Path:     "example.com/m/a",
			Afferent: 1,
			Efferent: 2,
			Metrics:  []reportmodel.MetricResult{applicable},
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

	got, err := encodeOrderedMetrics([]reportmodel.MetricResult{
		{
			Name:       string(reportmodel.MetricAbstractness),
			Scope:      reportmodel.ScopePackage,
			Value:      1,
			Applicable: true,
			Definition: "d",
		},
		{
			Name:       string(reportmodel.MetricDistance),
			Scope:      reportmodel.ScopePackage,
			Applicable: false,
			Reason:     "x",
			Definition: "d",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	s := string(got)

	if !strings.HasPrefix(s, `{"abstractness":`) ||
		strings.Index(s, "abstractness") > strings.Index(s, "distance") {

		t.Errorf("order not preserved: %s", s)
	}
}

// White-box: an unknown format is rejected.
func TestRenderUnknownFormat(t *testing.T) {
	t.Parallel()

	err := render(io.Discard, &renderOptions{
		report: sampleReport(),
		format: domain.Format("xml"),
	})
	if err == nil {
		t.Fatal("unknown format should error")
	}
}

type boundaryWriter struct {
	err    error
	closed bool
}

func (w *boundaryWriter) Write(p []byte) (int, error) { return 0, w.err }

func (w *boundaryWriter) Close() error { w.closed = true; return nil }

func TestOutputBoundarySanitizesErrorsAndCloses(t *testing.T) {
	hidden := errors.New("filesystem failure")
	for _, guide := range []bool{false, true} {
		writer := &boundaryWriter{err: hidden}
		var err error
		if guide {
			err = (Output{}).WriteDocs(writer, "dev")
		} else {
			err = (Output{}).Write(
				&reportmodel.Report{},
				writer,
				&reportmodel.WriteOptions{Format: "json"},
			)
		}
		if !writer.closed || !errors.Is(err, outbound.ErrOutput) || errors.Is(err, hidden) {
			t.Fatalf("boundary contract: closed=%v err=%v", writer.closed, err)
		}
	}
}

func TestOrderedMarshalRejectsInvalidValues(t *testing.T) {
	rows := orderedMetrics{{Name: "distance", Value: math.NaN(), Applicable: true}}
	if _, err := rows.MarshalJSON(); err == nil {
		t.Fatal("accepted NaN in ordered metrics")
	}
}

func TestWebOutputWriteFailure(t *testing.T) {
	failure := errors.New("destination failed")
	writer := &trackingWriteCloser{err: failure}
	if err := Write(
		sampleReport(),
		writer,
		&WriteOptions{Format: "web"},
	); !errors.Is(err, failure) ||
		!writer.closed {
		t.Fatalf("web write failure: %v closed=%v", err, writer.closed)
	}
}

func TestWebOutputShortWrite(t *testing.T) {
	writer := &trackingWriteCloser{}
	if err := Write(
		sampleReport(),
		writer,
		&WriteOptions{Format: "web"},
	); !errors.Is(err, errShortWrite) ||
		!writer.closed {
		t.Fatalf("short write: %v closed=%v", err, writer.closed)
	}
}

type closeFailureWriter struct{ err error }

func (w closeFailureWriter) Write(payload []byte) (int, error) { return len(payload), nil }

func (w closeFailureWriter) Close() error { return w.err }

func TestOutputCloseFailure(t *testing.T) {
	failure := errors.New("close failed")
	err := Write(sampleReport(), closeFailureWriter{err: failure}, &WriteOptions{Format: "json"})
	if !errors.Is(err, failure) {
		t.Fatalf("close failure = %v", err)
	}
}
