// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package splitpkg

import (
	"bytes"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type failingStream struct {
	readErr, writeErr, closeErr error
	short                       bool
	closed                      bool
}

func (s *failingStream) Read([]byte) (int, error) { return 0, s.readErr }

func (s *failingStream) Write(payload []byte) (int, error) {
	if s.short {
		return 0, nil
	}
	return len(payload), s.writeErr
}

func (s *failingStream) Close() error { s.closed = true; return s.closeErr }

func (s *failingStream) ReadDir(int) ([]fs.DirEntry, error) { return nil, s.readErr }

func openSplitRoot(t *testing.T) *packageSplit {
	t.Helper()
	directory := t.TempDir()
	root, err := os.OpenRoot(directory)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = root.Close() })
	return &packageSplit{root: root, dir: directory, pkgName: "sample"}
}

func requireError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected failure")
	}
}

func TestStreamLifecycleFailureMapping(t *testing.T) {
	failure := errors.New("I/O failed")
	for _, stream := range []*failingStream{{readErr: failure}, {readErr: io.EOF, closeErr: failure}} {
		_, err := readOpenedFile(stream, "input.go")
		if !errors.Is(err, failure) || !stream.closed {
			t.Fatalf("read failure = %v closed=%v", err, stream.closed)
		}
	}
	for _, stream := range []*failingStream{{writeErr: failure}, {closeErr: failure}, {short: true}} {
		err := writeOpenedFile(stream, "output.go", []byte("output"))
		requireError(t, err)
		if !stream.closed {
			t.Fatal("destination was not closed")
		}
	}
	if err := writeOpenedFile(&failingStream{}, "output.go", []byte("output")); err != nil {
		t.Fatal(err)
	}
	for _, stream := range []*failingStream{{readErr: failure}, {closeErr: failure}} {
		_, err := readDirectory(stream)
		if !errors.Is(err, failure) || !stream.closed {
			t.Fatalf("directory failure = %v closed=%v", err, stream.closed)
		}
	}
	if _, err := readDirectory(&failingStream{}); err != nil {
		t.Fatal(err)
	}
	if err := closeRootAfter(nil, &failingStream{closeErr: failure}); !errors.Is(err, failure) {
		t.Fatalf("root close = %v", err)
	}
	if err := closeRootAfter(failure, &failingStream{}); !errors.Is(err, failure) {
		t.Fatalf("prior failure = %v", err)
	}
}

