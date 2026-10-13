// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package presentation

import (
	"strconv"
	"strings"
	"unicode/utf8"

	reportmodel "github.com/gostafa/distance/internal/features/reporting/domain"
)

// CSVHeader returns the fixed CSV column names.
func CSVHeader() []string {
	return []string{
		DocScopePackage,
		csvColType,
		csvColMetric,
		csvColScope,
		csvColValue,
		csvColApplicable,
		csvColReason,
		csvColDefinition,
	}
}

// CSVRecords returns one CSV row per package metric.
func CSVRecords(report *reportmodel.Report) [][]string {
	records := make([][]string, zero, len(report.Packages))

	for i := range report.Packages {
		pkg := &report.Packages[i]

		records = append(records, metricCSVRows(pkg.Path, emptyString, pkg.Metrics)...)
	}

	return records
}

// FormatValue formats a metric float for CSV and similar outputs.
func FormatValue(value float64) string {
	return strconv.FormatFloat(value, 'g', -1, floatBitSize)
}

// MetricDocs returns the ordered metrics guide entries.
func MetricDocs() []MetricDoc {
	docs := []MetricDoc{
		abstractnessDoc(),
		instabilityDoc(),
		distanceDoc(),
		afferentDoc(),
		efferentDoc(),
	}

	return docs
}

// ParseFormat maps a format name to a Format value.
func ParseFormat(name string) (Format, bool) { return reportmodel.ParseFormat(name) }

// Text renders a human-readable report.
func Text(report *reportmodel.Report, opts *TextOptions) string {
	builder := []byte{}

	writeTextHeader(&builder, report, opts)

	if len(report.Packages) == zero {
		return string(builder)
	}

	writeBuilder(&builder, newline)

	table := newTextTable(report)
	writeAligned(table, &builder, opts)
	writeFooters(&writeFootersIn{
		table: table,
		buf:   &builder,
		rep:   report,
		opt:   opts,
	})

	return string(builder)
}

func (cell *tableCell) width() int {
	return utf8.RuneCountInString(cell.prefix) + utf8.RuneCountInString(cell.text)
}

func abbrev(name string) string {
	if name == reportmodel.MetricAbstractness {
		return "A"
	}

	if name == reportmodel.MetricInstability {
		return "I"
	}

	if name == reportmodel.MetricDistance {
		return "Dist"
	}

	return strings.ToUpper(name)
}

func abstractnessDoc() MetricDoc {
	return MetricDoc{
		Name:           reportmodel.MetricAbstractness,
		Label:          abbrev(reportmodel.MetricAbstractness),
		FullName:       "Abstractness",
		Scope:          DocScopePackage,
		Definition:     reportmodel.DefinitionAbstractness,
		FormulaMathML:  formulaAbstractness,
		FormulaLaTeX:   `A = \frac{N_{\text{interface}}}{N_{\text{named}}}`,
		Summary:        "The share of a package's named types that are interfaces.",
		HowCalculated:  howAbstractness,
		Interpretation: interpAbstractness,
		NotApplicable:  "When the package declares no relevant named types.",
		Direction:      DirectionNeutral,
		Bounded:        true,
		Example:        exampleAbstractness,
	}
}

func addStat(statsByName map[string]*columnStats, name string, value float64) {
	stats := statsByName[name]

	if stats == nil {
		stats = &columnStats{sum: zero, count: zero}
		statsByName[name] = stats
	}

	stats.sum += value
	stats.count++
}

func afferentDoc() MetricDoc {
	return MetricDoc{
		Name:           "ca",
		Label:          "Ca",
		FullName:       "Afferent coupling",
		Scope:          DocScopeStructural,
		Summary:        "How many analyzed packages import this package.",
		HowCalculated:  howAfferent,
		Interpretation: interpretAfferent,
		Direction:      DirectionNeutral,
		Example:        "If 3 analyzed packages import example.com/m/util, its Ca is 3.",
		FormulaLaTeX:   "",
		NotApplicable:  "",
		Definition:     "",
		FormulaMathML:  "",
		Bounded:        false,
	}
}

func aggregateTree(node *treeNode, summaries map[*treeNode]*treeSummary) *treeSummary {
	summary := &treeSummary{pkgAgg: make(map[string]*columnStats), pkgsTotal: zero}

	summaries[node] = summary

	if node.pkg != nil {
		summary.pkgsTotal = one
		collectPackageStats(node.pkg, summary.pkgAgg)
	}

	for i := range node.children {
		childSummary := aggregateTree(node.children[i], summaries)

		summary.pkgsTotal += childSummary.pkgsTotal
		mergeStats(summary.pkgAgg, childSummary.pkgAgg)
	}

	return summary
}

