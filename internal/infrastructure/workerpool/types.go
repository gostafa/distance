// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package workerpool

type (
	// WorkerRun configures the indexed task pool.
	WorkerRun = struct {
		Workers int
		Tasks   int
	}
)
