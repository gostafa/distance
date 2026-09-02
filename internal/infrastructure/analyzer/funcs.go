// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package analyzer

import (
	"github.com/gostafa/distance/internal/features/projectanalysis/application"
	typefacts "github.com/gostafa/distance/internal/features/typefacts/application"
	tfoutbound "github.com/gostafa/distance/internal/features/typefacts/ports/outbound"
)

// NewAnalyzer returns the default analysis pipeline backed by loader.
func NewAnalyzer(loader tfoutbound.FactSource) *application.Pipeline {
	return application.NewPipeline(typefacts.NewService(loader))
}
