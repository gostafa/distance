// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package main

import (
	"fmt"
	"io"
	"os"

	"github.com/gostafa/distance/internal/tooling/splitpkg"
)

func main() {
	os.Exit(run(os.Args))
}

func reportSplit(dir string) (abort, failed bool) {
	err := splitpkg.SplitPackage(dir)
	if err == nil {
		return false, false
	}

	writeErr := writeLine(os.Stderr, fmt.Sprintf("splitpkg %s: %v", dir, err))

	return writeErr != nil, true
}

func run(args []string) int {
	if len(args) < exitUsageCode {
		return usageExit()
	}

	if splitDirs(args[1:]) {
		return exitFail
	}

	return exitOK
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

func usageExit() int {
	err := writeLine(os.Stderr, "usage: splitpkg <dir> [<dir>...]")
	if err != nil {
		return exitFail
	}

	return exitUsageCode
}

func writeLine(destination io.Writer, text string) error {
	payload := text + "\n"

	written, err := io.WriteString(destination, payload)
	if err != nil {
		return fmt.Errorf(errFmtWriteLine, err)
	}

	if written != len(payload) {
		return fmt.Errorf(errFmtWriteLine, errShortWrite)
	}

	return nil
}
