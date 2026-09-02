// Package cli implements the distance command, which reports each package's
// distance from the main sequence.
//
//	distance [flags] [patterns...]
//	distance ./...
//	distance ./... --format=json
//	distance --web ./...
//	distance --help --web
//
// Logs go to stderr; the report goes to stdout or --output. --help --web
// opens an illustrated guide to the reported metrics in the browser.
package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gostafa/distance/distance"
	policydomain "github.com/gostafa/distance/internal/features/policy/domain"
	reporting "github.com/gostafa/distance/internal/features/reporting/application"
	reportingdomain "github.com/gostafa/distance/internal/features/reporting/domain"
	"github.com/gostafa/distance/internal/features/reporting/ports/outbound"
	"github.com/gostafa/distance/internal/infrastructure/browser"
	"github.com/gostafa/distance/internal/infrastructure/profiling"
	"github.com/gostafa/distance/internal/infrastructure/sinks"
	"github.com/gostafa/distance/internal/shared/version"
)

// Seams for tests that need to force exit, analysis, terminal, and I/O paths.
var (
	analyze        = distance.Analyze
	isTerminal     = stdoutIsTerminal
	createHelpTemp = os.CreateTemp
	closeHelpFile  = func(f *os.File) error { return f.Close() }
	writeDocs      = reporting.WriteDocs
	openBrowser    = browser.Open
	startCPU       = profiling.StartCPU
	writeHeap      = profiling.WriteHeap
)

// Run executes distance with args and returns its process exit code.
func Run(args []string) int { return run(args) }

func run(args []string) int {
	fs := flag.NewFlagSet("distance", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: distance [flags] [patterns...]\n\n")
		fs.PrintDefaults()
		fmt.Fprintf(
			os.Stderr,
			"\nFor an illustrated guide to the reported metrics:\n  distance --help --web\n",
		)
	}

	var (
		format    = fs.String("format", "text", "report format: text, json, csv, or web")
		webReport = fs.Bool(
			"web",
			false,
			"shorthand for -format=web: write a self-contained HTML report and open it",
		)
		output  = fs.String("output", "", "write the report to this file instead of stdout")
		explain = fs.Bool(
			"explain",
			false,
			"include reasons for n/a and dropped-component metrics in the text report",
		)
		workers = fs.Int(
			"workers",
			0,
			"concurrent package workers (0 = min(GOMAXPROCS, packages))",
		)
		dependencyScope = fs.String(
			"dependency-scope",
			"module",
			"dependency scope: project, module, or all",
		)
		buildTags       = fs.String("build-tags", "", "comma-separated build tags")
		includeTests    = fs.Bool("tests", false, "include test files and test packages")
		generated       = fs.Bool("generated", false, "include generated files")
		continueOnError = fs.Bool(
			"continue-on-error",
			false,
			"skip packages that fail to load or type-check",
		)
		cpuProfile    = fs.String("cpu-profile", "", "write a CPU profile to this file")
		memoryProfile = fs.String("memory-profile", "", "write a memory profile to this file")
		showVersion   = fs.Bool("version", false, "print the version and exit")
		verbose       = fs.Bool("verbose", false, "verbose logging to stderr")
		check         = fs.Bool(
			"check",
			false,
			"enforce --max-distance and exit 3 on violations (default threshold 0.5)",
		)
		maxDistance = fs.Float64(
			"max-distance",
			policydomain.DefaultMaxDistance,
			"maximum package distance; applies to every positional pattern (implies -check)",
		)
	)
	if err := fs.Parse(args); err != nil {
		// --help --web (either order) opens the metrics guide instead of
		// failing; parsing aborts at --help, so the raw args are scanned.
		if errors.Is(err, flag.ErrHelp) && wantsWebHelp(args) {
			return runWebHelp()
		}

		return 2
	}

	if *showVersion {
		fmt.Println("distance " + version.Version)

		return 0
	}

	level := slog.LevelInfo
	if *verbose {
		level = slog.LevelDebug
	}

	logger := newLogger(level)

	if *webReport {
		if formatWasSet(fs) && *format != string(reportingdomain.FormatWeb) {
			logger.Error("conflicting flags: -web implies -format=web", "format", *format)

			return 2
		}

		*format = string(reportingdomain.FormatWeb)
	}

	reportFormat, ok := reportingdomain.ParseFormat(*format)
	if !ok {
		logger.Error("invalid format", "format", *format, "want", "text, json, csv, or web")

		return 2
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if *cpuProfile != "" {
		stopProfile, err := startCPU(*cpuProfile)
		if err != nil {
			logger.Error("cpu profiling failed", "error", err)

			return 1
		}
		defer func() {
			err := stopProfile()
			if err != nil {
				logger.Error("cpu profiling failed", "error", err)
			}
		}()
	}

	// A policy gate runs only when explicitly requested: -check or
	// -max-distance.
	gating := *check || maxDistanceWasSet(fs)

	var (
		policy       policydomain.Policy
		policySource string
	)

	if gating {
		resolved, source, err := resolvePolicy(fs.Args(), *maxDistance)
		if err != nil {
			logger.Error("policy configuration failed", "error", err)

			return 2
		}

		policy, policySource = resolved, source
	}

	config := distance.Config{
		Patterns:         fs.Args(),
		IncludeTests:     *includeTests,
		IncludeGenerated: *generated,
		BuildTags:        splitList(*buildTags),
		Workers:          *workers,
		DependencyScope:  distance.DependencyScope(*dependencyScope),
		ContinueOnError:  *continueOnError,
	}

	start := time.Now()

	report, err := analyze(ctx, config)
	if err != nil {
		logger.Error("analysis failed", "error", err)

		// A signal (Ctrl-C / SIGTERM) cancels the context; report it with
		// the conventional 130 so CI can tell cancellation from failure.
		if errors.Is(err, context.Canceled) {
			return 130
		}

		return 1
	}

	logger.Debug("analysis complete",
		"packages", len(report.Packages), "duration", time.Since(start))

	if *memoryProfile != "" {
		err := writeHeap(*memoryProfile)
		if err != nil {
			logger.Error("memory profiling failed", "error", err)

			return 1
		}
	}

	outputPath := *output

	// A web report is unreadable on stdout: default it to a well-known file.
	webToDefaultFile := outputPath == "" && reportFormat == reportingdomain.FormatWeb
	if webToDefaultFile {
		outputPath = defaultWebReportName
	}

	var sink outbound.Sink = sinks.StdoutSink{}
	if outputPath != "" {
		sink = sinks.FileSink{Path: outputPath}
	}

	textOptions := reportingdomain.TextOptions{
		Color:   outputPath == "" && os.Getenv("NO_COLOR") == "" && isTerminal(),
		Explain: *explain,
	}
	if err := reporting.Write(report, reportFormat, sink, textOptions); err != nil {
		logger.Error("writing report failed", "error", err)

		return 1
	}

	// Open the defaulted web report only when a human is watching; an
	// explicit -output signals scripting, and pipes signal CI.
	if webToDefaultFile {
		logger.Info("report written", "path", outputPath)

		if isTerminal() {
			if err := openBrowser(outputPath); err != nil {
				logger.Warn("opening the report in a browser failed", "error", err)
			}
		}
	}

	// The report is on stdout; the violation summary goes to stderr so pipes
	// and CI stay clean, and a distinct exit code lets CI gate on it. Both
	// outcomes name the policy source so a run is never a silent no-op.
	if gating {
		violations := policydomain.Evaluate(report, policy)
		if len(violations) > 0 {
			logger.Error(
				"policy check failed",
				"source",
				policySource,
				"violations",
				len(violations),
			)
			fmt.Fprint(os.Stderr, policydomain.FormatViolations(violations))

			return 3
		}

		logger.Info("policy check passed", "source", policySource)
	}

	return 0
}

// newLogger builds the CLI's stderr logger. It drops slog's time and level
// attributes so status lines read as plain "msg key=value" text: a timestamp
// and a severity level are noise for an interactive CLI whose exit code
// already signals success or failure. A nil level selects the default (info).
func newLogger(level slog.Leveler) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if len(groups) == 0 && (a.Key == slog.TimeKey || a.Key == slog.LevelKey) {
				return slog.Attr{}
			}

			return a
		},
	}

	return slog.New(slog.NewTextHandler(os.Stderr, opts))
}

