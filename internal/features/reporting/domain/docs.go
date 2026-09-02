package domain

import "github.com/gostafa/distance/internal/shared/metrics"

// DocScope groups metrics-guide entries by the kind of entity they
// describe: package metrics, or the structural columns that are counted
// rather than computed.
type DocScope string

const (
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

const formulaAbstractness = `<math display="block" alttext="A = \frac{N_{interface}}{N_{named}}"><mrow><mi>A</mi><mo>=</mo><mfrac><msub><mi>N</mi><mtext>interface</mtext></msub><msub><mi>N</mi><mtext>named</mtext></msub></mfrac></mrow></math>`

const formulaInstability = `<math display="block" alttext="I = \frac{C_e}{C_a + C_e}"><mrow><mi>I</mi><mo>=</mo><mfrac><msub><mi>C</mi><mi>e</mi></msub><mrow><msub><mi>C</mi><mi>a</mi></msub><mo>+</mo><msub><mi>C</mi><mi>e</mi></msub></mrow></mfrac></mrow></math>`

const formulaDistance = `<math display="block" alttext="D = |A + I - 1|"><mrow><mi>D</mi><mo>=</mo><mo stretchy="false">|</mo><mi>A</mi><mo>+</mo><mi>I</mi><mo>−</mo><mn>1</mn><mo stretchy="false">|</mo></mrow></math>`

// MetricDocs returns the guide entries for the reported metrics and
// structural columns. Abstractness and instability are documented as
// reported fields; they are not selectable or gateable.
func MetricDocs() []MetricDoc {
	return []MetricDoc{
		{
			Name:           metrics.MetricAbstractness,
			Label:          abbrev(metrics.MetricAbstractness),
			FullName:       "Abstractness",
			Scope:          DocScopePackage,
			Definition:     metrics.DefinitionAbstractness,
			FormulaMathML:  formulaAbstractness,
			FormulaLaTeX:   `A = \frac{N_{\text{interface}}}{N_{\text{named}}}`,
			Summary:        "The share of a package's named types that are interfaces.",
			HowCalculated:  "Named interface types divided by the package's relevant named types. Type aliases are excluded from both counts. Reported next to instability and distance; not selectable or gateable on its own.",
			Interpretation: "A neutral ratio with no good/bad color: neither fully concrete nor fully abstract is universally better. Martin's main sequence wants A to balance instability I so that A + I ≈ 1.",
			NotApplicable:  "When the package declares no relevant named types.",
			Direction:      DirectionNeutral,
			Bounded:        true,
			Example:        "1 interface among 4 named types: A = 0.25.",
		},
		{
			Name:           metrics.MetricInstability,
			Label:          abbrev(metrics.MetricInstability),
			FullName:       "Instability",
			Scope:          DocScopePackage,
			Definition:     metrics.DefinitionInstability,
			FormulaMathML:  formulaInstability,
			FormulaLaTeX:   `I = \frac{C_e}{C_a + C_e}`,
			Summary:        "How independently a package can change, from its coupling.",
			HowCalculated:  "Ce / (Ca + Ce) within the configured dependency scope. An isolated package (Ca + Ce = 0) is treated as maximally stable: I = 0. Reported next to abstractness and distance; not selectable or gateable on its own.",
			Interpretation: "A neutral ratio with no good/bad color: high I means many outgoing reasons to change; low I means others depend on this package. Martin's main sequence wants I to balance abstractness A so that A + I ≈ 1.",
			Direction:      DirectionNeutral,
			Bounded:        true,
			Example:        "Ca = 1 and Ce = 3: I = 0.75. An isolated package is I = 0.",
		},
		{
			Name:           metrics.MetricDistance,
			Label:          abbrev(metrics.MetricDistance),
			FullName:       "Distance from the Main Sequence",
			Scope:          DocScopePackage,
			Definition:     metrics.DefinitionDistance,
			FormulaMathML:  formulaDistance,
			FormulaLaTeX:   `D = \lvert A + I - 1 \rvert`,
			Summary:        "How far a package sits from the ideal abstractness–instability balance.",
			HowCalculated:  "The absolute distance of the package's abstractness A and instability I from the 'main sequence' line A + I = 1, where abstraction and stability balance. A and I are reported beside D. An isolated package (Ca + Ce = 0) is treated as maximally stable (I = 0). Only distance is policy-gateable.",
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
	}
}
