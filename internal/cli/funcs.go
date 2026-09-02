// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package cli

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gostafa/distance/distance"
	"github.com/gostafa/distance/distance/wire"
	policydomain "github.com/gostafa/distance/internal/features/policy/domain"
	reporting "github.com/gostafa/distance/internal/features/reporting/application"
	reportingdomain "github.com/gostafa/distance/internal/features/reporting/domain"
	"github.com/gostafa/distance/internal/shared/version"
)

// Run is the distance CLI entrypoint; it returns a process exit code.
func Run(args []string) int {
	runtime := defaultRuntime()

	return runtime.execute(args)
}

func (runtime *cliRuntime) announceWeb(ctx context.Context, logger *slog.Logger, path string) {
	logger.InfoContext(ctx, msgReportWritten, slog.String(keyPath, path))
	runtime.maybeOpen(ctx, &openArgs{log: logger, path: path})
}

func attachPolicy(ctx context.Context, session *runSession) int {
	session.gating = session.opts.check

	if !session.gating {
		return zero
	}

	return fillPolicy(ctx, session)
}

func (runtime *cliRuntime) execute(args []string) int {
	opts, code := parseCLI(args)

	if opts == nil {
		return code
	}

	if opts.showVersion {
		return handleVersion()
	}

	return runtime.runWithOptions(opts)
}

func fillPolicy(ctx context.Context, session *runSession) int {
	policy, source, policyErr := resolvePolicy(session.opts.rules)
	if policyErr != nil {
		session.logger.ErrorContext(
			ctx,
			msgPolicyFailed,
			slog.String(keyError, policyErr.Error()),
		)

		return exitUsage
	}

	session.policy = policy
	session.policySource = source

	return zero
}

func (runtime *cliRuntime) finishReport(ctx context.Context, args *reportArgs) int {
	heapCode := runtime.maybeWriteHeap(ctx, args.session.opts.memoryProfile, args.session.logger)

	if heapCode != zero {
		return heapCode
	}

	writeCode := runtime.writeReport(ctx, args)

	if writeCode != zero {
		return writeCode
	}

	return policyExit(ctx, args.report, args.session)
}

func (runtime *cliRuntime) maybeOpen(ctx context.Context, args *openArgs) {
	openErr := runtime.openBrowserIfTTY(args.path)
	if openErr == nil {
		return
	}

	if args.guide {
		args.log.WarnContext(ctx, msgOpenGuideFailed, slog.String(keyError, openErr.Error()))

		return
	}

	args.log.WarnContext(ctx, msgOpenReportFailed, slog.String(keyError, openErr.Error()))
}

func (runtime *cliRuntime) openBrowserIfTTY(path string) error {
	if !runtime.isTerminal() {
		return nil
	}

	openErr := runtime.openBrowser(path)
	if openErr != nil {
		return fmt.Errorf("cli open browser: %w", openErr)
	}

	return nil
}

func (runtime *cliRuntime) maybeStartCPU(path string, logger *slog.Logger) (stop func(), code int) {
	if path == emptyString {
		return noopStop, zero
	}

	stopProfile, startErr := runtime.startCPU(path)
	if startErr != nil {
		logger.ErrorContext(
			context.Background(),
			msgCPUFailed,
			slog.String(keyError, startErr.Error()),
		)

		return noopStop, one
	}

	return stopCPUWithLog(stopProfile, logger), zero
}

func (runtime *cliRuntime) maybeWriteHeap(ctx context.Context, path string, log *slog.Logger) int {
	if path == emptyString {
		return zero
	}

	heapErr := runtime.writeHeap(path)
	if heapErr != nil {
		log.ErrorContext(ctx, msgHeapFailed, slog.String(keyError, heapErr.Error()))

		return one
	}

	return zero
}

func policyExit(ctx context.Context, report *distance.Report, session *runSession) int {
	if !session.gating {
		return zero
	}

	return enforcePolicy(ctx, report, session)
}

func (runtime *cliRuntime) runJob(session *runSession) int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	defer stop()

	stopCPU, cpuCode := runtime.maybeStartCPU(session.opts.cpuProfile, session.logger)

	if cpuCode != zero {
		return cpuCode
	}

	defer stopCPU()

	return runtime.analyzeAndReport(ctx, session)
}

