// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package main

import (
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
