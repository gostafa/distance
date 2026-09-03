// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package goloader

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"runtime"
	"slices"
	"strings"
	"sync"

	tfdom "github.com/gostafa/distance/internal/features/typefacts/domain"
	"github.com/gostafa/distance/internal/features/typefacts/domain/model"
	fo "github.com/gostafa/distance/internal/features/typefacts/ports/outbound"
	"golang.org/x/tools/go/packages"
)

func extractPackage(pkg *packages.Package, opts *extractorOptions) model.PackageExtract {
	generated := generatedFiles(pkg)

	return model.PackageExtract{
		Path:     pkg.PkgPath,
		InModule: inModule(pkg, opts.modulePath),
		Imports:  importPaths(pkg),
		Types:    pkgTypes(pkg, opts, generated),
	}
}

func pkgTypes(pkg *packages.Package, opt *extractorOptions, gen map[string]bool) []model.TypeName {
	return extractNamedTypes(pkg, &skipFilter{fset: pkg.Fset, generated: gen, opts: opt})
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

func generatedFiles(pkg *packages.Package) map[string]bool {
	generated := make(map[string]bool, len(pkg.Syntax))

	for i := range pkg.Syntax {
		filename := pkg.Fset.Position(pkg.Syntax[i].Package).Filename

		generated[filename] = ast.IsGenerated(pkg.Syntax[i])
	}

	return generated
}

func skipPos(filter *skipFilter, pos token.Pos) bool {
	if filter.opts.includeGenerated {
		return false
	}

	return filter.generated[filter.fset.Position(pos).Filename]
}

func inModule(pkg *packages.Package, modulePath string) bool {
	return pkg.Module != nil && modulePath != emptyString && pkg.Module.Path == modulePath
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

func typeKind(named *types.Named) uint8 {
	switch named.Underlying().(type) {
	case *types.Struct:
		return tfdom.KindStruct
	case *types.Interface:
		return tfdom.KindInterface
	default:
		return tfdom.KindOther
	}
}

// New returns a Loader that extracts type facts via go/packages.
func New() *Loader { return &Loader{} }

func defaultLoaderRuntime() loaderRuntime {
	return loaderRuntime{
		packagesLoad:      packages.Load,
		runExtractWorkers: RunWorkers,
	}
}

// Load loads packages and returns the module path plus one extract per package.
func (*Loader) Load(ctx context.Context, opt *fo.FactOptions) (mod string, ext Loaded, err error) {
	out := defaultLoaderRuntime().load(ctx, opt)

	if out.err != nil {
		return emptyString, nil, fmt.Errorf(errWrapLoad, out.err)
	}

	return out.mod, out.ext, nil
}

func (rtm loaderRuntime) load(ctx context.Context, opt *factOpts) loadOut {
	pkgs, err := rtm.loadPackages(ctx, opt)
	if err != nil {
		return loadOut{err: fmt.Errorf(errWrapLoad, err)}
	}

	modulePath := mainModulePath(pkgs)

	extracts, err := rtm.extractAll(ctx, &extractJob{pkgs: pkgs, opts: opt, modulePath: modulePath})
	if err != nil {
		return loadOut{err: fmt.Errorf(errWrapLoad, err)}
	}

	return loadOut{mod: modulePath, ext: extracts}
}

func (rtm loaderRuntime) loadPackages(ctx context.Context, opt *factOpts) (loadedPkgs, error) {
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

func packagesConfig(ctx context.Context, opts *fo.FactOptions) *packages.Config {
	cfg := &packages.Config{
		Context: ctx,
		Dir:     opts.Directory,
		Mode:    loadMode,
		Tests:   opts.IncludeTests,
	}

	if len(opts.BuildTags) > emptyLen {
		cfg.BuildFlags = []string{"-tags=" + strings.Join(opts.BuildTags, ",")}
	}

	return cfg
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

func (rtm loaderRuntime) extractAll(ctx context.Context, job *extractJob) (pkgExtracts, error) {
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

func emptyExtracts(count int) []model.PackageExtract {
	extracts := make([]model.PackageExtract, emptyLen, count)

	for range count {
		extracts = append(extracts, model.PackageExtract{})
	}

	return extracts
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

func selectPackages(loaded loadedPkgs, opts *fo.FactOptions) (loadedPkgs, error) {
	byPath, order := indexPackages(loaded)

	var errs []string

	pkgs, errs := filterPkgs(byPath, order, opts)

	if len(errs) > emptyLen {
		return nil, fmt.Errorf("goloader select packages: %w", packageLoadError(errs))
	}

	return pkgs, nil
}

func indexPackages(loaded loadedPkgs) (byPath pkgByPath, order []string) {
	byPath = make(pkgByPath, len(loaded))
	order = make([]string, emptyLen, len(loaded))

	for i := range loaded {
		recordPackage(byPath, &order, loaded[i])
	}

	return byPath, order
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

func filterPkgs(byPath pkgByPath, order []string, opt *factOpts) (pkgs loadedPkgs, errs []string) {
	out := filterOut{pkgs: make([]*packages.Package, emptyLen, len(order))}

	for i := range order {
		out.absorb(byPath[order[i]], order[i], opt)
	}

	return out.pkgs, out.errs
}

func (out *filterOut) absorb(pkg *packages.Package, path string, opts *fo.FactOptions) {
	if !packageBroken(pkg) {
		out.pkgs = append(out.pkgs, pkg)

		return
	}

	if opts.ContinueOnError {
		return
	}

	out.errs = append(out.errs, packageErrorMsgs(path, pkg)...)
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

func mainModulePath(pkgs []*packages.Package) string {
	if path := firstMainModule(pkgs); path != emptyString {
		return path
	}

	return firstAnyModule(pkgs)
}

func firstMainModule(pkgs []*packages.Package) string {
	for i := range pkgs {
		if m := pkgs[i].Module; m != nil && m.Main {
			return m.Path
		}
	}

	return emptyString
}

func firstAnyModule(pkgs []*packages.Package) string {
	for i := range pkgs {
		if m := pkgs[i].Module; m != nil {
			return m.Path
		}
	}

	return emptyString
}

// RunWorkers executes task(i) for each i in [0, cfg.Tasks) using cfg.Workers goroutines.
func RunWorkers(ctx context.Context, cfg WorkerRun, task func(int) error) error {
	runErr := runWorkersEmpty(ctx)

	if cfg.Tasks != emptyLen {
		runErr = runWorkersIndexed(ctx, &cfg, task)
	}

	if runErr != nil {
		return fmt.Errorf(errWrapRunWorkers, runErr)
	}

	return nil
}

// Workers returns how many goroutines to use for taskCount tasks.
func Workers(configured, taskCount int) int {
	workers := min(runtime.GOMAXPROCS(emptyLen), taskCount)

	if configured > emptyLen {
		workers = min(configured, taskCount)
	}

	return max(workers, minWorkers)
}

func runWorkersEmpty(ctx context.Context) error {
	emptyErr := workerContextError(ctx, "goloader workers")
	if emptyErr != nil {
		return fmt.Errorf(errWrapRunWorkers, emptyErr)
	}

	return nil
}

func workerContextError(ctx context.Context, prefix string) error {
	err := ctx.Err()
	if err != nil {
		return fmt.Errorf("%s: %w", prefix, err)
	}

	return nil
}

func emptyWorkerErrors(count int) []error {
	errs := make([]error, emptyLen, count)

	for range count {
		errs = append(errs, nil)
	}

	return errs
}

func foldWorkerErrors(ctx context.Context, errs []error) error {
	err := workerContextError(ctx, "goloader workers foldErrors")
	if err != nil {
		return fmt.Errorf(errWrapFoldWorkerErrors, err)
	}

	first := firstWorkerError(errs)
	if first != nil {
		return fmt.Errorf(errWrapFoldWorkerErrors, first)
	}

	return nil
}

func firstWorkerError(errs []error) error {
	for i := range errs {
		if errs[i] != nil {
			return fmt.Errorf("goloader workers firstError: %w", errs[i])
		}
	}

	return nil
}

func runWorkersIndexed(ctx context.Context, cfg *WorkerRun, task func(int) error) error {
	errs := emptyWorkerErrors(cfg.Tasks)
	tasks := make(chan int)

	waitGroup := startWorkerDraining(cfg, func() {
		drainWorkerTasks(tasks, errs, task)
	})
	sendWorkerTasks(ctx, tasks, cfg.Tasks)
	close(tasks)
	waitGroup.Wait()

	foldErr := foldWorkerErrors(ctx, errs)
	if foldErr != nil {
		return fmt.Errorf("goloader workers runIndexed: %w", foldErr)
	}

	return nil
}

func startWorkerDraining(cfg *WorkerRun, start func()) *sync.WaitGroup {
	waitGroup := new(sync.WaitGroup)

	startWorkerPool(waitGroup, Workers(cfg.Workers, cfg.Tasks), start)

	return waitGroup
}

func startWorkerPool(waitGroup *sync.WaitGroup, count int, start func()) {
	waitGroup.Add(count)
	launchWorkerGoroutines(waitGroup, count, start)
}

func sendOneWorkerTask(ctx context.Context, tasks chan<- int, index int) bool {
	select {
	case tasks <- index:
		return true
	case <-ctx.Done():
		return false
	}
}

func sendWorkerTasks(ctx context.Context, tasks chan<- int, count int) {
	for index := range count {
		if !sendOneWorkerTask(ctx, tasks, index) {
			return
		}
	}
}

func launchWorkerGoroutines(waitGroup *sync.WaitGroup, count int, start func()) {
	for range count {
		go func() {
			defer waitGroup.Done()

			start()
		}()
	}
}

func drainWorkerTasks(tasks <-chan int, errs []error, task func(int) error) {
	for index := range tasks {
		errs[index] = task(index)
	}
}
