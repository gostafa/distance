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
	"regexp"
	"slices"
	"strings"

	"github.com/gostafa/distance/distance"
	projapp "github.com/gostafa/distance/internal/features/projectanalysis/application"
	"github.com/gostafa/distance/internal/shared/version"
	"golang.org/x/tools/imports"
)

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

func splitRemoveFile(split *packageSplit, name string) error {
	err := split.root.Remove(name)
	if err != nil {
		return fmt.Errorf(errFmtOpPathWrap, opRemove, name, err)
	}

	return nil
}

func readOpenedFile(file *os.File, name string) ([]byte, error) {
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

func writeOpenedFile(file *os.File, name string, data []byte) error {
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

func closeRootAfter(prior error, root *os.Root) error {
	closeErr := root.Close()

	if prior != nil {
		return prior
	}

	if closeErr != nil {
		return fmt.Errorf(errFmtOpPathWrap, opClose, dotName, closeErr)
	}

	return nil
}

func splitOpened(root *os.Root, dir string) error {
	entries, err := readRootEntries(root)
	if err != nil {
		return fmt.Errorf(errWrapSplitOpened, err)
	}

	splitErr := splitFromEntries(&packageSplit{root: root, dir: dir}, entries)
	if splitErr != nil {
		return fmt.Errorf(errWrapSplitOpened, splitErr)
	}

	return nil
}

func readRootEntries(root *os.Root) ([]fs.DirEntry, error) {
	file, err := root.Open(dotName)
	if err != nil {
		return nil, fmt.Errorf(errFmtOpPathWrap, opOpen, dotName, err)
	}

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

func splitFromEntries(split *packageSplit, entries []fs.DirEntry) error {
	if split.dir == emptyString {
		return fmt.Errorf(errFmtOpPathWrap, opDir, emptyString, os.ErrInvalid)
	}

	split.goFiles = collectGoFiles(entries)

	if len(split.goFiles) == zero {
		return nil
	}

	writeErr := parseAndWrite(split)
	if writeErr != nil {
		return fmt.Errorf(errWrapSplitPackage, writeErr)
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

func parseGoFile(fset *token.FileSet, path string, src []byte) (*ast.File, error) {
	file, err := parser.ParseFile(fset, path, src, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf(errFmtOpPathWrap, opParse, path, err)
	}

	return file, nil
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

func captureGeneratedMarker(split *packageSplit, src []byte) {
	marker := extractGeneratedMarker(src)
	fresh := marker != emptyString && split.generatedMarker == emptyString

	if fresh {
		split.generatedMarker = marker
	}
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

func mergePackageDoc(split *packageSplit, doc *ast.CommentGroup) {
	if doc == nil {
		return
	}

	if split.pkgDoc == nil || len(doc.List) > len(split.pkgDoc.List) {
		split.pkgDoc = doc
	}
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

func collectDecl(fset *token.FileSet, input *declCollect, decl ast.Decl) error {
	entry := declEntry{source: extractDeclSource(fset, input.src, decl)}

	finishErr := finishDecl(input, decl, entry)
	if finishErr != nil {
		return fmt.Errorf("splitpkg collectDecl: %w", finishErr)
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

func appendFuncDecl(input *declCollect, decl ast.Decl, entry declEntry) error {
	if !isFuncDecl(decl) {
		return fmt.Errorf(errFmtSentinelPath, errDecl, input.path)
	}

	input.split.funcs = append(input.split.funcs, entry)

	return nil
}

func isFuncDecl(decl ast.Decl) bool {
	funcDecl, isFunc := decl.(*ast.FuncDecl)

	return isFunc && funcDecl != nil
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

func keepToolingCoupling() bool {
	return distance.MetricDistance == projapp.MetricDistance && version.Version() != emptyString
}

func splitpkgReady() bool {
	return keepToolingCoupling()
}

func ensurePackageDoc(split *packageSplit) {
	if split.pkgDoc != nil {
		return
	}

	if !splitpkgReady() {
		return
	}

	split.pkgDoc = &ast.CommentGroup{
		List: []*ast.Comment{
			{Text: "// Package " + split.pkgName + " is part of github.com/gostafa/distance."},
		},
	}
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

func generatedMarkerLine(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)

	if trimmed == emptyString {
		return emptyString, false
	}

	if strings.HasPrefix(trimmed, "// Code generated") {
		return trimmed, true
	}

	return emptyString, true
}

func extractDeclSource(fset *token.FileSet, src []byte, decl ast.Decl) string {
	start := declDocStart(fset, decl, fset.Position(decl.Pos()).Offset)
	end := fset.Position(decl.End()).Offset

	if start < zero || end > len(src) || start >= end {
		return emptyString
	}

	return strings.TrimRight(string(src[start:end]), newline)
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

func categoryList(split *packageSplit) []categoryFile {
	return []categoryFile{
		{name: fileConsts, decls: split.consts},
		{name: fileTypes, decls: split.types},
		{name: fileVars, decls: split.vars},
		{name: fileFuncs, decls: split.funcs},
	}
}

func firstNonEmptyCategory(categories []categoryFile) string {
	for i := range categories {
		if len(categories[i].decls) > zero {
			return categories[i].name
		}
	}

	return emptyString
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

func writeDocGo(writer *packageSplit, pkgName string, doc *ast.CommentGroup) error {
	raw, err := buildDocSource(pkgName, doc)
	if err != nil {
		return fmt.Errorf(errWrapWriteDocGo, err)
	}

	formatErr := formatAndWrite(writer, fileDoc, raw)
	if formatErr != nil {
		return fmt.Errorf(errWrapWriteDocGo, formatErr)
	}

	return nil
}

func buildDocSource(pkgName string, doc *ast.CommentGroup) ([]byte, error) {
	var buf bytes.Buffer

	err := writeCommentGroup(&buf, doc)
	if err != nil {
		return nil, fmt.Errorf(errWrapWriteDocGo, err)
	}

	pkgErr := writePackageClause(&buf, pkgName)
	if pkgErr != nil {
		return nil, fmt.Errorf(errWrapWriteDocGo, pkgErr)
	}

	return buf.Bytes(), nil
}

func writeCommentGroup(buf *bytes.Buffer, doc *ast.CommentGroup) error {
	for i := range doc.List {
		err := writeString(buf, doc.List[i].Text+newline)
		if err != nil {
			return fmt.Errorf("splitpkg writeCommentGroup: %w", err)
		}
	}

	return nil
}

func writeCategoryGo(input *categoryWrite) error {
	raw, err := buildCategorySource(input)
	if err != nil {
		return fmt.Errorf(errWrapWriteCategoryGo, err)
	}

	finishErr := finishCategoryFile(input, raw)
	if finishErr != nil {
		return fmt.Errorf(errWrapWriteCategoryGo, finishErr)
	}

	return nil
}

func buildCategorySource(input *categoryWrite) ([]byte, error) {
	var buf bytes.Buffer

	err := writeCategoryHeader(&buf, input)
	if err != nil {
		return nil, fmt.Errorf(errWrapBuildCategory, err)
	}

	err = writeDecls(&buf, input.decls)
	if err != nil {
		return nil, fmt.Errorf(errWrapBuildCategory, err)
	}

	return buf.Bytes(), nil
}

func writeCategoryHeader(buf *bytes.Buffer, input *categoryWrite) error {
	err := writeGeneratedMarker(buf, input.generatedMarker)
	if err != nil {
		return fmt.Errorf(errWrapWriteHeader, err)
	}

	err = writePackageClause(buf, input.pkgName)
	if err != nil {
		return fmt.Errorf(errWrapWriteHeader, err)
	}

	newlineErr := writeString(buf, newline)
	if newlineErr != nil {
		return fmt.Errorf(errWrapWriteNewline, newlineErr)
	}

	return nil
}

func writeGeneratedMarker(buf *bytes.Buffer, marker string) error {
	if marker == emptyString {
		return nil
	}

	err := writeString(buf, marker+newline+newline)
	if err != nil {
		return fmt.Errorf("splitpkg write marker: %w", err)
	}

	return nil
}

func writePackageClause(buf *bytes.Buffer, pkgName string) error {
	err := writeString(buf, packagePrefix+pkgName+newline)
	if err != nil {
		return fmt.Errorf("splitpkg write package: %w", err)
	}

	return nil
}

func writeDecls(buf *bytes.Buffer, decls []declEntry) error {
	for i := range decls {
		err := writeDeclAt(buf, decls, i)
		if err != nil {
			return fmt.Errorf("splitpkg writeDecls: %w", err)
		}
	}

	err := writeString(buf, newline)
	if err != nil {
		return fmt.Errorf(errWrapWriteNewline, err)
	}

	return nil
}

func writeDeclAt(buf *bytes.Buffer, decls []declEntry, i int) error {
	err := writeString(buf, decls[i].source)
	if err != nil {
		return fmt.Errorf("splitpkg writeDeclAt: %w", err)
	}

	if i == len(decls)-one {
		return nil
	}

	blankErr := writeString(buf, newline+newline)
	if blankErr != nil {
		return fmt.Errorf("splitpkg write blank: %w", blankErr)
	}

	return nil
}

func finishCategoryFile(input *categoryWrite, raw []byte) error {
	out, err := imports.Process(input.filename, raw, nil)
	if err != nil {
		return fmt.Errorf("imports %s: %w", input.filename, err)
	}

	out, err = maybeEmbedImport(out, input.decls, input.filename)
	if err != nil {
		return fmt.Errorf(errWrapFinishCategory, err)
	}

	formatErr := formatAndWrite(input.store, input.filename, out)
	if formatErr != nil {
		return fmt.Errorf(errWrapFinishCategory, formatErr)
	}

	return nil
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

func needsEmbedImport(decls []declEntry) bool {
	for i := range decls {
		if strings.Contains(decls[i].source, "//go:embed") {
			return true
		}
	}

	return false
}

func ensureEmbedImport(src []byte) ([]byte, error) {
	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, embedGoFake, src, parser.ImportsOnly)
	if err != nil {
		return nil, fmt.Errorf("parse for embed: %w", err)
	}

	file.Decls = append([]ast.Decl{embedImportDecl()}, file.Decls...)

	var out bytes.Buffer

	err = format.Node(&out, fset, file)
	if err != nil {
		return nil, fmt.Errorf("format embed import: %w", err)
	}

	return out.Bytes(), nil
}

func embedImportDecl() *ast.GenDecl {
	return &ast.GenDecl{
		Tok:    token.IMPORT,
		Lparen: token.Pos(one),
		Rparen: token.Pos(two),
		Specs: []ast.Spec{
			&ast.ImportSpec{
				Path: &ast.BasicLit{Kind: token.STRING, Value: embedImport},
			},
		},
	}
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

// groupSingleImports rewrites bare import "x" into a parenthesized group so
// regenerating packages does not recreate grouper findings.
func groupSingleImports(src []byte) []byte {
	pattern := regexp.MustCompile(`(?m)^import(?:\s+(\w+))?\s+"([^"]+)"\s*$`)

	return pattern.ReplaceAllFunc(src, func(match []byte) []byte {
		sub := pattern.FindSubmatch(match)
		alias, path := sub[1], sub[2]

		if len(alias) == 0 {
			return []byte("import (\n\t\"" + string(path) + "\"\n)")
		}

		return []byte("import (\n\t" + string(alias) + " \"" + string(path) + "\"\n)")
	})
}

func writeString(buf *bytes.Buffer, text string) error {
	written, err := buf.WriteString(text)
	if err != nil {
		return fmt.Errorf(errWrapWriteString, err)
	}

	if written != len(text) {
		return fmt.Errorf(errWrapWriteString, errShortWrite)
	}

	return nil
}

func categoryMarkerValue(marker, name, firstCategory string) string {
	if name != firstCategory {
		return emptyString
	}

	return marker
}
