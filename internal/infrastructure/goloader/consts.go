// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package goloader

import (
	"golang.org/x/tools/go/packages"
)

const (
	loadMode = packages.NeedName |
		packages.NeedModule |
		packages.NeedCompiledGoFiles |
		packages.NeedImports |
		packages.NeedSyntax |
		packages.NeedTypes |
		packages.NeedTypesInfo |
		packages.NeedTypesSizes

	emptyLen          = 0
	emptyString       = ""
	defaultPattern    = "./..."
	testPackageSuffix = ".test"
	maxShownErrors    = 10
	errWrapLoad       = "goloader load: %w"
	errWrapPatterns   = "%w %v"
	errWrapRunWorkers = "goloader RunWorkers: %w"
	errWrapFoldWorkerErrors = "goloader foldWorkerErrors: %w"

	workerZero = 0
	minWorkers = 1
)
