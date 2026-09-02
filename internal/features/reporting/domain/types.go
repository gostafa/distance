// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	"strings"

	"github.com/gostafa/distance/distance"
)

type (

	// DocScope is the applicability scope of a metric guide entry.
	DocScope string

	// MetricDoc is one metrics guide entry shown in docs and the web report.
	MetricDoc struct {
		FormulaLaTeX   string
		NotApplicable  string
		FullName       string
		Scope          DocScope
		Definition     string
		FormulaMathML  string
		HowCalculated  string
		Example        string
		Label          string
		Interpretation string
		Name           string
		Direction      string
		Summary        string
		Bounded        bool
	}

	// Format is a report output format name.
	Format string

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

	colorist interface {
		paint(text, style string) string
	}

	childFinder interface {
		child(name string) *treeNode
	}

	columnMeasurer interface {
		measureColumns()
	}

	rowEmitter interface {
		headerRow() []tableCell
	}

	tableCell struct {
		prefix string
		text   string
		style  string
	}

	treeNode struct {
		name     string
		pkg      *distance.PackageReport
		children []*treeNode
	}

	treeSummary struct {
		pkgAgg    map[string]*columnStats
		pkgsTotal int
	}

	textTable struct {
		pkgCols []string
		rows    [][]tableCell
		widths  []int
		sawNA   bool
	}

	columnStats struct {
		sum   float64
		count int
	}

	walkFrame struct {
		prefix    string
		connector string
	}

	noteIn struct {
		buf   *strings.Builder
		pkg   *distance.PackageReport
		notes []string
	}
)
