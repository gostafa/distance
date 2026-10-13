// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	"fmt"
	"io"

	"github.com/gostafa/distance/internal/features/reporting/domain"
	"github.com/gostafa/distance/internal/features/reporting/ports/outbound"
)

// New injects the renderer and destination lifecycle implementation.
func New(output outbound.Output) *Service {
	return &Service{output: renderOperations{Write: output.Write, WriteDocs: output.WriteDocs}}
}

// Write renders a report and closes the provided destination.
func (s *Service) Write(
	report *domain.Report,
	destination io.WriteCloser,
	opts *domain.WriteOptions,
) error {
	err := s.output.Write(report, destination, opts)
	if err != nil {
		return fmt.Errorf("reporting write: %w", err)
	}

	return nil
}

// WriteDocs renders a standalone metric guide and closes the destination.
func (s *Service) WriteDocs(destination io.WriteCloser, version string) error {
	err := s.output.WriteDocs(destination, version)
	if err != nil {
		return fmt.Errorf("reporting guide: %w", err)
	}

	return nil
}
