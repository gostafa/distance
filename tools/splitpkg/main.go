// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package main

import (
	"errors"
	"fmt"
	"os"
)

var errShortWrite = errors.New("short write")

func main() {
	os.Exit(run(os.Args))
}

func run(args []string) int {
	if len(args) < two {
		return usageExit()
	}

	if splitDirs(args[one:]) {
		return one
	}

	return zero
}

func usageExit() int {
	err := writeLine(os.Stderr, "usage: splitpkg <dir> [<dir>...]")
	if err != nil {
		return one
	}

	return two
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
	err := splitPackage(dir)
	if err == nil {
		return false, false
	}

	writeErr := writeLine(os.Stderr, fmt.Sprintf("splitpkg %s: %v", dir, err))

	return writeErr != nil, true
}

func writeLine(file *os.File, text string) error {
	written, err := fmt.Fprintln(file, text)
	if err != nil {
		return fmt.Errorf(errWrapWriteLine, err)
	}

	if written == zero {
		return fmt.Errorf(errWrapWriteLine, errShortWrite)
	}

	return nil
}
