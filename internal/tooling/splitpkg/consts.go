// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package splitpkg

const (
	zero = 0
	one  = 1
	two  = 2

	fileDoc    = "doc.go"
	fileConsts = "consts.go"
	fileTypes  = "types.go"
	fileVars   = "vars.go"
	fileFuncs  = "funcs.go"

	goFileSuffix   = ".go"
	testFileSuffix = "_test.go"

	packagePrefix = "package "
	embedImport   = `"embed"`
	embedGoFake   = "embed.go"

	emptyString = ""
	newline     = "\n"
	dotName     = "."

	opOpen  = "open"
	opClose = "close"
	opWrite = "write"

	errWrapWriteLine        = "splitpkg writeLine: %w"
	errWrapSplitPackage     = "splitpkg splitPackage: %w"
	errWrapSplitOpened      = "splitpkg splitOpened: %w"
	errWrapIngestGoFile     = "splitpkg ingestGoFile: %w"
	errWrapAcceptParsedFile = "splitpkg acceptParsedFile: %w"
	errWrapWriteSplit       = "splitpkg writeSplit: %w"
	errWrapWriteDocGo       = "splitpkg writeDocGo: %w"
	errWrapWriteCategoryGo  = "splitpkg writeCategoryGo: %w"
	errWrapBuildCategory    = "splitpkg buildCategorySource: %w"
	errWrapWriteHeader      = "splitpkg writeCategoryHeader: %w"
	errWrapWriteNewline     = "splitpkg write newline: %w"
	errWrapFinishCategory   = "splitpkg finishCategoryFile: %w"
	errWrapWriteString      = "write string: %w"

	errShortWrite = "short write"
)
