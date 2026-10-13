// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	"io"

	"github.com/gostafa/distance/internal/features/reporting/domain"
)

type (
	// Service coordinates report rendering and destination lifecycle.
	Service struct{ output renderOperations }

	renderOperations = struct {
		Write     func(*domain.Report, io.WriteCloser, *domain.WriteOptions) error
		WriteDocs func(io.WriteCloser, string) error
	}
)
