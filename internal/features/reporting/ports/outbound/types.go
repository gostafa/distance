// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package outbound

import (
	"fmt"
	"io"
)

type (

	// Stream is a concrete write destination returned by Sink.Open.
	Stream struct {
		writer io.Writer
		closer io.Closer
	}

	// Sink opens a destination for a rendered report.
	Sink interface {
		// Open returns the stream; the caller closes it when rendering is done.
		Open() (Stream, error)
	}
)

// NewStream wraps a write-closer as a Stream.
func NewStream(writeCloser io.WriteCloser) Stream {
	return Stream{writer: writeCloser, closer: writeCloser}
}

// Close closes the underlying closer when present.
func (stream *Stream) Close() error {
	if stream.closer == nil {
		return nil
	}

	closeErr := stream.closer.Close()
	if closeErr != nil {
		return fmt.Errorf("outbound Stream: %w", closeErr)
	}

	return nil
}

// Write writes p to the underlying writer.
func (stream *Stream) Write(p []byte) (int, error) {
	count, writeErr := stream.writer.Write(p)
	if writeErr != nil {
		return count, fmt.Errorf("outbound Stream write: %w", writeErr)
	}

	return count, nil
}
