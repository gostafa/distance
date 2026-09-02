// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package cli

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gostafa/distance/distance"
	"github.com/gostafa/distance/internal/features/reporting/ports/outbound"
)

func execute(args []string) int {
	runtime := defaultRuntime()

	return runtime.execute(args)
}

func stubAnalyze(runtime *cliRuntime) {
	runtime.analyze = func(context.Context, *distance.Config) (distance.Report, error) {
		return distance.Report{
			SchemaVersion: "6",
			Tool:          distance.ToolIdent("distance", "test"),
			Module:        "example.com/m",
		}, nil
	}
}

func TestResolvePolicyDefaultsAndPatterns(t *testing.T) {
	policy, source, err := resolvePolicy(nil, 0.5)
	if err != nil {
		t.Fatal(err)
	}

	if source != "flag thresholds" || len(policy.Packages) != 1 ||
		policy.Packages[0].Pattern != "./..." || policy.Packages[0].MaxDistance != 0.5 {

		t.Fatalf("default policy = %+v source = %q", policy.Packages, source)
	}

	policy, _, err = resolvePolicy([]string{"./internal/...", "./..."}, 0.3)
	if err != nil {
		t.Fatal(err)
	}

	if len(policy.Packages) != 2 || policy.Packages[0].MaxDistance != 0.3 {
		t.Fatalf("pattern policy = %+v", policy.Packages)
	}

	if _, _, err := resolvePolicy([]string{""}, 0.5); err == nil {
		t.Fatal("empty pattern succeeded")
	}
}

func TestResolvePolicyOverrideSource(t *testing.T) {
	_, source, err := resolvePolicy([]string{"./..."}, 0.4)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(source, "flag thresholds") {
		t.Fatalf("source = %q", source)
	}
}

func TestRunEarlyErrorPaths(t *testing.T) {
	if code := execute([]string{"--verbose", "--dependency-scope=nope"}); code != 1 {
		t.Fatalf("invalid dependency scope exit = %d, want 1", code)
	}

	badProfile := filepath.Join(t.TempDir(), "missing", "cpu.prof")

	if code := execute([]string{"--cpu-profile=" + badProfile}); code != 1 {
		t.Fatalf("bad CPU profile exit = %d, want 1", code)
	}

	badTemp := filepath.Join(t.TempDir(), "missing")
	t.Setenv("TMPDIR", badTemp)
	t.Setenv("TMP", badTemp)

	runtime := defaultRuntime()
	if code := runtime.runWebHelp(); code != 1 {
		t.Fatalf("web help with bad temp dir exit = %d, want 1", code)
	}
}

func TestRunCanceledAnalysis(t *testing.T) {
	runtime := defaultRuntime()

	runtime.analyze = func(context.Context, *distance.Config) (distance.Report, error) {
		return distance.Report{}, context.Canceled
	}

	if code := runtime.execute([]string{"./..."}); code != 130 {
		t.Fatalf("exit = %d, want 130", code)
	}
}

func TestRunMemoryProfileAndReportWriteErrors(t *testing.T) {
	runtime := defaultRuntime()
	stubAnalyze(&runtime)

	badHeap := filepath.Join(t.TempDir(), "missing", "heap.prof")

	if code := runtime.execute([]string{"--memory-profile=" + badHeap, "./..."}); code != 1 {
		t.Fatalf("bad memory profile exit = %d", code)
	}

	outDir := t.TempDir()

	err := os.Chmod(outDir, 0o500)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = os.Chmod(outDir, 0o700) })

	out := filepath.Join(outDir, "report.json")

	if code := runtime.execute([]string{"--format=json", "--output=" + out, "./..."}); code != 1 {
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

	if code := runtime.execute([]string{"--format=web", "./..."}); code != 0 {
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

	if code := runtime.execute(
		[]string{"--cpu-profile=" + filepath.Join(t.TempDir(), "cpu.prof"), "./..."},
	); code != 0 {
		t.Fatalf("exit = %d, want 0 (stop error is logged only)", code)
	}
}

func TestRunWebHelpTerminalBrowserWarn(t *testing.T) {
	runtime := defaultRuntime()

	runtime.isTerminal = func() bool { return true }
	runtime.openBrowser = func(string) error { return errors.New("open failed") }

	if code := runtime.runWebHelp(); code != 0 {
		t.Fatalf("runWebHelp exit = %d", code)
	}
}

func TestWriteHelpDocsCloseAndWriteErrors(t *testing.T) {
	runtime := defaultRuntime()

	runtime.closeHelpFile = func(*os.File) error { return errors.New("close failed") }

	if _, err := runtime.writeHelpDocs(); err == nil {
		t.Fatal("want close error")
	}

	runtime = defaultRuntime()
	runtime.writeDocs = func(outbound.Sink, string) error { return errors.New("docs failed") }

	if _, err := runtime.writeHelpDocs(); err == nil {
		t.Fatal("want docs write error")
	}
}
