// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package cli

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"flag"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/gostafa/distance/distance"
	policydomain "github.com/gostafa/distance/internal/features/policy/public"
)

// White-box: comma-list parsing trims blanks and empties.
func TestSplitList(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		want []string
	}{
		{"", nil},
		{"a,b,c", []string{"a", "b", "c"}},
		{" a , ,b ", []string{"a", "b"}},
	}

	for _, tt := range tests {
		if got := splitList(tt.in); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("splitList(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

// White-box: the raw-args scan behind --help --web accepts truthy -web
// tokens before a "--" terminator and nothing else.
func TestWantsWebHelp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		args []string
		want bool
	}{
		{[]string{"--help", "--web"}, true},
		{[]string{"--web", "--help"}, true},
		{[]string{"-web", "-h"}, true},
		{[]string{"--web=true", "--help"}, true},
		{[]string{"--web=false", "--help"}, false},
		{[]string{"--web=nonsense", "--help"}, false},
		{[]string{"--help"}, false},
		{[]string{"--help", "--", "--web"}, false},
		{nil, false},
	}

	for _, tt := range tests {
		if got := wantsWebHelp(tt.args); got != tt.want {
			t.Errorf("wantsWebHelp(%v) = %v, want %v", tt.args, got, tt.want)
		}
	}
}

// White-box: bad flags and formats exit with code 2 before any loading.
func TestRunUsageErrors(t *testing.T) {
	t.Parallel()

	if code := executeDefault([]string{"--format=xml"}); code != 2 {
		t.Errorf("invalid format exit = %d, want 2", code)
	}

	if code := executeDefault([]string{"--this-flag-does-not-exist"}); code != 2 {
		t.Errorf("unknown flag exit = %d, want 2", code)
	}
}

func executeDefault(args []string) int {
	runtime := defaultRuntime()

	return execute(&runtime, args)
}

func stubAnalyze(runtime *runtimeConfig) {
	runtime.analyze = func(context.Context, *distance.Config) (distance.Report, error) {
		return distance.Report{
			SchemaVersion: "6",
			ToolName:      "distance", ToolVersion: "test",
			Module: "example.com/m",
		}, nil
	}
}

func TestResolvePolicyRequiresRules(t *testing.T) {
	if _, _, err := resolvePolicy(nil); err == nil {
		t.Fatal("empty rules succeeded")
	}

	rules := []ruleSpec{{pattern: "**/internal/**", maximum: 0.2}}

	policy, source, err := resolvePolicy(rules)
	if err != nil {
		t.Fatal(err)
	}

	if source != policySourceFlagRules || len(policy) != 1 ||
		policy[0].Pattern != "**/internal/**" || policy[0].Max != 0.2 {

		t.Fatalf("rule policy = %+v source = %q", policy, source)
	}

	bad := []ruleSpec{{pattern: "**", maximum: 2}}
	if _, _, err := resolvePolicy(bad); err == nil {
		t.Fatal("out-of-range max succeeded")
	}
}

func TestParseRuleSpec(t *testing.T) {
	t.Parallel()

	got, err := parseRuleSpec("**:0.5")
	if err != nil || got.pattern != "**" || got.maximum != 0.5 {
		t.Fatalf("parseRuleSpec = %+v err = %v", got, err)
	}

	if _, err := parseRuleSpec("not-a-number"); err == nil {
		t.Fatal("missing colon succeeded")
	}

	if _, err := parseRuleSpec(":0.5"); err == nil {
		t.Fatal("empty pattern succeeded")
	}

	if _, err := parseRuleSpec("**:nope"); err == nil {
		t.Fatal("invalid number succeeded")
	}
}

func TestRuleFlagAccumulation(t *testing.T) {
	t.Parallel()

	var rules []ruleSpec

	if err := appendRule(&rules, "**:0.5"); err != nil {
		t.Fatal(err)
	}

	if err := appendRule(&rules, "**/internal/**:0.2"); err != nil {
		t.Fatal(err)
	}

	if len(rules) != 2 ||
		rules[0].pattern != "**" || rules[0].maximum != 0.5 ||
		rules[1].pattern != "**/internal/**" || rules[1].maximum != 0.2 {
		t.Fatalf("rules = %+v", rules)
	}

	if err := appendRule(&rules, "bad"); err == nil {
		t.Fatal("invalid spec succeeded")
	}
}

