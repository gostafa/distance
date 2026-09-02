package analyzer

import (
	"cmp"
	"fmt"

	"github.com/gostafa/distance/distance"
)

// Settings configures the distance policy analyzer. Analysis fields map to
// the distance.Config facade. Policy fields use the same package min/max
// shape as CLI thresholds, but are decoded directly from golangci-lint's
// linters.settings.custom.distance.settings block.
type Settings struct {
	// Directory is the working directory for package loading. Empty means the
	// process working directory.
	Directory string `json:"directory"`
	// Patterns are the package patterns to analyze. Empty means ["./..."].
	Patterns []string `json:"patterns"`
	// Tests includes test files and test packages.
	Tests bool `json:"tests"`
	// Generated includes files with the standard generated-code marker.
	Generated bool `json:"generated"`
	// DependencyScope is "project", "module", or "all". Empty means "module".
	DependencyScope string `json:"dependency-scope"`
	// Workers bounds analysis concurrency. Zero selects the facade default.
	Workers int `json:"workers"`
	// ContinueOnError skips packages that fail to load or type-check.
	ContinueOnError bool `json:"continue-on-error"`
	// BuildTags are extra build tags for package loading.
	BuildTags []string `json:"build-tags"`
	// Package configures package-level structural and metric limits. Nil,
	// together with nil Type, Funcs, and Metrics, selects the recommended
	// defaults.
	Package *PackageSettings `json:"package"`
	// Type configures type-level structural limits.
	Type *TypeSettings `json:"type"`
	// Funcs configures function and method detail limits.
	Funcs *FuncSettings `json:"funcs"`
	// Metrics configures legacy/global metric limits. Prefer the scoped metric
	// maps under Package for new configurations.
	Metrics map[string]LimitSettings `json:"metrics"`
}

func (s Settings) withDefaults() Settings {
	if len(s.Patterns) == 0 {
		s.Patterns = []string{"./..."}
	}

	s.DependencyScope = cmp.Or(s.DependencyScope, string(distance.DependencyScopeModule))

	return s
}

func (s Settings) validate() error {
	if err := validateDependencyScope(s.DependencyScope); err != nil {
		return err
	}

	_, err := s.policy()

	return err
}

func validateDependencyScope(value string) error {
	switch distance.DependencyScope(value) {
	case distance.DependencyScopeProject,
		distance.DependencyScopeModule,
		distance.DependencyScopeAll:
		return nil
	default:
		return fmt.Errorf(
			"invalid dependency-scope %q (want project, module, or all)",
			value,
		)
	}
}

func (s Settings) toConfig() distance.Config {
	return distance.Config{
		Directory:        s.Directory,
		Patterns:         append([]string(nil), s.Patterns...),
		IncludeTests:     s.Tests,
		IncludeGenerated: s.Generated,
		BuildTags:        append([]string(nil), s.BuildTags...),
		Workers:          s.Workers,
		DependencyScope:  distance.DependencyScope(s.DependencyScope),
		ContinueOnError:  s.ContinueOnError,
	}
}
