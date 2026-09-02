package analyzer

import (
	"github.com/gostafa/distance/internal/features/projectanalysis/application"
	"github.com/gostafa/distance/internal/features/projectanalysis/ports/inbound"
	typefacts "github.com/gostafa/distance/internal/features/typefacts/application"
	"github.com/gostafa/distance/internal/infrastructure/goloader"
)

// NewAnalyzer returns the production analyzer: go/packages fact extraction
// feeding the metric pipeline.
func NewAnalyzer() inbound.Analyzer {
	return application.NewPipeline(typefacts.NewService(goloader.New()))
}
