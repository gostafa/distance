// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// TestForbiddenIdentifier enforces the project constraint that a certain
// legacy identifier never appears anywhere in the tree. The needle is
// assembled at runtime so this file cannot violate the rule itself.
func TestForbiddenIdentifier(t *testing.T) {
	needle := "go" + "metrics"

	err := filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if name := d.Name(); name == ".git" || name == ".qodo" {
				return filepath.SkipDir
			}

			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		if strings.Contains(strings.ToLower(string(content)), needle) {
			t.Errorf("%s contains the forbidden identifier", path)
		}

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestMainDelegatesToCLI(t *testing.T) {
	var gotArgs []string

	var gotCode int

	runtime := mainRuntime{
		run: func(args []string) int {
			gotArgs = append([]string(nil), args...)

			return 7
		},
		exit: func(code int) { gotCode = code },
	}

	runtime.start([]string{"--version"})

	if len(gotArgs) != 1 || gotArgs[0] != "--version" {
		t.Fatalf("args = %v", gotArgs)
	}

	if gotCode != 7 {
		t.Fatalf("exit code = %d, want 7", gotCode)
	}
}

func TestRepositoryArchitecture(t *testing.T) {
	for _, violation := range architectureViolations(".") {
		t.Error(violation)
	}
}

func architectureViolations(root string) []string {
	findings := []string{}
	packages := map[string]map[string]*ast.File{}
	graph := map[string][]string{}
	set := token.NewFileSet()
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if entry.Name() == ".git" || entry.Name() == ".task" || entry.Name() == ".qodo" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		file, err := parser.ParseFile(set, path, nil, parser.ParseComments)
		if err != nil {
			findings = append(findings, fmt.Sprintf("%s: parse: %v", path, err))
			return nil
		}
		relative, err := filepath.Rel(root, filepath.Dir(path))
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		if packages[relative] == nil {
			packages[relative] = map[string]*ast.File{}
		}
		name := filepath.Base(path)
		packages[relative][name] = file
		findings = append(findings, layoutViolations(relative, name, file)...)
		if strings.HasSuffix(name, "_test.go") {
			return nil
		}
		for _, imp := range file.Imports {
			dependency, _ := strconv.Unquote(imp.Path.Value)
			if reason := forbiddenDependency(relative, dependency); reason != "" {
				findings = append(findings, fmt.Sprintf("%s: %s: %s", path, reason, dependency))
			}
			if strings.HasPrefix(dependency, repositoryModule+"/") {
				graph[relative] = append(
					graph[relative],
					strings.TrimPrefix(dependency, repositoryModule+"/"),
				)
			}
		}
		return nil
	})
	if err != nil {
		findings = append(findings, err.Error())
	}
	for dir, files := range packages {
		doc, ok := files["doc.go"]
		if !ok || doc.Doc == nil || strings.TrimSpace(doc.Doc.Text()) == "" {
			findings = append(findings, dir+": missing documented doc.go")
		}
		if isPublished(dir) {
			for name, file := range files {
				if !strings.HasSuffix(name, "_test.go") {
					findings = append(
						findings,
						signatureViolations(filepath.Join(dir, name), file, files)...)
				}
			}
		}

	}
	for _, cycle := range dependencyCycles(graph) {
		findings = append(findings, "dependency cycle: "+cycle)
	}
	sort.Strings(findings)
	return findings
}

const repositoryModule = "github.com/gostafa/distance"

func layoutViolations(dir, name string, file *ast.File) []string {
	findings := []string{}
	pkg := strings.TrimSuffix(file.Name.Name, "_test")
	if strings.HasSuffix(name, "_test.go") {
		expected := pkg + "_test.go"
		if file.Name.Name != pkg {
			expected = pkg + "_integration_test.go"
		}
		if name != expected && name != pkg+"_bench_test.go" {
			findings = append(findings, dir+": invalid test filename "+name)
		}
		return findings
	}
	expected := map[token.Token]string{
		token.TYPE:  "types.go",
		token.CONST: "consts.go",
		token.VAR:   "vars.go",
	}
	allowed := name == "doc.go" || name == "funcs.go" || name == "types.go" ||
		name == "consts.go" ||
		name == "vars.go"
	if !allowed {
		findings = append(findings, dir+": invalid production filename "+name)
	}
	for _, decl := range file.Decls {
		target := "funcs.go"
		if generic, ok := decl.(*ast.GenDecl); ok {
			if generic.Tok == token.IMPORT {
				if name == "doc.go" {
					findings = append(findings, dir+": doc.go contains imports")
				}
				continue
			}
			target = expected[generic.Tok]
		}
		if name != target {
			findings = append(findings, dir+": declaration belongs in "+target+", found in "+name)
		}
	}
	return findings
}

