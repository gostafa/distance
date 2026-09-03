// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package cli

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/gostafa/distance/distance"
)

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