func attachAtPath(root *treeNode, pkg *reportmodel.PackageReport, rel string) {
	node := root

	for seg := range strings.SplitSeq(rel, pathSep) {
		node = findChild(node, seg)
	}

	node.pkg = toTreePkg(pkg)
}

func boundedColor(name string, value float64) string {
	direction, bounded, ok := qualityFor(name)

	if !ok || !bounded {
		return emptyString
	}

	return thresholdColor(direction, value)
}

func boundedColorFor(name string) func(float64) string {
	return func(value float64) string { return boundedColor(name, value) }
}

func branchGlyph(index, total int) string {
	if index == total-one {
		return branchLast
	}

	return branchMid
}

func buildTree(report *reportmodel.Report) *treeNode {
	root := &treeNode{name: "", pkg: nil, children: nil}

	for i := range report.Packages {
		insertPackage(root, &report.Packages[i], report.Module)
	}

	compressChildren(root)

	return root
}

func cellWidth(sizer cellSizer) int {
	return sizer.width()
}

func childIndent(prefix, connector string) string {
	switch connector {
	case branchMid:
		return prefix + indentMid
	case branchLast:
		return prefix + noteLineIndent
	default:
		return prefix
	}
}

func collectPackageStats(pkg *treePkg, pkgAgg map[string]*columnStats) {
	for i := range pkg.metrics {
		if pkg.metrics[i].applicable {
			addStat(pkgAgg, pkg.metrics[i].name, pkg.metrics[i].value)
		}
	}
}

func compressChildren(root *treeNode) {
	for i := range root.children {
		compressNode(root.children[i])
	}
}

func compressNode(node *treeNode) {
	for node.pkg == nil && len(node.children) == one {
		only := node.children[zero]

		node.name = node.name + pathSep + only.name
		node.pkg = only.pkg
		node.children = only.children
	}

	for i := range node.children {
		compressNode(node.children[i])
	}
}

func distanceDoc() MetricDoc {
	return MetricDoc{
		Name:           reportmodel.MetricDistance,
		Label:          abbrev(reportmodel.MetricDistance),
		FullName:       "Distance from the Main Sequence",
		Scope:          DocScopePackage,
		Definition:     reportmodel.DefinitionDistance,
		FormulaMathML:  formulaDistance,
		FormulaLaTeX:   `D = \lvert A + I - 1 \rvert`,
		Summary:        summaryDistance,
		HowCalculated:  howDistance,
		Interpretation: interpDistance,
		NotApplicable:  naDistance,
		Direction:      DirectionLower,
		Bounded:        true,
		Example:        exampleDistance,
	}
}

func efferentDoc() MetricDoc {
	return MetricDoc{
		Name:     "ce",
		Label:    "Ce",
		FullName: "Efferent coupling",
		Scope:    DocScopeStructural,
		Summary:  "How many packages this package imports, within the dependency scope.",
		HowCalculated: "The package's imports that fall in the configured " +
			"-dependency-scope: project counts only other analyzed packages, " +
			"module counts packages of the main module, all counts every " +
			"import. Duplicates and self-imports are ignored.",
		Interpretation: "A neutral count with no good/bad color. High Ce means " +
			"the package has many reasons to change. It is the outgoing half " +
			"of instability.",
		Direction: DirectionNeutral,
		Example: "A package importing 2 in-scope packages has Ce = 2 " +
			"regardless of how often each is imported.",
		FormulaLaTeX: "", NotApplicable: "", Definition: "", FormulaMathML: "", Bounded: false,
	}
}

func emitChildren(args *emitChildrenIn) {
	total := len(args.node.children)

	for i := range args.node.children {
		if i > zero {
			args.table.rows = append(
				args.table.rows,
				[]tableCell{{prefix: args.prefix + pipeGlyph, text: "", style: ""}},
			)
		}

		emitNode(&emitNodeIn{
			table: args.table,
			node:  args.node.children[i],
			sums:  args.sums,
			frame: &walkFrame{
				prefix:    args.prefix,
				connector: branchGlyph(i, total),
			},
		})
	}
}

func emitModuleSummary(table *textTable, node *treeNode, summary *treeSummary) {
	nodeRow(&nodeRowIn{
		table:   table,
		node:    node,
		summary: summary,
		label:   emptyString,
	})
}

