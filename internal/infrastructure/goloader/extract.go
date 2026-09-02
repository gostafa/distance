package goloader

import (
	"go/ast"
	"go/token"
	"go/types"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gostafa/distance/internal/features/typefacts/domain"
	"golang.org/x/tools/go/packages"
)

type extractorOptions struct {
	includeGenerated bool
	analyzed         map[string]bool // PkgPath set defining the CBO scope
	modulePath       string
	baseDir          string
}

// docRange records whether the struct field declared in [start, end] carries
// documentation (a doc comment or a trailing line comment, both of which
// godoc renders).
type docRange struct {
	start, end token.Pos
	documented bool
}

// extractPackage walks one loaded package and produces its facts. Each call
// is confined to one worker goroutine and only reads its own package's data.
func extractPackage(pkg *packages.Package, opts extractorOptions) domain.PackageExtract {
	generated, funcDecls, typeDocs, fieldDocs := indexSyntax(pkg)

	variables, constants, functions := packageDecls(pkg, opts, generated)
	exported, unexported := functionExportCounts(pkg, opts.includeGenerated, generated)

	out := domain.PackageExtract{
		Path:                pkg.PkgPath,
		InModule:            inModule(pkg, opts.modulePath),
		Imports:             importPaths(pkg),
		ExportedFuncCount:   exported,
		UnexportedFuncCount: unexported,
		VarCount:            len(variables),
		ConstCount:          len(constants),
		Variables:           variables,
		Constants:           constants,
		Functions:           functions,
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

		out.Types = append(out.Types,
			extractType(pkg, opts, generated, funcDecls, typeDocs, fieldDocs, tn, named))
	}

	return out
}

// functionExportCounts counts declared functions and methods in non-excluded
// files, split by exportedness. It preserves the historical package count.
func functionExportCounts(
	pkg *packages.Package,
	includeGenerated bool,
	generated map[string]bool,
) (exported, unexported int) {
	for _, file := range pkg.Syntax {
		if !includeGenerated && generated[pkg.Fset.Position(file.Package).Filename] {
			continue
		}

		for _, decl := range file.Decls {
			if decl, ok := decl.(*ast.FuncDecl); ok {
				if decl.Name.IsExported() {
					exported++
				} else {
					unexported++
				}
			}
		}
	}

	return exported, unexported
}

// packageDecls extracts top-level declarations that belong directly to the
// package detail view: vars, consts, and free functions.
func packageDecls(
	pkg *packages.Package,
	opts extractorOptions,
	generated map[string]bool,
) (variables, constants []domain.DeclarationFacts, functions []domain.FunctionFacts) {
	for _, file := range pkg.Syntax {
		if !opts.includeGenerated && generated[pkg.Fset.Position(file.Package).Filename] {
			continue
		}

		for _, decl := range file.Decls {
			switch decl := decl.(type) {
			case *ast.FuncDecl:
				if decl.Recv == nil {
					functions = append(functions, functionFacts(pkg.Fset, opts.baseDir, decl))
				}
			case *ast.GenDecl:
				switch decl.Tok {
				case token.VAR:
					variables = append(variables, valueDecls(pkg.Fset, opts.baseDir, decl)...)
				case token.CONST:
					constants = append(constants, valueDecls(pkg.Fset, opts.baseDir, decl)...)
				}
			}
		}
	}

	sort.Slice(variables, func(i, j int) bool { return declLess(variables[i], variables[j]) })
	sort.Slice(constants, func(i, j int) bool { return declLess(constants[i], constants[j]) })
	sort.Slice(functions, func(i, j int) bool { return functionLess(functions[i], functions[j]) })

	return variables, constants, functions
}

func valueDecls(fset *token.FileSet, baseDir string, decl *ast.GenDecl) []domain.DeclarationFacts {
	var out []domain.DeclarationFacts
	for _, spec := range decl.Specs {
		value, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}

		for _, name := range value.Names {
			if name.Name != "_" {
				out = append(out, domain.DeclarationFacts{
					Name:     name.Name,
					Exported: name.IsExported(),
					Pos:      position(fset, baseDir, name.Pos()),
				})
			}
		}
	}

	return out
}

