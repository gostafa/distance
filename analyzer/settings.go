package analyzer

import (
	"cmp"
	"fmt"

	"github.com/gostafa/distance/distance"
	policydomain "github.com/gostafa/distance/internal/features/policy/domain"
)

// Settings configures the distance policy analyzer. Analysis fields map to
// the distance.Config facade. Policy is a list of package-path patterns,
// each with a maximum distance. Empty Packages selects the recommended
// default: [{pattern: "./...", max-distance: 0.5}].
type Settings struct {
	// Directory is the working directory for package loading. Empty means the
	// process working directory.
	Directory string `json:"directory"`
	// Packages are first-match policy rules. Load patterns are the union of
	// the rule Pattern fields.
	Packages []policydomain.PackageRule `json:"packages"`
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
}

func (s Settings) withDefaults() Settings {
	if len(s.Packages) == 0 {
		s.Packages = []policydomain.PackageRule{{
			Pattern:     "./...",
			MaxDistance: policydomain.DefaultMaxDistance,
		}}
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
		Patterns:         policydomain.LoadPatterns(s.Packages),
		IncludeTests:     s.Tests,
		IncludeGenerated: s.Generated,
		BuildTags:        append([]string(nil), s.BuildTags...),
		Workers:          s.Workers,
		DependencyScope:  distance.DependencyScope(s.DependencyScope),
		ContinueOnError:  s.ContinueOnError,
	}
}

// policy returns the inline package-rule policy. With no rules configured,
// the recommended defaults apply. It never reads or discovers a policy file.
func (s Settings) policy() (policydomain.Policy, error) {
	policy := policydomain.Policy{Packages: s.Packages}
	if err := policydomain.Validate(policy); err != nil {
		return policydomain.Policy{}, err
	}

	return policy, nil
}
