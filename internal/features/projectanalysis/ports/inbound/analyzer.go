package inbound

import (
	"context"
	"fmt"

	typefacts "github.com/gostafa/distance/internal/features/typefacts/domain"
	"github.com/gostafa/distance/internal/shared/metrics"
)

// Options is a fully validated, defaults-applied analysis request.
type Options struct {
	// Directory is the working directory package loading runs from.
	Directory string
	// Patterns are the package patterns to analyze (e.g. "./...").
	Patterns []string
	// IncludeTests also analyzes test files and test packages.
	IncludeTests bool
	// IncludeGenerated also analyzes generated files.
	IncludeGenerated bool
	// BuildTags are extra build tags applied while loading.
	BuildTags []string
	// Workers bounds package-level concurrency; 0 selects a default.
	Workers int
	// DependencyScope is "project", "module", or "all".
	DependencyScope string
	// ContinueOnError skips packages that fail to load or type-check.
	ContinueOnError bool
}

// TypeResult carries one type's structural facts. Types carry no reported
// metric in this linter.
type TypeResult struct {
	// Name is the type's declared name.
	Name string
	// Exported reports whether the type name is exported.
	Exported bool
	// Kind classifies the type's underlying type.
	Kind typefacts.TypeKind
	// Pos locates the type declaration in source.
	Pos typefacts.Position
	// Fields is the type's struct field count (embedded fields count one).
	Fields int
	// FieldDetails holds the struct fields in declaration order.
	FieldDetails []typefacts.FieldFacts
	// Methods is the type's declared method count.
	Methods int
	// MethodDetails holds the declared methods in report order.
	MethodDetails []FunctionResult
}

// DeclarationResult carries one top-level variable or constant declaration.
type DeclarationResult struct {
	// Name is the declared identifier.
	Name string
	// Exported reports whether the declaration name is exported.
	Exported bool
	// Pos locates the declaration in source.
	Pos typefacts.Position
}

// FunctionResult carries one function or method's structural details.
type FunctionResult struct {
	// Name is the declared function or method name.
	Name string
	// Exported reports whether the function name is exported.
	Exported bool
	// Receiver is the declaring type name for methods; empty for free funcs.
	Receiver string
	// Pos locates the declaration in source.
	Pos typefacts.Position
	// Lines is the inclusive source line count from func keyword to end.
	Lines int
}

// PackageResult carries one package's display metrics and analyzed types.
type PackageResult struct {
	// Path is the package's import path.
	Path string
	// Afferent counts analyzed packages importing this package (Ca).
	Afferent int
	// Efferent counts this package's in-scope imports (Ce).
	Efferent int
	// ExportedFuncs counts the package's declared functions and methods with
	// an exported name.
	ExportedFuncs int
	// UnexportedFuncs counts the package's declared functions and methods with
	// an unexported name.
	UnexportedFuncs int
	// Vars counts top-level variable names declared in the package.
	Vars int
	// Consts counts top-level constant names declared in the package.
	Consts int
	// Variables are the package's top-level variable declarations.
	Variables []DeclarationResult
	// Constants are the package's top-level constant declarations.
	Constants []DeclarationResult
	// Functions are the package's top-level functions, excluding methods.
	Functions []FunctionResult
	// Metrics holds the package's display metrics in the fixed order.
	Metrics []metrics.MetricResult
	// Types are the package's analyzed types, sorted by name.
	Types []TypeResult
}

// String summarizes the package result for debugging.
func (p PackageResult) String() string {
	return fmt.Sprintf("%s: %d package metrics, %d types", p.Path, len(p.Metrics), len(p.Types))
}

// Result is a deterministic analysis outcome: packages sorted by import
// path, types by name, metrics in the fixed order.
type Result struct {
	// ModulePath is the analyzed main module's path, when known.
	ModulePath string
	// Packages are the analyzed packages, sorted by import path.
	Packages []PackageResult
}

// Analyzer runs the full analysis pipeline. It implements no metric
// formula itself.
type Analyzer interface {
	Analyze(ctx context.Context, opts Options) (Result, error)
}
