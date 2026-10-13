// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package splitpkg

import (
	"errors"
	"os"
)

var (
	errShortWrite = errors.New("short write")

	errPackageName = errors.New("package-name")

	errDecl = errors.New("decl")

	errGenDecl = errors.New("gen-decl")

	errTestPackage = errors.New("conflicting benchmark packages")

	errMixedGenerated = errors.New(
		"mixed generated and ordinary declarations cannot share one file",
	)

	errBuildConstraint = errors.New(
		"build constraints cannot be preserved by a single category file",
	)

	errImportAlias = errors.New("conflicting test import alias")

	_ directoryReader = (*os.File)(nil)

	_ packageFiles = (*os.Root)(nil)
)
