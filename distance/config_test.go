package distance

import "testing"

func TestWithDefaults(t *testing.T) {
	cfg := configWithDefaults(Config{})
	if len(cfg.Patterns) != 1 || cfg.Patterns[0] != "./..." {
		t.Fatalf("patterns = %v", cfg.Patterns)
	}

	if cfg.DependencyScope != DependencyScopeModule {
		t.Fatalf("scope = %q", cfg.DependencyScope)
	}
}

func TestValidate(t *testing.T) {
	valid := configWithDefaults(Config{})
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
	got := AllMetrics()
	if len(got) != 1 || got[0] != MetricDistance {
		t.Fatalf("AllMetrics() = %v, want [%s]", got, MetricDistance)
	}

	if def := DefaultMetrics(); len(def) != 1 || def[0] != MetricDistance {
		t.Fatalf("DefaultMetrics() = %v", def)
	}
}
