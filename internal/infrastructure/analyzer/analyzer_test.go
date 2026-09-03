// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package analyzer

import (
	"testing"

	"github.com/gostafa/distance/internal/features/projectanalysis/application"
	"github.com/gostafa/distance/internal/infrastructure/goloader"
)

func TestNewAnalyzerImplementsPort(t *testing.T) {
	t.Parallel()

	requireAnalyzer(t, NewAnalyzer(goloader.New()))
}

func requireAnalyzer(t *testing.T, a *application.Pipeline) {
	t.Helper()

	if a == nil {
		t.Fatal("NewAnalyzer returned nil")
	}
}
