// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package metrics_test

import (
	"testing"

	"github.com/gostafa/distance/internal/shared/metrics"
)

// Black-box: a package with balanced abstractness/instability sits on the main
// sequence (distance 0).
func TestDistanceOnMainSequence(t *testing.T) {
	t.Parallel()

	a := metrics.Abstractness(1, 2) // 0.5
	i := metrics.Instability(1, 1)  // 0.5

	d := metrics.Distance(&a, &i)

	if !d.Applicable || d.Value != 0 {
		t.Fatalf("distance = %+v, want 0", d)
	}
}
