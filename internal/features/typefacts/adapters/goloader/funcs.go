// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package goloader

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"slices"
	"strings"

	"github.com/gostafa/distance/internal/features/typefacts/domain/model"
	fo "github.com/gostafa/distance/internal/features/typefacts/ports/outbound"
	"github.com/gostafa/distance/internal/infrastructure/workerpool"
	"golang.org/x/tools/go/packages"
)

// New returns a Loader that extracts type facts via go/packages.
func New() *Loader { return &Loader{extract: extractFacts} }

// RunWorkers delegates extraction work to the reusable pool.
func RunWorkers(ctx context.Context, cfg WorkerRun, task func(int) error) error {
	err := workerpool.RunWorkers(ctx, workerpool.WorkerRun(cfg), task)
	if err != nil {
		return fmt.Errorf("extraction workers: %w", err)
	}

	return nil
}

// Workers selects the bounded worker count.
func Workers(configured, taskCount int) int { return workerpool.Workers(configured, taskCount) }

// Load loads packages and returns the module path plus one extract per package.
func (loader *Loader) Load(ctx context.Context, opt *factOpts) (mod string, ext Loaded, err error) {
	mod, ext, err = loader.extract(ctx, opt)
	if err != nil {
		canceled := ctx.Err()
		if canceled != nil {
			err = canceled
		}

		return emptyString, nil, fmt.Errorf(errWrapLoad, loadFailure(err))
	}

	return mod, ext, nil
}

func absorb(out *filterOut, pkg *packages.Package, opts *fo.FactOptions) {
	if !packageBroken(pkg) {
		out.pkgs = append(out.pkgs, pkg)

		return
	}

	if opts.ContinueOnError {
		return
	}

	out.errs = append(out.errs, packageErrorMsgs(pkg.PkgPath, pkg)...)
}

func defaultLoaderRuntime() loaderRuntime {
	return loaderRuntime{
		packagesLoad:      packages.Load,
		runExtractWorkers: RunWorkers,
	}
}

func emptyExtracts(count int) []model.PackageExtract {
	extracts := make([]model.PackageExtract, emptyLen, count)

	for range count {
		extracts = append(
			extracts,
			model.PackageExtract{Path: "", Imports: nil, Types: nil, InModule: false},
		)
	}

	return extracts
}

func extractAll(ctx context.Context, rtm loaderRuntime, job *extractJob) (pkgExtracts, error) {
	extracts := emptyExtracts(len(job.pkgs))

	err := rtm.runExtractWorkers(
		ctx,
		WorkerRun{
			Workers: Workers(job.opts.Workers, len(job.pkgs)),
			Tasks:   len(job.pkgs),
		},
		extractAt(job, extracts),
	)
	if err != nil {
		return nil, fmt.Errorf("goloader extractAll: %w", err)
	}

	return extracts, nil
}

func extractAt(job *extractJob, extracts []model.PackageExtract) func(int) error {
	return func(index int) error {
		pkg := job.pkgs[index]

		extracts[index] = extractPackage(pkg, &extractorOptions{
			includeGenerated: job.opts.IncludeGenerated,
			modulePath:       job.modulePath,
		})
		pkg.Syntax = nil
		pkg.TypesInfo = nil
		pkg.Types = nil

		return nil
	}
}

func extractFacts(ctx context.Context, opt *fo.FactOptions) (mod string, ext Loaded, err error) {
	out := load(ctx, defaultLoaderRuntime(), opt)

	if out.err != nil {
		return emptyString, nil, fmt.Errorf(errWrapLoad, out.err)
	}

	return out.mod, out.ext, nil
}

func extractNamedTypes(pkg *packages.Package, flt *skipFilter) []model.TypeName {
	scope := pkg.Types.Scope()
	out := make([]model.TypeName, emptyLen, len(scope.Names()))

	for i := range scope.Names() { // already sorted
		extracted, ok := namedTypeAt(scope, scope.Names()[i], flt)

		if ok {
			out = append(out, extracted)
		}
	}

	return out
}

