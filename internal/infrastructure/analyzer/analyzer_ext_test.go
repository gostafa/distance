// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package analyzer_test

import (
	"path/filepath"
	"testing"

	"github.com/gostafa/distance/internal/features/projectanalysis/ports/inbound"
	"github.com/gostafa/distance/internal/infrastructure/analyzer"
)

// Black-box: the wired analyzer runs the real pipeline over the fixture module
// end to end (compiler load → facts → metrics).
func TestAnalyzeFixture(t *testing.T) {
	t.Parallel()

	result, err := analyzer.NewAnalyzer().Analyze(t.Context(), &inbound.Options{
		Directory:       filepath.Join("..", "..", "..", "testdata", "fixture"),
		Patterns:        []string{"./..."},
		DependencyScope: "module",
	})
	if err != nil {
		t.Fatal(err)
	}

	if result.ModulePath != "example.com/fixture" {
		t.Fatalf("module = %q", result.ModulePath)
	}

	if len(result.Packages) < 7 {
		t.Fatalf("packages = %d, want >= 7", len(result.Packages))
	}

	// Packages come back sorted by import path.
	for i := 1; i < len(result.Packages); i++ {
		if result.Packages[i-1].Path > result.Packages[i].Path {
			t.Fatalf(
				"packages not sorted: %s before %s",
				result.Packages[i-1].Path,
				result.Packages[i].Path,
			)
		}
	}
}
