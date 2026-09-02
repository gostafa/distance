// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package analyzer

import (
	"github.com/gostafa/distance/internal/features/projectanalysis/application"
	typefacts "github.com/gostafa/distance/internal/features/typefacts/application"
	"github.com/gostafa/distance/internal/infrastructure/goloader"
)

// NewAnalyzer returns the default analysis pipeline backed by goloader.
func NewAnalyzer() *application.Pipeline {
	return application.NewPipeline(typefacts.NewService(goloader.New()))
}
