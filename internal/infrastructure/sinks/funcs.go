// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package sinks

import (
	"bufio"
	"fmt"
	"os"

	"github.com/gostafa/distance/internal/features/reporting/ports/outbound"
)

// Open returns a buffered writer for standard output.
func (StdoutSink) Open() (outbound.Stream, error) {
	return outbound.NewStream(stdoutStream{w: bufio.NewWriter(os.Stdout)}), nil
}

func (stream stdoutStream) Close() error {
	flushErr := stream.w.Flush()
	if flushErr != nil {
		return fmt.Errorf("stdout flush: %w", flushErr)
	}

	return nil
}

func (stream stdoutStream) Write(p []byte) (int, error) {
	count, writeErr := stream.w.Write(p)
	if writeErr != nil {
		return count, fmt.Errorf("stdout write: %w", writeErr)
	}

	return count, nil
}

// Open creates (or truncates) the sink's destination file.
func (sink FileSink) Open() (outbound.Stream, error) {
	file, createErr := os.Create(sink.Path)
	if createErr != nil {
		return outbound.Stream{}, fmt.Errorf("create report file: %w", createErr)
	}

	return outbound.NewStream(file), nil
}
