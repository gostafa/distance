// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package splitpkg

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"golang.org/x/tools/imports"
)

// SplitPackage rewrites Go files in dir into consts, funcs, types, and vars files.
func SplitPackage(dir string) error {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return fmt.Errorf(errFmtOpPathWrap, opOpen, dir, err)
	}

	closeErr := closeRootAfter(splitOpened(root, dir), root)
	if closeErr != nil {
		return fmt.Errorf(errWrapSplitPackage, closeErr)
	}

	return nil
}

func acceptPackageName(split *packageSplit, path, name string) error {
	if split.pkgName == emptyString {
		split.pkgName = name

		return nil
	}

	if name != split.pkgName {
		return fmt.Errorf(errFmtSentinelPath, errPackageName, path)
	}

	return nil
}

func acceptParsedFile(fset *token.FileSet, file *ast.File, input *declCollect) error {
	err := acceptPackageName(input.split, input.path, file.Name.Name)
	if err != nil {
		return fmt.Errorf(errWrapAcceptParsedFile, err)
	}

	mergePackageDoc(input.split, file.Doc)

	collectErr := collectDecls(fset, input, file.Decls)
	if collectErr != nil {
		return fmt.Errorf(errWrapAcceptParsedFile, collectErr)
	}

	return nil
}

func appendFuncDecl(input *declCollect, decl ast.Decl, entry declEntry) error {
	if !isFuncDecl(decl) {
		return fmt.Errorf(errFmtSentinelPath, errDecl, input.path)
	}

	input.split.funcs = append(input.split.funcs, entry)

	return nil
}

func appendTestDeclaration(
	bundle *testBundle,
	input *testInput,
	declaration *testDeclarationInput,
) {
	decl := input.file.Decls[declaration.index]

	if group, ok := decl.(*ast.GenDecl); ok && group.Tok == token.IMPORT {
		return
	}

	bundle.declarations = append(bundle.declarations, testDeclarationSource(input, declaration))
}

func appendTestFile(split *packageSplit, bundles map[string]*testBundle, input *testInput) error {
	err := acceptPackageName(
		split,
		input.name,
		strings.TrimSuffix(input.file.Name.Name, testPackageSuffix),
	)
	if err != nil {
		return fmt.Errorf("test package: %w", err)
	}

	bundle, err := testBundleFor(bundles, input)
	if err != nil {
		return fmt.Errorf("test bundle: %w", err)
	}

	err = collectTestDeclarations(bundle, input)
	if err != nil {
		return fmt.Errorf("test declarations: %w", err)
	}

	return nil
}

func applyDecl(input *declCollect, decl ast.Decl, entry declEntry) error {
	var err error

	switch typed := decl.(type) {
	case *ast.GenDecl:
		err = collectGenDecl(input, typed.Tok, entry)
	default:
		err = appendFuncDecl(input, decl, entry)
	}

	if err != nil {
		return fmt.Errorf(errWrapApplyDecl, err)
	}

	return nil
}

func applyTestRewrites(text string, edits []testRewrite) string {
	slices.SortFunc(edits, func(a, b testRewrite) int { return b.start - a.start })

	for i := range edits {
		edit := &edits[i]

		text = text[:edit.start] + edit.value + text[edit.end:]
	}

	return text
}

func bindTestImport(
	bundle *testBundle,
	bindings map[string]string,
	specification *ast.ImportSpec,
) error {
	path, err := strconv.Unquote(specification.Path.Value)
	if err != nil {
		return fmt.Errorf("test import path: %w", err)
	}

	alias := testImportAlias(specification, path)

	canonical, err := canonicalTestAlias(bundle, path, alias)
	if err != nil {
		return fmt.Errorf("test import alias: %w", err)
	}

	bindings[alias] = canonical

	return nil
}

func buildCategorySource(input *categoryWrite) []byte {
	return []byte(categoryHeader(input) + declarationSource(input.decls))
}

func buildDocSource(pkgName string, doc *ast.CommentGroup) []byte {
	lines := make([]string, zero, len(doc.List)+two)

	lines = append(lines, goHeader)

	for i := range doc.List {
		lines = append(lines, doc.List[i].Text+newline)
	}

	lines = append(lines, packagePrefix+pkgName+newline)

	return []byte(strings.Join(lines, emptyString))
}