func emitNode(args *emitNodeIn) {
	nodeRow(&nodeRowIn{
		table:   args.table,
		node:    args.node,
		summary: args.sums[args.node],
		label:   args.frame.prefix + args.frame.connector,
	})
	emitChildren(&emitChildrenIn{
		table:  args.table,
		node:   args.node,
		sums:   args.sums,
		prefix: childIndent(args.frame.prefix, args.frame.connector),
	})
}

func emitTreeBody(table *textTable, root *treeNode, summaries map[*treeNode]*treeSummary) {
	if root.pkg != nil {
		emitModuleSummary(table, root, summaries[root])
	}

	for i := range root.children {
		if root.pkg != nil || i > zero {
			table.rows = append(table.rows, nil)
		}

		emitNode(&emitNodeIn{
			table: table,
			node:  root.children[i],
			sums:  summaries,
			frame: &walkFrame{prefix: "", connector: ""},
		})
	}
}

func filterReportedMetrics(present map[string]bool) []string {
	var cols []string

	metrics := []string{
		reportmodel.MetricAbstractness,
		reportmodel.MetricInstability,
		reportmodel.MetricDistance,
	}

	for i := range metrics {
		key := metrics[i]

		if present[key] {
			cols = append(cols, key)
		}
	}

	return cols
}

func findChild(node *treeNode, name string) *treeNode {
	for i := range node.children {
		if node.children[i].name == name {
			return node.children[i]
		}
	}

	created := &treeNode{name: name, pkg: nil, children: nil}

	node.children = append(node.children, created)

	return created
}

func formatCell(value float64) string {
	return strconv.FormatFloat(value, 'f', two, floatBitSize)
}

func headerRow(table *textTable) []tableCell {
	header := make([]tableCell, zero, one+len(table.pkgCols))

	header = append(header, tableCell{text: pathHeader, style: ansiDim, prefix: ""})

	for i := range table.pkgCols {
		header = append(
			header,
			tableCell{text: abbrev(table.pkgCols[i]), style: ansiDim, prefix: ""},
		)
	}

	return header
}

func insertPackage(root *treeNode, pkg *reportmodel.PackageReport, module string) {
	rel := relPath(pkg.Path, module)

	if rel == pathDot {
		root.pkg = toTreePkg(pkg)

		return
	}

	attachAtPath(root, pkg, rel)
}

func instabilityDoc() MetricDoc {
	return MetricDoc{
		Name:           reportmodel.MetricInstability,
		Label:          abbrev(reportmodel.MetricInstability),
		FullName:       "Instability",
		Scope:          DocScopePackage,
		Definition:     reportmodel.DefinitionInstability,
		FormulaMathML:  formulaInstability,
		FormulaLaTeX:   `I = \frac{C_e}{C_a + C_e}`,
		Summary:        "How independently a package can change, from its coupling.",
		HowCalculated:  howInstability,
		Interpretation: interpInstability,
		Direction:      DirectionNeutral,
		Bounded:        true,
		Example:        exampleInstability, NotApplicable: "",
	}
}

func lastNonEmptyCell(row []tableCell) int {
	last := len(row) - one

	for last > zero && row[last].text == emptyString && row[last].prefix == emptyString {
		last--
	}

	return last
}

func markMetricNames(present map[string]bool, results []reportmodel.MetricResult) {
	for i := range results {
		present[results[i].Name] = true
	}
}

func meanCell(stats *columnStats, color func(float64) string) tableCell {
	if stats == nil || stats.count == zero {
		return naTableCell()
	}

	value := stats.sum / float64(stats.count)

	return tableCell{text: formatCell(value), style: ansiBold + color(value), prefix: ""}
}

func meanPkgCells(summary *treeSummary, pkgCols []string) []tableCell {
	cells := make([]tableCell, zero, len(pkgCols))

	for i := range pkgCols {
		cells = append(cells, meanCell(summary.pkgAgg[pkgCols[i]], boundedColorFor(pkgCols[i])))
	}

	return cells
}

func measureColumns(table *textTable) {
	count := one + len(table.pkgCols)
	widths := make([]int, zero, count)

	for range count {
		widths = append(widths, zero)
	}

	for i := range table.rows {
		measureRow(table, table.rows[i], widths)
	}

	table.widths = widths
}

func measureRow(table *textTable, row []tableCell, widths []int) {
	for col := range row {
		widths[col] = max(widths[col], cellWidth(&row[col]))

		if row[col].text == naCell {
			table.sawNA = true
		}
	}
}

