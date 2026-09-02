// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package profiling

const (
	emptyString = ""
	currentDir  = "."

	errWrapOpenProfileDir = "open profile directory: %w"
	errWrapCreateProfile  = "create profile file: %w"
	errWrapStartCPU       = "profiling startCPU: %w"
	errWrapStartCPUProf   = "start cpu profile: %w"
	errWrapWriteHeap      = "profiling writeHeap: %w"
)
