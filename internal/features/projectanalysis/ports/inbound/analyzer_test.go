// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package inbound

import (
	"strings"
	"testing"

	"github.com/gostafa/distance/internal/shared/metrics"
)

// White-box: the debug Stringer summarizes a package result.
func TestPackageResultString(t *testing.T) {
	t.Parallel()

	pr := PackageResult{
		Path:    "example.com/m/p",
		Metrics: make([]metrics.MetricResult, 2),
	}

	s := pr.String()

	for _, want := range []string{"example.com/m/p", "2 package metrics"} {
		if !strings.Contains(s, want) {
			t.Errorf("String()=%q missing %q", s, want)
		}
	}
}