func TestRunEarlyErrorPaths(t *testing.T) {
	if code := executeDefault([]string{"--verbose", "--dependency-scope=nope"}); code != 1 {
		t.Fatalf("invalid dependency scope exit = %d, want 1", code)
	}

	badProfile := filepath.Join(t.TempDir(), "missing", "cpu.prof")

	if code := executeDefault([]string{"--cpu-profile=" + badProfile}); code != 1 {
		t.Fatalf("bad CPU profile exit = %d, want 1", code)
	}

	badTemp := filepath.Join(t.TempDir(), "missing")
	t.Setenv("TMPDIR", badTemp)
	t.Setenv("TMP", badTemp)

	runtime := defaultRuntime()
	if code := runWebHelp(&runtime); code != 1 {
		t.Fatalf("web help with bad temp dir exit = %d, want 1", code)
	}
}

func TestRunCanceledAnalysis(t *testing.T) {
	runtime := defaultRuntime()

	runtime.analyze = func(context.Context, *distance.Config) (distance.Report, error) {
		return distance.Report{}, context.Canceled
	}

	if code := execute(&runtime, []string{"./..."}); code != 130 {
		t.Fatalf("exit = %d, want 130", code)
	}
}

func TestRunMemoryProfileAndReportWriteErrors(t *testing.T) {
	runtime := defaultRuntime()
	stubAnalyze(&runtime)

	badHeap := filepath.Join(t.TempDir(), "missing", "heap.prof")

	if code := execute(&runtime, []string{"--memory-profile=" + badHeap, "./..."}); code != 1 {
		t.Fatalf("bad memory profile exit = %d", code)
	}

	outDir := t.TempDir()

	err := os.Chmod(outDir, 0o500)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = os.Chmod(outDir, 0o700) })

	out := filepath.Join(outDir, "report.json")

	if code := execute(&runtime, []string{"--format=json", "--output=" + out, "./..."}); code != 1 {
		t.Fatalf("unwritable output exit = %d", code)
	}
}

func TestRunWebDefaultOpensBrowser(t *testing.T) {
	runtime := defaultRuntime()
	stubAnalyze(&runtime)

	dir := t.TempDir()
	t.Chdir(dir)

	runtime.isTerminal = func() bool { return true }
	runtime.openBrowser = func(string) error { return errors.New("no browser") }

	if code := execute(&runtime, []string{"--format=web", "./..."}); code != 0 {
		t.Fatalf("web default exit = %d", code)
	}

	if _, err := os.Stat(filepath.Join(dir, defaultWebReportName)); err != nil {
		t.Fatalf("default web report missing: %v", err)
	}
}

func TestRunCPUStopProfileError(t *testing.T) {
	runtime := defaultRuntime()
	stubAnalyze(&runtime)

	runtime.startCPU = func(string) (func() error, error) {
		return func() error { return errors.New("stop failed") }, nil
	}

	if code := execute(
		&runtime,
		[]string{"--cpu-profile=" + filepath.Join(t.TempDir(), "cpu.prof"), "./..."},
	); code != 0 {
		t.Fatalf("exit = %d, want 0 (stop error is logged only)", code)
	}
}

func TestRunWebHelpTerminalBrowserWarn(t *testing.T) {
	runtime := defaultRuntime()

	runtime.isTerminal = func() bool { return true }
	runtime.openBrowser = func(string) error { return errors.New("open failed") }

	if code := runWebHelp(&runtime); code != 0 {
		t.Fatalf("runWebHelp exit = %d", code)
	}
}

func TestWriteHelpDocsCloseAndWriteErrors(t *testing.T) {
	runtime := defaultRuntime()

	runtime.closeHelpFile = func(*os.File) error { return errors.New("close failed") }

	if _, err := writeHelpDocs(&runtime); err == nil {
		t.Fatal("want close error")
	}

	runtime = defaultRuntime()
	runtime.writeDocs = func(io.WriteCloser, string) error { return errors.New("docs failed") }

	if _, err := writeHelpDocs(&runtime); err == nil {
		t.Fatal("want docs write error")
	}
}

type commandWriter struct {
	err           error
	failAt, calls int
	short         bool
}