// defaultWebReportName is where the web report lands when -output is unset.
const defaultWebReportName = "distance-report.html"

// formatWasSet reports whether the -format flag was given explicitly.
func formatWasSet(fs *flag.FlagSet) bool {
	set := false
	fs.Visit(func(f *flag.Flag) { set = set || f.Name == "format" })

	return set
}

// maxDistanceWasSet reports whether -max-distance was given explicitly.
func maxDistanceWasSet(fs *flag.FlagSet) bool {
	set := false
	fs.Visit(func(f *flag.Flag) { set = set || f.Name == "max-distance" })

	return set
}

// wantsWebHelp reports whether the raw arguments request the web metrics
// guide: a truthy -web / --web token before any "--" terminator. Raw args
// are scanned because --help aborts flag parsing before -web is seen.
func wantsWebHelp(args []string) bool {
	for _, arg := range args {
		if arg == "--" {
			return false
		}

		if arg == "-web" || arg == "--web" {
			return true
		}

		value, found := strings.CutPrefix(arg, "-web=")
		if !found {
			value, found = strings.CutPrefix(arg, "--web=")
		}

		if found {
			truthy, err := strconv.ParseBool(value)

			return err == nil && truthy
		}
	}

	return false
}

// runWebHelp writes the metrics guide to the OS temp dir and, when a human
// is watching, opens it in the browser. The file is always written so
// scripts and CI can read the logged path.
func runWebHelp() int {
	logger := newLogger(nil)

	path, err := writeHelpDocs()
	if err != nil {
		logger.Error("writing the metrics guide failed", "error", err)

		return 1
	}

	logger.Info("metrics guide written", "path", path)

	if isTerminal() {
		if err := openBrowser(path); err != nil {
			logger.Warn("opening the metrics guide in a browser failed", "error", err)
		}
	}

	return 0
}

// writeHelpDocs renders the metrics guide into a fresh temp file and
// returns its path.
func writeHelpDocs() (string, error) {
	file, err := createHelpTemp("", "distance-help-*.html")
	if err != nil {
		return "", err
	}

	path := file.Name()
	if err := closeHelpFile(file); err != nil {
		return "", err
	}

	if err := writeDocs(sinks.FileSink{Path: path}, version.Version); err != nil {
		return "", err
	}

	return path, nil
}

// stdoutIsTerminal reports whether stdout is a character device, so ANSI
// colors never leak into pipes or redirected files.
func stdoutIsTerminal() bool {
	info, err := os.Stdout.Stat()

	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

func splitList(list string) []string {
	if list == "" {
		return nil
	}

	parts := strings.Split(list, ",")

	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}

	return out
}
