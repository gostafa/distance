// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package splitpkg

const (
	goHeader = "// Gostafa 2026.\n// SPDX-License-Identifier: Apache-2.0.\n\n"

	zero = 0

	one = 1

	two = 2

	fileDoc = "doc.go"

	fileConsts = "consts.go"

	fileTypes = "types.go"

	fileVars = "vars.go"

	fileFuncs = "funcs.go"

	goFileSuffix = ".go"

	testFileSuffix = "_test.go"

	packagePrefix = "package "

	embedImport = `"embed"`

	embedGoFake = "embed.go"

	emptyString = ""

	newline = "\n"

	dotName = "."

	opOpen = "open"

	opClose = "close"

	opWrite = "write"

	opCreate = "create"

	opRemove = "remove"

	opRead = "read"

	opReaddir = "readdir"

	opDir = "dir"

	opParse = "parse"

	errFmtOpPathWrap = "%s %s: %w"

	errFmtSentinelPath = "%w %s"

	errWrapSplitPackage = "splitpkg splitPackage: %w"

	errWrapSplitOpened = "splitpkg splitOpened: %w"

	errWrapIngestGoFile = "splitpkg ingestGoFile: %w"

	errWrapAcceptParsedFile = "splitpkg acceptParsedFile: %w"

	errWrapFinishDecl = "splitpkg finishDecl: %w"

	errWrapApplyDecl = "splitpkg applyDecl: %w"

	errWrapWriteSplit = "splitpkg writeSplit: %w"

	errWrapWriteDocGo = "splitpkg writeDocGo: %w"

	errWrapWriteCategoryGo = "splitpkg writeCategoryGo: %w"

	errWrapFinishCategory = "splitpkg finishCategoryFile: %w"

	testPackageSuffix = "_test"

	errSentinelDetails = "%w: %s"

	errWriteTests = "write tests: %w"
)
