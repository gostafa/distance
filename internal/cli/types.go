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
	policydomain "github.com/gostafa/distance/internal/features/policy/public"
	reporting "github.com/gostafa/distance/internal/features/reporting/public"
)

type (
	flagBindings = struct {
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
		rules           []ruleSpec
	}

	ruleSpec = struct {
		pattern string
		maximum float64
	}

	cliOptions = struct {
		flagSet         *flag.FlagSet
		cpuProfile      string
		output          string
		dependencyScope string
		buildTags       string
		format          string
		memoryProfile   string
		patterns        []string
		rules           []ruleSpec
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

	runSession = struct {
		opts         *cliOptions
		logger       *slog.Logger
		format       string
		policySource string
		policy       []policydomain.Rule
		gating       bool
	}

	runtimeConfig = struct {
		analyze        func(context.Context, *distance.Config) (distance.Report, error)
		isTerminal     func() bool
		createHelpTemp func(string, string) (*os.File, error)
		closeHelpFile  func(*os.File) error
		writeReport    func(*distance.Report, io.WriteCloser, *reportOptions) error
		writeDocs      func(io.WriteCloser, string) error
		openBrowser    func(string) error
		startCPU       func(string) (func() error, error)
		writeHeap      func(string) error
	}

	stdoutSink struct {
		w *bufio.Writer
	}

	reportSink = struct {
		io.WriteCloser
	}

	analyzeArgs = struct {
		cfg *distance.Config
		log *slog.Logger
	}

	analyzeOut = struct {
		report distance.Report
		code   int
	}

	formatOut = struct {
		format string
		code   int
	}

	reportArgs = struct {
		report  *distance.Report
		session *runSession
	}

	openArgs struct {
		log   *slog.Logger
		path  string
		guide bool
	}

	heapArgs = struct {
		log  *slog.Logger
		path string
	}

	textOptionsValue = struct {
		Color   bool
		Explain bool
	}

	reportOptions = struct {
		Format string
		Text   reporting.TextOptions
	}
)
