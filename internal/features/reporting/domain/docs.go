package domain

import "github.com/gostafa/distance/internal/shared/metrics"

// DocScope groups metrics-guide entries by the kind of entity they
// describe: type metrics, package metrics, or the structural columns that
// are counted rather than computed.
type DocScope string

const (
	// DocScopeType marks a type-level metric entry.
	DocScopeType DocScope = "type"
	// DocScopePackage marks a package-level metric entry.
	DocScopePackage DocScope = "package"
	// DocScopeStructural marks a counted column (Ca, Funcs, Fields, …).
	DocScopeStructural DocScope = "structural"
)

// Doc directions: whether smaller or larger values are better, or no
// universal direction exists. They mirror qualityByMetric.
const (
	DirectionLower   = "lower"
	DirectionHigher  = "higher"
	DirectionNeutral = "neutral"
)

// MetricDoc explains one reported metric or structural field to a human:
// what it means, how it is computed, and how to judge its values. It is
// the single source behind the standalone metrics guide (--help --web) and
// the report page's per-column explanations.
type MetricDoc struct {
	// Name is the metric or column key, e.g. "amc" or "ca".
	Name string
	// Label is the column heading, e.g. "AMC".
	Label string
	// FullName spells the metric out, e.g. "Average Method Complexity".
	FullName string
	// Scope groups the entry: type metric, package metric, or structural.
	Scope DocScope
	// Definition is the versioned formula id; empty for structural fields.
	Definition string
	// FormulaMathML holds display-mode <math> markup. MathML Core only, so
	// browsers typeset it natively and the page stays self-contained.
	// Empty for structural fields.
	FormulaMathML string
	// FormulaLaTeX is the LaTeX source of record behind FormulaMathML.
	FormulaLaTeX string
	// Summary is the one-sentence meaning.
	Summary string
	// HowCalculated spells out the inputs and mechanics.
	HowCalculated string
	// Interpretation explains when values are good or bad, and why.
	Interpretation string
	// NotApplicable states when the metric is n/a; empty means always
	// applicable.
	NotApplicable string
	// Direction is "lower", "higher", or "neutral", matching the quality
	// coloring of the renderers.
	Direction string
	// Bounded reports whether values live in [0, 1].
	Bounded bool
	// Example is a small worked numeric example.
	Example string
}






// LaTeX: D = \lvert A + I - 1 \rvert
const formulaDistance = `<math display="block" alttext="D = |A + I - 1|"><mrow><mi>D</mi><mo>=</mo><mo stretchy="false">|</mo><mi>A</mi><mo>+</mo><mi>I</mi><mo>−</mo><mn>1</mn><mo stretchy="false">|</mo></mrow></math>
<math display="block" alttext="A = \frac{N_{interface}}{N_{named}}"><mrow><mi>A</mi><mo>=</mo><mfrac><msub><mi>N</mi><mtext>interface</mtext></msub><msub><mi>N</mi><mtext>named</mtext></msub></mfrac></mrow></math>
<math display="block" alttext="I = \frac{C_e}{C_a + C_e}"><mrow><mi>I</mi><mo>=</mo><mfrac><msub><mi>C</mi><mi>e</mi></msub><mrow><msub><mi>C</mi><mi>a</mi></msub><mo>+</mo><msub><mi>C</mi><mi>e</mi></msub></mrow></mfrac></mrow></math>`

