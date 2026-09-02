package analyzer

import (
	"go/token"

	"golang.org/x/tools/go/analysis"
)

// packagePos returns a position for package-scoped diagnostics: the package
// clause of the first file, or token.NoPos when the pass has no files.
func packagePos(pass *analysis.Pass) token.Pos {
	for _, file := range pass.Files {
		if file != nil {
			return file.Package
		}
	}

	return token.NoPos
}
