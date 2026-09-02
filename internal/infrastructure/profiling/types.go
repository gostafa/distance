// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package profiling

import (
	"io"
	"os"
)

type (
	heapSink interface {
		writeTo(writer io.Writer) error
	}

	fileCloser interface {
		closeNamed(file *os.File) error
	}

	profilingRuntime struct {
		writeHeapProfile func(io.Writer) error
		closeFile        func(*os.File) error
	}
)