// MetricDocs returns the guide entries for the reported metric and
// structural columns. Abstractness and instability are described as
// inputs to distance; they are not documented as selectable metrics.
func MetricDocs() []MetricDoc {
	return []MetricDoc{
		{
			Name:           metrics.MetricDistance,
			Label:          abbrev(metrics.MetricDistance),
			FullName:       "Distance from the Main Sequence",
			Scope:          DocScopePackage,
			Definition:     metrics.DefinitionDistance,
			FormulaMathML:  formulaDistance,
			FormulaLaTeX:   `D = \lvert A + I - 1 \rvert` + "\n" + `A = \frac{N_{\text{interface}}}{N_{\text{named}}}` + "\n" + `I = \frac{C_e}{C_a + C_e}`,
			Summary:        "How far a package sits from the ideal abstractness–instability balance.",
			HowCalculated:  "The absolute distance of the package's abstractness A and instability I from the 'main sequence' line A + I = 1, where abstraction and stability balance. A is the share of named types that are interfaces; I is Ce/(Ca+Ce) within the configured dependency scope. An isolated package (Ca + Ce = 0) is treated as maximally stable (I = 0). Abstractness and instability are computed internally and are not reported, selectable, or gateable on their own.",
			Interpretation: "0 is on the main sequence. High distance means the package is either concrete and stable (rigid — everything depends on its details) or abstract and unstable (abstractions nobody depends on). Mind the isolated-package convention: a concrete isolated package has A = 0 and I = 0, so D = 1 by definition, not necessarily by design fault.",
			NotApplicable:  "When abstractness is not applicable — for example, the package declares no relevant named types.",
			Direction:      DirectionLower,
			Bounded:        true,
			Example:        "A = 0.25 and I = 0.5: D = |0.25 + 0.5 − 1| = 0.25.",
		},
		{
			Name:           "ca",
			Label:          "Ca",
			FullName:       "Afferent coupling",
			Scope:          DocScopeStructural,
			Summary:        "How many analyzed packages import this package.",
			HowCalculated:  "Counted within the analyzed set only — importers outside the analysis are not observable, so the value depends on the patterns you analyze.",
			Interpretation: "A neutral count with no good/bad color. High Ca marks load-bearing packages: many others break when this one changes, so it should be stable and well tested. It is the incoming half of instability.",
			Direction:      DirectionNeutral,
			Example:        "If 3 analyzed packages import example.com/m/util, its Ca is 3.",
		},
		{
			Name:           "ce",
			Label:          "Ce",
			FullName:       "Efferent coupling",
			Scope:          DocScopeStructural,
			Summary:        "How many packages this package imports, within the dependency scope.",
			HowCalculated:  "The package's imports that fall in the configured -dependency-scope: project counts only other analyzed packages, module counts packages of the main module, all counts every import. Duplicates and self-imports are ignored.",
			Interpretation: "A neutral count with no good/bad color. High Ce means the package has many reasons to change. It is the outgoing half of instability.",
			Direction:      DirectionNeutral,
			Example:        "A package importing 2 in-scope packages has Ce = 2 regardless of how often each is imported.",
		},
		{
			Name:           "funcs",
			Label:          "Funcs",
			FullName:       "Functions",
			Scope:          DocScopeStructural,
			Summary:        "Declared functions and methods in the package.",
			HowCalculated:  "Counted over the package's analyzed files — excluded files (tests or generated code, unless included by flag) do not contribute.",
			Interpretation: "A neutral size measure: use it to weigh the metrics — a package with 3 funcs and a package with 300 deserve different scrutiny at the same scores.",
			Direction:      DirectionNeutral,
			Example:        "A package with 4 functions and 6 methods across its types shows Funcs = 10.",
		},
		{
			Name:           "vars",
			Label:          "Vars",
			FullName:       "Variables",
			Scope:          DocScopeStructural,
			Summary:        "Top-level variable names declared in the package.",
			HowCalculated:  "Counts each non-blank identifier in package-level var declarations over analyzed files. Local variables inside functions and methods do not contribute.",
			Interpretation: "A neutral size measure. High values can signal broad package-level mutable state, but context matters.",
			Direction:      DirectionNeutral,
			Example:        "var a, b int contributes Vars = 2; var _ = setup() contributes 0.",
		},
		{
			Name:           "consts",
			Label:          "Consts",
			FullName:       "Constants",
			Scope:          DocScopeStructural,
			Summary:        "Top-level constant names declared in the package.",
			HowCalculated:  "Counts each non-blank identifier in package-level const declarations over analyzed files. Local constants inside functions and methods do not contribute.",
			Interpretation: "A neutral size measure. Many constants may be harmless domain vocabulary or a sign that related values could be grouped.",
			Direction:      DirectionNeutral,
			Example:        "const A, B = 1, 2 contributes Consts = 2; const _ = iota contributes 0.",
		},
		{
			Name:           "types",
			Label:          "Types",
			FullName:       "Named types",
			Scope:          DocScopeStructural,
			Summary:        "Analyzed named types declared in the package.",
			HowCalculated:  "Counts the package's named type declarations that enter the analysis; type aliases never enter the model.",
			Interpretation: "A neutral size measure, shown in the Packages view. Many types with poor cohesion scores is a stronger signal than one outlier.",
			Direction:      DirectionNeutral,
			Example:        "A package declaring Service, Config, and an Option interface shows Types = 3.",
		},
		{
			Name:           "fields",
			Label:          "Fields",
			FullName:       "Struct fields",
			Scope:          DocScopeStructural,
			Summary:        "The type's struct field count.",
			HowCalculated:  "An embedded field counts as one; members promoted through embedding are not counted. Non-struct types show 0.",
			Interpretation: "A neutral count that sizes the cohesion metrics: LCOM and TCC both reason about how methods use these fields.",
			Direction:      DirectionNeutral,
			Example:        "struct { ID int; Name string; sync.Mutex } has Fields = 3 — the embedded mutex counts as one.",
		},
		{
			Name:           "methods",
			Label:          "Methods",
			FullName:       "Declared methods",
			Scope:          DocScopeStructural,
			Summary:        "The type's declared method count.",
			HowCalculated:  "Value- and pointer-receiver methods are counted alike; methods promoted from embedded types are excluded.",
			Interpretation: "A neutral count that sizes the cohesion and complexity metrics: most of them are n/a below 1 or 2 methods, by design rather than as a gap.",
			Direction:      DirectionNeutral,
			Example:        "A type with func (s *S) Open() and func (s S) Close() has Methods = 2.",
		},
	}
}