func (runtime *cliRuntime) analyzeAndReport(ctx context.Context, session *runSession) int {
	policyCode := attachPolicy(ctx, session)

	if policyCode != zero {
		return policyCode
	}

	out := runtime.runAnalyze(ctx, &analyzeArgs{
		cfg: buildAnalyzeConfig(session.opts),
		log: session.logger,
	})

	if out.code != zero {
		return out.code
	}

	return runtime.finishReport(ctx, &reportArgs{report: &out.report, session: session})
}

func (runtime *cliRuntime) runAnalyze(ctx context.Context, args *analyzeArgs) analyzeOut {
	start := time.Now()

	rep, analyzeErr := runtime.analyze(ctx, args.cfg)
	if analyzeErr != nil {
		return analyzeOut{report: rep, code: analyzeExitCode(ctx, args.log, analyzeErr)}
	}

	if rep.ToolVersion == "" {
		rep.ToolVersion = version.Version()
	}

	args.log.DebugContext(
		ctx,
		msgAnalysisComplete,
		slog.Int(keyPackages, len(rep.Packages)),
		slog.Duration(keyDuration, time.Since(start)),
	)

	return analyzeOut{report: rep}
}

func (runtime *cliRuntime) runWebHelp() int {
	logger := newLogger(nil)
	ctx := context.Background()

	path, helpErr := runtime.writeHelpDocs()
	if helpErr != nil {
		logger.ErrorContext(ctx, msgGuideFailed, slog.String(keyError, helpErr.Error()))

		return one
	}

	logger.InfoContext(ctx, msgGuideWritten, slog.String(keyPath, path))
	runtime.maybeOpen(ctx, &openArgs{log: logger, path: path, guide: true})

	return zero
}

func (runtime *cliRuntime) runWithOptions(opts *cliOptions) int {
	logger := newLogger(logLevel(opts))
	resolved := resolveReportFormat(opts, logger)

	if resolved.code != zero {
		return resolved.code
	}

	return runtime.runJob(&runSession{opts: opts, format: resolved.format, logger: logger})
}

func (runtime *cliRuntime) writeHelpDocs() (string, error) {
	file, createErr := runtime.createHelpTemp(emptyString, helpTempPattern)
	if createErr != nil {
		return emptyString, fmt.Errorf(errWrapWriteHelp, createErr)
	}

	path, finishErr := runtime.finishHelpDocs(file)
	if finishErr != nil {
		return emptyString, fmt.Errorf(errWrapWriteHelp, finishErr)
	}

	return path, nil
}

func (runtime *cliRuntime) finishHelpDocs(file *os.File) (string, error) {
	path := file.Name()

	closeErr := runtime.closeHelpFile(file)
	if closeErr != nil {
		return emptyString, fmt.Errorf(errWrapWriteHelp, closeErr)
	}

	writer, openErr := openOutput(path)
	if openErr != nil {
		return emptyString, fmt.Errorf(errWrapWriteHelp, openErr)
	}

	docsErr := runtime.writeDocs(writer, version.Version())
	if docsErr != nil {
		return emptyString, fmt.Errorf(errWrapWriteHelp, docsErr)
	}

	return path, nil
}

func (runtime *cliRuntime) writeReport(ctx context.Context, args *reportArgs) int {
	outputPath, webDefault := resolveOutputPath(args.session.opts.output, args.session.format)

	writer, openErr := openOutput(outputPath)
	if openErr != nil {
		return failWrite(ctx, args.session, openErr)
	}

	writeErr := reporting.Write(args.report, writer, &reporting.WriteOptions{
		Format: args.session.format,
		Text:   textOptions(outputPath, args.session, runtime.isTerminal),
	})
	if writeErr != nil {
		return failWrite(ctx, args.session, writeErr)
	}

	if webDefault {
		runtime.announceWeb(ctx, args.session.logger, outputPath)
	}

	return zero
}

func failWrite(ctx context.Context, session *runSession, writeErr error) int {
	session.logger.ErrorContext(ctx, msgWriteFailed, slog.String(keyError, writeErr.Error()))

	return one
}