func extractPackage(pkg *packages.Package, opts *extractorOptions) model.PackageExtract {
	generated := generatedFiles(pkg)

	return model.PackageExtract{
		Path:     pkg.PkgPath,
		InModule: inModule(pkg, opts.modulePath),
		Imports:  importPaths(pkg),
		Types:    pkgTypes(pkg, opts, generated),
	}
}

func filterPkgs(byPath pkgByPath, order []string, opt *factOpts) (pkgs loadedPkgs, errs []string) {
	out := filterOut{pkgs: make([]*packages.Package, emptyLen, len(order)), errs: nil}

	for i := range order {
		absorb(&out, byPath[order[i]], opt)
	}

	return out.pkgs, out.errs
}

func finishLoad(loaded loadedPkgs, pats []string, opts *fo.FactOptions) (loadedPkgs, error) {
	if len(loaded) == emptyLen {
		return nil, fmt.Errorf(errWrapPatterns, errNoMatchedPackages, pats)
	}

	pkgs, err := selectPackages(loaded, opts)
	if err != nil {
		return nil, fmt.Errorf("goloader finishLoad: %w", err)
	}

	if len(pkgs) == emptyLen {
		return nil, fmt.Errorf(errWrapPatterns, errNoLoadablePackages, pats)
	}

	return pkgs, nil
}

func firstAnyModule(pkgs []*packages.Package) string {
	for i := range pkgs {
		if m := pkgs[i].Module; m != nil {
			return m.Path
		}
	}

	return emptyString
}

func firstMainModule(pkgs []*packages.Package) string {
	for i := range pkgs {
		if m := pkgs[i].Module; m != nil && m.Main {
			return m.Path
		}
	}

	return emptyString
}

func generatedFiles(pkg *packages.Package) map[string]bool {
	generated := make(map[string]bool, len(pkg.Syntax))

	for i := range pkg.Syntax {
		filename := pkg.Fset.Position(pkg.Syntax[i].Package).Filename

		generated[filename] = ast.IsGenerated(pkg.Syntax[i])
	}

	return generated
}

func importPaths(pkg *packages.Package) []string {
	if len(pkg.Imports) == emptyLen {
		return nil
	}

	paths := make([]string, emptyLen, len(pkg.Imports))

	for path := range pkg.Imports {
		paths = append(paths, path)
	}

	slices.Sort(paths)

	return paths
}

func inModule(pkg *packages.Package, modulePath string) bool {
	return pkg.Module != nil && modulePath != emptyString && pkg.Module.Path == modulePath
}

func indexPackages(loaded loadedPkgs) (byPath pkgByPath, order []string) {
	byPath = make(pkgByPath, len(loaded))
	order = make([]string, emptyLen, len(loaded))

	for i := range loaded {
		recordPackage(byPath, &order, loaded[i])
	}

	return byPath, order
}

func load(ctx context.Context, rtm loaderRuntime, opt *factOpts) loadOut {
	pkgs, err := loadPackages(ctx, rtm, opt)
	if err != nil {
		return loadOut{err: fmt.Errorf(errWrapLoad, err), mod: "", ext: nil}
	}

	modulePath := mainModulePath(pkgs)

	extracts, err := extractAll(
		ctx,
		rtm,
		&extractJob{pkgs: pkgs, opts: opt, modulePath: modulePath},
	)
	if err != nil {
		return loadOut{err: fmt.Errorf(errWrapLoad, err), mod: "", ext: nil}
	}

	return loadOut{mod: modulePath, ext: extracts, err: nil}
}

func loadFailure(err error) error {
	if errors.Is(err, context.Canceled) {
		return fmt.Errorf("load canceled: %w", context.Canceled)
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("load deadline: %w", context.DeadlineExceeded)
	}

	return fmt.Errorf("%w: %s", fo.ErrLoad, err)
}

func loadPackages(ctx context.Context, rtm loaderRuntime, opt *factOpts) (loadedPkgs, error) {
	patterns := loadPatterns(opt)

	loaded, err := rtm.packagesLoad(packagesConfig(ctx, opt), patterns...)
	if err != nil {
		return nil, fmt.Errorf("load packages: %w", err)
	}

	pkgs, err := finishLoad(loaded, patterns, opt)
	if err != nil {
		return nil, fmt.Errorf("goloader finish load: %w", err)
	}

	return pkgs, nil
}

