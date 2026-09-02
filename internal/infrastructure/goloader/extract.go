package goloader

import (
	"go/ast"
	"go/token"
	"go/types"
	"sort"

	"github.com/gostafa/distance/internal/features/typefacts/domain"
	"golang.org/x/tools/go/packages"
)

type extractorOptions struct {
	includeGenerated bool
	modulePath       string
}

// extractPackage walks one loaded package and produces its facts. Each call
// is confined to one worker goroutine and only reads its own package's data.
func extractPackage(pkg *packages.Package, opts extractorOptions) domain.PackageExtract {
	generated := generatedFiles(pkg)

	out := domain.PackageExtract{
		Path:     pkg.PkgPath,
		InModule: inModule(pkg, opts.modulePath),
		Imports:  importPaths(pkg),
	}

	scope := pkg.Types.Scope()
	for _, name := range scope.Names() { // already sorted
		tn, ok := scope.Lookup(name).(*types.TypeName)
		if !ok || tn.IsAlias() {
			continue
		}

		named, ok := tn.Type().(*types.Named)
		if !ok {
			continue
		}

		if skipPos(pkg.Fset, opts.includeGenerated, generated, tn.Pos()) {
			continue
		}

		out.Types = append(out.Types, domain.TypeExtract{
			Name: tn.Name(),
			Kind: typeKind(named),
		})
	}

	return out
}

func generatedFiles(pkg *packages.Package) map[string]bool {
	generated := make(map[string]bool, len(pkg.Syntax))
	for _, file := range pkg.Syntax {
		filename := pkg.Fset.Position(file.Package).Filename
		generated[filename] = ast.IsGenerated(file)
	}

	return generated
}

// skipPos reports whether the declaration at pos lives in a generated file
// that the run excludes.
func skipPos(
	fset *token.FileSet,
	includeGenerated bool,
	generated map[string]bool,
	pos token.Pos,
) bool {
	if includeGenerated {
		return false
	}

	return generated[fset.Position(pos).Filename]
}

// inModule reports whether the package belongs to the main module.
func inModule(pkg *packages.Package, modulePath string) bool {
	return pkg.Module != nil && modulePath != "" && pkg.Module.Path == modulePath
}

// importPaths returns the package's distinct import paths, sorted.
func importPaths(pkg *packages.Package) []string {
	if len(pkg.Imports) == 0 {
		return nil
	}

	paths := make([]string, 0, len(pkg.Imports))
	for path := range pkg.Imports {
		paths = append(paths, path)
	}

	sort.Strings(paths)

	return paths
}

func typeKind(named *types.Named) domain.TypeKind {
	switch named.Underlying().(type) {
	case *types.Struct:
		return domain.KindStruct
	case *types.Interface:
		return domain.KindInterface
	default:
		return domain.KindOther
	}
}
