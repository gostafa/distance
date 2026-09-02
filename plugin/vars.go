//go:build distplugin

// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package plugin

import (
	"github.com/golangci/plugin-module-register/register"
	"github.com/gostafa/distance/analyzer"
)

var _ = registerDistance()

func registerDistance() int {
	register.Plugin(analyzer.Name, func(raw any) (register.LinterPlugin, error) {
		return New(raw)
	})

	return 0
}
