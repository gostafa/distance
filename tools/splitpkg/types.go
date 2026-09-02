// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package main

import (
	"go/ast"
	"os"
)

type (
	fileReader interface {
		ReadFile(name string) ([]byte, error)
	}

	fileWriter interface {
		WriteFile(name string, data []byte) error
	}

	fileRemover interface {
		RemoveFile(name string) error
	}

	dirHolder interface {
		Dir() string
	}

	sourceLister interface {
		GoFiles() []string
	}

	markerHolder interface {
		Marker() string
	}

	declEntry struct {
		source string
	}

	categoryFile struct {
		name  string
		decls []declEntry
	}

	packageSplit struct {
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

	categoryWrite struct {
		store           fileWriter
		pkgName         string
		filename        string
		generatedMarker string
		decls           []declEntry
	}

	declCollect struct {
		split *packageSplit
		path  string
		src   []byte
	}

	pathError struct {
		err  error
		op   string
		path string
	}
)