func featurePart(path string) (string, string) {
	parts := strings.Split(strings.TrimPrefix(path, repositoryModule+"/"), "/")
	if len(parts) >= 4 && parts[0] == "internal" && parts[1] == "features" {
		return parts[2], strings.Join(parts[3:], "/")
	}
	return "", ""
}

func forbiddenDependency(source, dependency string) string {
	feature, layer := featurePart(source)
	target, targetLayer := featurePart(dependency)
	internal := strings.HasPrefix(dependency, repositoryModule+"/")
	if feature != "" && target != "" && feature != target &&
		!(targetLayer == "public" || strings.HasPrefix(targetLayer, "public/")) {
		return "private cross-feature dependency"
	}
	domain := layer == "domain" || strings.HasPrefix(layer, "domain/")
	application := layer == "application" || strings.HasPrefix(layer, "application/")
	if domain {
		if target != "" &&
			(target != feature || !(targetLayer == "domain" || strings.HasPrefix(targetLayer, "domain/"))) {
			return "domain depends outside its domain"
		}
		if internal && !strings.HasPrefix(dependency, repositoryModule+"/internal/shared/") &&
			target == "" {
			return "domain imports outer implementation"
		}
		if !internal && strings.Contains(strings.Split(dependency, "/")[0], ".") {
			return "domain imports third-party code"
		}
		for _, prefix := range []string{"go", "golang.org/x/tools", "encoding/json", "encoding/csv", "os", "io", "path/filepath", "flag", "log", "runtime/debug", "net", "embed"} {
			if dependency == prefix || strings.HasPrefix(dependency, prefix+"/") {
				return "domain imports boundary technology"
			}
		}
	}
	if application {
		if target == feature &&
			(targetLayer == "adapters" || strings.HasPrefix(targetLayer, "adapters/")) {
			return "application imports concrete adapter"
		}
		if internal && strings.HasPrefix(dependency, repositoryModule+"/internal/infrastructure/") {
			return "application imports concrete infrastructure"
		}
		if !internal && strings.Contains(strings.Split(dependency, "/")[0], ".") {
			return "application imports third-party code"
		}
		if target != feature && target != "" &&
			dependency != repositoryModule+"/internal/features/packagemetrics/public" {
			return "application integration requires consumer port"
		}
		for _, prefix := range []string{"os", "net", "encoding/json", "encoding/csv", "flag", "log", "runtime/pprof", "embed"} {
			if dependency == prefix || strings.HasPrefix(dependency, prefix+"/") {
				return "application imports boundary technology"
			}
		}
	}
	if strings.HasPrefix(source, "internal/shared/") && internal &&
		!strings.HasPrefix(dependency, repositoryModule+"/internal/shared/") {
		return "shared kernel depends on feature implementation"
	}
	if strings.HasPrefix(source, "internal/infrastructure/") && target != "" &&
		!(targetLayer == "ports" || strings.HasPrefix(targetLayer, "ports/") || targetLayer == "public" || strings.HasPrefix(targetLayer, "public/")) {
		return "infrastructure depends on private feature implementation"
	}
	return ""
}

func isPublished(path string) bool {
	_, layer := featurePart(path)
	return path == "distance" || path == "plugin" || layer == "public" ||
		strings.HasPrefix(layer, "public/")
}

type signatureDefinition = struct {
	expression ast.Node
	imports    map[string]string
}

func fileImports(file *ast.File) map[string]string {
	imports := map[string]string{}
	for _, specification := range file.Imports {
		path, _ := strconv.Unquote(specification.Path.Value)
		alias := filepath.Base(path)
		if specification.Name != nil {
			alias = specification.Name.Name
		}
		imports[alias] = path
	}
	return imports
}

