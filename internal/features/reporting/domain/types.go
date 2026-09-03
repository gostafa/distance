// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	"strings"

	"github.com/gostafa/distance/distance"
)

type (

	// DocScope is the applicability scope of a metric guide entry.
	DocScope = string

	// MetricDoc is one metrics guide entry shown in docs and the web report.
	MetricDoc struct {
		// FormulaLaTeX is the metric formula in LaTeX.
		FormulaLaTeX string
		// NotApplicable explains when the metric does not apply.
		NotApplicable string
		// FullName is the human-readable metric title.
		FullName string
		// Scope classifies whether the metric is package- or structural-level.
		Scope DocScope
		// Definition is the formula/version identifier for the metric.
		Definition string
		// FormulaMathML is the metric formula in MathML.
		FormulaMathML string
		// HowCalculated describes how the tool computes the value.
		HowCalculated string
		// Example is a short worked example of the metric.
		Example string
		// Label is the short column header for the metric.
		Label string
		// Interpretation explains how to read high versus low scores.
		Interpretation string
		// Name is the canonical metric identifier.
		Name string
		// Direction is whether higher or lower values are better.
		Direction string
		// Summary is a one-line description of the metric.
		Summary string
		// Bounded is true when the metric is normalized to [0, 1].
		Bounded bool
	}

	// Format is a report output format name.
	Format = string

	// TextOptions configures human-readable text rendering.
	TextOptions struct {
		// Color wraps values in ANSI quality colors. Callers enable it only
		// when the destination understands escapes (a terminal).
		Color bool
		// Explain appends a notes section with the reasons behind n/a cells
		// and dropped metric components.
		Explain bool
	}

	cellSizer interface {
		width() int
	}

	tableCell struct {
		prefix string
		text   string
		style  string
	}

	// treeNode is one path segment in the text report tree.
	// Kept as a named type because Go forbids recursive type aliases.
	treeNode struct {
		name     string
		pkg      *treePkg
		children []*treeNode
	}

	// treePkg holds leaf package metrics without naming distance.PackageReport
	// (which would raise treeNode CBO above the gate).
	treePkg = struct {
		metrics []treeMetric
	}

	treeMetric = struct {
		name       string
		value      float64
		applicable bool
	}

	treeSummary = struct {
		pkgAgg    map[string]*columnStats
		pkgsTotal int
	}

	textTable = struct {
		pkgCols []string
		rows    [][]tableCell
		widths  []int
		sawNA   bool
	}

	columnStats = struct {
		sum   float64
		count int
	}

	walkFrame = struct {
		prefix    string
		connector string
	}

	noteIn = struct {
		buf   *strings.Builder
		pkg   *distance.PackageReport
		notes []string
	}

	writeRowIn = struct {
		table   *textTable
		builder *strings.Builder
		opts    *TextOptions
		row     []tableCell
	}

	writeRowCellsIn = struct {
		opts    *TextOptions
		builder *strings.Builder
		cells   []tableCell
		widths  []int
	}

	writeFootersIn = struct {
		table *textTable
		buf   *strings.Builder
		rep   *distance.Report
		opt   *TextOptions
	}

	emitNodeIn = struct {
		table *textTable
		node  *treeNode
		sums  map[*treeNode]*treeSummary
		frame *walkFrame
	}

	emitChildrenIn = struct {
		table  *textTable
		node   *treeNode
		sums   map[*treeNode]*treeSummary
		prefix string
	}

	nodeRowIn = struct {
		table   *textTable
		node    *treeNode
		summary *treeSummary
		label   string
	}
)