func buildTestSource(bundle *testBundle) string {
	lines := make([]string, zero, len(bundle.imports))

	for path := range bundle.imports {
		alias := bundle.imports[path]

		lines = append(lines, alias+" "+strconv.Quote(path))
	}

	slices.Sort(lines)

	return goHeader + bundle.generatedMarker + "\npackage " + bundle.packageName + "\nimport (\n" + strings.Join(
		lines,
		"\n",
	) + "\n)\n" + strings.Join(
		bundle.declarations,
		"\n\n",
	) + "\n"
}

func canonicalTestAlias(bundle *testBundle, path, alias string) (string, error) {
	canonical, exists := bundle.imports[path]

	if !exists {
		canonical = alias
	}

	if conflictsTestAlias(bundle, canonical, path) {
		return emptyString, fmt.Errorf(errSentinelDetails, errImportAlias, canonical)
	}

	bundle.imports[path] = canonical
	bundle.aliases[canonical] = path

	return canonical, nil
}

func captureGeneratedMarker(split *packageSplit, src []byte) {
	marker := extractGeneratedMarker(src)
	fresh := marker != emptyString && split.generatedMarker == emptyString

	if fresh {
		split.generatedMarker = marker
	}
}

func categoryHeader(input *categoryWrite) string {
	marker := input.generatedMarker

	if marker != emptyString {
		marker += newline + newline
	}

	return goHeader + marker + packagePrefix + input.pkgName + newline + newline
}

func categoryList(split *packageSplit) []categoryFile {
	return []categoryFile{
		{name: fileConsts, decls: split.consts},
		{name: fileTypes, decls: split.types},
		{name: fileVars, decls: split.vars},
		{name: fileFuncs, decls: split.funcs},
	}
}

func categoryMarkerValue(marker, name, firstCategory string) string {
	if name != firstCategory {
		return emptyString
	}

	return marker
}

func checkGeneratedCategories(input *generatedCheck) error {
	file, err := parser.ParseFile(token.NewFileSet(), input.name, input.source, zero)
	if err != nil {
		return fmt.Errorf("generated source parse: %w", err)
	}

	for i := range file.Decls {
		category := sourceCategory(input.name, file.Name.Name, file.Decls[i])

		err = recordGeneratedCategory(input, category)
		if err != nil {
			return fmt.Errorf("generated source: %w", err)
		}
	}

	return nil
}

func checkSourceConstraints(name string, data []byte) error {
	file, err := parser.ParseFile(token.NewFileSet(), name, data, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("preflight parse: %w", err)
	}

	if slices.ContainsFunc(file.Comments, hasBuildConstraint) {
		return fmt.Errorf(errSentinelDetails, errBuildConstraint, name)
	}

	return nil
}

func closeRootAfter(prior error, root io.Closer) error {
	closeErr := root.Close()

	if prior != nil {
		return prior
	}

	if closeErr != nil {
		return fmt.Errorf(errFmtOpPathWrap, opClose, dotName, closeErr)
	}

	return nil
}

func collectDecl(fset *token.FileSet, input *declCollect, decl ast.Decl) error {
	entry := declEntry{source: extractDeclSource(fset, input.src, decl)}

	finishErr := finishDecl(input, decl, entry)
	if finishErr != nil {
		return fmt.Errorf("splitpkg collectDecl: %w", finishErr)
	}

	return nil
}

func collectDecls(fset *token.FileSet, input *declCollect, decls []ast.Decl) error {
	for i := range decls {
		err := collectDecl(fset, input, decls[i])
		if err != nil {
			return fmt.Errorf("splitpkg collectDecls: %w", err)
		}
	}

	return nil
}

func collectGenDecl(input *declCollect, tok token.Token, entry declEntry) error {
	target := genDeclTarget(input.split, tok)

	if target != nil {
		*target = append(*target, entry)

		return nil
	}

	if tok == token.IMPORT {
		return nil
	}

	return fmt.Errorf(errFmtSentinelPath, errGenDecl, input.path)
}

func collectGoFiles(entries []fs.DirEntry) []string {
	goFiles := make([]string, zero, len(entries))

	for i := range entries {
		name, ok := goSourceName(entries[i])

		if ok {
			goFiles = append(goFiles, name)
		}
	}

	slices.Sort(goFiles)

	return goFiles
}

