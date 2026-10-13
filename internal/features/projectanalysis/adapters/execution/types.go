// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package execution

import (
	"context"

	provider "github.com/gostafa/distance/internal/features/typefacts/public"
)

type (
	// Collector implements the consuming analysis fact port.
	Collector struct {
		collect func(context.Context, *provider.FactOptions) (provider.ProjectFacts, error)
	}

	// Executor adapts the shared worker pool to analysis execution.
	Executor struct{}
)