func loadPatterns(opts *fo.FactOptions) []string {
	if len(opts.Patterns) == emptyLen {
		return []string{defaultPattern}
	}

	return opts.Patterns
}

func mainModulePath(pkgs []*packages.Package) string {
	if path := firstMainModule(pkgs); path != emptyString {
		return path
	}

	return firstAnyModule(pkgs)
}

func namedTypeAt(scope *types.Scope, name string, filter *skipFilter) (model.TypeName, bool) {
	typeName, ok := scope.Lookup(name).(*types.TypeName)

	if !ok || typeName.IsAlias() {
		return model.Named(emptyString, model.KindOther), false
	}

	named, ok := typeName.Type().(*types.Named)

	if !ok {
		return model.Named(emptyString, model.KindOther), false
	}

	if skipPos(filter, typeName.Pos()) {
		return model.Named(emptyString, model.KindOther), false
	}

	return model.Named(typeName.Name(), typeKind(named)), true
}

func packageBroken(pkg *packages.Package) bool {
	return len(pkg.Errors) > emptyLen || pkg.Types == nil || pkg.TypesInfo == nil
}

func packageErrorMsgs(path string, pkg *packages.Package) []string {
	errs := make([]string, emptyLen, len(pkg.Errors)+1)

	for i := range pkg.Errors {
		errs = append(errs, fmt.Sprintf("%s: %s", path, pkg.Errors[i].Msg))
	}

	if len(pkg.Errors) == emptyLen {
		errs = append(errs, path+": type information unavailable")
	}

	return errs
}

func packageLoadError(errs []string) error {
	shown := errs
	suffix := emptyString

	if len(shown) > maxShownErrors {
		shown = shown[:maxShownErrors]
		suffix = fmt.Sprintf("\n… and %d more", len(errs)-maxShownErrors)
	}

	return fmt.Errorf("%w:\n%s%s", errPackageLoadFailures, strings.Join(shown, "\n"), suffix)
}

func packagesConfig(ctx context.Context, opts *fo.FactOptions) *packages.Config {
	cfg := &packages.Config{
		Context:    ctx,
		Dir:        opts.Directory,
		Mode:       loadMode,
		Tests:      opts.IncludeTests,
		Logf:       nil,
		Env:        nil,
		BuildFlags: nil,
		Fset:       nil,
		ParseFile:  nil,
		Overlay:    nil,
	}

	if len(opts.BuildTags) > emptyLen {
		cfg.BuildFlags = []string{"-tags=" + strings.Join(opts.BuildTags, ",")}
	}

	return cfg
}

func pkgTypes(pkg *packages.Package, opt *extractorOptions, gen map[string]bool) []model.TypeName {
	return extractNamedTypes(pkg, &skipFilter{fset: pkg.Fset, generated: gen, opts: opt})
}

func recordPackage(byPath pkgByPath, order *[]string, pkg *packages.Package) {
	if strings.HasSuffix(pkg.PkgPath, testPackageSuffix) {
		return // synthesized test main package
	}

	existing, ok := byPath[pkg.PkgPath]

	if !ok {
		byPath[pkg.PkgPath] = pkg
		*order = append(*order, pkg.PkgPath)

		return
	}

	if len(pkg.CompiledGoFiles) > len(existing.CompiledGoFiles) {
		byPath[pkg.PkgPath] = pkg
	}
}

func selectPackages(loaded loadedPkgs, opts *fo.FactOptions) (loadedPkgs, error) {
	byPath, order := indexPackages(loaded)

	var errs []string

	pkgs, errs := filterPkgs(byPath, order, opts)

	if len(errs) > emptyLen {
		return nil, fmt.Errorf("goloader select packages: %w", packageLoadError(errs))
	}

	return pkgs, nil
}

func skipPos(filter *skipFilter, pos token.Pos) bool {
	if filter.opts.includeGenerated {
		return false
	}

	return filter.generated[filter.fset.Position(pos).Filename]
}

func typeKind(named *types.Named) uint8 {
	switch named.Underlying().(type) {
	case *types.Struct:
		return model.KindStruct
	case *types.Interface:
		return model.KindInterface
	default:
		return model.KindOther
	}
}
