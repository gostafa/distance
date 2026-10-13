// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package plugin

import (
	"fmt"

	"github.com/golangci/plugin-module-register/register"
	analyzer "github.com/gostafa/distance/internal/features/projectanalysis/adapters/goanalysis"
	"golang.org/x/tools/go/analysis"
)

// BuildAnalyzers returns the distance go/analysis Analyzer.
func (plugin *pluginInstance) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	analyzers, err := runBuild(plugin.build)
	if err != nil {
		return nil, fmt.Errorf(errFmtBuildAnalyzers, err)
	}

	return analyzers, nil
}

// GetLoadMode requests type information so diagnostics can locate package
// positions accurately.
func (plugin *pluginInstance) GetLoadMode() string {
	return loadModeFor(plugin.build)
}

func analyzerBuilder(settings *analyzer.Settings) func() ([]*analysis.Analyzer, error) {
	return func() ([]*analysis.Analyzer, error) {
		analyzerInstance, err := analyzer.New(settings)
		if err != nil {
			return nil, fmt.Errorf(errFmtBuildAnalyzers, err)
		}

		return []*analysis.Analyzer{analyzerInstance}, nil
	}
}

func loadModeFor(func() ([]*analysis.Analyzer, error)) string {
	return register.LoadModeTypesInfo
}

// newPlugin constructs the module plugin from inline settings.
func newPlugin(raw any) (*pluginInstance, error) {
	settings, err := analyzer.DecodeSettings(raw)
	if err != nil {
		return nil, fmt.Errorf("new plugin: %w", err)
	}

	return &pluginInstance{build: analyzerBuilder(&settings)}, nil
}

func registerDistance() int {
	register.Plugin(analyzer.Name, func(raw any) (register.LinterPlugin, error) {
		pluginInstance, err := newPlugin(raw)
		if err != nil {
			return nil, fmt.Errorf("registerDistance: %w", err)
		}

		return pluginInstance, nil
	})

	return registerDone
}

func runBuild(build func() ([]*analysis.Analyzer, error)) ([]*analysis.Analyzer, error) {
	analyzers, err := build()
	if err != nil {
		return nil, fmt.Errorf("build analyzers: %w", err)
	}

	return analyzers, nil
}
