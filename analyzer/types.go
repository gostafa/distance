// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package analyzer

import (
	"context"
	"sync"

	"github.com/gostafa/distance/distance"
	policydomain "github.com/gostafa/distance/internal/features/policy/domain"
)

type (
	reportAnalyzer = func(ctx context.Context, cfg *distance.Config) (distance.Report, error)

	pkgViolations = map[string][]policydomain.Violation

	runner = struct {
		analyzer reportAnalyzer
		err      error
		byPkg    pkgViolations
		settings Settings
		once     sync.Once
	}

	runResult struct{}

	scopeError struct {
		value string
	}

	// RuleSettings is one inline policy rule in analyzer settings.
	RuleSettings struct {
		// Max is the exclusive maximum distance for matching packages.
		Max *float64 `json:"max"`
		// Pattern is a package-path glob (* one segment, ** any depth).
		Pattern string `json:"pattern"`
	}

	// Settings configures the golangci-lint / go/analysis adapter.
	Settings = struct {
		// Directory is the module root used for analysis (empty = cwd).
		Directory string `json:"directory"`
		// DependencyScope selects which imports count toward coupling.
		DependencyScope string `json:"dependency_scope"`
		// Patterns are package patterns passed to the loader (default ./...).
		Patterns []string `json:"patterns"`
		// BuildTags are extra build tags for package loading.
		BuildTags []string `json:"build_tags"`
		// Rules are inline policy thresholds; empty uses recommended defaults.
		Rules []RuleSettings `json:"rules"`
		// Workers is the parallel worker count (0 = runtime default).
		Workers int `json:"workers"`
		// Tests includes _test.go packages when true.
		Tests bool `json:"tests"`
		// Generated includes generated files when true.
		Generated bool `json:"generated"`
		// ContinueOnError keeps analyzing after package load failures.
		ContinueOnError bool `json:"continue_on_error"`
	}
)
