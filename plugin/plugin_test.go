// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package plugin

import (
	"testing"

	"github.com/golangci/plugin-module-register/register"
	analyzer "github.com/gostafa/distance/internal/features/projectanalysis/adapters/goanalysis"
)

func TestNewBuildAnalyzersAndLoadMode(t *testing.T) {
	t.Parallel()

	p, err := newPlugin(map[string]any{
		"dependency-scope": "module",
		"patterns":         []any{"./..."},
		"rules": []any{
			map[string]any{"pattern": "**", "max": 0.5},
			map[string]any{"pattern": "**/internal/**", "max": 0.2},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if mode := p.GetLoadMode(); mode != register.LoadModeTypesInfo {
		t.Fatalf("GetLoadMode = %q, want %q", mode, register.LoadModeTypesInfo)
	}

	analyzers, err := p.BuildAnalyzers()
	if err != nil {
		t.Fatal(err)
	}

	if len(analyzers) != 1 {
		t.Fatalf("len(analyzers) = %d, want 1", len(analyzers))
	}

	if analyzers[0].Name != analyzer.Name {
		t.Fatalf("Name = %q, want %q", analyzers[0].Name, analyzer.Name)
	}
}

func TestNewNilSettings(t *testing.T) {
	t.Parallel()

	p, err := newPlugin(nil)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := p.BuildAnalyzers(); err != nil {
		t.Fatal(err)
	}
}

func TestNewRejectsUnknownSettings(t *testing.T) {
	t.Parallel()

	_, err := newPlugin(map[string]any{"not-a-real-setting": true})
	if err == nil {
		t.Fatal("expected error for unknown settings key")
	}
}

func TestNewRejectsPolicyFileSetting(t *testing.T) {
	t.Parallel()

	_, err := newPlugin(map[string]any{"config": ".modularity.yml"})
	if err == nil {
		t.Fatal("expected error for removed config file setting")
	}
}

func TestNewRejectsUnknownInlinePolicySettings(t *testing.T) {
	t.Parallel()

	if _, err := newPlugin(map[string]any{
		"package": map[string]any{"types": 12},
	}); err == nil {
		t.Fatal("expected decoding error for removed package settings")
	}
}

func TestNewRejectsAbstractnessAndInstabilitySettings(t *testing.T) {
	t.Parallel()

	for _, key := range []string{"abstractness", "instability", "metrics"} {
		if _, err := newPlugin(map[string]any{key: 0.5}); err == nil {
			t.Errorf("%s: expected error for unknown policy metric setting", key)
		}
	}

	if _, err := newPlugin(map[string]any{
		"rules": []any{
			map[string]any{
				"pattern":      "**",
				"max":          0.5,
				"abstractness": 0.3,
			},
		},
	}); err == nil {
		t.Fatal("expected error for abstractness on a rule")
	}
}

func TestNewRejectsLegacyPackagesAndMaxDistance(t *testing.T) {
	t.Parallel()

	if _, err := newPlugin(map[string]any{
		"packages": []any{
			map[string]any{"pattern": "./...", "max-distance": 0.5},
		},
	}); err == nil {
		t.Fatal("expected error for removed packages setting")
	}

	if _, err := newPlugin(map[string]any{
		"rules": []any{
			map[string]any{"pattern": "**", "max-distance": 0.5},
		},
	}); err == nil {
		t.Fatal("expected error for removed max-distance key")
	}
}

func TestNewRejectsInvalidAnalyzerSettings(t *testing.T) {
	t.Parallel()

	p, err := newPlugin(map[string]any{"dependency-scope": "bogus"})
	if err != nil {
		t.Fatal(err)
	}

	_, err = p.BuildAnalyzers()
	if err == nil {
		t.Fatal("expected BuildAnalyzers error for invalid dependency-scope")
	}

	p, err = newPlugin(map[string]any{
		"rules": []any{
			map[string]any{"pattern": "", "max": 0.5},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err = p.BuildAnalyzers(); err == nil {
		t.Fatal("expected BuildAnalyzers error for empty pattern")
	}
}

func TestRegisteredFactory(t *testing.T) {
	factory, err := register.GetPlugin(analyzer.Name)
	if err != nil {
		t.Fatal(err)
	}
	instance, err := factory(nil)
	if err != nil {
		t.Fatal(err)
	}
	if instance.GetLoadMode() != register.LoadModeTypesInfo {
		t.Fatal("registration load mode")
	}
	if _, err := instance.BuildAnalyzers(); err != nil {
		t.Fatal(err)
	}
	if _, err := factory(map[string]any{"unknown": true}); err == nil {
		t.Fatal("registered factory accepted unknown settings")
	}
	if _, err := factory(make(chan int)); err == nil {
		t.Fatal("registered factory accepted unserializable settings")
	}
}
