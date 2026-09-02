// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package plugin

import (
	"github.com/gostafa/distance/analyzer"
)

type (

	// Plugin is the golangci-lint LinterPlugin for the distance analyzer.
	Plugin struct {
		loadMode

		settings analyzer.Settings
	}

	loadMode struct{}
)
