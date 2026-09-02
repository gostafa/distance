// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"os"
	"slices"
	"strings"
)

func (err *pathError) Error() string {
	return err.op + " " + err.path + ": " + err.err.Error()
}

func (err *pathError) Unwrap() error {
	return err.err
}

func (split *packageSplit) Marker() string {
	return split.generatedMarker
}

func markerOf(holder markerHolder) string {
	return holder.Marker()
}

func (split *packageSplit) Dir() string {
	return split.dir
}

func (split *packageSplit) GoFiles() []string {
	return split.goFiles
}

func (split *packageSplit) ReadFile(name string) ([]byte, error) {
	file, err := split.root.Open(name)
	if err != nil {
		return nil, &pathError{op: opOpen, path: name, err: err}
	}

	data, readErr := readOpenedFile(file, name)
	if readErr != nil {
		return nil, fmt.Errorf("splitpkg ReadFile: %w", readErr)
	}

	return data, nil
}

func (split *packageSplit) WriteFile(name string, data []byte) error {
	file, err := split.root.Create(name)
	if err != nil {
		return &pathError{op: "create", path: name, err: err}
	}

	writeErr := writeOpenedFile(file, name, data)
	if writeErr != nil {
		return fmt.Errorf("splitpkg WriteFile: %w", writeErr)
	}

	return nil
}

func (split *packageSplit) RemoveFile(name string) error {
	err := split.root.Remove(name)
	if err != nil {
		return &pathError{op: "remove", path: name, err: err}
	}

	return nil
}

func readNamed(reader fileReader, name string) ([]byte, error) {
	data, err := reader.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("splitpkg readNamed: %w", err)
	}

	return data, nil
}

func writeNamed(writer fileWriter, name string, data []byte) error {
	err := writer.WriteFile(name, data)
	if err != nil {
		return fmt.Errorf("splitpkg writeNamed: %w", err)
	}

	return nil
}

func removeNamed(remover fileRemover, name string) error {
	err := remover.RemoveFile(name)
	if err != nil {
		return fmt.Errorf("splitpkg removeNamed: %w", err)
	}

	return nil
}

func storeDir(holder dirHolder) string {
	return holder.Dir()
}

func fileNames(lister sourceLister) []string {
	return lister.GoFiles()
}

func readOpenedFile(file *os.File, name string) ([]byte, error) {
	data, err := io.ReadAll(file)
	closeErr := file.Close()

	if err != nil {
		return nil, &pathError{op: "read", path: name, err: err}
	}

	if closeErr != nil {
		return nil, &pathError{op: opClose, path: name, err: closeErr}
	}

	return data, nil
}

func writeOpenedFile(file *os.File, name string, data []byte) error {
	written, err := file.Write(data)
	closeErr := file.Close()

	if err != nil {
		return &pathError{op: opWrite, path: name, err: err}
	}

	if written != len(data) {
		return &pathError{op: opWrite, path: name, err: errShortWrite}
	}

	if closeErr != nil {
		return &pathError{op: opClose, path: name, err: closeErr}
	}

	return nil
}

func splitPackage(dir string) error {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return &pathError{op: opOpen, path: dir, err: err}
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
		return &pathError{op: opClose, path: dotName, err: closeErr}
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
		return nil, &pathError{op: opOpen, path: dotName, err: err}
	}

	entries, readErr := file.ReadDir(-one)
	closeErr := file.Close()

	if readErr != nil {
		return nil, &pathError{op: "readdir", path: dotName, err: readErr}
	}

	if closeErr != nil {
		return nil, &pathError{op: opClose, path: dotName, err: closeErr}
	}

	return entries, nil
}

func splitFromEntries(split *packageSplit, entries []fs.DirEntry) error {
	if storeDir(split) == emptyString {
		return &pathError{op: "dir", path: emptyString, err: os.ErrInvalid}
	}

	split.goFiles = collectGoFiles(entries)

	if len(fileNames(split)) == zero {
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
	names := fileNames(split)

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
	src, err := readNamed(split, name)
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
		return nil, &pathError{op: "parse", path: path, err: err}
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
		return &pathError{op: "package-name", path: path}
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
	gen, isGen := decl.(*ast.GenDecl)
	err := appendFuncDecl(input, decl, entry)

	if isGen {
		err = collectGenDecl(input, gen.Tok, entry)
	}

	if err != nil {
		return fmt.Errorf("splitpkg finishDecl: %w", err)
	}

	return nil
}

func appendFuncDecl(input *declCollect, decl ast.Decl, entry declEntry) error {
	funcDecl, isFunc := decl.(*ast.FuncDecl)

	if !isFunc || funcDecl == nil {
		return &pathError{op: "decl", path: input.path}
	}

	input.split.funcs = append(input.split.funcs, entry)

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

	return &pathError{op: "gen-decl", path: input.path}
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

func ensurePackageDoc(split *packageSplit) {
	if split.pkgDoc != nil {
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
