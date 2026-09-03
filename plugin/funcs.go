// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package plugin

import (
	"encoding/json"
	"fmt"

	"github.com/golangci/plugin-module-register/register"
	"github.com/gostafa/distance/analyzer"
	"golang.org/x/tools/go/analysis"
)

func registerDistance() int {
	register.Plugin(analyzer.Name, func(raw any) (register.LinterPlugin, error) {
		pluginInstance, err := New(raw)
		if err != nil {
			return nil, fmt.Errorf("registerDistance: %w", err)
		}

		return pluginInstance, nil
	})

	return registerDone
}

// New constructs the Module Plugin from golangci-lint custom settings.
func New(raw any) (*Plugin, error) {
	settings, err := decodePluginSettings(raw)
	if err != nil {
		return nil, fmt.Errorf("New: %w", err)
	}

	return &Plugin{build: analyzerBuilder(&settings)}, nil
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

func decodePluginSettings(raw any) (analyzer.Settings, error) {
	if raw == nil {
		return analyzer.Settings{}, nil
	}

	data, err := json.Marshal(raw)
	if err != nil {
		return analyzer.Settings{}, fmt.Errorf("marshal settings: %w", err)
	}

	var settings analyzer.Settings

	err = analyzer.UnmarshalSettings(data, &settings)
	if err != nil {
		return analyzer.Settings{}, fmt.Errorf("decode settings: %w", err)
	}

	return settings, nil
}

// BuildAnalyzers returns the distance go/analysis Analyzer.
func (plugin *Plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	analyzers, err := runBuild(plugin.build)
	if err != nil {
		return nil, fmt.Errorf(errFmtBuildAnalyzers, err)
	}

	return analyzers, nil
}

// GetLoadMode requests type information so diagnostics can locate package
// positions accurately.
func (plugin *Plugin) GetLoadMode() string {
	return loadModeFor(plugin.build)
}

func runBuild(build func() ([]*analysis.Analyzer, error)) ([]*analysis.Analyzer, error) {
	analyzers, err := build()
	if err != nil {
		return nil, fmt.Errorf("build analyzers: %w", err)
	}

	return analyzers, nil
}

func loadModeFor(func() ([]*analysis.Analyzer, error)) string {
	return register.LoadModeTypesInfo
}
