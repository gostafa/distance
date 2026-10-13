// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package goanalysis

import (
	"context"
	"encoding/json"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/gostafa/distance/distance"
	policydomain "github.com/gostafa/distance/internal/features/policy/public"
	"golang.org/x/tools/go/analysis"
)

func floatPtr(value float64) *float64 {
	return &value
}

func TestNewRejectsInvalidSettings(t *testing.T) {
	t.Parallel()

	_, err := New(&Settings{DependencyScope: "nope"})
	if err == nil {
		t.Fatal("expected error for invalid dependency-scope")
	}
}

func TestNewAcceptsDefaults(t *testing.T) {
	t.Parallel()

	a, err := New(&Settings{})
	if err != nil {
		t.Fatal(err)
	}

	if a.Name != Name {
		t.Fatalf("Name = %q, want %q", a.Name, Name)
	}
}

func TestRunnerLoadGroupsViolations(t *testing.T) {
	fixtureDir := filepath.Join(repoRoot(t), "testdata", "fixture")

	settings := Settings{
		Directory: fixtureDir,
		Patterns:  []string{"./isolated"},
		Rules: []RuleSettings{{
			Pattern: "**/isolated",
			Max:     floatPtr(0),
		}},
	}
	s := settingsWithDefaults(&settings)
	r := newRunner(&s)

	loadRunner(r)

	if r.err != nil {
		t.Fatal(r.err)
	}

	got := r.byPkg["example.com/fixture/isolated"]

	if len(got) == 0 {
		t.Fatal("expected distance violations for isolated with max 0")
	}

	for _, item := range got {
		if item.Rule != "**/isolated" {
			t.Fatalf("unexpected rule %q in %#v", item.Rule, got)
		}
	}
}

func TestFormatViolation(t *testing.T) {
	t.Parallel()

	msg := formatViolation(&policydomain.Violation{
		Package:   "example.com/p",
		Value:     0.9,
		Threshold: 0.5,
		Rule:      "**",
	})

	want := "example.com/p (package): distance 0.90 exceeds max 0.50 (rule **)"

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

	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", "..", ".."))
}

