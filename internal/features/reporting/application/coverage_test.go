// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"
	"testing"

	"github.com/gostafa/distance/distance"
	"github.com/gostafa/distance/internal/features/reporting/domain"
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

	_, err := encodeOrderedMetrics([]distance.MetricResult{{
		Name:       string(distance.MetricDistance),
		Scope:      distance.ScopePackage,
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
		big.Packages = append(big.Packages, distance.PackageReport{
			Path: fmt.Sprintf("example.com/m/p%d", i),
			Metrics: []distance.MetricResult{
				{
					Name: string(
						distance.MetricDistance,
					),
					Scope:      distance.ScopePackage,
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

	if _, err := encodeOrderedMetricsWith(runtime, []distance.MetricResult{
		{
			Name: string(
				distance.MetricDistance,
			),
			Scope:      distance.ScopePackage,
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