func functionFacts(fset *token.FileSet, baseDir string, decl *ast.FuncDecl) domain.FunctionFacts {
	return domain.FunctionFacts{
		Name:     decl.Name.Name,
		Exported: decl.Name.IsExported(),
		Pos:      position(fset, baseDir, decl.Pos()),
		Lines:    lineCount(fset, decl.Pos(), decl.End()),
	}
}

func declLess(a, b domain.DeclarationFacts) bool {
	if a.Pos.File != b.Pos.File {
		return a.Pos.File < b.Pos.File
	}
	if a.Pos.Line != b.Pos.Line {
		return a.Pos.Line < b.Pos.Line
	}
	if a.Pos.Column != b.Pos.Column {
		return a.Pos.Column < b.Pos.Column
	}

	return a.Name < b.Name
}

func functionLess(a, b domain.FunctionFacts) bool {
	if a.Pos.File != b.Pos.File {
		return a.Pos.File < b.Pos.File
	}
	if a.Pos.Line != b.Pos.Line {
		return a.Pos.Line < b.Pos.Line
	}
	if a.Pos.Column != b.Pos.Column {
		return a.Pos.Column < b.Pos.Column
	}

	return a.Name < b.Name
}

// indexSyntax walks the ASTs once, recording generated files, method
// declarations, and documentation facts.
func indexSyntax(
	pkg *packages.Package,
) (generated map[string]bool, funcDecls map[*types.Func]*ast.FuncDecl, typeDocs map[types.Object]bool, fieldDocs []docRange) {
	generated = make(map[string]bool)
	funcDecls = make(map[*types.Func]*ast.FuncDecl)
	typeDocs = make(map[types.Object]bool)

	for _, file := range pkg.Syntax {
		filename := pkg.Fset.Position(file.Package).Filename
		generated[filename] = ast.IsGenerated(file)

		for _, decl := range file.Decls {
			switch decl := decl.(type) {
			case *ast.FuncDecl:
				if decl.Recv == nil {
					continue // free function, never a method
				}

				if fn, ok := pkg.TypesInfo.Defs[decl.Name].(*types.Func); ok {
					funcDecls[fn] = decl
				}
			case *ast.GenDecl:
				fieldDocs = indexTypeDecl(pkg.TypesInfo, typeDocs, fieldDocs, decl)
			}
		}
	}

	sort.Slice(fieldDocs, func(i, j int) bool { return fieldDocs[i].start < fieldDocs[j].start })

	return generated, funcDecls, typeDocs, fieldDocs
}

// indexTypeDecl records type documentation and struct field doc ranges from
// one general declaration, returning the grown field-doc list.
func indexTypeDecl(
	info *types.Info,
	typeDocs map[types.Object]bool,
	fieldDocs []docRange,
	decl *ast.GenDecl,
) []docRange {
	if decl.Tok != token.TYPE {
		return fieldDocs
	}

	for _, spec := range decl.Specs {
		spec, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}

		documented := spec.Doc != nil || (len(decl.Specs) == 1 && decl.Doc != nil)
		if obj := info.Defs[spec.Name]; obj != nil {
			typeDocs[obj] = documented
		}

		if st, ok := spec.Type.(*ast.StructType); ok && st.Fields != nil {
			for _, field := range st.Fields.List {
				fieldDocs = append(fieldDocs, docRange{
					start:      field.Pos(),
					end:        field.End(),
					documented: field.Doc != nil || field.Comment != nil,
				})
			}
		}
	}

	return fieldDocs
}

// extractType produces one named type's facts.
func extractType(
	pkg *packages.Package,
	opts extractorOptions,
	generated map[string]bool,
	funcDecls map[*types.Func]*ast.FuncDecl,
	typeDocs map[types.Object]bool,
	fieldDocs []docRange,
	tn *types.TypeName,
	named *types.Named,
) domain.TypeExtract {
	out := domain.TypeExtract{
		Name:     tn.Name(),
		Exported: tn.Exported(),
		Kind:     typeKind(named),
		Pos:      position(pkg.Fset, opts.baseDir, tn.Pos()),
	}

	out.Fields = structFields(named)

	methods := sortedMethods(pkg.Fset, opts, generated, funcDecls, named)

	out.Methods = make([]domain.MethodFacts, 0, len(methods))
	for _, m := range methods {
		out.Methods = append(out.Methods, methodFacts(pkg, opts, m))
	}

	return out
}

