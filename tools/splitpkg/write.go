// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"regexp"
	"strings"

	"golang.org/x/tools/imports"
)

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
		generatedMarker: categoryMarker(split, category.name, firstCategory),
	})
	if err != nil {
		return fmt.Errorf("splitpkg write category: %w", err)
	}

	return nil
}

func categoryMarker(split *packageSplit, name, firstCategory string) string {
	if name != firstCategory {
		return emptyString
	}

	return markerOf(split)
}

func removeNonCategoryFiles(split *packageSplit) error {
	allowed := map[string]bool{
		fileDoc: true, fileConsts: true, fileTypes: true, fileVars: true, fileFuncs: true,
	}
	names := fileNames(split)

	for i := range names {
		name := names[i]

		if allowed[name] {
			continue
		}

		err := removeNamed(split, name)
		if err != nil {
			return fmt.Errorf("splitpkg remove files: %w", err)
		}
	}

	return nil
}

func writeDocGo(writer fileWriter, pkgName string, doc *ast.CommentGroup) error {
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

func formatAndWrite(writer fileWriter, name string, src []byte) error {
	src = groupSingleImports(src)

	formatted, err := format.Source(src)
	if err != nil {
		return fmt.Errorf("format %s: %w", name, err)
	}

	formatted = groupSingleImports(formatted)

	writeErr := writeNamed(writer, name, formatted)
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