func (w *commandWriter) Write(data []byte) (int, error) {
	w.calls++
	if w.err != nil && (w.failAt == 0 || w.failAt == w.calls) {
		return 0, w.err
	}
	if w.short {
		return len(data) / 2, nil
	}
	return len(data), nil
}

func TestCommandOutputFailures(t *testing.T) {
	failure := errors.New("output failed")
	if code := writeVersion(&commandWriter{err: failure}); code != 1 {
		t.Fatalf("version exit = %d", code)
	}
	if err := writeOutputText(
		&commandWriter{short: true},
		"output",
	); !errors.Is(
		err,
		errShortWrite,
	) {
		t.Fatalf("short output = %v", err)
	}
	for _, failAt := range []int{1, 2} {
		writer := &commandWriter{err: failure, failAt: failAt}
		flags := flag.NewFlagSet("distance", flag.ContinueOnError)
		flags.SetOutput(io.Discard)
		writeUsage(flags, writer)
		if writer.calls != failAt {
			t.Fatalf("usage writes = %d want %d", writer.calls, failAt)
		}
	}
	sink := &stdoutSink{w: bufio.NewWriter(&commandWriter{err: failure})}
	if _, err := sink.Write(bytes.Repeat([]byte("x"), 8192)); !errors.Is(err, failure) {
		t.Fatalf("stdout write = %v", err)
	}
	buffered := &stdoutSink{w: bufio.NewWriter(&commandWriter{err: failure})}
	if _, err := buffered.Write([]byte("x")); err != nil {
		t.Fatal(err)
	}
	if err := buffered.Close(); !errors.Is(err, failure) {
		t.Fatalf("stdout close = %v", err)
	}
}

func TestClosedStderrAndFile(t *testing.T) {
	closed, err := os.CreateTemp(t.TempDir(), "closed")
	if err != nil {
		t.Fatal(err)
	}
	if err := closed.Close(); err != nil {
		t.Fatal(err)
	}
	if err := closeFile(closed); !errors.Is(err, os.ErrClosed) {
		t.Fatalf("close = %v", err)
	}
	original := os.Stderr
	os.Stderr = closed
	t.Cleanup(func() { os.Stderr = original })
	if err := writeStderr("policy"); !errors.Is(err, os.ErrClosed) {
		t.Fatalf("stderr = %v", err)
	}
}

func TestBrowserCommands(t *testing.T) {
	cases := []struct {
		platform, name string
		arguments      []string
	}{
		{"darwin", "open", []string{"report.html"}},
		{"windows", "rundll32", []string{"url.dll,FileProtocolHandler", "report.html"}},
		{"linux", "xdg-open", []string{"report.html"}},
	}
	for _, test := range cases {
		command := browserCommandFor(test.platform, "report.html")
		name, arguments := command.Args[0], command.Args[1:]
		if name != test.name || !reflect.DeepEqual(arguments, test.arguments) {
			t.Fatalf("browser command = %s %v", name, arguments)
		}
	}
	t.Setenv("PATH", t.TempDir())
	if err := openBrowser("report.html"); err == nil {
		t.Fatal("missing browser command succeeded")
	}
	directory := t.TempDir()
	name, _ := browserOpenCommand("darwin", "report.html")
	if err := os.WriteFile(
		filepath.Join(directory, name),
		[]byte("#!/bin/sh\nexit 0\n"),
		0o700,
	); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", directory)
	if err := openBrowser("report.html"); err != nil {
		t.Fatal(err)
	}
}