func analyzeExitCode(ctx context.Context, logger *slog.Logger, analyzeErr error) int {
	logger.ErrorContext(ctx, msgAnalysisFailed, slog.String(keyError, analyzeErr.Error()))

	if errors.Is(analyzeErr, context.Canceled) {
		return exitInterrupted
	}

	return one
}

func applyWebFlag(opts *cliOptions, logger *slog.Logger) int {
	if !opts.webReport {
		return zero
	}

	return applyWebFormat(opts, logger)
}

func applyWebFormat(opts *cliOptions, logger *slog.Logger) int {
	conflict := flagWasSet(opts.flagSet, keyFormat) &&
		opts.format != string(reportingdomain.FormatWeb)

	if conflict {
		logger.ErrorContext(
			context.Background(),
			msgConflictWeb,
			slog.String(keyFormat, opts.format),
		)

		return exitUsage
	}

	opts.format = string(reportingdomain.FormatWeb)

	return zero
}

func bindAnalysisFlags(flagSet *flag.FlagSet, bindings *flagBindings) {
	bindings.workers = flagSet.Int(
		"workers",
		zero,
		"concurrent package workers (0 = min(GOMAXPROCS, packages))",
	)
	bindings.dependencyScope = flagSet.String(
		"dependency-scope",
		"module",
		"dependency scope: project, module, or all",
	)
	bindings.buildTags = flagSet.String("build-tags", emptyString, "comma-separated build tags")
	bindings.includeTests = flagSet.Bool("tests", false, "include test files and test packages")
	bindings.generated = flagSet.Bool("generated", false, "include generated files")
	bindings.continueOnError = flagSet.Bool(
		"continue-on-error",
		false,
		"skip packages that fail to load or type-check",
	)
}

func bindFlags(flagSet *flag.FlagSet) *flagBindings {
	bindings := &flagBindings{}

	bindOutputFlags(flagSet, bindings)
	bindAnalysisFlags(flagSet, bindings)
	bindProfileFlags(flagSet, bindings)
	bindPolicyFlags(flagSet, bindings)

	return bindings
}

func bindOutputFlags(flagSet *flag.FlagSet, bindings *flagBindings) {
	bindings.format = flagSet.String(
		keyFormat,
		"text",
		"report format: text, json, csv, or web",
	)
	bindings.webReport = flagSet.Bool(
		"web",
		false,
		"shorthand for -format=web: write a self-contained HTML report and open it",
	)
	bindings.output = flagSet.String(
		"output",
		emptyString,
		"write the report to this file instead of stdout",
	)
	bindings.explain = flagSet.Bool(
		"explain",
		false,
		"include reasons for n/a and dropped-component metrics in the text report",
	)
}

func bindPolicyFlags(flagSet *flag.FlagSet, bindings *flagBindings) {
	bindings.showVersion = flagSet.Bool("version", false, "print the version and exit")
	bindings.verbose = flagSet.Bool("verbose", false, "verbose logging to stderr")
	bindings.check = flagSet.Bool(
		"check",
		false,
		"enforce -rule thresholds and exit 3 on violations",
	)
	flagSet.Var(
		&bindings.rules,
		flagNameRule,
		"policy rule pattern:max (repeatable; e.g. '**/internal/**':0.2; requires -check)",
	)
}

func bindProfileFlags(flagSet *flag.FlagSet, bindings *flagBindings) {
	bindings.cpuProfile = flagSet.String(
		"cpu-profile",
		emptyString,
		"write a CPU profile to this file",
	)
	bindings.memoryProfile = flagSet.String(
		"memory-profile",
		emptyString,
		"write a memory profile to this file",
	)
}

func buildAnalyzeConfig(opts *cliOptions) *distance.Config {
	return &distance.Config{
		Patterns:         opts.patterns,
		IncludeTests:     opts.includeTests,
		IncludeGenerated: opts.generated,
		BuildTags:        splitList(opts.buildTags),
		Workers:          opts.workers,
		DependencyScope:  distance.DependencyScope(opts.dependencyScope),
		ContinueOnError:  opts.continueOnError,
	}
}

