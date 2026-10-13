// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package plugin

import (
	"golang.org/x/tools/go/analysis"
)

type (
	// pluginInstance is the golangci-lint module plugin for distance.
	pluginInstance struct {
		build func() ([]*analysis.Analyzer, error)
	}
)
