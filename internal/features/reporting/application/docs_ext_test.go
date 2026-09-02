// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	reporting "github.com/gostafa/distance/internal/features/reporting/application"
	"github.com/gostafa/distance/distance"
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

	for _, name := range distance.AllMetrics() {
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
