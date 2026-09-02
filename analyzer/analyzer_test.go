package analyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"runtime"
	"testing"

	policydomain "github.com/gostafa/distance/internal/features/policy/domain"
	"golang.org/x/tools/go/analysis"
)

func TestNewRejectsInvalidSettings(t *testing.T) {
	t.Parallel()

	_, err := New(Settings{DependencyScope: "nope"})
	if err == nil {
		t.Fatal("expected error for invalid dependency-scope")
	}
}

func TestNewAcceptsDefaults(t *testing.T) {
	t.Parallel()

	a, err := New(Settings{})
	if err != nil {
		t.Fatal(err)
	}

	if a.Name != Name {
		t.Fatalf("Name = %q, want %q", a.Name, Name)
	}
}

func TestRunnerLoadGroupsViolations(t *testing.T) {
	fixtureDir := filepath.Join(repoRoot(t), "testdata", "fixture")

	r := newRunner(Settings{
		Directory: fixtureDir,
		Packages: []policydomain.PackageRule{{
			Pattern:     "./isolated",
			MaxDistance: 0,
		}},
	}.withDefaults())

	r.load()
	if r.err != nil {
		t.Fatal(r.err)
	}

	got := r.byPkg["example.com/fixture/isolated"]
	if len(got) == 0 {
		t.Fatal("expected distance violations for isolated with max-distance 0")
	}

	for _, v := range got {
		if v.Key != "distance" {
			t.Fatalf("unexpected key %q in %#v", v.Key, got)
		}
	}
}

func TestFormatViolation(t *testing.T) {
	t.Parallel()

	msg := formatViolation(policydomain.Violation{
		Package:    "example.com/p",
		Key:        "distance",
		Value:      0.9,
		Comparator: policydomain.ComparatorMax,
		Threshold:  0.5,
	})

	want := "example.com/p (package): distance 0.90 exceeds max 0.50"
	if msg != want {
		t.Fatalf("formatViolation = %q, want %q", msg, want)
	}
}

func TestPackagePos(t *testing.T) {
	t.Parallel()

	src := `package p

type Widget struct{}
`
	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, "p.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}

	pass := &analysis.Pass{Files: []*ast.File{file}, Fset: fset}

	if pos := packagePos(pass); pos != file.Package {
		t.Fatalf("packagePos = %v, want %v", pos, file.Package)
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(file), ".."))
}
