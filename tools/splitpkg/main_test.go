// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package main

import (
	"errors"
	"flag"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCommandUsage(t *testing.T) {
	if os.Getenv("DISTANCE_SPLITTER_TEST") == "1" {
		os.Args = []string{"splitpkg"}
		main()
		return
	}
	command := exec.CommandContext(
		t.Context(),
		os.Args[0],
		"-test.run=^TestCommandUsage$",
		"-test.paniconexit0=false",
	)
	command.Env = append(os.Environ(), "DISTANCE_SPLITTER_TEST=1")
	coverage := flag.Lookup("test.gocoverdir")
	if coverage != nil && coverage.Value.String() != "" {
		command.Env = append(command.Env, "GOCOVERDIR="+coverage.Value.String())
	}
	output, err := command.CombinedOutput()
	exit, ok := err.(*exec.ExitError)
	if !ok || exit.ExitCode() != 2 {
		t.Fatalf("usage exit = %v\n%s", err, output)
	}
	if !strings.Contains(string(output), "usage: splitpkg") {
		t.Fatalf("usage output: %s", output)
	}
}

type toolWriter struct{ err error }

func (w toolWriter) Write([]byte) (int, error) { return 0, w.err }

func TestSplitterCommandResults(t *testing.T) {
	directory := t.TempDir()
	if err := os.WriteFile(
		filepath.Join(directory, "old.go"),
		[]byte("package sample\nvar Value = 7\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	if code := run([]string{"splitpkg", directory}); code != 0 {
		t.Fatalf("successful split exit = %d", code)
	}
	if code := run([]string{"splitpkg", filepath.Join(directory, "missing")}); code != 1 {
		t.Fatalf("failed split exit = %d", code)
	}
	if code := run(nil); code != 2 {
		t.Fatalf("usage exit = %d", code)
	}
	if err := writeLine(io.Discard, "output"); err != nil {
		t.Fatal(err)
	}
	if err := writeLine(toolWriter{}, "output"); !errors.Is(err, errShortWrite) {
		t.Fatalf("short line = %v", err)
	}
	failure := errors.New("write failed")
	if err := writeLine(toolWriter{err: failure}, "output"); !errors.Is(err, failure) {
		t.Fatalf("line error = %v", err)
	}
	closed, err := os.CreateTemp(t.TempDir(), "closed")
	if err != nil {
		t.Fatal(err)
	}
	if err := closed.Close(); err != nil {
		t.Fatal(err)
	}
	original := os.Stderr
	os.Stderr = closed
	t.Cleanup(func() { os.Stderr = original })
	if code := run(nil); code != 1 {
		t.Fatalf("failed usage output exit = %d", code)
	}
	if !splitDirs([]string{filepath.Join(directory, "missing")}) {
		t.Fatal("failed stderr should abort split")
	}
}