func collectTestBundles(
	split *packageSplit,
	bundles map[string]*testBundle,
	entries []fs.DirEntry,
) error {
	for i := range entries {
		err := collectTestEntry(split, bundles, entries[i])
		if err != nil {
			return fmt.Errorf("collect test entry: %w", err)
		}
	}

	return nil
}

func collectTestDeclarations(bundle *testBundle, input *testInput) error {
	bindings, err := testBindings(bundle, input.file.Imports)
	if err != nil {
		return fmt.Errorf("test imports: %w", err)
	}

	for i := range input.file.Decls {
		appendTestDeclaration(bundle, input, &testDeclarationInput{bindings: bindings, index: i})
	}

	bundle.files = append(bundle.files, input.name)

	return nil
}

func collectTestEntry(
	split *packageSplit,
	bundles map[string]*testBundle,
	entry fs.DirEntry,
) error {
	if entry.IsDir() || !strings.HasSuffix(entry.Name(), testFileSuffix) {
		return nil
	}

	err := collectTestFile(split, bundles, entry.Name())
	if err != nil {
		return fmt.Errorf("collect test file: %w", err)
	}

	return nil
}

func collectTestFile(split *packageSplit, bundles map[string]*testBundle, name string) error {
	input, err := readTestInput(split, name)
	if err != nil {
		return fmt.Errorf("load tests: %w", err)
	}

	err = appendTestFile(split, bundles, input)
	if err != nil {
		return fmt.Errorf("append test file: %w", err)
	}

	return nil
}

func conflictsTestAlias(bundle *testBundle, alias, path string) bool {
	if alias == "_" || alias == "." {
		return false
	}

	prior, used := bundle.aliases[alias]

	return used && prior != path
}

func consolidateTests(split *packageSplit, entries []fs.DirEntry) error {
	bundles := map[string]*testBundle{}

	err := collectTestBundles(split, bundles, entries)
	if err != nil {
		return fmt.Errorf("collect tests: %w", err)
	}

	err = writeTestBundles(split, bundles)
	if err != nil {
		return fmt.Errorf(errWriteTests, err)
	}

	return nil
}

func declDoc(decl ast.Decl) *ast.CommentGroup {
	gen, ok := decl.(*ast.GenDecl)

	if ok {
		return gen.Doc
	}

	funcDecl, isFunc := decl.(*ast.FuncDecl)

	if isFunc {
		return funcDecl.Doc
	}

	return nil
}

func declDocStart(fset *token.FileSet, decl ast.Decl, start int) int {
	doc := declDoc(decl)

	if doc == nil {
		return start
	}

	docStart := fset.Position(doc.Pos()).Offset

	if docStart >= zero && docStart < start {
		return docStart
	}

	return start
}

func declarationSource(decls []declEntry) string {
	declarations := make([]string, zero, len(decls))

	for i := range decls {
		declarations = append(declarations, decls[i].source)
	}

	return strings.Join(declarations, newline+newline) + newline
}

func documentTestPackage(split *packageSplit) error {
	if split.pkgDoc != nil || split.pkgName == emptyString {
		return nil
	}

	ensurePackageDoc(split)

	err := writeDocGo(split, split.pkgName, split.pkgDoc)
	if err != nil {
		return fmt.Errorf("document test-only package: %w", err)
	}

	return nil
}

func embedImportDecl() *ast.GenDecl {
	return &ast.GenDecl{
		Tok:    token.IMPORT,
		Lparen: token.Pos(one),
		Rparen: token.Pos(two),
		Specs: []ast.Spec{
			&ast.ImportSpec{
				Path: &ast.BasicLit{
					Kind:     token.STRING,
					Value:    embedImport,
					ValuePos: zero,
					ValueEnd: zero,
				},
				Doc:     nil,
				Name:    nil,
				Comment: nil,
				EndPos:  zero,
			},
		}, Doc: nil, TokPos: zero,
	}
}

func embedImportWith(
	src []byte,
	formatNode func(io.Writer, *token.FileSet, any) error,
) ([]byte, error) {
	set := token.NewFileSet()

	file, err := parser.ParseFile(set, embedGoFake, src, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse for embed: %w", err)
	}

	file.Decls = append([]ast.Decl{embedImportDecl()}, file.Decls...)

	var out bytes.Buffer

	err = formatNode(&out, set, file)
	if err != nil {
		return nil, fmt.Errorf("format embed import: %w", err)
	}

	return out.Bytes(), nil
}

