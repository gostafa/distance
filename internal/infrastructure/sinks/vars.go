// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package sinks

import (
	"github.com/gostafa/distance/internal/features/reporting/ports/outbound"
)

var (
	_ outbound.Sink = StdoutSink{}
	_ outbound.Sink = FileSink{}
)