// structFields extracts the struct's field slots in declaration order. An
// embedded field is one slot of the outer type; promoted members are never
// represented here (§ promoted policy). Non-struct types yield no fields.
func structFields(named *types.Named) []domain.FieldFacts {
	st, ok := named.Underlying().(*types.Struct)
	if !ok {
		return nil
	}

	fields := make([]domain.FieldFacts, 0, st.NumFields())
	for i := 0; i < st.NumFields(); i++ {
		field := st.Field(i)
		fields = append(fields, domain.FieldFacts{
			Name:     field.Name(),
			Exported: field.Exported(),
			Embedded: field.Anonymous(),
		})
	}

	return fields
}

// methodDecl pairs a method object with its declaration site.
type methodDecl struct {
	fn   *types.Func
	decl *ast.FuncDecl
}

// sortedMethods collects the explicitly declared, non-skipped methods
// (receiver-carrying functions; pointer and value receivers both resolve
// here) sorted by name then source position. Promoted methods never appear
// in named.Method.
func sortedMethods(
	fset *token.FileSet,
	opts extractorOptions,
	generated map[string]bool,
	funcDecls map[*types.Func]*ast.FuncDecl,
	named *types.Named,
) []methodDecl {
	methods := make([]methodDecl, 0, named.NumMethods())
	for fn := range named.Methods() {
		fn := fn

		decl, ok := funcDecls[fn]
		if !ok || skipPos(fset, opts.includeGenerated, generated, decl.Pos()) {
			continue
		}

		methods = append(methods, methodDecl{fn: fn, decl: decl})
	}

	// Method names on a named type are unique, so name order is enough.
	sort.Slice(methods, func(i, j int) bool {
		return methods[i].fn.Name() < methods[j].fn.Name()
	})

	return methods
}

// methodFacts extracts one method's facts.
func methodFacts(
	pkg *packages.Package,
	opts extractorOptions,
	m methodDecl,
) domain.MethodFacts {
	return domain.MethodFacts{
		Name:     m.fn.Name(),
		Exported: m.fn.Exported(),
		Pos:      position(pkg.Fset, opts.baseDir, m.decl.Pos()),
		Lines:    lineCount(pkg.Fset, m.decl.Pos(), m.decl.End()),
	}
}

// memberDocs counts exported members (the type itself, exported fields,
// exported declared methods) and how many of them are documented.
func memberDocs(
	typeDocs map[types.Object]bool,
	fieldDocs []docRange,
	tn *types.TypeName,
	fields []domain.FieldFacts,
	fieldPositions []token.Pos,
	methods []methodDocInput,
) (exported, documented int) {
	if tn.Exported() {
		exported++

		if typeDocs[tn] {
			documented++
		}
	}

	for i, f := range fields {
		if !f.Exported {
			continue
		}

		exported++

		if fieldDocumented(fieldDocs, fieldPositions[i]) {
			documented++
		}
	}

	for _, m := range methods {
		if !m.exported {
			continue
		}

		exported++

		if m.documented {
			documented++
		}
	}

	return exported, documented
}

type methodDocInput struct {
	exported   bool
	documented bool
}

