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
	reportAnalyzer interface {
		// Analyze evaluates one distance configuration.
		Analyze(ctx context.Context, cfg *distance.Config) (distance.Report, error)
	}

	analyzeFunc func(ctx context.Context, cfg *distance.Config) (distance.Report, error)

	pkgViolations map[string][]policydomain.Violation

	runner struct {
		analyzer reportAnalyzer
		err      error
		byPkg    pkgViolations
		settings *Settings
		once     sync.Once
	}

	runResult struct{}

	scopeError struct {
		value string
	}

	unknownFieldError struct {
		key string
	}

	// Settings is the golangci-lint configuration block for the distance analyzer.
	Settings struct {
		Directory       string                     `json:"directory"`
		DependencyScope string                     `json:"dependency_scope"`
		Packages        []policydomain.PackageRule `json:"packages"`
		BuildTags       []string                   `json:"build_tags"`
		Workers         int                        `json:"workers"`
		Tests           bool                       `json:"tests"`
		Generated       bool                       `json:"generated"`
		ContinueOnError bool                       `json:"continue_on_error"`
	}
)
