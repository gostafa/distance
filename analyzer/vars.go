// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package analyzer

import (
	"golang.org/x/tools/go/analysis"
)

func bindRun(built *analysis.Analyzer, runner *runner) {
	built.Run = func(pass *analysis.Pass) (any, error) {
		return runner.run(pass)
	}
}