func ensureEmbedImport(src []byte) ([]byte, error) {
	out, err := embedImportWith(src, format.Node)
	if err != nil {
		return nil, fmt.Errorf("embed import: %w", err)
	}

	return out, nil
}

func ensurePackageDoc(split *packageSplit) {
	if split.pkgDoc != nil {
		return
	}

	split.pkgDoc = &ast.CommentGroup{
		List: []*ast.Comment{
			{
				Slash: zero,
				Text:  "// Package " + split.pkgName + " is part of github.com/gostafa/distance.",
			},
		},
	}
}

func extractDeclSource(fset *token.FileSet, src []byte, decl ast.Decl) string {
	start := declDocStart(fset, decl, fset.Position(decl.Pos()).Offset)
	end := fset.Position(decl.End()).Offset

	if start < zero || end > len(src) || start >= end {
		return emptyString
	}

	return strings.TrimRight(string(src[start:end]), newline)
}

func extractGeneratedMarker(src []byte) string {
	for line := range strings.SplitSeq(string(src), newline) {
		marker, done := generatedMarkerLine(line)

		if done {
			return marker
		}
	}

	return emptyString
}

func finishCategoryFile(input *categoryWrite, raw []byte) error {
	err := finishCategoryWith(input, raw, maybeEmbedImport)
	if err != nil {
		return fmt.Errorf(errWrapFinishCategory, err)
	}

	return nil
}

func finishCategoryWith(
	input *categoryWrite,
	raw []byte,
	addEmbed func([]byte, []declEntry, string) ([]byte, error),
) error {
	out, err := imports.Process(input.filename, raw, nil)
	if err != nil {
		return fmt.Errorf("imports %s: %w", input.filename, err)
	}

	out, err = addEmbed(out, input.decls, input.filename)
	if err != nil {
		return fmt.Errorf(errWrapFinishCategory, err)
	}

	formatErr := formatAndWrite(input.store, input.filename, out)
	if formatErr != nil {
		return fmt.Errorf(errWrapFinishCategory, formatErr)
	}

	return nil
}

func finishDecl(input *declCollect, decl ast.Decl, entry declEntry) error {
	err := applyDecl(input, decl, entry)
	if err != nil {
		return fmt.Errorf(errWrapFinishDecl, err)
	}

	return nil
}

func finishTestBundles(split *packageSplit, bundles map[string]*testBundle) error {
	for name := range bundles {
		err := removeOldTests(split, name, bundles[name].files)
		if err != nil {
			return fmt.Errorf("remove old tests: %w", err)
		}
	}

	err := documentTestPackage(split)
	if err != nil {
		return fmt.Errorf("document tests: %w", err)
	}

	return nil
}

func firstNonEmptyCategory(categories []categoryFile) string {
	for i := range categories {
		if len(categories[i].decls) > zero {
			return categories[i].name
		}
	}

	return emptyString
}

func formatAndWrite(writer *packageSplit, name string, src []byte) error {
	src = groupSingleImports(src)

	formatted, err := format.Source(src)
	if err != nil {
		return fmt.Errorf("format %s: %w", name, err)
	}

	formatted = groupSingleImports(formatted)

	writeErr := splitWriteFile(writer, name, formatted)
	if writeErr != nil {
		return fmt.Errorf("splitpkg formatAndWrite: %w", writeErr)
	}

	return nil
}

func genDeclTarget(split *packageSplit, tok token.Token) *[]declEntry {
	if tok == token.CONST {
		return &split.consts
	}

	if tok == token.VAR {
		return &split.vars
	}

	if tok == token.TYPE {
		return &split.types
	}

	return nil
}

func generatedMarkerLine(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)

	if trimmed == emptyString {
		return emptyString, false
	}

	if strings.HasPrefix(trimmed, "// Code generated") {
		return trimmed, true
	}

	return emptyString, strings.HasPrefix(trimmed, "package")
}

func goSourceName(entry fs.DirEntry) (string, bool) {
	name := entry.Name()

	if entry.IsDir() || !strings.HasSuffix(name, goFileSuffix) {
		return emptyString, false
	}

	if strings.HasSuffix(name, testFileSuffix) {
		return emptyString, false
	}

	return name, true
}