func signatureDefinitions(files map[string]*ast.File) map[string]signatureDefinition {
	definitions := map[string]signatureDefinition{}
	for name, file := range files {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		imports := fileImports(file)
		for _, decl := range file.Decls {
			if group, ok := decl.(*ast.GenDecl); ok {
				for _, spec := range group.Specs {
					if typed, ok := spec.(*ast.TypeSpec); ok {
						definitions[typed.Name.Name] = signatureDefinition{typed.Type, imports}
					}
				}
			}
			if function, ok := decl.(*ast.FuncDecl); ok && function.Recv == nil {
				definitions[function.Name.Name] = signatureDefinition{
					function.Type.Results,
					imports,
				}
			}
		}
	}
	return definitions
}

func signatureViolations(path string, file *ast.File, files map[string]*ast.File) []string {
	findings := []string{}
	definitions := signatureDefinitions(files)
	visited := map[string]bool{}
	var inspect func(ast.Node, map[string]string)
	inspect = func(node ast.Node, imports map[string]string) {
		if node == nil {
			return
		}
		ast.Inspect(node, func(child ast.Node) bool {
			switch typed := child.(type) {
			case *ast.StructType:
				for _, field := range typed.Fields.List {
					if len(field.Names) == 0 || field.Names[0].IsExported() {
						inspect(field.Type, imports)
					}
				}
				return false
			case *ast.Field:
				inspect(typed.Type, imports)
				return false
			case *ast.Ident:
				if definition, found := definitions[typed.Name]; found && !visited[typed.Name] {
					visited[typed.Name] = true
					inspect(definition.expression, definition.imports)
				}
			case *ast.SelectorExpr:
				alias, ok := typed.X.(*ast.Ident)
				if !ok {
					return true
				}
				dependency := imports[alias.Name]
				if strings.HasPrefix(dependency, repositoryModule+"/") {
					if !isPublished(strings.TrimPrefix(dependency, repositoryModule+"/")) {
						findings = append(findings, path+": published signature leaks "+dependency)
					}
				} else if strings.Contains(strings.Split(dependency, "/")[0], ".") {
					findings = append(
						findings,
						path+": published signature leaks third-party type "+dependency,
					)
				}
				return false
			}
			return true
		})
	}
	imports := fileImports(file)
	for _, decl := range file.Decls {
		switch value := decl.(type) {
		case *ast.FuncDecl:
			if !value.Name.IsExported() {
				continue
			}
			if value.Recv != nil {
				receiver := value.Recv.List[0].Type
				if pointer, ok := receiver.(*ast.StarExpr); ok {
					receiver = pointer.X
				}
				if id, ok := receiver.(*ast.Ident); ok && !id.IsExported() {
					continue
				}
			}
			inspect(value.Type, imports)
		case *ast.GenDecl:
			for _, specification := range value.Specs {
				if typed, ok := specification.(*ast.TypeSpec); ok && typed.Name.IsExported() {
					inspect(typed.Type, imports)
				}
				if variable, ok := specification.(*ast.ValueSpec); ok {
					for _, name := range variable.Names {
						if !name.IsExported() {
							continue
						}
						inspect(variable.Type, imports)
						if variable.Type == nil {
							for _, expression := range variable.Values {
								inspect(expression, imports)
							}
						}
					}
				}
			}
		}
	}
	return findings
}

func dependencyCycles(graph map[string][]string) []string {
	state := map[string]int{}
	cycles := []string{}
	var visit func(string, []string)
	visit = func(node string, path []string) {
		if state[node] == 1 {
			cycles = append(cycles, strings.Join(append(path, node), " -> "))
			return
		}
		if state[node] == 2 {
			return
		}
		state[node] = 1
		for _, next := range graph[node] {
			visit(next, append(path, node))
		}
		state[node] = 2
	}
	for node := range graph {
		visit(node, nil)
	}
	return cycles
}

