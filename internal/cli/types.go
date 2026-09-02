// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package cli

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"github.com/gostafa/distance/distance"
	policydomain "github.com/gostafa/distance/internal/features/policy/domain"
	reportingdomain "github.com/gostafa/distance/internal/features/reporting/domain"
	"github.com/gostafa/distance/internal/features/reporting/ports/outbound"
)

type (
	flagBindings struct {
		format          *string
		webReport       *bool
		output          *string
		explain         *bool
		workers         *int
		dependencyScope *string
		buildTags       *string
		includeTests    *bool
		generated       *bool
		continueOnError *bool
		cpuProfile      *string
		memoryProfile   *string
		showVersion     *bool
		verbose         *bool
		check           *bool
		maxDistance     *float64
	}

	cliOptions struct {
		flagSet         *flag.FlagSet
		cpuProfile      string
		output          string
		dependencyScope string
		buildTags       string
		format          string
		memoryProfile   string
		patterns        []string
		workers         int
		maxDistance     float64
		generated       bool
		continueOnError bool
		showVersion     bool
		verbose         bool
		check           bool
		includeTests    bool
		explain         bool
		webReport       bool
	}

	runSession struct {
		opts         *cliOptions
		logger       *slog.Logger
		format       reportingdomain.Format
		policySource string
		policy       policydomain.Policy
		gating       bool
	}

	cliRuntime struct {
		analyze        func(context.Context, *distance.Config) (distance.Report, error)
		isTerminal     func() bool
		createHelpTemp func(string, string) (*os.File, error)
		closeHelpFile  func(*os.File) error
		writeDocs      func(outbound.Sink, string) error
		openBrowser    func(string) error
		startCPU       func(string) (func() error, error)
		writeHeap      func(string) error
	}

	sinkFactory struct {
		path string
	}

	analyzeArgs struct {
		cfg *distance.Config
		log *slog.Logger
	}

	analyzeOut struct {
		report distance.Report
		code   int
	}

	formatOut struct {
		format reportingdomain.Format
		code   int
	}

	reportArgs struct {
		report  *distance.Report
		session *runSession
	}

	openArgs struct {
		log   *slog.Logger
		path  string
		guide bool
	}
)
