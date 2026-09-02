// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package outbound

import (
	"io"
	"testing"
)

type nopWriteCloser struct{ io.Writer }

func (nopWriteCloser) Close() error { return nil }

type stubSink struct{}

func (stubSink) Open() (Stream, error) { return NewStream(nopWriteCloser{io.Discard}), nil }

var _ Sink = stubSink{}

// White-box: the Sink port is satisfiable and yields a usable stream.
func TestSinkContract(t *testing.T) {
	t.Parallel()

	var s Sink = stubSink{}

	w, err := s.Open()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := io.WriteString(&w, "hello"); err != nil {
		t.Fatal(err)
	}

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
}