// fieldDocumented finds the (non-overlapping) field declaration range that
// contains pos and reports its documentation flag.
func fieldDocumented(fieldDocs []docRange, pos token.Pos) bool {
	i := sort.Search(len(fieldDocs), func(i int) bool { return fieldDocs[i].start > pos })
	if i == 0 {
		return false
	}

	r := fieldDocs[i-1]

	return pos >= r.start && pos <= r.end && r.documented
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

// position locates pos, relative to baseDir when possible so output is
// machine-independent.
func position(fset *token.FileSet, baseDir string, pos token.Pos) domain.Position {
	p := fset.Position(pos)

	file := p.Filename
	if baseDir != "" {
		if rel, err := filepath.Rel(baseDir, file); err == nil && !strings.HasPrefix(rel, "..") {
			file = filepath.ToSlash(rel)
		}
	}

	return domain.Position{File: file, Line: p.Line, Column: p.Column}
}

func lineCount(fset *token.FileSet, start, end token.Pos) int {
	startLine := fset.Position(start).Line
	endLine := fset.Position(end).Line
	if startLine <= 0 || endLine < startLine {
		return 0
	}

	return endLine - startLine + 1
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

// refCollector accumulates the CBO fact: the distinct other analyzed named
// types a type references through fields, method parameters, method
// returns, and embedded types.
type refCollector struct {
	self     *types.TypeName
	analyzed map[string]bool
	seen     map[string]bool
	visited  map[types.Type]bool
}

func newRefCollector(self *types.TypeName, analyzed map[string]bool) *refCollector {
	return &refCollector{
		self:     self,
		analyzed: analyzed,
		seen:     make(map[string]bool),
		visited:  make(map[types.Type]bool),
	}
}

// addType records the analyzed named types reachable through the structure
// of t (pointers, containers, function types, anonymous structs and
// interfaces, and generic type arguments). It does not descend into a named
// type's underlying type: transitive references belong to the referenced
// type, not this one.
func (r *refCollector) addType(t types.Type) {
	t = types.Unalias(t)
	if r.visited[t] {
		return
	}

	r.visited[t] = true

	if named, ok := t.(*types.Named); ok {
		recordNamedRef(r.seen, r.self, r.analyzed, named)
		addTypeArgRefs(r, named)

		return
	}

	descendRef(r, t)
}

// descendRef records references reachable through t's container structure
// (pointers, maps, function types, anonymous structs and interfaces),
// recursing through r.addType so the visited guard short-circuits cycles.
func descendRef(r *refCollector, t types.Type) {
	switch t := t.(type) {
	case *types.Map:
		r.addType(t.Key())
		r.addType(t.Elem())
	case interface{ Elem() types.Type }:
		// Pointers, slices, arrays, and channels.
		r.addType(t.Elem())
	case *types.Signature:
		addSignatureRefs(r, t)
	case *types.Struct:
		addStructRefs(r, t)
	case *types.Interface:
		addInterfaceRefs(r, t)
	}
}

// recordNamedRef marks one named type when it is another analyzed type.
func recordNamedRef(
	seen map[string]bool,
	self *types.TypeName,
	analyzed map[string]bool,
	t *types.Named,
) {
	tn := t.Origin().Obj()
	if tn != self && tn.Pkg() != nil && analyzed[tn.Pkg().Path()] {
		seen[domain.TypeKey(tn.Pkg().Path(), tn.Name())] = true
	}
}

// addTypeArgRefs descends into a named type's generic type arguments.
func addTypeArgRefs(r *refCollector, t *types.Named) {
	args := t.TypeArgs()
	for t := range args.Types() {
		r.addType(t)
	}
}

// addSignatureRefs descends into a function type's parameters and results.
func addSignatureRefs(r *refCollector, sig *types.Signature) {
	for v := range sig.Params().Variables() {
		r.addType(v.Type())
	}

	for v := range sig.Results().Variables() {
		r.addType(v.Type())
	}
}

// addStructRefs descends into an anonymous struct's field types.
func addStructRefs(r *refCollector, st *types.Struct) {
	for field := range st.Fields() {
		r.addType(field.Type())
	}
}

// addInterfaceRefs descends into an anonymous interface's embeds and
// explicit method signatures.
func addInterfaceRefs(r *refCollector, iface *types.Interface) {
	for etyp := range iface.EmbeddedTypes() {
		r.addType(etyp)
	}

	for method := range iface.ExplicitMethods() {
		if sig, ok := method.Type().(*types.Signature); ok {
			addSignatureRefs(r, sig)
		}
	}
}

// sortedRefKeys flattens the collected reference set deterministically.
func sortedRefKeys(seen map[string]bool) []string {
	if len(seen) == 0 {
		return nil
	}

	keys := make([]string, 0, len(seen))
	for key := range seen {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return keys
}