// groupSingleImports rewrites bare import "x" into a parenthesized group so
// regenerating packages does not recreate grouper findings.
func groupSingleImports(src []byte) []byte {
	pattern := regexp.MustCompile(`(?m)^import(?:\s+(\w+))?\s+"([^"]+)"\s*$`)

	return pattern.ReplaceAllFunc(src, func(match []byte) []byte {
		sub := pattern.FindSubmatch(match)
		alias, path := sub[1], sub[2]

		if len(alias) == zero {
			return []byte("import (\n\t\"" + string(path) + "\"\n)")
		}

		return []byte("import (\n\t" + string(alias) + " \"" + string(path) + "\"\n)")
	})
}

func hasBuildConstraint(group *ast.CommentGroup) bool {
	for i := range group.List {
		text := group.List[i].Text

		if strings.HasPrefix(text, "//go:build") || strings.HasPrefix(text, "// +build") {
			return true
		}
	}

	return false
}

func identifierRewrite(input *testSelectorInput, identifier *ast.Ident) (testRewrite, bool) {
	alias, exists := input.bindings[identifier.Name]
	start := input.set.Position(identifier.Pos()).Offset - input.start

	return testRewrite{
		value: alias,
		start: start,
		end:   start + len(identifier.Name),
	}, exists &&
		alias != identifier.Name
}

func ingestGoFile(fset *token.FileSet, split *packageSplit, name string) error {
	src, err := splitReadFile(split, name)
	if err != nil {
		return fmt.Errorf(errWrapIngestGoFile, err)
	}

	captureGeneratedMarker(split, src)

	acceptErr := parseAndAccept(fset, &declCollect{split: split, path: name, src: src})
	if acceptErr != nil {
		return fmt.Errorf(errWrapIngestGoFile, acceptErr)
	}

	return nil
}

func isFuncDecl(decl ast.Decl) bool {
	funcDecl, isFunc := decl.(*ast.FuncDecl)

	return isFunc && funcDecl != nil
}

func maybeEmbedImport(out []byte, decls []declEntry, filename string) ([]byte, error) {
	if !needsEmbedImport(decls) || bytes.Contains(out, []byte(embedImport)) {
		return out, nil
	}

	out, err := ensureEmbedImport(out)
	if err != nil {
		return nil, fmt.Errorf("embed import %s: %w", filename, err)
	}

	return out, nil
}

func mergePackageDoc(split *packageSplit, doc *ast.CommentGroup) {
	if doc == nil {
		return
	}

	if split.pkgDoc == nil || len(doc.List) > len(split.pkgDoc.List) {
		split.pkgDoc = doc
	}
}

func needsEmbedImport(decls []declEntry) bool {
	for i := range decls {
		if strings.Contains(decls[i].source, "//go:embed") {
			return true
		}
	}

	return false
}

func newTestBundle(input *testInput) *testBundle {
	return &testBundle{
		packageName: input.file.Name.Name, generatedMarker: extractGeneratedMarker(input.source),
		imports: map[string]string{}, aliases: map[string]string{}, declarations: nil, files: nil,
	}
}

func parseAndAccept(fset *token.FileSet, input *declCollect) error {
	file, err := parseGoFile(fset, input.path, input.src)
	if err != nil {
		return fmt.Errorf(errWrapIngestGoFile, err)
	}

	acceptErr := acceptParsedFile(fset, file, input)
	if acceptErr != nil {
		return fmt.Errorf(errWrapIngestGoFile, acceptErr)
	}

	return nil
}

func parseAndWrite(split *packageSplit) error {
	err := parsePackageFiles(split)
	if err != nil {
		return fmt.Errorf(errWrapSplitPackage, err)
	}

	writeErr := writeSplit(split)
	if writeErr != nil {
		return fmt.Errorf(errWrapSplitPackage, writeErr)
	}

	return nil
}

func parseGoFile(fset *token.FileSet, path string, src []byte) (*ast.File, error) {
	file, err := parser.ParseFile(fset, path, src, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf(errFmtOpPathWrap, opParse, path, err)
	}

	return file, nil
}

func parsePackageFiles(split *packageSplit) error {
	fset := token.NewFileSet()
	names := split.goFiles

	for i := range names {
		err := ingestGoFile(fset, split, names[i])
		if err != nil {
			return fmt.Errorf("splitpkg parsePackageFiles: %w", err)
		}
	}

	ensurePackageDoc(split)

	return nil
}

