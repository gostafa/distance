// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package cli

import (
	"bufio"
	"context"
	"flag"
	"io"
	"log/slog"
	"os"

	"github.com/gostafa/distance/distance"
	policydomain "github.com/gostafa/distance/internal/features/policy/domain"
	reportingdomain "github.com/gostafa/distance/internal/features/reporting/domain"
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
		rules           ruleList
	}

	ruleSpec struct {
		pattern string
		maximum float64
	}

	ruleList struct {
		items []ruleSpec
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
		rules           ruleList
		workers         int
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
		policy       []policydomain.Rule
		gating       bool
	}

	cliRuntime struct {
		analyze        func(context.Context, *distance.Config) (distance.Report, error)
		isTerminal     func() bool
		createHelpTemp func(string, string) (*os.File, error)
		closeHelpFile  func(*os.File) error
		writeDocs      func(io.WriteCloser, string) error
		openBrowser    func(string) error
		startCPU       func(string) (func() error, error)
		writeHeap      func(string) error
	}

	stdoutSink struct {
		w *bufio.Writer
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