func TestParserAndDeclarationFailures(t *testing.T) {
	split := openSplitRoot(t)
	set := token.NewFileSet()
	input := &declCollect{split: split, path: "input.go", src: []byte("package other\n")}
	requireError(t, parseAndAccept(set, input))
	input.src = []byte("package invalid{")
	requireError(t, parseAndAccept(set, input))
	bad := &ast.BadDecl{}
	if got := extractDeclSource(set, nil, bad); got != "" {
		t.Fatalf("invalid span = %q", got)
	}
	if doc := declDoc(bad); doc != nil {
		t.Fatalf("bad declaration doc = %v", doc)
	}
	input.src = nil
	file := &ast.File{Name: ast.NewIdent("sample"), Decls: []ast.Decl{bad}}
	requireError(t, acceptParsedFile(set, file, input))
	requireError(t, collectDecls(set, input, []ast.Decl{bad}))
	requireError(t, collectGenDecl(input, token.ILLEGAL, declEntry{}))
	if err := collectGenDecl(input, token.IMPORT, declEntry{}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(split.dir, "broken.go"),
		[]byte("package invalid{"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	split.goFiles = []string{"broken.go"}
	requireError(t, parseAndWrite(split))
	requireError(t, SplitPackage(split.dir))
	check := &generatedCheck{
		name:    "broken.go",
		source:  []byte("package invalid{"),
		markers: map[string]bool{},
		seen:    map[string]bool{},
	}
	requireError(t, checkGeneratedCategories(check))
	requireError(t, checkSourceConstraints(check.name, check.source))
}

func TestFileOperationFailures(t *testing.T) {
	split := openSplitRoot(t)
	if err := os.Mkdir(filepath.Join(split.dir, "directory.go"), 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := splitReadFile(split, "directory.go"); err == nil {
		t.Fatal("directory read succeeded")
	}
	requireError(t, splitRemoveFile(split, "missing"))
	if err := split.root.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := readRootEntries(split.root); err == nil {
		t.Fatal("closed root read succeeded")
	}
	requireError(t, splitOpened(split.root, split.dir))
	requireError(t, splitWriteFile(split, "output.go", []byte("package sample")))
	requireError(t, ingestGoFile(token.NewFileSet(), split, "missing.go"))
	if _, err := readTestInput(split, "missing_test.go"); err == nil {
		t.Fatal("closed test read succeeded")
	}
	requireError(t, collectTestFile(split, map[string]*testBundle{}, "missing_test.go"))
	requireError(t, removeOldTests(split, "sample_test.go", []string{"old_test.go"}))
	requireError(
		t,
		finishTestBundles(
			split,
			map[string]*testBundle{"sample_test.go": {files: []string{"old_test.go"}}},
		),
	)
	split.pkgDoc = nil
	requireError(t, documentTestPackage(split))
	split.pkgDoc = nil
	requireError(t, finishTestBundles(split, map[string]*testBundle{}))
	requireError(t, writeSplit(split))
	requireError(t, formatAndWrite(split, "doc.go", []byte("invalid")))
	requireError(t, formatAndWrite(split, "doc.go", []byte("package sample\n")))
	requireError(
		t,
		writeTestBundles(
			split,
			map[string]*testBundle{
				"sample_test.go": {
					packageName:  "sample",
					imports:      map[string]string{},
					declarations: []string{"func TestValid() {}"},
				},
			},
		),
	)
}

func TestConsolidationRejectsInvalidImportsAndPackages(t *testing.T) {
	bundle := &testBundle{
		packageName: "sample",
		imports:     map[string]string{},
		aliases:     map[string]string{},
	}
	if _, err := testBindings(
		bundle,
		[]*ast.ImportSpec{{Path: &ast.BasicLit{Value: "invalid"}}},
	); err == nil {
		t.Fatal("accepted invalid import literal")
	}
	if _, err := canonicalTestAlias(bundle, "fmt", "f"); err != nil {
		t.Fatal(err)
	}
	if _, err := canonicalTestAlias(bundle, "io", "f"); err == nil {
		t.Fatal("accepted conflicting aliases")
	}
	if _, err := canonicalTestAlias(bundle, "io", "_"); err != nil {
		t.Fatal(err)
	}
	input := &testInput{name: "bench_test.go", file: &ast.File{Name: ast.NewIdent("sample_test")}}
	if _, err := testBundleFor(
		map[string]*testBundle{"sample_bench_test.go": bundle},
		input,
	); err == nil {
		t.Fatal("accepted mixed benchmark packages")
	}
	split := openSplitRoot(t)
	set := token.NewFileSet()
	file, err := parser.ParseFile(set, "source_test.go", "package other\n", 0)
	if err != nil {
		t.Fatal(err)
	}
	requireError(
		t,
		appendTestFile(
			split,
			map[string]*testBundle{},
			&testInput{set: set, file: file, name: "source_test.go"},
		),
	)
	file = &ast.File{
		Name: ast.NewIdent("sample"),
		Imports: []*ast.ImportSpec{
			{Name: ast.NewIdent("f"), Path: &ast.BasicLit{Value: `"strings"`}},
		},
	}
	requireError(t, collectTestDeclarations(bundle, &testInput{set: set, file: file}))
	malformed := map[string]*testBundle{
		"sample_test.go": {
			packageName:  "sample",
			imports:      map[string]string{},
			declarations: []string{"invalid source"},
		},
	}
	requireError(t, writeTestBundles(split, malformed))
	if err := os.WriteFile(
		filepath.Join(split.dir, "source_test.go"),
		[]byte("package invalid{"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(split.dir)
	if err != nil {
		t.Fatal(err)
	}
	requireError(t, consolidateTests(split, entries))
	requireError(t, splitFromEntries(split, entries))
}

func TestWriteFailuresPreserveOriginalSources(t *testing.T) {
	split := openSplitRoot(t)
	split.pkgDoc = &ast.CommentGroup{
		List: []*ast.Comment{{Text: "// Package sample exercises safe rewriting."}},
	}
	split.funcs = []declEntry{{source: "func Example() {}"}}
	if err := os.Mkdir(filepath.Join(split.dir, fileFuncs), 0o700); err != nil {
		t.Fatal(err)
	}
	requireError(t, writeSplit(split))
	requireError(
		t,
		writeCategoryGo(
			&categoryWrite{
				store:    split,
				pkgName:  "sample",
				filename: fileFuncs,
				decls:    split.funcs,
			},
		),
	)
	if err := split.root.Remove(fileFuncs); err != nil {
		t.Fatal(err)
	}
	split.types = []declEntry{{source: "type invalid{"}}
	requireError(t, writeSplit(split))
	split.goFiles = []string{fileFuncs}
	requireError(t, splitProduction(split))
	split.types = nil
	split.funcs = nil
	if err := os.Mkdir(filepath.Join(split.dir, "old.go"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(split.dir, "old.go", "keep"),
		[]byte("original"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	split.goFiles = []string{"old.go"}
	requireError(t, writeSplit(split))
	if _, err := os.Stat(filepath.Join(split.dir, "old.go", "keep")); err != nil {
		t.Fatalf("lost original file: %v", err)
	}
	split.dir = ""
	requireError(t, splitFromEntries(split, nil))
}

func TestGeneratedHeadersAndEmbedImports(t *testing.T) {
	marker := "// Code generated by a tool. DO NOT EDIT."
	source := []byte(goHeader + marker + "\npackage sample\n")
	if got := extractGeneratedMarker(source); got != marker {
		t.Fatalf("marker after header = %q", got)
	}
	if got := categoryMarkerValue(marker, fileVars, fileVars); got != marker {
		t.Fatalf("category marker = %q", got)
	}
	if got := firstNonEmptyCategory(nil); got != "" {
		t.Fatalf("empty category = %q", got)
	}
	check := &generatedCheck{
		name:    "source.go",
		source:  source,
		seen:    map[string]bool{"var": true},
		markers: map[string]bool{"var": false},
	}
	requireError(t, recordGeneratedCategory(check, "var"))
	if err := recordGeneratedCategory(check, ""); err != nil {
		t.Fatal(err)
	}
	raw := []byte("package sample\n//go:embed input.txt\nvar Payload string\n")
	decls := []declEntry{{source: "//go:embed input.txt\nvar Payload string"}}
	out, err := maybeEmbedImport(raw, decls, fileVars)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(out, []byte(`"embed"`)) ||
		!bytes.Contains(out, []byte("var Payload string")) {
		t.Fatalf("missing embed import: %s", out)
	}
	if _, err := maybeEmbedImport(out, decls, fileVars); err != nil {
		t.Fatal(err)
	}
	if _, err := maybeEmbedImport([]byte("invalid"), decls, fileVars); err == nil {
		t.Fatal("accepted invalid embed source")
	}
	split := openSplitRoot(t)
	if err := writeDocGo(
		split,
		"sample",
		&ast.CommentGroup{List: []*ast.Comment{{Text: "// Package sample documents the package."}}},
	); err != nil {
		t.Fatal(err)
	}
	if err := finishCategoryFile(
		&categoryWrite{store: split, filename: fileVars, decls: decls},
		raw,
	); err != nil {
		t.Fatal(err)
	}
	doc, err := os.ReadFile(filepath.Join(split.dir, fileDoc))
	if err != nil || !strings.Contains(string(doc), "documents the package") {
		t.Fatalf("package documentation: %s err=%v", doc, err)
	}
}

type testPackageFiles struct {
	root                 packageFiles
	openFile, createFile *os.File
}

func (f testPackageFiles) Open(name string) (*os.File, error) {
	if f.openFile != nil {
		return f.openFile, nil
	}
	return f.root.Open(name)
}

func (f testPackageFiles) Create(name string) (*os.File, error) {
	if f.createFile != nil {
		return f.createFile, nil
	}
	return f.root.Create(name)
}

func (f testPackageFiles) Remove(name string) error { return f.root.Remove(name) }

func (f testPackageFiles) Close() error { return f.root.Close() }

func TestFinalSplitterBoundaries(t *testing.T) {
	split := openSplitRoot(t)
	closed, err := os.CreateTemp(t.TempDir(), "closed")
	if err != nil {
		t.Fatal(err)
	}
	if err := closed.Close(); err != nil {
		t.Fatal(err)
	}
	fake := testPackageFiles{root: split.root, openFile: closed, createFile: closed}
	if _, err := readRootEntries(fake); err == nil {
		t.Fatal("failed directory read succeeded")
	}
	split.root = fake
	requireError(t, splitWriteFile(split, "output", []byte("data")))
	failure := errors.New("formatter failed")
	if _, err := embedImportWith(
		[]byte("package sample\n"),
		func(io.Writer, *token.FileSet, any) error { return failure },
	); !errors.Is(
		err,
		failure,
	) {
		t.Fatalf("formatter failure = %v", err)
	}
	requireError(
		t,
		finishCategoryWith(
			&categoryWrite{store: split, filename: "vars.go"},
			[]byte("package sample\n"),
			func([]byte, []declEntry, string) ([]byte, error) { return nil, failure },
		),
	)
}

func TestSplitterCommentsMarkersAndSelectors(t *testing.T) {
	set := token.NewFileSet()
	source := []byte("package sample\n// Example documents the function.\nfunc Example() {}\n")
	file, err := parser.ParseFile(set, "source.go", source, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	decl := file.Decls[0]
	text := extractDeclSource(set, source, decl)
	if !strings.HasPrefix(text, "// Example") {
		t.Fatalf("lost declaration comment: %s", text)
	}
	if got := declDocStart(set, decl, 0); got != 0 {
		t.Fatalf("invalid comment offset = %d", got)
	}
	if got := groupSingleImports(
		[]byte("package sample\nimport \"fmt\"\n"),
	); !bytes.Contains(
		got,
		[]byte("import ("),
	) {
		t.Fatalf("ungrouped import: %s", got)
	}
	if got := extractGeneratedMarker([]byte("// ordinary comment")); got != "" {
		t.Fatalf("unexpected marker %q", got)
	}
	split := openSplitRoot(t)
	marker := "// Code generated by test. DO NOT EDIT."
	captureGeneratedMarker(split, []byte(marker+"\npackage sample\n"))
	if split.generatedMarker != marker {
		t.Fatalf("captured marker = %q", split.generatedMarker)
	}
	if !strings.Contains(
		categoryHeader(&categoryWrite{generatedMarker: marker, pkgName: "sample"}),
		marker,
	) {
		t.Fatal("lost generated header")
	}
	rewritten := applyTestRewrites(
		"old.One old.Two",
		[]testRewrite{{start: 0, end: 3, value: "new"}, {start: 8, end: 11, value: "new"}},
	)
	if rewritten != "new.One new.Two" {
		t.Fatalf("selector edits = %q", rewritten)
	}
	if got := sourceCategory("source.go", "sample", &ast.GenDecl{Tok: token.IMPORT}); got != "" {
		t.Fatalf("import category = %s", got)
	}
}

func TestConsolidationFailurePropagation(t *testing.T) {
	split := openSplitRoot(t)
	set := token.NewFileSet()
	file := &ast.File{Name: ast.NewIdent("sample_test")}
	input := &testInput{set: set, file: file, name: "bench_test.go"}
	bundles := map[string]*testBundle{"sample_bench_test.go": {packageName: "sample"}}
	requireError(t, appendTestFile(split, bundles, input))
	file.Name = ast.NewIdent("sample")
	file.Imports = []*ast.ImportSpec{{Path: &ast.BasicLit{Value: "invalid"}}}
	requireError(t, appendTestFile(split, map[string]*testBundle{}, input))
	if err := os.WriteFile(
		filepath.Join(split.dir, "other_test.go"),
		[]byte("package other\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	requireError(t, collectTestFile(split, map[string]*testBundle{}, "other_test.go"))
	if err := split.root.Remove("other_test.go"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(split.dir, "old.go"),
		[]byte("package sample\nfunc Example() {}\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(split.dir, fileFuncs), 0o700); err != nil {
		t.Fatal(err)
	}
	split.goFiles = []string{"old.go"}
	requireError(t, parseAndWrite(split))
	entries, err := os.ReadDir(split.dir)
	if err != nil {
		t.Fatal(err)
	}
	requireError(t, splitDirectory(split, entries))
	testFile := filepath.Join(split.dir, "old_test.go")
	if err := os.WriteFile(
		testFile,
		[]byte("package sample\nfunc TestExample() {}\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(split.dir, "sample_test.go"), 0o700); err != nil {
		t.Fatal(err)
	}
	entries, err = os.ReadDir(split.dir)
	if err != nil {
		t.Fatal(err)
	}
	requireError(t, consolidateTests(split, entries))
	if err := split.root.Close(); err != nil {
		t.Fatal(err)
	}
	check := &generatedCheck{markers: map[string]bool{}, seen: map[string]bool{}}
	for _, entry := range entries {
		if entry.Name() == "old.go" {
			requireError(t, preflightEntry(split, check, entry))
		}
	}
}

func TestFormatAndCleanupFailures(t *testing.T) {
	split := openSplitRoot(t)
	requireError(t, formatAndWrite(split, "funcs.go", []byte("package sample\nfunc Broken(")))
	bundles := map[string]*testBundle{"sample_test.go": {
		packageName:  "sample",
		imports:      map[string]string{},
		declarations: []string{"func TestExample() {}"},
		files:        []string{"missing_test.go"},
	}}
	requireError(t, writeTestBundles(split, bundles))
	if _, err := os.Stat(filepath.Join(split.dir, "sample_test.go")); err != nil {
		t.Fatalf("consolidated output was not written: %v", err)
	}
}
