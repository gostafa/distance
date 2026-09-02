package analyzer

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"
	"testing"

	policydomain "github.com/gostafa/distance/internal/features/policy/domain"
	"golang.org/x/tools/go/analysis"
)

func TestRunnerRunReportsPackageViolations(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "p.go", "package p\n\ntype Widget struct{}\n", 0)
	if err != nil {
		t.Fatal(err)
	}

	r := &runner{byPkg: map[string][]policydomain.Violation{
		"example.com/p": {{
			Package:    "example.com/p",
			Key:        "distance",
			Value:      0.9,
			Comparator: policydomain.ComparatorMax,
			Threshold:  0.5,
		}},
	}}
	r.once.Do(func() {})

	var diagnostics []analysis.Diagnostic
	pass := &analysis.Pass{
		Fset:   fset,
		Files:  []*ast.File{file},
		Pkg:    types.NewPackage("example.com/p", "p"),
		Report: func(d analysis.Diagnostic) { diagnostics = append(diagnostics, d) },
	}

	if _, err := r.run(pass); err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %#v, want one", diagnostics)
	}
	if diagnostics[0].Pos != file.Package {
		t.Errorf("package diagnostic position = %v, want %v", diagnostics[0].Pos, file.Package)
	}
	if !strings.Contains(diagnostics[0].Message, "exceeds max 0.50") {
		t.Errorf("diagnostic = %q", diagnostics[0].Message)
	}

	sentinel := errors.New("cached load error")
	failing := &runner{err: sentinel}
	failing.once.Do(func() {})
	if _, err := failing.run(pass); !errors.Is(err, sentinel) {
		t.Fatalf("run error = %v, want sentinel", err)
	}
}

func TestRunnerLoadErrors(t *testing.T) {
	r := newRunner(Settings{Packages: []policydomain.PackageRule{{
		Pattern:     "",
		MaxDistance: 0.5,
	}}}.withDefaults())
	r.load()
	if r.err == nil || !strings.Contains(r.err.Error(), "distance policy") {
		t.Fatalf("policy load error = %v", r.err)
	}

	r = newRunner(Settings{
		Directory: filepath.Join(t.TempDir(), "missing"),
		Packages:  []policydomain.PackageRule{{Pattern: "./...", MaxDistance: 0.5}},
	}.withDefaults())
	r.load()
	if r.err == nil || !strings.Contains(r.err.Error(), "distance analyze") {
		t.Fatalf("analysis load error = %v", r.err)
	}
}

func TestInlinePolicyDefaultsAndIgnoresModularityFile(t *testing.T) {
	dir := t.TempDir()
	config := filepath.Join(dir, ".modularity.yml")
	if err := os.WriteFile(
		config,
		[]byte("version: 1\npackage:\n  types: 3\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	defaults, err := (Settings{Directory: dir}).withDefaults().policy()
	if err != nil {
		t.Fatal(err)
	}
	if len(defaults.Packages) != 1 ||
		defaults.Packages[0].Pattern != "./..." ||
		defaults.Packages[0].MaxDistance != 0.5 {
		t.Fatalf("default policy = %+v", defaults.Packages)
	}

	inline, err := (Settings{Packages: []policydomain.PackageRule{
		{Pattern: "./internal/...", MaxDistance: 0.2},
		{Pattern: "./...", MaxDistance: 0.5},
	}}).policy()
	if err != nil {
		t.Fatal(err)
	}
	if len(inline.Packages) != 2 || inline.Packages[0].MaxDistance != 0.2 {
		t.Fatalf("inline policy = %+v", inline.Packages)
	}
}

func TestEmptyPackagePosition(t *testing.T) {
	if pos := packagePos(&analysis.Pass{}); pos != token.NoPos {
		t.Fatalf("packagePos(empty) = %v, want NoPos", pos)
	}
}

func TestSettingsToConfigUnionsPatterns(t *testing.T) {
	cfg := Settings{Packages: []policydomain.PackageRule{
		{Pattern: "./internal/...", MaxDistance: 0.2},
		{Pattern: "./internal/...", MaxDistance: 0.1},
		{Pattern: "./...", MaxDistance: 0.5},
	}}.toConfig()

	if len(cfg.Patterns) != 2 || cfg.Patterns[0] != "./internal/..." || cfg.Patterns[1] != "./..." {
		t.Fatalf("load patterns = %v", cfg.Patterns)
	}
}

func TestSettingsRejectsEmptyPattern(t *testing.T) {
	if err := (Settings{Packages: []policydomain.PackageRule{{Pattern: "", MaxDistance: 0.5}}}).validate(); err == nil {
		t.Fatal("empty pattern accepted")
	}
}
