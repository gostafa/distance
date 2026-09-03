// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package analyzer

import (
	"errors"

	"golang.org/x/tools/go/analysis"
)

var errRuleMaxRequired = errors.New("max is required")

func bindRun(built *analysis.Analyzer, runner *runner) {
	built.Run = func(pass *analysis.Pass) (any, error) {
		return runner.run(pass)
	}
}