func mergeStats(dst, src map[string]*columnStats) {
	for name := range src {
		dest := dst[name]

		if dest == nil {
			c := *src[name]

			dst[name] = &c

			continue
		}

		dest.sum += src[name].sum
		dest.count += src[name].count
	}
}

func metricCSVRow(pkgPath, typeName string, result *reportmodel.MetricResult) []string {
	value := emptyString

	if result.Applicable {
		value = FormatValue(result.Value)
	}

	return []string{
		pkgPath, typeName, result.Name, result.Scope, value,
		strconv.FormatBool(result.Applicable), result.Reason, result.Definition,
	}
}

func metricCSVRows(pkgPath, typeName string, results []reportmodel.MetricResult) [][]string {
	rows := make([][]string, zero, len(results))

	for i := range results {
		rows = append(rows, metricCSVRow(pkgPath, typeName, &results[i]))
	}

	return rows
}

func metricCell(byName map[string]treeMetric, name string) tableCell {
	result, ok := byName[name]

	switch {
	case !ok:
		return tableCell{prefix: "", text: "", style: ""}
	case !result.applicable:
		return naTableCell()
	default:
		return tableCell{
			text:  formatCell(result.value),
			style: ansiBold + boundedColor(name, result.value), prefix: "",
		}
	}
}

func metricNamesPresent(report *reportmodel.Report) map[string]bool {
	present := make(map[string]bool)

	for i := range report.Packages {
		markMetricNames(present, report.Packages[i].Metrics)
	}

	return present
}

func metricsByName(results []treeMetric) map[string]treeMetric {
	byName := make(map[string]treeMetric, len(results))

	for i := range results {
		byName[results[i].name] = results[i]
	}

	return byName
}

func naTableCell() tableCell {
	return tableCell{text: naCell, style: ansiDim, prefix: ""}
}

func newTextTable(report *reportmodel.Report) *textTable {
	table := &textTable{pkgCols: packageColumns(report), rows: nil, widths: nil, sawNA: false}
	summaries := make(map[*treeNode]*treeSummary)

	table.rows = append(table.rows, headerRow(table))

	root := buildTree(report)

	root.name = pathDot
	aggregateTree(root, summaries)
	emitTreeBody(table, root, summaries)

	return table
}

func nodePkgCells(node *treeNode, summary *treeSummary, pkgCols []string) []tableCell {
	if node.pkg != nil && len(node.children) == zero {
		return packageMetricCells(node.pkg, pkgCols)
	}

	return meanPkgCells(summary, pkgCols)
}

func nodeRow(args *nodeRowIn) {
	row := make([]tableCell, zero, one+len(args.table.pkgCols))

	row = append(row, tableCell{prefix: args.label, text: args.node.name, style: ansiBold})
	row = append(row, nodePkgCells(args.node, args.summary, args.table.pkgCols)...)

	args.table.rows = append(args.table.rows, row)
}

func packageColumns(report *reportmodel.Report) []string {
	return filterReportedMetrics(metricNamesPresent(report))
}

func packageMetricCells(pkg *treePkg, cols []string) []tableCell {
	byName := metricsByName(pkg.metrics)
	cells := make([]tableCell, zero, len(cols))

	for i := range cols {
		cells = append(cells, metricCell(byName, cols[i]))
	}

	return cells
}

func packageNotes(pkg *reportmodel.PackageReport) []string {
	var notes []string

	for i := range pkg.Metrics {
		if pkg.Metrics[i].Reason != emptyString {
			notes = append(notes, pkg.Metrics[i].Name+": "+pkg.Metrics[i].Reason)
		}
	}

	return notes
}

func paint(opts *TextOptions, text, style string) string {
	if !opts.Color || style == emptyString {
		return text
	}

	return style + text + ansiReset
}

func qualityFor(name string) (direction string, bounded, ok bool) {
	if name == reportmodel.MetricDistance {
		return DirectionLower, true, true
	}

	return emptyString, false, false
}

func relPath(path, module string) string {
	if module == emptyString {
		return path
	}

	if path == module {
		return pathDot
	}

	if strings.HasPrefix(path, module+pathSep) {
		return path[len(module)+one:]
	}

	return path
}

func thresholdColor(direction string, score float64) string {
	if direction == DirectionLower {
		score = one - score
	}

	switch {
	case score >= qualityHighThreshold:
		return ansiGreen
	case score >= qualityMidThreshold:
		return ansiYellow
	default:
		return ansiRed
	}
}

