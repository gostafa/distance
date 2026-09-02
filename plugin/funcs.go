// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package plugin

import (
	"fmt"

	"github.com/golangci/plugin-module-register/register"
	"github.com/gostafa/distance/analyzer"
	"golang.org/x/tools/go/analysis"
)

var _ register.LinterPlugin = (*Plugin)(nil)

func (loadMode) GetLoadMode() string {
	return register.LoadModeTypesInfo
}

// BuildAnalyzers constructs the go/analysis analyzers for this plugin.
func (plugin *Plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	built, buildErr := analyzer.New(&plugin.settings)
	if buildErr != nil {
		return nil, fmt.Errorf("plugin BuildAnalyzers: %w", buildErr)
	}

	return []*analysis.Analyzer{built}, nil
}

// New decodes analyzer settings and returns the distance plugin.
func New(raw any) (*Plugin, error) {
	settings, decodeErr := register.DecodeSettings[analyzer.Settings](raw)
	if decodeErr != nil {
		return nil, fmt.Errorf("plugin New: %w", decodeErr)
	}

	return &Plugin{settings: settings}, nil
}
