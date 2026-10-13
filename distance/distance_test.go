// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package distance

import (
	"context"
	"errors"
	"testing"
)

func TestWithDefaults(t *testing.T) {
	cfg := configWithDefaults(&Config{})

	if len(cfg.Patterns) != 1 || cfg.Patterns[0] != "./..." {
		t.Fatalf("patterns = %v", cfg.Patterns)
	}

	if cfg.DependencyScope != DependencyScopeModule {
		t.Fatalf("scope = %q", cfg.DependencyScope)
	}
}

func TestValidate(t *testing.T) {
	valid := configWithDefaults(&Config{})

	err := validateConfig(valid)
	if err != nil {
		t.Fatal(err)
	}

	bad := valid

	bad.DependencyScope = "galaxy"

	err = validateConfig(bad)
	if err == nil {
		t.Fatal("invalid scope accepted")
	}

	bad = valid

	bad.Patterns = []string{""}

	err = validateConfig(bad)
	if err == nil {
		t.Fatal("empty pattern accepted")
	}
}

func TestAllMetrics(t *testing.T) {
	want := []string{MetricAbstractness, MetricInstability, MetricDistance}
	got := AllMetrics()

	if len(got) != len(want) {
		t.Fatalf("AllMetrics() = %v, want %v", got, want)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("AllMetrics() = %v, want %v", got, want)
		}
	}
}

func TestConfigurationOwnershipAndNilDefaults(t *testing.T) {
	defaults := configWithDefaults(nil)
	if defaults.DependencyScope != DependencyScopeModule || len(defaults.Patterns) != 1 {
		t.Fatalf("nil defaults: %+v", defaults)
	}
	original := Config{Patterns: []string{"./a"}, BuildTags: []string{"tag"}}
	copied := configWithDefaults(&original)
	copied.Patterns[0] = "changed"
	copied.BuildTags[0] = "changed"
	if original.Patterns[0] != "./a" || original.BuildTags[0] != "tag" {
		t.Fatal("caller-owned slices changed")
	}
	empty := Config{}
	_ = configWithDefaults(&empty)
	if empty.DependencyScope != "" || empty.Patterns != nil {
		t.Fatal("caller-owned defaults changed")
	}
}

func TestPublicErrorContracts(t *testing.T) {
	hidden := errors.New("private backend failure")
	for _, test := range []struct{ cause, want error }{{hidden, ErrAnalysis}, {context.Canceled, context.Canceled}, {context.DeadlineExceeded, context.DeadlineExceeded}} {
		err := publicAnalysisError(test.cause)
		if !errors.Is(err, ErrAnalysis) || !errors.Is(err, test.want) {
			t.Fatalf("identity lost: %v", err)
		}
		if errors.Is(err, hidden) {
			t.Fatal("private failure leaked into public chain")
		}
	}
	report, err := Analyze(t.Context(), &Config{DependencyScope: "invalid"})
	if !errors.Is(err, ErrInvalidConfig) || report.SchemaVersion != "" || report.Packages != nil {
		t.Fatalf("invalid config: report=%+v err=%v", report, err)
	}
	_, err = Analyze(t.Context(), &Config{Patterns: []string{""}})
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("invalid pattern: %v", err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = Analyze(ctx, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("nil canceled config: %v", err)
	}
}
