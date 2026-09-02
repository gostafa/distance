// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	"github.com/gostafa/distance/internal/shared/metrics"
)

type (

	// Result holds the three package-level metrics for one package.
	Result struct {
		// Abstractness is the package's interface ratio.
		Abstractness metrics.MetricResult
		// Instability is the package's efferent coupling ratio.
		Instability metrics.MetricResult
		// Distance is the package's distance from the main sequence.
		Distance metrics.MetricResult
	}
)
