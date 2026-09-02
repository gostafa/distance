// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

const (

	// DocScopePackage marks a package-level metric entry.
	DocScopePackage DocScope = "package"
	// DocScopeStructural marks a counted column (Ca, Funcs, Fields, …).
	DocScopeStructural DocScope = "structural"

	// DirectionLower means lower values are better.
	DirectionLower = "lower"
	// DirectionHigher means higher values are better.
	DirectionHigher = "higher"
	// DirectionNeutral means the metric has no good/bad direction.
	DirectionNeutral = "neutral"

	formulaAbstractness = `<math display="block" ` +
		`alttext="A = \frac{N_{interface}}{N_{named}}">` +
		`<mrow><mi>A</mi><mo>=</mo><mfrac>` +
		`<msub><mi>N</mi><mtext>interface</mtext></msub>` +
		`<msub><mi>N</mi><mtext>named</mtext></msub>` +
		`</mfrac></mrow></math>`

	formulaInstability = `<math display="block" ` +
		`alttext="I = \frac{C_e}{C_a + C_e}">` +
		`<mrow><mi>I</mi><mo>=</mo><mfrac>` +
		`<msub><mi>C</mi><mi>e</mi></msub>` +
		`<mrow><msub><mi>C</mi><mi>a</mi></msub><mo>+</mo>` +
		`<msub><mi>C</mi><mi>e</mi></msub></mrow>` +
		`</mfrac></mrow></math>`

	formulaDistance = `<math display="block" alttext="D = |A + I - 1|">` +
		`<mrow><mi>D</mi><mo>=</mo><mo stretchy="false">|</mo>` +
		`<mi>A</mi><mo>+</mo><mi>I</mi><mo>−</mo><mn>1</mn>` +
		`<mo stretchy="false">|</mo></mrow></math>`

	// FormatText renders a human-readable report.
	FormatText Format = "text"
	// FormatJSON renders the versioned JSON schema.
	FormatJSON Format = "json"
	// FormatCSV renders one row per entity and metric.
	FormatCSV Format = "csv"
	// FormatWeb renders a self-contained interactive HTML report.
	FormatWeb Format = "web"

	ansiReset  = "\x1b[0m"
	ansiBold   = "\x1b[1m"
	ansiDim    = "\x1b[2m"
	ansiRed    = "\x1b[31m"
	ansiGreen  = "\x1b[32m"
	ansiYellow = "\x1b[33m"

	naCell           = "–"
	emptyString      = ""
	newline          = "\n"
	spaceSep         = " "
	pathDot          = "."
	pathSep          = "/"
	modulePrefix     = "module "
	schemaLabel      = " — schema "
	pathHeader       = "PATH"
	naLegendSuffix   = " = not applicable"
	notesTitle       = "notes"
	notePkgIndent    = "  "
	noteLineIndent   = "    "
	branchMid        = "├── "
	branchLast       = "└── "
	indentMid        = "│   "
	pipeGlyph        = "│"
	csvColType       = "type"
	csvColMetric     = "metric"
	csvColScope      = "scope"
	csvColValue      = "value"
	csvColApplicable = "applicable"
	csvColReason     = "reason"
	csvColDefinition = "definition"

	zero                 = 0
	one                  = 1
	two                  = 2
	floatBitSize         = 64
	qualityHighThreshold = 0.66
	qualityMidThreshold  = 0.33

	howAbstractness = "Named interface types divided by the package's relevant " +
		"named types. Type aliases are excluded from both counts. " +
		"Reported next to instability and distance; not selectable or " +
		"gateable on its own."
	interpAbstractness = "A neutral ratio with no good/bad color: neither fully " +
		"concrete nor fully abstract is universally better. Martin's main " +
		"sequence wants A to balance instability I so that A + I ≈ 1."
	exampleAbstractness = "1 interface among 4 named types: A = 0.25."

	howInstability = "Ce / (Ca + Ce) within the configured dependency scope. " +
		"An isolated package (Ca + Ce = 0) is treated as maximally stable: " +
		"I = 0. Reported next to abstractness and distance; not selectable " +
		"or gateable on its own."
	interpInstability = "A neutral ratio with no good/bad color: high I means " +
		"many outgoing reasons to change; low I means others depend on " +
		"this package. Martin's main sequence wants I to balance " +
		"abstractness A so that A + I ≈ 1."
	exampleInstability = "Ca = 1 and Ce = 3: I = 0.75. An isolated package is I = 0."

	howDistance = "The absolute distance of the package's abstractness A " +
		"and instability I from the 'main sequence' line A + I = 1, where " +
		"abstraction and stability balance. A and I are reported beside D. " +
		"An isolated package (Ca + Ce = 0) is treated as maximally stable " +
		"(I = 0). Only distance is policy-gateable."
	interpDistance = "0 is on the main sequence. High distance means the " +
		"package is either concrete and stable (rigid — everything depends " +
		"on its details) or abstract and unstable (abstractions nobody " +
		"depends on). Mind the isolated-package convention: a concrete " +
		"isolated package has A = 0 and I = 0, so D = 1 by definition, " +
		"not necessarily by design fault."
	naDistance = "When abstractness is not applicable — for example, " +
		"the package declares no relevant named types."
	exampleDistance = "A = 0.25 and I = 0.5: D = |0.25 + 0.5 − 1| = 0.25."
	summaryDistance = "How far a package sits from the ideal " +
		"abstractness–instability balance."
)
