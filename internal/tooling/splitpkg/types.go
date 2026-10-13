// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package splitpkg

import (
	"go/ast"
	"go/token"
	"io/fs"
	"os"
)

type (
	declEntry = struct {
		source string
	}

	categoryFile = struct {
		name  string
		decls []declEntry
	}

	packageSplit = struct {
		root            packageFiles
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

	testBundle = struct {
		packageName     string
		generatedMarker string
		imports         map[string]string
		aliases         map[string]string
		declarations    []string
		files           []string
	}

	testRewrite = struct {
		value string
		start int
		end   int
	}

	testInput = struct {
		set    *token.FileSet
		file   *ast.File
		name   string
		source []byte
	}

	testSelectorInput = struct {
		set      *token.FileSet
		decl     ast.Decl
		bindings map[string]string
		start    int
	}

	generatedCheck = struct {
		markers map[string]bool
		seen    map[string]bool
		name    string
		source  []byte
	}

	testDeclarationInput = struct {
		bindings map[string]string
		index    int
	}

	directoryReader = interface {
		ReadDir(count int) ([]fs.DirEntry, error)
		Close() error
	}

	// packageFiles scopes the I/O used by package rewriting.
	packageFiles = interface {
		Open(name string) (*os.File, error)
		Create(name string) (*os.File, error)
		Remove(name string) error
		Close() error
	}
)
