// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package composition

import (
	"testing"

	application "github.com/gostafa/distance/internal/features/projectanalysis/public"
	"github.com/gostafa/distance/internal/features/typefacts/adapters/goloader"
)

func TestNewAnalyzerImplementsPort(t *testing.T) {
	t.Parallel()

	requireAnalyzer(t, NewAnalyzer(goloader.New()))
}

func requireAnalyzer(t *testing.T, a application.Pipeline) {
	t.Helper()

	if a == nil {
		t.Fatal("NewAnalyzer returned nil")
	}
}
