// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package analyzer

import (
	"testing"

	"github.com/gostafa/distance/internal/features/projectanalysis/ports/inbound"
)

// White-box: the composition root wires up an analyzer satisfying the port.
func TestNewAnalyzerImplementsPort(t *testing.T) {
	t.Parallel()

	requireAnalyzer(t, NewAnalyzer())
}

func requireAnalyzer(t *testing.T, a inbound.Analyzer) {
	t.Helper()

	if a == nil {
		t.Fatal("NewAnalyzer returned nil")
	}
}