func defaultRuntime() cliRuntime {
	return cliRuntime{
		analyze:        wire.AnalyzeWithDefault,
		isTerminal:     stdoutIsTerminal,
		createHelpTemp: os.CreateTemp,
		closeHelpFile:  closeFile,
		writeDocs:      reporting.WriteDocs,
		openBrowser:    openBrowser,
		startCPU:       startCPUProfile,
		writeHeap:      writeHeapProfile,
	}
}

func closeFile(file *os.File) error {
	closeErr := file.Close()
	if closeErr != nil {
		return fmt.Errorf("cli close help file: %w", closeErr)
	}

	return nil
}

func enforcePolicy(ctx context.Context, report *distance.Report, session *runSession) int {
	violations := policydomain.Evaluate(report, session.policy)

	if len(violations) == zero {
		session.logger.InfoContext(
			ctx,
			msgPolicySucceeded,
			slog.String(keyPolicyOrigin, session.policySource),
		)

		return zero
	}

	return failPolicy(ctx, session, violations)
}

func failPolicy(ctx context.Context, session *runSession, violations []policydomain.Violation) int {
	session.logger.ErrorContext(
		ctx,
		msgPolicyCheckFail,
		slog.String(keyPolicyOrigin, session.policySource),
		slog.Int(keyViolations, len(violations)),
	)

	writeErr := writeStderr(policydomain.FormatViolations(violations))
	if writeErr != nil {
		return one
	}

	return exitPolicy
}

func flagWasSet(flagSet *flag.FlagSet, name string) bool {
	set := false

	flagSet.Visit(func(item *flag.Flag) { set = set || item.Name == name })

	return set
}

func handleParseError(parseErr error, args []string) int {
	if errors.Is(parseErr, flag.ErrHelp) && wantsWebHelp(args) {
		runtime := defaultRuntime()

		return runtime.runWebHelp()
	}

	return exitUsage
}

func handleVersion() int {
	written, writeErr := fmt.Fprintln(os.Stdout, versionPrefix+version.Version())
	if writeErr != nil {
		return one
	}

	if written == zero {
		return one
	}

	return zero
}

func isWebFlag(arg string) bool {
	if arg == webFlagShort || arg == webFlagLong {
		return true
	}

	return webFlagTruthy(arg)
}

func logLevel(opts *cliOptions) slog.Level {
	if opts.verbose {
		return slog.LevelDebug
	}

	return slog.LevelInfo
}

func newLogger(level slog.Leveler) *slog.Logger {
	opts := &slog.HandlerOptions{Level: level}

	return slog.New(slog.NewTextHandler(os.Stderr, opts))
}

func noopStop() {}

func resolveOutputPath(outputPath string, format reportingdomain.Format) (string, bool) {
	webDefault := outputPath == emptyString && format == reportingdomain.FormatWeb

	if webDefault {
		outputPath = defaultWebReportName
	}

	return outputPath, webDefault
}

func openOutput(path string) (io.WriteCloser, error) {
	if path == emptyString {
		return &stdoutSink{w: bufio.NewWriter(os.Stdout)}, nil
	}

	file, createErr := os.Create(path)
	if createErr != nil {
		return nil, fmt.Errorf("create report file: %w", createErr)
	}

	return file, nil
}

func (stream *stdoutSink) Close() error {
	flushErr := stream.w.Flush()
	if flushErr != nil {
		return fmt.Errorf("stdout flush: %w", flushErr)
	}

	return nil
}

func (stream *stdoutSink) Write(p []byte) (int, error) {
	count, writeErr := stream.w.Write(p)
	if writeErr != nil {
		return count, fmt.Errorf("stdout write: %w", writeErr)
	}

	return count, nil
}

func openBrowser(path string) error {
	name, args := browserOpenCommand(runtime.GOOS, path)

	cmd := &exec.Cmd{
		Path: name,
		Args: append([]string{name}, args...),
	}

	resolved, lookErr := exec.LookPath(name)
	if lookErr == nil {
		cmd.Path = resolved
	}

	startErr := cmd.Start()
	if startErr != nil {
		return fmt.Errorf("open %s in browser: %w", path, startErr)
	}

	return nil
}

func browserOpenCommand(goos, path string) (name string, args []string) {
	switch goos {
	case "darwin":
		return "open", []string{path}
	case "windows":
		return "rundll32", []string{"url.dll,FileProtocolHandler", path}
	default:
		return "xdg-open", []string{path}
	}
}

