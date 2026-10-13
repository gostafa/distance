// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package composition

import (
	"fmt"
	"os"

	analysisruntime "github.com/gostafa/distance/internal/features/projectanalysis/adapters/execution"
	projectapp "github.com/gostafa/distance/internal/features/projectanalysis/application"
	"github.com/gostafa/distance/internal/features/reporting/adapters/destinations"
	"github.com/gostafa/distance/internal/features/reporting/adapters/formats"
	reportapp "github.com/gostafa/distance/internal/features/reporting/application"
	loader "github.com/gostafa/distance/internal/features/typefacts/adapters/goloader"
	factsapp "github.com/gostafa/distance/internal/features/typefacts/application"
	factport "github.com/gostafa/distance/internal/features/typefacts/ports/outbound"
)

// NewAnalyzer wires a loader to the analysis use case.
func NewAnalyzer(source factport.FactSource) *projectapp.Pipeline {
	return projectapp.NewPipeline(
		analysisruntime.New(factsapp.NewService(source)),
		analysisruntime.Executor{},
	)
}

// DefaultAnalyzer constructs the built-in analysis use case.
func DefaultAnalyzer() *projectapp.Pipeline { return NewAnalyzer(loader.New()) }

// FileOutput constructs a file destination for reports and profiling output.
func FileOutput(path string) (*os.File, error) {
	file, err := destinations.CreateFile(path)
	if err != nil {
		return nil, fmt.Errorf("file output: %w", err)
	}

	return file, nil
}

// ReportWriter constructs the built-in reporting use case.
func ReportWriter() *reportapp.Service { return reportapp.New(formats.Output{}) }