func preflightEntry(split *packageSplit, input *generatedCheck, entry fs.DirEntry) error {
	if entry.IsDir() || !strings.HasSuffix(entry.Name(), goFileSuffix) {
		return nil
	}

	data, err := splitReadFile(split, entry.Name())
	if err != nil {
		return fmt.Errorf("preflight read: %w", err)
	}

	input.name, input.source = entry.Name(), data

	err = validateSource(input)
	if err != nil {
		return fmt.Errorf("preflight source: %w", err)
	}

	return nil
}

func preflightSources(split *packageSplit, entries []fs.DirEntry) error {
	input := &generatedCheck{
		markers: map[string]bool{},
		seen:    map[string]bool{},
		name:    emptyString,
		source:  nil,
	}

	for i := range entries {
		err := preflightEntry(split, input, entries[i])
		if err != nil {
			return fmt.Errorf("preflight entry: %w", err)
		}
	}

	return nil
}

func prepareTestBundles(
	split *packageSplit,
	bundles map[string]*testBundle,
) (map[string][]byte, error) {
	prepared := map[string][]byte{}

	for name := range bundles {
		source, err := imports.Process(
			filepath.Join(split.dir, name),
			[]byte(buildTestSource(bundles[name])),
			nil,
		)
		if err != nil {
			return nil, fmt.Errorf("format consolidated tests: %w", err)
		}

		prepared[name] = source
	}

	return prepared, nil
}

func readDirectory(file directoryReader) ([]fs.DirEntry, error) {
	entries, readErr := file.ReadDir(-one)
	closeErr := file.Close()

	if readErr != nil {
		return nil, fmt.Errorf(errFmtOpPathWrap, opReaddir, dotName, readErr)
	}

	if closeErr != nil {
		return nil, fmt.Errorf(errFmtOpPathWrap, opClose, dotName, closeErr)
	}

	return entries, nil
}

func readOpenedFile(file io.ReadCloser, name string) ([]byte, error) {
	data, err := io.ReadAll(file)
	closeErr := file.Close()

	if err != nil {
		return nil, fmt.Errorf(errFmtOpPathWrap, opRead, name, err)
	}

	if closeErr != nil {
		return nil, fmt.Errorf(errFmtOpPathWrap, opClose, name, closeErr)
	}

	return data, nil
}

func readRootEntries(root packageFiles) ([]fs.DirEntry, error) {
	file, err := root.Open(dotName)
	if err != nil {
		return nil, fmt.Errorf(errFmtOpPathWrap, opOpen, dotName, err)
	}

	entries, err := readDirectory(file)
	if err != nil {
		return nil, fmt.Errorf("read root entries: %w", err)
	}

	return entries, nil
}

func readTestInput(split *packageSplit, name string) (*testInput, error) {
	source, err := splitReadFile(split, name)
	if err != nil {
		return nil, fmt.Errorf("read tests: %w", err)
	}

	set := token.NewFileSet()

	file, err := parser.ParseFile(set, name, source, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse tests: %w", err)
	}

	return &testInput{name: name, source: source, set: set, file: file}, nil
}

func recordGeneratedCategory(input *generatedCheck, category string) error {
	if category == emptyString {
		return nil
	}

	generated := extractGeneratedMarker(input.source) != emptyString

	if input.seen[category] && input.markers[category] != generated {
		return fmt.Errorf(errSentinelDetails, errMixedGenerated, category)
	}

	input.seen[category] = true
	input.markers[category] = generated

	return nil
}

func removeNonCategoryFiles(split *packageSplit) error {
	allowed := map[string]bool{
		fileDoc: true, fileConsts: true, fileTypes: true, fileVars: true, fileFuncs: true,
	}
	names := split.goFiles

	for i := range names {
		name := names[i]

		if allowed[name] {
			continue
		}

		err := splitRemoveFile(split, name)
		if err != nil {
			return fmt.Errorf("splitpkg remove files: %w", err)
		}
	}

	return nil
}

func removeOldTests(split *packageSplit, target string, files []string) error {
	for i := range files {
		if files[i] == target {
			continue
		}

		err := splitRemoveFile(split, files[i])
		if err != nil {
			return fmt.Errorf("remove old test: %w", err)
		}
	}

	return nil
}

func rewriteTestSelectors(text string, input *testSelectorInput) string {
	edits := []testRewrite{}

	ast.Inspect(input.decl, func(node ast.Node) bool {
		edit, found := selectorRewrite(input, node)

		if found {
			edits = append(edits, edit)
		}

		return true
	})

	return applyTestRewrites(text, edits)
}