func TestArchitectureRejectsViolations(t *testing.T) {
	cases := []struct{ name, path, source, want string }{
		{
			"private integration",
			"internal/features/a/application/funcs.go",
			"package application\nimport _ \"github.com/gostafa/distance/internal/features/b/domain\"",
			"private cross-feature dependency",
		},
		{
			"concrete application",
			"internal/features/a/application/funcs.go",
			"package application\nimport _ \"github.com/gostafa/distance/internal/infrastructure/workerpool\"",
			"application imports concrete infrastructure",
		},
		{
			"nested domain",
			"internal/features/a/domain/model/funcs.go",
			"package model\nimport _ \"go/types\"",
			"domain imports boundary technology",
		},
		{
			"facade leak",
			"distance/types.go",
			"package distance\nimport x \"golang.org/x/tools/go/analysis\"\ntype Result struct { Analyzer *x.Analyzer }",
			"published signature leaks third-party type",
		},
		{
			"wrong filename",
			"tools/helper/main.go",
			"package main\nfunc main(){}",
			"invalid production filename",
		},
		{
			"wrong declaration",
			"internal/features/a/domain/funcs.go",
			"package domain\nvar value=1",
			"declaration belongs in vars.go",
		},
		{
			"missing documentation",
			"only/funcs.go",
			"package only\nfunc Work(){}",
			"missing documented doc.go",
		},
		{"wrong test filename", "only/more_test.go", "package only\n", "invalid test filename"},
		{
			"hidden alias leak",
			"distance/types.go",
			"package distance\nimport x \"github.com/gostafa/distance/internal/features/a/domain\"\ntype hidden = x.Value\ntype Result struct { Value hidden }",
			"published signature leaks",
		},
		{
			"private parameter",
			"distance/funcs.go",
			"package distance\nimport x \"github.com/gostafa/distance/internal/features/a/domain\"\nfunc Analyze(config *x.Config){}",
			"published signature leaks",
		},
		{
			"published prefix",
			"internal/features/a/application/funcs.go",
			"package application\nimport _ \"github.com/gostafa/distance/internal/features/b/publicity\"",
			"private cross-feature dependency",
		},
		{
			"private inferred variable",
			"distance/vars.go",
			"package distance\nimport x \"github.com/gostafa/distance/internal/features/a/domain\"\nvar Default = x.Config{}",
			"published signature leaks",
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, test.path)
			if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(test.source), 0o600); err != nil {
				t.Fatal(err)
			}
			if got := strings.Join(
				architectureViolations(root),
				"\n",
			); !strings.Contains(
				got,
				test.want,
			) {
				t.Fatalf("findings %q missing %q", got, test.want)
			}
		})
	}
}

func TestArchitectureCycleDetection(t *testing.T) {
	if cycles := dependencyCycles(map[string][]string{"a": {"b"}, "b": {"a"}}); len(cycles) == 0 {
		t.Fatal("cycle not detected")
	}
	if cycles := dependencyCycles(map[string][]string{"a": {"b"}}); len(cycles) != 0 {
		t.Fatalf("acyclic graph reported %v", cycles)
	}
}

func TestCommandProcess(t *testing.T) {
	if os.Getenv("DISTANCE_COMMAND_TEST") == "1" {
		os.Args = []string{"distance", "--version"}
		main()
		return
	}
	command := exec.CommandContext(
		t.Context(),
		os.Args[0],
		"-test.run=^TestCommandProcess$",
		"-test.paniconexit0=false",
	)
	command.Env = append(os.Environ(), "DISTANCE_COMMAND_TEST=1")
	coverage := flag.Lookup("test.gocoverdir")
	if coverage != nil && coverage.Value.String() != "" {
		command.Env = append(command.Env, "GOCOVERDIR="+coverage.Value.String())
	}
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("command: %v\n%s", err, output)
	}
	if !strings.Contains(string(output), "distance") {
		t.Fatalf("version output: %s", output)
	}
}

func TestArchitectureRejectsSignatureLeakAcrossFiles(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "distance")
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"doc.go":   "// Package distance publishes the facade.\npackage distance\n",
		"types.go": "package distance\nimport sdk \"golang.org/x/tools/go/analysis\"\ntype hidden = sdk.Analyzer\n",
		"funcs.go": "package distance\nfunc Leak(input *hidden){}\n",
	}
	for name, source := range files {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(source), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	findings := strings.Join(architectureViolations(root), "\n")
	if !strings.Contains(findings, "funcs.go: published signature leaks") {
		t.Fatalf("missed cross-file signature: %s", findings)
	}
}