func TestRunnerRunReportsPackageViolations(t *testing.T) {
	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, "p.go", "package p\n\ntype Widget struct{}\n", 0)
	if err != nil {
		t.Fatal(err)
	}

	r := &runner{byPkg: map[string][]policydomain.Violation{
		"example.com/p": {{
			Package:   "example.com/p",
			Value:     0.9,
			Threshold: 0.5,
			Rule:      "**",
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

	if _, err := runRunner(r, pass); err != nil {
		t.Fatal(err)
	}

	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %#v, want one", diagnostics)
	}

	if diagnostics[0].Pos != file.Package {
		t.Errorf("package diagnostic position = %v, want %v", diagnostics[0].Pos, file.Package)
	}

	if !strings.Contains(diagnostics[0].Message, "exceeds max 0.50 (rule **)") {
		t.Errorf("diagnostic = %q", diagnostics[0].Message)
	}

	sentinel := errors.New("cached load error")
	failing := &runner{err: sentinel}
	failing.once.Do(func() {})

	if _, err := runRunner(failing, pass); !errors.Is(err, sentinel) {
		t.Fatalf("run error = %v, want sentinel", err)
	}
}

func TestRunnerLoadErrors(t *testing.T) {
	settings := Settings{Rules: []RuleSettings{{
		Pattern: "",
		Max:     floatPtr(0.5),
	}}}
	s := settingsWithDefaults(&settings)
	r := newRunner(&s)
	r.analyzer = func(context.Context, *distance.Config) (distance.Report, error) { return distance.Report{}, nil }
	loadRunner(r)

	if r.err == nil || !strings.Contains(r.err.Error(), "distance policy") {
		t.Fatalf("policy load error = %v", r.err)
	}

	settings = Settings{
		Directory: filepath.Join(t.TempDir(), "missing"),
		Rules:     []RuleSettings{{Pattern: "**", Max: floatPtr(0.5)}},
	}
	s = settingsWithDefaults(&settings)
	r = newRunner(&s)
	loadRunner(r)

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

	defaultsSettings := Settings{Directory: dir}
	defaults, err := settingsRules(&defaultsSettings)
	if err != nil {
		t.Fatal(err)
	}

	if len(defaults) != 1 ||
		defaults[0].Pattern != "**" ||
		defaults[0].Max != 0.5 {

		t.Fatalf("default policy = %+v", defaults)
	}

	inlineSettings := Settings{Rules: []RuleSettings{
		{Pattern: "**/internal/**", Max: floatPtr(0.2)},
		{Pattern: "**", Max: floatPtr(0.5)},
	}}
	inline, err := settingsRules(&inlineSettings)
	if err != nil {
		t.Fatal(err)
	}

	if len(inline) != 2 || inline[0].Max != 0.2 {
		t.Fatalf("inline policy = %+v", inline)
	}
}

func TestEmptyPackagePosition(t *testing.T) {
	if pos := packagePos(&analysis.Pass{}); pos != token.NoPos {
		t.Fatalf("packagePos(empty) = %v, want NoPos", pos)
	}
}

func TestSettingsToConfigUsesLoadPatterns(t *testing.T) {
	settings := Settings{
		Patterns: []string{"./internal/...", "./..."},
		Rules: []RuleSettings{
			{Pattern: "**/internal/**", Max: floatPtr(0.2)},
			{Pattern: "**", Max: floatPtr(0.5)},
		},
	}
	s := settingsWithDefaults(&settings)
	cfg := settingsToConfig(&s)

	if len(cfg.Patterns) != 2 || cfg.Patterns[0] != "./internal/..." || cfg.Patterns[1] != "./..." {
		t.Fatalf("load patterns = %v", cfg.Patterns)
	}
}

func TestSettingsRejectsEmptyPatternAndMissingMax(t *testing.T) {
	settings := Settings{Rules: []RuleSettings{{Pattern: "", Max: floatPtr(0.5)}}}
	err := validateSettings(&settings)
	if err == nil {
		t.Fatal("empty pattern accepted")
	}

	settings = Settings{Rules: []RuleSettings{{Pattern: "**"}}}
	err = validateSettings(&settings)
	if err == nil {
		t.Fatal("missing max accepted")
	}
}

func TestSettingsRejectsLegacyKeys(t *testing.T) {
	t.Parallel()

	var settings Settings

	if err := UnmarshalSettings([]byte(`{"packages":[]}`), &settings); err == nil {
		t.Fatal("packages key accepted")
	}

	if err := UnmarshalSettings(
		[]byte(`{"rules":[{"pattern":"**","max-distance":0.5}]}`),
		&settings,
	); err == nil {
		t.Fatal("max-distance key accepted")
	}
}

func TestSettingsFailurePaths(t *testing.T) {
	if _, err := DecodeSettings(make(chan int)); err == nil {
		t.Fatal("accepted unserializable settings")
	}
	if _, err := DecodeSettings(map[string]any{"unknown": true}); err == nil {
		t.Fatal("accepted unknown setting")
	}
	if err := UnmarshalSettings([]byte(`invalid`), new(Settings)); err == nil {
		t.Fatal("accepted invalid JSON")
	}
	if _, err := marshalRemappedKeys(
		map[string]json.RawMessage{"x": []byte(`invalid`)},
	); err == nil {
		t.Fatal("accepted invalid remapped payload")
	}
	if _, err := settingsRules(&Settings{Rules: []RuleSettings{{Pattern: "**"}}}); err == nil {
		t.Fatal("accepted missing maximum")
	}
	if got := formatNumber(1); got != "1" {
		t.Fatalf("integral value: %s", got)
	}
	if got := packagePos(&analysis.Pass{Files: []*ast.File{nil}}); got != token.NoPos {
		t.Fatalf("position = %v", got)
	}
}

func TestAnalyzerRunFailureIsCached(t *testing.T) {
	instance, err := New(&Settings{Directory: filepath.Join(t.TempDir(), "missing")})
	if err != nil {
		t.Fatal(err)
	}
	pass := &analysis.Pass{Pkg: types.NewPackage("example.com/m", "m")}
	for range 2 {
		if _, err := instance.Run(pass); !errors.Is(err, distance.ErrAnalysis) {
			t.Fatalf("analysis failure = %v", err)
		}
	}
}

func TestDecodedSettingsMapping(t *testing.T) {
	settings, err := DecodeSettings(
		map[string]any{
			"directory":        "source",
			"dependency-scope": "all",
			"build-tags":       []string{"tag"},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	config := settingsToConfig(&settings)
	if config.Directory != "source" || config.DependencyScope != "all" ||
		len(config.BuildTags) != 1 ||
		config.BuildTags[0] != "tag" {
		t.Fatalf("settings mapping = %+v", config)
	}
}

func TestRemappingEncoderFailure(t *testing.T) {
	failure := errors.New("encoder unavailable")
	_, err := remapKeysWith(
		[]byte(`{}`),
		func(map[string]json.RawMessage) ([]byte, error) { return nil, failure },
	)
	if !errors.Is(err, failure) {
		t.Fatalf("encoder failure = %v", err)
	}
}

func TestMissingRuleMaximumBeforeAnalysis(t *testing.T) {
	settings := Settings{DependencyScope: "module", Rules: []RuleSettings{{Pattern: "**"}}}
	if err := validateSettings(&settings); err == nil {
		t.Fatal("missing maximum accepted")
	}
	_, err := computeViolations(
		&settings,
		func(context.Context, *distance.Config) (distance.Report, error) {
			t.Fatal("analysis ran before shape validation")
			return distance.Report{}, nil
		},
	)
	if err == nil {
		t.Fatal("missing maximum accepted by runner")
	}
}
