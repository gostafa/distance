// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package workerpool

type (

	// PoolConfig supplies worker and task counts for a parallel run.
	PoolConfig interface {
		TaskCount() int
		WorkerLimit() int
	}

	// Config configures a parallel indexed run.
	Config struct {
		// Workers is the maximum number of goroutines to spawn.
		Workers int
		// Tasks is the number of indexed tasks in [0, Tasks).
		Tasks int
	}
)