func selectorRewrite(input *testSelectorInput, node ast.Node) (testRewrite, bool) {
	selector, ok := node.(*ast.SelectorExpr)

	if !ok {
		return testRewrite{value: emptyString, start: zero, end: zero}, false
	}

	identifier, ok := selector.X.(*ast.Ident)

	if !ok || identifier.Obj != nil {
		return testRewrite{value: emptyString, start: zero, end: zero}, false
	}

	return identifierRewrite(input, identifier)
}

func sourceCategory(name, packageName string, decl ast.Decl) string {
	if strings.HasSuffix(name, testFileSuffix) {
		return testTarget(name, packageName)
	}

	group, ok := decl.(*ast.GenDecl)

	if !ok {
		return "func"
	}

	if group.Tok == token.IMPORT {
		return emptyString
	}

	return group.Tok.String()
}

func splitDirectory(split *packageSplit, entries []fs.DirEntry) error {
	err := preflightSources(split, entries)
	if err != nil {
		return fmt.Errorf("preflight Go sources: %w", err)
	}

	err = splitFromEntries(split, entries)
	if err != nil {
		return fmt.Errorf(errWrapSplitOpened, err)
	}

	return nil
}

func splitFromEntries(split *packageSplit, entries []fs.DirEntry) error {
	if split.dir == emptyString {
		return fmt.Errorf(errFmtOpPathWrap, opDir, emptyString, os.ErrInvalid)
	}

	split.goFiles = collectGoFiles(entries)

	err := splitProduction(split)
	if err != nil {
		return fmt.Errorf(errWrapSplitPackage, err)
	}

	err = consolidateTests(split, entries)
	if err != nil {
		return fmt.Errorf("split tests: %w", err)
	}

	return nil
}

func splitOpened(root packageFiles, dir string) error {
	entries, err := readRootEntries(root)
	if err != nil {
		return fmt.Errorf(errWrapSplitOpened, err)
	}

	split := new(packageSplit)

	split.root, split.dir = root, dir

	err = splitDirectory(split, entries)
	if err != nil {
		return fmt.Errorf(errWrapSplitOpened, err)
	}

	return nil
}

func splitProduction(split *packageSplit) error {
	if len(split.goFiles) == zero {
		return nil
	}

	err := parseAndWrite(split)
	if err != nil {
		return fmt.Errorf(errWrapSplitPackage, err)
	}

	return nil
}

func splitReadFile(split *packageSplit, name string) ([]byte, error) {
	file, err := split.root.Open(name)
	if err != nil {
		return nil, fmt.Errorf(errFmtOpPathWrap, opOpen, name, err)
	}

	data, readErr := readOpenedFile(file, name)
	if readErr != nil {
		return nil, fmt.Errorf("splitpkg ReadFile: %w", readErr)
	}

	return data, nil
}

func splitRemoveFile(split *packageSplit, name string) error {
	err := split.root.Remove(name)
	if err != nil {
		return fmt.Errorf(errFmtOpPathWrap, opRemove, name, err)
	}

	return nil
}

func splitWriteFile(split *packageSplit, name string, data []byte) error {
	file, err := split.root.Create(name)
	if err != nil {
		return fmt.Errorf(errFmtOpPathWrap, opCreate, name, err)
	}

	writeErr := writeOpenedFile(file, name, data)
	if writeErr != nil {
		return fmt.Errorf("splitpkg WriteFile: %w", writeErr)
	}

	return nil
}

func testBindings(bundle *testBundle, specifications []*ast.ImportSpec) (map[string]string, error) {
	bindings := map[string]string{}

	for i := range specifications {
		err := bindTestImport(bundle, bindings, specifications[i])
		if err != nil {
			return nil, fmt.Errorf("bind test import: %w", err)
		}
	}

	return bindings, nil
}

func testBundleFor(bundles map[string]*testBundle, input *testInput) (*testBundle, error) {
	target := testTarget(input.name, input.file.Name.Name)
	bundle := bundles[target]

	if bundle == nil {
		bundle = newTestBundle(input)
		bundles[target] = bundle
	}

	if bundle.packageName != input.file.Name.Name {
		return nil, fmt.Errorf(errSentinelDetails, errTestPackage, target)
	}

	return bundle, nil
}