func TestProfilingLifecycleFailures(t *testing.T) {
	first, err := startCPUProfile(filepath.Join(t.TempDir(), "first.pprof"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = first() })
	if _, err := startCPUProfile(filepath.Join(t.TempDir(), "second.pprof")); err == nil {
		t.Fatal("overlapping CPU profiling succeeded")
	}
	if err := first(); err != nil {
		t.Fatal(err)
	}
	if err := first(); !errors.Is(err, os.ErrClosed) {
		t.Fatalf("repeated stop = %v", err)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	stopCPUWithLog(func() error { return errors.New("stop failed") }, logger)()
}

func TestCommandPolicyFailure(t *testing.T) {
	session := &runSession{
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		policy: []policydomain.Rule{{Pattern: "**", Max: 2}},
	}
	if code := enforcePolicy(t.Context(), new(distance.Report), session); code != 1 {
		t.Fatalf("policy failure exit = %d", code)
	}
	runtime := defaultRuntime()
	runtime.isTerminal = func() bool { return false }
	if err := openBrowserIfTTY(&runtime, "report"); err != nil {
		t.Fatal(err)
	}
	if _, err := writeHelpDocsAt(
		&runtime,
		filepath.Join(t.TempDir(), "missing", "guide"),
	); err == nil {
		t.Fatal("missing help path succeeded")
	}
	runtime.writeDocs = func(writer io.WriteCloser, _ string) error { _ = writer.Close(); return errors.New("guide failed") }
	if _, err := writeHelpDocsAt(&runtime, filepath.Join(t.TempDir(), "guide")); err == nil {
		t.Fatal("failed guide write succeeded")
	}
	runtime.analyze = func(context.Context, *distance.Config) (distance.Report, error) { return distance.Report{}, nil }
	output := runAnalyze(
		t.Context(),
		&runtime,
		&analyzeArgs{cfg: new(distance.Config), log: session.logger},
	)
	if output.code != 0 || output.report.ToolVersion == "" {
		t.Fatalf("analysis metadata = %+v", output)
	}
}

func TestWebHelpAndHeapFailurePaths(t *testing.T) {
	if value, found := cutWebValue("-web=true"); !found || value != "true" {
		t.Fatalf("short web flag = %q found=%v", value, found)
	}
	runtime := defaultRuntime()
	path := filepath.Join(t.TempDir(), "heap.pprof")
	if err := writeHeapProfile(path); err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	runtime.writeDocs = func(writer io.WriteCloser, _ string) error { return writer.Close() }
	if got, err := writeHelpDocsAt(
		&runtime,
		filepath.Join(directory, "guide"),
	); err != nil ||
		got == "" {
		t.Fatalf("guide path = %q err=%v", got, err)
	}
}

func TestPolicySummaryOutputFailure(t *testing.T) {
	closed, err := os.CreateTemp(t.TempDir(), "stderr")
	if err != nil {
		t.Fatal(err)
	}
	if err := closed.Close(); err != nil {
		t.Fatal(err)
	}
	original := os.Stderr
	os.Stderr = closed
	t.Cleanup(func() { os.Stderr = original })
	session := &runSession{logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	violations := []policydomain.Violation{
		{Package: "example.com/m", Rule: "**", Value: 1, Threshold: 0.5},
	}
	if code := failPolicy(t.Context(), session, violations); code != 1 {
		t.Fatalf("summary failure exit = %d", code)
	}
}

type failingDestination struct{ writeErr, closeErr error }

func (d failingDestination) Write(payload []byte) (int, error) { return len(payload), d.writeErr }

func (d failingDestination) Close() error { return d.closeErr }

func TestFinalCommandFailurePaths(t *testing.T) {
	failure := errors.New("destination failed")
	for _, destination := range []failingDestination{{writeErr: failure}, {closeErr: failure}} {
		if err := writeHeapTo(destination); !errors.Is(err, failure) {
			t.Fatalf("heap failure = %v", err)
		}
	}
	runtime := defaultRuntime()
	runtime.isTerminal = func() bool { return true }
	runtime.openBrowser = func(string) error { return nil }
	if err := openBrowserIfTTY(&runtime, "report"); err != nil {
		t.Fatal(err)
	}
	runtime.writeReport = func(_ *distance.Report, destination io.WriteCloser, _ *reportOptions) error {
		_ = destination.Close()
		return failure
	}
	session := &runSession{
		opts:   &cliOptions{output: filepath.Join(t.TempDir(), "report")},
		format: "json",
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	if code := writeReport(
		t.Context(),
		&runtime,
		&reportArgs{report: new(distance.Report), session: session},
	); code != 1 {
		t.Fatalf("write failure exit = %d", code)
	}
}

func TestHeapEncoderFailure(t *testing.T) {
	failure := errors.New("heap encoder failed")
	err := writeHeapWith(
		filepath.Join(t.TempDir(), "heap"),
		func(destination io.WriteCloser) error { _ = destination.Close(); return failure },
	)
	if !errors.Is(err, failure) {
		t.Fatalf("heap encoder = %v", err)
	}
}