func startCPUProfile(path string) (stop func() error, err error) {
	file, err := createProfileFile(path)
	if err != nil {
		return nil, fmt.Errorf("profiling startCPU: %w", err)
	}

	startErr := pprof.StartCPUProfile(file)
	if startErr != nil {
		closeErr := file.Close()

		return nil, fmt.Errorf("profiling startCPU: %w", errors.Join(startErr, closeErr))
	}

	return func() error {
		pprof.StopCPUProfile()

		closeErr := file.Close()
		if closeErr != nil {
			return fmt.Errorf("profiling stopCPU: %w", closeErr)
		}

		return nil
	}, nil
}

func writeHeapProfile(path string) error {
	file, err := createProfileFile(path)
	if err != nil {
		return fmt.Errorf("profiling writeHeap: %w", err)
	}

	writeErr := pprof.WriteHeapProfile(file)
	closeErr := file.Close()

	if writeErr != nil {
		return fmt.Errorf("profiling writeHeap: %w", writeErr)
	}

	if closeErr != nil {
		return fmt.Errorf("profiling writeHeap: %w", closeErr)
	}

	return nil
}

func createProfileFile(path string) (*os.File, error) {
	dir, name := filepath.Split(path)

	if dir == emptyString {
		dir = "."
	}

	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, fmt.Errorf("open profile directory: %w", err)
	}

	file, createErr := root.Create(name)
	rootCloseErr := root.Close()

	if createErr != nil {
		return nil, fmt.Errorf("create profile file: %w", createErr)
	}

	if rootCloseErr != nil {
		discardErr := file.Close()

		return nil, fmt.Errorf("create profile file: %w", errors.Join(rootCloseErr, discardErr))
	}

	return file, nil
}

func optionsFrom(flagSet *flag.FlagSet, bindings *flagBindings) *cliOptions {
	return &cliOptions{
		format:          *bindings.format,
		webReport:       *bindings.webReport,
		output:          *bindings.output,
		explain:         *bindings.explain,
		workers:         *bindings.workers,
		dependencyScope: *bindings.dependencyScope,
		buildTags:       *bindings.buildTags,
		includeTests:    *bindings.includeTests,
		generated:       *bindings.generated,
		continueOnError: *bindings.continueOnError,
		cpuProfile:      *bindings.cpuProfile,
		memoryProfile:   *bindings.memoryProfile,
		showVersion:     *bindings.showVersion,
		verbose:         *bindings.verbose,
		check:           *bindings.check,
		rules:           bindings.rules,
		patterns:        flagSet.Args(),
		flagSet:         flagSet,
	}
}

func parseCLI(args []string) (opts *cliOptions, code int) {
	flagSet := flag.NewFlagSet(toolName, flag.ContinueOnError)
	flagSet.SetOutput(os.Stderr)

	flagSet.Usage = usagePrinter(flagSet)

	bindings := bindFlags(flagSet)

	parseErr := flagSet.Parse(args)
	if parseErr != nil {
		return nil, handleParseError(parseErr, args)
	}

	return optionsFrom(flagSet, bindings), zero
}

func parseRuleSpec(value string) (ruleSpec, error) {
	pattern, number, ok := strings.Cut(value, ":")

	if !ok {
		return ruleSpec{}, fmt.Errorf("%w: got %q", errExpectedPatternMax, value)
	}

	pattern = strings.TrimSpace(pattern)

	if pattern == emptyString {
		return ruleSpec{}, fmt.Errorf("%w: %q", errEmptyPattern, value)
	}

	parsed, err := strconv.ParseFloat(strings.TrimSpace(number), floatBits)
	if err != nil {
		return ruleSpec{}, fmt.Errorf("invalid number in %q: %w", value, err)
	}

	return ruleSpec{pattern: pattern, maximum: parsed}, nil
}

func policyDomainRules(specs []ruleSpec) []policydomain.Rule {
	rules := make([]policydomain.Rule, zero, len(specs))

	for i := range specs {
		rules = append(rules, policydomain.Rule{Pattern: specs[i].pattern, Max: specs[i].maximum})
	}

	return rules
}

