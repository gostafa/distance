// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package splitpkg

import (
	"go/ast"
	"os"
)

type (
	declEntry struct {
		source string
	}

	categoryFile = struct {
		name  string
		decls []declEntry
	}

	packageSplit = struct {
		root            *os.Root
		dir             string
		pkgName         string
		generatedMarker string
		pkgDoc          *ast.CommentGroup
		consts          []declEntry
		types           []declEntry
		vars            []declEntry
		funcs           []declEntry
		goFiles         []string
	}

	categoryWrite = struct {
		store           *packageSplit
		pkgName         string
		filename        string
		generatedMarker string
		decls           []declEntry
	}

	declCollect = struct {
		split *packageSplit
		path  string
		src   []byte
	}
)
