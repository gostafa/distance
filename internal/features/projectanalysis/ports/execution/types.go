// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package execution

import (
	"context"
)

type (
	// WorkerRun configures bounded execution of indexed tasks.
	WorkerRun = struct {
		Workers int
		Tasks   int
	}

	// WorkerExecutor runs independent metric computations.
	WorkerExecutor interface {
		// Run executes each index once, bounds concurrency, and returns task or cancellation failures.
		Run(ctx context.Context, cfg WorkerRun, task func(index int) error) error
	}
)