func toTreePkg(pkg *reportmodel.PackageReport) *treePkg {
	metrics := make([]treeMetric, zero, len(pkg.Metrics))

	for i := range pkg.Metrics {
		metrics = append(metrics, treeMetric{
			name:       pkg.Metrics[i].Name,
			value:      pkg.Metrics[i].Value,
			applicable: pkg.Metrics[i].Applicable,
		})
	}

	return &treePkg{metrics: metrics}
}

func writeAligned(table *textTable, builder *[]byte, opts *TextOptions) {
	measureColumns(table)

	for i := range table.rows {
		writeRow(&writeRowIn{
			table:   table,
			builder: builder,
			opts:    opts,
			row:     table.rows[i],
		})
	}
}

func writeBuilder(builder *[]byte, text string) { *builder = append(*builder, text...) }

func writeCellContent(builder *[]byte, opts *TextOptions, cell *tableCell) {
	writeBuilder(builder, cell.prefix)
	writeBuilder(builder, paint(opts, cell.text, cell.style))
}

func writeCellPad(builder *[]byte, pad int) {
	if pad <= zero {
		return
	}

	writeBuilder(builder, strings.Repeat(spaceSep, pad))
}

func writeFooters(args *writeFootersIn) {
	if args.table.sawNA {
		writeNALegend(args.buf, args.opt)
	}

	if args.opt.Explain {
		writeNotes(args.buf, args.rep, args.opt)
	}
}

func writeModuleLine(builder *[]byte, report *reportmodel.Report, opts *TextOptions) {
	if report.Module == emptyString {
		return
	}

	writeBuilder(builder, modulePrefix)
	writeBuilder(builder, paint(opts, report.Module, ansiBold))
	writeBuilder(builder, newline)
}

func writeNALegend(builder *[]byte, opts *TextOptions) {
	writeBuilder(builder, newline)
	writeBuilder(builder, paint(opts, naCell+naLegendSuffix, ansiDim))
	writeBuilder(builder, newline)
}

func writeNoteBlock(opts *TextOptions, input *noteIn) {
	writeBuilder(input.buf, notePkgIndent)
	writeBuilder(input.buf, paint(opts, input.pkg.Path, ansiDim))
	writeBuilder(input.buf, newline)

	for i := range input.notes {
		writeNoteLine(input.buf, input.notes[i], opts)
	}
}

func writeNoteLine(builder *[]byte, note string, opts *TextOptions) {
	writeBuilder(builder, noteLineIndent)
	writeBuilder(builder, paint(opts, note, ansiDim))
	writeBuilder(builder, newline)
}

func writeNotes(builder *[]byte, report *reportmodel.Report, opts *TextOptions) {
	headerDone := false

	for i := range report.Packages {
		notes := packageNotes(&report.Packages[i])

		if len(notes) == zero {
			continue
		}

		if !headerDone {
			writeNotesHeader(builder, opts)

			headerDone = true
		}

		writeNoteBlock(opts, &noteIn{buf: builder, pkg: &report.Packages[i], notes: notes})
	}
}

func writeNotesHeader(builder *[]byte, opts *TextOptions) {
	writeBuilder(builder, newline)
	writeBuilder(builder, paint(opts, notesTitle, ansiDim))
	writeBuilder(builder, newline)
}

func writeRow(args *writeRowIn) {
	if len(args.row) == zero {
		writeBuilder(args.builder, newline)

		return
	}

	last := lastNonEmptyCell(args.row)
	writeRowCells(&writeRowCellsIn{
		opts:    args.opts,
		builder: args.builder,
		cells:   args.row[:last+one],
		widths:  args.table.widths,
	})
	writeBuilder(args.builder, newline)
}

func writeRowCells(args *writeRowCellsIn) {
	last := len(args.cells) - one

	for col := range args.cells {
		writeCellContent(args.builder, args.opts, &args.cells[col])

		pad := zero

		if col < last {
			pad = args.widths[col] - cellWidth(&args.cells[col]) + two
		}

		writeCellPad(args.builder, pad)
	}
}

func writeTextHeader(builder *[]byte, report *reportmodel.Report, opts *TextOptions) {
	writeBuilder(builder, report.ToolName)
	writeBuilder(builder, spaceSep)
	writeBuilder(builder, report.ToolVersion)
	writeBuilder(builder, schemaLabel)
	writeBuilder(builder, report.SchemaVersion)
	writeBuilder(builder, newline)
	writeModuleLine(builder, report, opts)
}
