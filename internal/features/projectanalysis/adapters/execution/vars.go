// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package execution

import (
	executionport "github.com/gostafa/distance/internal/features/projectanalysis/ports/execution"
	"github.com/gostafa/distance/internal/features/projectanalysis/ports/outbound"
)

var (
	_ outbound.FactCollector = (*Collector)(nil)

	_ executionport.WorkerExecutor = Executor{}
)