func testDeclarationSource(input *testInput, declaration *testDeclarationInput) string {
	decl := input.file.Decls[declaration.index]
	text := extractDeclSource(input.set, input.source, decl)
	start := declDocStart(input.set, decl, input.set.Position(decl.Pos()).Offset)

	return rewriteTestSelectors(
		text,
		&testSelectorInput{
			set:      input.set,
			decl:     decl,
			bindings: declaration.bindings,
			start:    start,
		},
	)
}

func testImportAlias(specification *ast.ImportSpec, path string) string {
	if specification.Name != nil {
		return specification.Name.Name
	}

	return filepath.Base(path)
}

func testTarget(name, packageName string) string {
	base := strings.TrimSuffix(packageName, testPackageSuffix)

	if strings.Contains(name, "bench") {
		return base + "_bench_test.go"
	}

	if packageName != base {
		return base + "_integration_test.go"
	}

	return base + testFileSuffix
}

func validateSource(input *generatedCheck) error {
	err := checkSourceConstraints(input.name, input.source)
	if err != nil {
		return fmt.Errorf("source constraints: %w", err)
	}

	err = checkGeneratedCategories(input)
	if err != nil {
		return fmt.Errorf("generated constraints: %w", err)
	}

	return nil
}

func writeCategories(split *packageSplit) error {
	categories := categoryList(split)
	firstCategory := firstNonEmptyCategory(categories)

	for i := range categories {
		err := writeOneCategory(split, &categories[i], firstCategory)
		if err != nil {
			return fmt.Errorf("splitpkg writeCategories: %w", err)
		}
	}

	return nil
}

func writeCategoryGo(input *categoryWrite) error {
	err := finishCategoryFile(input, buildCategorySource(input))
	if err != nil {
		return fmt.Errorf(errWrapWriteCategoryGo, err)
	}

	return nil
}

func writeDocGo(writer *packageSplit, pkgName string, doc *ast.CommentGroup) error {
	err := formatAndWrite(writer, fileDoc, buildDocSource(pkgName, doc))
	if err != nil {
		return fmt.Errorf(errWrapWriteDocGo, err)
	}

	return nil
}

func writeOneCategory(split *packageSplit, category *categoryFile, firstCategory string) error {
	if len(category.decls) == zero {
		return nil
	}

	err := writeCategoryGo(&categoryWrite{
		store:           split,
		pkgName:         split.pkgName,
		filename:        category.name,
		decls:           category.decls,
		generatedMarker: categoryMarkerValue(split.generatedMarker, category.name, firstCategory),
	})
	if err != nil {
		return fmt.Errorf("splitpkg write category: %w", err)
	}

	return nil
}

func writeOpenedFile(file io.WriteCloser, name string, data []byte) error {
	written, err := file.Write(data)
	closeErr := file.Close()

	if err != nil {
		return fmt.Errorf(errFmtOpPathWrap, opWrite, name, err)
	}

	if written != len(data) {
		return fmt.Errorf(errFmtOpPathWrap, opWrite, name, errShortWrite)
	}

	if closeErr != nil {
		return fmt.Errorf(errFmtOpPathWrap, opClose, name, closeErr)
	}

	return nil
}

func writePreparedTests(split *packageSplit, prepared map[string][]byte) error {
	for name := range prepared {
		err := splitWriteFile(split, name, prepared[name])
		if err != nil {
			return fmt.Errorf("write consolidated tests: %w", err)
		}
	}

	return nil
}

func writeSplit(split *packageSplit) error {
	err := writeDocGo(split, split.pkgName, split.pkgDoc)
	if err != nil {
		return fmt.Errorf(errWrapWriteSplit, err)
	}

	err = writeCategories(split)
	if err != nil {
		return fmt.Errorf(errWrapWriteSplit, err)
	}

	removeErr := removeNonCategoryFiles(split)
	if removeErr != nil {
		return fmt.Errorf(errWrapWriteSplit, removeErr)
	}

	return nil
}

func writeTestBundles(split *packageSplit, bundles map[string]*testBundle) error {
	prepared, err := prepareTestBundles(split, bundles)
	if err != nil {
		return fmt.Errorf("prepare tests: %w", err)
	}

	err = writePreparedTests(split, prepared)
	if err != nil {
		return fmt.Errorf(errWriteTests, err)
	}

	err = finishTestBundles(split, bundles)
	if err != nil {
		return fmt.Errorf("finish tests: %w", err)
	}

	return nil
}
