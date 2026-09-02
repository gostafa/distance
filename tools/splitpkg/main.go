// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/gostafa/distance/internal/tooling/splitpkg"
)

var errShortWrite = errors.New("short write")

func main() {
	os.Exit(run(os.Args))
}

func run(args []string) int {
	if len(args) < 2 {
		return usageExit()
	}

	if splitDirs(args[1:]) {
		return 1
	}

	return 0
}

func usageExit() int {
	err := writeLine(os.Stderr, "usage: splitpkg <dir> [<dir>...]")
	if err != nil {
		return 1
	}

	return 2
}

func splitDirs(dirs []string) bool {
	failed := false

	for i := range dirs {
		abort, oneFailed := reportSplit(dirs[i])

		failed = failed || oneFailed

		if abort {
			return true
		}
	}

	return failed
}

func reportSplit(dir string) (abort, failed bool) {
	err := splitpkg.SplitPackage(dir)
	if err == nil {
		return false, false
	}

	writeErr := writeLine(os.Stderr, fmt.Sprintf("splitpkg %s: %v", dir, err))

	return writeErr != nil, true
}

func writeLine(file *os.File, text string) error {
	written, err := fmt.Fprintln(file, text)
	if err != nil {
		return fmt.Errorf("splitpkg writeLine: %w", err)
	}

	if written == 0 {
		return fmt.Errorf("splitpkg writeLine: %w", errShortWrite)
	}

	return nil
}