func resolvePolicy(rules ruleList) ([]policydomain.Rule, string, error) {
	if len(rules.items) == zero {
		return nil, emptyString, errNoPolicyRules
	}

	domainRules := policyDomainRules(rules.items)

	validateErr := policydomain.Validate(domainRules)
	if validateErr != nil {
		return nil, emptyString, fmt.Errorf(errWrapPolicy, validateErr)
	}

	return domainRules, policySourceFlagRules, nil
}

func (list *ruleList) Set(value string) error {
	spec, err := parseRuleSpec(value)
	if err != nil {
		return fmt.Errorf("set rule: %w", err)
	}

	list.items = append(list.items, spec)

	return nil
}

func (list *ruleList) String() string {
	parts := make([]string, zero, len(list.items))

	for i := range list.items {
		parts = append(parts, list.items[i].pattern+":"+
			strconv.FormatFloat(list.items[i].maximum, 'g', -one, floatBits))
	}

	return strings.Join(parts, commaSep)
}

func resolveReportFormat(opts *cliOptions, log *slog.Logger) formatOut {
	webCode := applyWebFlag(opts, log)

	if webCode != zero {
		return formatOut{code: webCode}
	}

	return parseReportFormat(opts, log)
}

func parseReportFormat(opts *cliOptions, log *slog.Logger) formatOut {
	reportFormat, ok := reportingdomain.ParseFormat(opts.format)

	if !ok {
		log.ErrorContext(
			context.Background(),
			msgInvalidFormat,
			slog.String(keyFormat, opts.format),
			slog.String(keyWant, wantFormats),
		)

		return formatOut{code: exitUsage}
	}

	return formatOut{format: reportFormat}
}

func splitList(list string) []string {
	if list == emptyString {
		return nil
	}

	return splitNonEmpty(strings.Split(list, ","))
}

func splitNonEmpty(parts []string) []string {
	out := make([]string, zero, len(parts))

	for i := range parts {
		part := strings.TrimSpace(parts[i])

		if part != emptyString {
			out = append(out, part)
		}
	}

	return out
}

func stdoutIsTerminal() bool {
	info, statErr := os.Stdout.Stat()

	return statErr == nil && info.Mode()&os.ModeCharDevice != zero
}

func stopCPUWithLog(stopProfile func() error, logger *slog.Logger) func() {
	return func() {
		stopErr := stopProfile()
		if stopErr != nil {
			logger.ErrorContext(
				context.Background(),
				msgCPUFailed,
				slog.String(keyError, stopErr.Error()),
			)
		}
	}
}

func textOptions(path string, ses *runSession, term func() bool) reportingdomain.TextOptions {
	color := path == emptyString && os.Getenv("NO_COLOR") == emptyString && term()

	return reportingdomain.TextOptions{
		Color:   color,
		Explain: ses.opts.explain,
	}
}

func usagePrinter(flagSet *flag.FlagSet) func() {
	return func() { printUsage(flagSet) }
}

func printUsage(flagSet *flag.FlagSet) {
	headerErr := writeStderr(usageHeader)
	if headerErr != nil {
		return
	}

	flagSet.PrintDefaults()

	hintErr := writeStderr(usageWebHint)
	if hintErr != nil {
		return
	}
}

func wantsWebHelp(args []string) bool {
	for i := range args {
		if args[i] == argTerminator {
			return false
		}

		if isWebFlag(args[i]) {
			return true
		}
	}

	return false
}

func webFlagTruthy(arg string) bool {
	value, found := cutWebValue(arg)

	if !found {
		return false
	}

	truthy, parseErr := strconv.ParseBool(value)

	return parseErr == nil && truthy
}

func cutWebValue(arg string) (string, bool) {
	value, found := strings.CutPrefix(arg, webFlagShortEq)

	if found {
		return value, true
	}

	return strings.CutPrefix(arg, webFlagLongEq)
}

func writeStderr(text string) error {
	written, writeErr := fmt.Fprint(os.Stderr, text)
	if writeErr != nil {
		return fmt.Errorf(errWrapStderr, writeErr)
	}

	if written != len(text) {
		return fmt.Errorf(errWrapStderr, errShortWrite)
	}

	return nil
}
