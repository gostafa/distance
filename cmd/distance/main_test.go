// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package main

import (
	"flag"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestMainDelegatesToCLI(t *testing.T) {
	var gotArgs []string

	var gotCode int

	runtime := mainRuntime{
		run: func(args []string) int {
			gotArgs = append([]string(nil), args...)

			return 7
		},
		exit: func(code int) { gotCode = code },
	}

	runtime.start([]string{"--version"})

	if len(gotArgs) != 1 || gotArgs[0] != "--version" {
		t.Fatalf("args = %v", gotArgs)
	}

	if gotCode != 7 {
		t.Fatalf("exit code = %d, want 7", gotCode)
	}
}

func TestCommandProcess(t *testing.T) {
	if os.Getenv("DISTANCE_COMMAND_TEST") == "1" {
		os.Args = []string{"distance", "--version"}
		main()
		return
	}
	command := exec.CommandContext(
		t.Context(),
		os.Args[0],
		"-test.run=^TestCommandProcess$",
		"-test.paniconexit0=false",
	)
	command.Env = append(os.Environ(), "DISTANCE_COMMAND_TEST=1")
	coverage := flag.Lookup("test.gocoverdir")
	if coverage != nil && coverage.Value.String() != "" {
		command.Env = append(command.Env, "GOCOVERDIR="+coverage.Value.String())
	}
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("command: %v\n%s", err, output)
	}
	if !strings.Contains(string(output), "distance") {
		t.Fatalf("version output: %s", output)
	}
}
