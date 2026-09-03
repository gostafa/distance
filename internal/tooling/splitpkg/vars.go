// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package splitpkg

import (
	"errors"
)

var (
	errShortWrite  = errors.New("short write")
	errPackageName = errors.New("package-name")
	errDecl        = errors.New("decl")
	errGenDecl     = errors.New("gen-decl")
)
