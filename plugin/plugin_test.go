package plugin_test

import (
	"testing"

	"github.com/golangci/plugin-module-register/register"
	"github.com/gostafa/distance/analyzer"
	"github.com/gostafa/distance/plugin"
)

func TestNewBuildAnalyzersAndLoadMode(t *testing.T) {
	t.Parallel()

	p, err := plugin.New(map[string]any{
		"dependency-scope": "module",
		"packages": []any{
			map[string]any{"pattern": "./internal/...", "max-distance": 0.2},
			map[string]any{"pattern": "./...", "max-distance": 0.5},
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

	p, err := plugin.New(nil)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := p.BuildAnalyzers(); err != nil {
		t.Fatal(err)
	}
}

func TestNewRejectsUnknownSettings(t *testing.T) {
	t.Parallel()

	_, err := plugin.New(map[string]any{"not-a-real-setting": true})
	if err == nil {
		t.Fatal("expected error for unknown settings key")
	}
}

func TestNewRejectsPolicyFileSetting(t *testing.T) {
	t.Parallel()

	_, err := plugin.New(map[string]any{"config": ".modularity.yml"})
	if err == nil {
		t.Fatal("expected error for removed config file setting")
	}
}

func TestNewRejectsUnknownInlinePolicySettings(t *testing.T) {
	t.Parallel()

	if _, err := plugin.New(map[string]any{
		"package": map[string]any{"types": 12},
	}); err == nil {
		t.Fatal("expected decoding error for removed package settings")
	}
}

func TestNewRejectsAbstractnessAndInstabilitySettings(t *testing.T) {
	t.Parallel()

	for _, key := range []string{"abstractness", "instability", "metrics"} {
		if _, err := plugin.New(map[string]any{key: 0.5}); err == nil {
			t.Errorf("%s: expected error for unknown policy metric setting", key)
		}
	}

	if _, err := plugin.New(map[string]any{
		"packages": []any{
			map[string]any{
				"pattern":      "./...",
				"max-distance": 0.5,
				"abstractness": 0.3,
			},
		},
	}); err == nil {
		t.Fatal("expected error for abstractness on a package rule")
	}
}

func TestNewRejectsInvalidAnalyzerSettings(t *testing.T) {
	t.Parallel()

	p, err := plugin.New(map[string]any{"dependency-scope": "bogus"})
	if err != nil {
		t.Fatal(err)
	}

	_, err = p.BuildAnalyzers()
	if err == nil {
		t.Fatal("expected BuildAnalyzers error for invalid dependency-scope")
	}

	p, err = plugin.New(map[string]any{
		"packages": []any{
			map[string]any{"pattern": "", "max-distance": 0.5},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err = p.BuildAnalyzers(); err == nil {
		t.Fatal("expected BuildAnalyzers error for empty pattern")
	}
}
