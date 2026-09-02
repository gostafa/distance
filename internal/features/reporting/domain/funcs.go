// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gostafa/distance/distance"
)

// MetricDocs returns the ordered metrics guide entries.
func MetricDocs() []MetricDoc {
	return []MetricDoc{
		abstractnessDoc(),
		instabilityDoc(),
		distanceDoc(),
		afferentDoc(),
		efferentDoc(),
	}
}

func abstractnessDoc() MetricDoc {
	return MetricDoc{
		Name:           string(distance.MetricAbstractness),
		Label:          abbrev(string(distance.MetricAbstractness)),
		FullName:       "Abstractness",
		Scope:          DocScopePackage,
		Definition:     distance.DefinitionAbstractness,
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

func instabilityDoc() MetricDoc {
	return MetricDoc{
		Name:           string(distance.MetricInstability),
		Label:          abbrev(string(distance.MetricInstability)),
		FullName:       "Instability",
		Scope:          DocScopePackage,
		Definition:     distance.DefinitionInstability,
		FormulaMathML:  formulaInstability,
		FormulaLaTeX:   `I = \frac{C_e}{C_a + C_e}`,
		Summary:        "How independently a package can change, from its coupling.",
		HowCalculated:  howInstability,
		Interpretation: interpInstability,
		Direction:      DirectionNeutral,
		Bounded:        true,
		Example:        exampleInstability,
	}
}

func distanceDoc() MetricDoc {
	return MetricDoc{
		Name:           string(distance.MetricDistance),
		Label:          abbrev(string(distance.MetricDistance)),
		FullName:       "Distance from the Main Sequence",
		Scope:          DocScopePackage,
		Definition:     distance.DefinitionDistance,
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

func afferentDoc() MetricDoc {
	return MetricDoc{
		Name:     "ca",
		Label:    "Ca",
		FullName: "Afferent coupling",
		Scope:    DocScopeStructural,
		Summary:  "How many analyzed packages import this package.",
		HowCalculated: "Counted within the analyzed set only — importers " +
			"outside the analysis are not observable, so the value depends " +
			"on the patterns you analyze.",
		Interpretation: "A neutral count with no good/bad color. High Ca marks " +
			"load-bearing packages: many others break when this one changes, " +
			"so it should be stable and well tested. It is the incoming half " +
			"of instability.",
		Direction: DirectionNeutral,
		Example:   "If 3 analyzed packages import example.com/m/util, its Ca is 3.",
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
	}
}

// ParseFormat maps a format name to a Format value.
func ParseFormat(name string) (Format, bool) {
	switch Format(name) {
	case FormatText, FormatJSON, FormatCSV, FormatWeb:
		return Format(name), true
	default:
		return emptyString, false
	}
}

// FormatValue formats a metric float for CSV and similar outputs.
func FormatValue(value float64) string {
	return strconv.FormatFloat(value, 'g', -1, floatBitSize)
}

func abbrev(name string) string {
	if name == string(distance.MetricAbstractness) {
		return "A"
	}

	if name == string(distance.MetricInstability) {
		return "I"
	}

	if name == string(distance.MetricDistance) {
		return "Dist"
	}

	return strings.ToUpper(name)
}

func qualityFor(name string) (direction string, bounded, ok bool) {
	if name == string(distance.MetricDistance) {
		return DirectionLower, true, true
	}

	return emptyString, false, false
}

func writeBuilder(builder *strings.Builder, text string) {
	written, err := builder.WriteString(text)

	if err != nil || written != len(text) {
		builder.Reset()
	}
}

func cellWidth(sizer cellSizer) int {
	return sizer.width()
}

func paintWith(painter colorist, text, style string) string {
	return painter.paint(text, style)
}

func findChild(finder childFinder, name string) *treeNode {
	return finder.child(name)
}

func measureTable(measurer columnMeasurer) {
	measurer.measureColumns()
}

func headerCells(emitter rowEmitter) []tableCell {
	return emitter.headerRow()
}

func (cell *tableCell) width() int {
	return utf8.RuneCountInString(cell.prefix) + utf8.RuneCountInString(cell.text)
}

func (opts *TextOptions) paint(text, style string) string {
	if !opts.Color || style == emptyString {
		return text
	}

	return style + text + ansiReset
}

func (node *treeNode) child(name string) *treeNode {
	for i := range node.children {
		if node.children[i].name == name {
			return node.children[i]
		}
	}

	created := &treeNode{name: name}

	node.children = append(node.children, created)

	return created
}

func aggregateTree(node *treeNode, summaries map[*treeNode]*treeSummary) *treeSummary {
	summary := &treeSummary{pkgAgg: make(map[string]*columnStats)}

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

func collectPackageStats(pkg *distance.PackageReport, pkgAgg map[string]*columnStats) {
	for i := range pkg.Metrics {
		if pkg.Metrics[i].Applicable {
			addStat(pkgAgg, pkg.Metrics[i].Name, pkg.Metrics[i].Value)
		}
	}
}

func addStat(statsByName map[string]*columnStats, name string, value float64) {
	stats := statsByName[name]

	if stats == nil {
		stats = &columnStats{}
		statsByName[name] = stats
	}

	stats.sum += value
	stats.count++
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

func (node *treeNode) compress() {
	for node.pkg == nil && len(node.children) == one {
		only := node.children[zero]

		node.name = node.name + pathSep + only.name
		node.pkg = only.pkg
		node.children = only.children
	}

	for i := range node.children {
		node.children[i].
			compress()
	}
}

func buildTree(report *distance.Report) *treeNode {
	root := &treeNode{}

	for i := range report.Packages {
		insertPackage(root, &report.Packages[i], report.Module)
	}

	compressChildren(root)

	return root
}

func insertPackage(root *treeNode, pkg *distance.PackageReport, module string) {
	rel := relPath(pkg.Path, module)

	if rel == pathDot {
		root.pkg = pkg

		return
	}

	attachAtPath(root, pkg, rel)
}

func attachAtPath(root *treeNode, pkg *distance.PackageReport, rel string) {
	node := root

	for seg := range strings.SplitSeq(rel, pathSep) {
		node = findChild(node, seg)
	}

	node.pkg = pkg
}

func compressChildren(root *treeNode) {
	for i := range root.children {
		root.children[i].
			compress()
	}
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

// Text renders a human-readable report.
func Text(report *distance.Report, opts *TextOptions) string {
	var builder strings.Builder

	writeTextHeader(&builder, report, opts)

	if len(report.Packages) == zero {
		return builder.String()
	}

	writeBuilder(&builder, newline)

	table := newTextTable(report)
	table.writeAligned(&builder, opts)
	table.writeFooters(&builder, report, opts)

	return builder.String()
}

func writeTextHeader(builder *strings.Builder, report *distance.Report, opts *TextOptions) {
	writeBuilder(builder, report.ToolName)
	writeBuilder(builder, spaceSep)
	writeBuilder(builder, report.ToolVersion)
	writeBuilder(builder, schemaLabel)
	writeBuilder(builder, report.SchemaVersion)
	writeBuilder(builder, newline)
	writeModuleLine(builder, report, opts)
}

func writeModuleLine(builder *strings.Builder, report *distance.Report, opts *TextOptions) {
	if report.Module == emptyString {
		return
	}

	writeBuilder(builder, modulePrefix)
	writeBuilder(builder, paintWith(opts, report.Module, ansiBold))
	writeBuilder(builder, newline)
}

func newTextTable(report *distance.Report) *textTable {
	table := &textTable{pkgCols: packageColumns(report)}
	summaries := make(map[*treeNode]*treeSummary)

	table.rows = append(table.rows, headerCells(table))

	root := buildTree(report)

	root.name = pathDot
	aggregateTree(root, summaries)
	table.emitTreeBody(root, summaries)

	return table
}

func (t *textTable) headerRow() []tableCell {
	header := make([]tableCell, zero, one+len(t.pkgCols))

	header = append(header, tableCell{text: pathHeader, style: ansiDim})

	for i := range t.pkgCols {
		header = append(header, tableCell{text: abbrev(t.pkgCols[i]), style: ansiDim})
	}

	return header
}

func (t *textTable) emitTreeBody(root *treeNode, summaries map[*treeNode]*treeSummary) {
	if root.pkg != nil {
		t.emitModuleSummary(root, summaries[root])
	}

	for i := range root.children {
		if root.pkg != nil || i > zero {
			t.rows = append(t.rows, nil)
		}

		t.emitNode(root.children[i], summaries, &walkFrame{})
	}
}

func (t *textTable) writeAligned(builder *strings.Builder, opts *TextOptions) {
	measureTable(t)

	for i := range t.rows {
		t.writeRow(builder, opts, t.rows[i])
	}
}

func (t *textTable) measureColumns() {
	count := one + len(t.pkgCols)
	widths := make([]int, zero, count)

	for range count {
		widths = append(widths, zero)
	}

	for i := range t.rows {
		t.measureRow(t.rows[i], widths)
	}

	t.storeWidths(widths)
}

func (t *textTable) measureRow(row []tableCell, widths []int) {
	for col := range row {
		widths[col] = max(widths[col], cellWidth(&row[col]))

		if row[col].text == naCell {
			t.sawNA = true
		}
	}
}

func (t *textTable) storeWidths(widths []int) {
	t.widths = widths
}

func (t *textTable) writeRow(builder *strings.Builder, opts *TextOptions, row []tableCell) {
	if len(row) == zero {
		writeBuilder(builder, newline)

		return
	}

	last := lastNonEmptyCell(row)
	opts.writeRowCells(builder, row[:last+one], t.widths)
	writeBuilder(builder, newline)
}

func lastNonEmptyCell(row []tableCell) int {
	last := len(row) - one

	for last > zero && row[last].text == emptyString && row[last].prefix == emptyString {
		last--
	}

	return last
}

func (opts *TextOptions) writeRowCells(builder *strings.Builder, cells []tableCell, widths []int) {
	last := len(cells) - one

	for col := range cells {
		writeCellContent(builder, opts, &cells[col])

		pad := zero

		if col < last {
			pad = widths[col] - cellWidth(&cells[col]) + two
		}

		writeCellPad(builder, pad)
	}
}

func writeCellContent(builder *strings.Builder, opts *TextOptions, cell *tableCell) {
	writeBuilder(builder, cell.prefix)
	writeBuilder(builder, paintWith(opts, cell.text, cell.style))
}

func writeCellPad(builder *strings.Builder, pad int) {
	if pad <= zero {
		return
	}

	writeBuilder(builder, strings.Repeat(spaceSep, pad))
}

func (t *textTable) writeFooters(buf *strings.Builder, rep *distance.Report, opt *TextOptions) {
	if t.sawNA {
		writeNALegend(buf, opt)
	}

	if opt.Explain {
		writeNotes(buf, rep, opt)
	}
}

func writeNALegend(builder *strings.Builder, opts *TextOptions) {
	writeBuilder(builder, newline)
	writeBuilder(builder, paintWith(opts, naCell+naLegendSuffix, ansiDim))
	writeBuilder(builder, newline)
}

func (t *textTable) emitModuleSummary(node *treeNode, summary *treeSummary) {
	t.nodeRow(node, summary, emptyString)
}

func naTableCell() tableCell {
	return tableCell{text: naCell, style: ansiDim}
}

func (t *textTable) emitNode(node *treeNode, sums map[*treeNode]*treeSummary, frame *walkFrame) {
	t.nodeRow(node, sums[node], frame.prefix+frame.connector)
	t.emitChildren(node, sums, childIndent(frame.prefix, frame.connector))
}

func (t *textTable) emitChildren(node *treeNode, sums map[*treeNode]*treeSummary, prefix string) {
	total := len(node.children)

	for i := range node.children {
		if i > zero {
			t.rows = append(t.rows, []tableCell{{prefix: prefix + pipeGlyph}})
		}

		t.emitNode(node.children[i], sums, &walkFrame{
			prefix:    prefix,
			connector: branchGlyph(i, total),
		})
	}
}

func branchGlyph(index, total int) string {
	if index == total-one {
		return branchLast
	}

	return branchMid
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

func (t *textTable) nodeRow(node *treeNode, summary *treeSummary, label string) {
	row := make([]tableCell, zero, one+len(t.pkgCols))

	row = append(row, tableCell{prefix: label, text: node.name, style: ansiBold})
	row = append(row, nodePkgCells(node, summary, t.pkgCols)...)

	t.rows = append(t.rows, row)
}

func nodePkgCells(node *treeNode, summary *treeSummary, pkgCols []string) []tableCell {
	if node.pkg != nil && len(node.children) == zero {
		return packageMetricCells(node.pkg, pkgCols)
	}

	return meanPkgCells(summary, pkgCols)
}

func meanPkgCells(summary *treeSummary, pkgCols []string) []tableCell {
	cells := make([]tableCell, zero, len(pkgCols))

	for i := range pkgCols {
		cells = append(cells, meanCell(summary.pkgAgg[pkgCols[i]], boundedColorFor(pkgCols[i])))
	}

	return cells
}

func metricsByName(results []distance.MetricResult) map[string]distance.MetricResult {
	byName := make(map[string]distance.MetricResult, len(results))

	for i := range results {
		byName[results[i].Name] = results[i]
	}

	return byName
}

func packageMetricCells(pkg *distance.PackageReport, cols []string) []tableCell {
	byName := metricsByName(pkg.Metrics)
	cells := make([]tableCell, zero, len(cols))

	for i := range cols {
		cells = append(cells, metricCell(byName, cols[i]))
	}

	return cells
}

func metricCell(byName map[string]distance.MetricResult, name string) tableCell {
	result, ok := byName[name]

	switch {
	case !ok:
		return tableCell{}
	case !result.Applicable:
		return naTableCell()
	default:
		return tableCell{
			text:  formatCell(result.Value),
			style: ansiBold + boundedColor(name, result.Value),
		}
	}
}

func meanCell(stats *columnStats, color func(float64) string) tableCell {
	if stats == nil || stats.count == zero {
		return naTableCell()
	}

	value := stats.sum / float64(stats.count)

	return tableCell{text: formatCell(value), style: ansiBold + color(value)}
}

func boundedColorFor(name string) func(float64) string {
	return func(value float64) string { return boundedColor(name, value) }
}

func packageColumns(report *distance.Report) []string {
	return filterReportedMetrics(metricNamesPresent(report))
}

func metricNamesPresent(report *distance.Report) map[string]bool {
	present := make(map[string]bool)

	for i := range report.Packages {
		markMetricNames(present, report.Packages[i].Metrics)
	}

	return present
}

func markMetricNames(present map[string]bool, results []distance.MetricResult) {
	for i := range results {
		present[results[i].Name] = true
	}
}

func filterReportedMetrics(present map[string]bool) []string {
	var cols []string

	for _, name := range distance.AllMetrics() {
		key := string(name)

		if present[key] {
			cols = append(cols, key)
		}
	}

	return cols
}

func writeNotes(builder *strings.Builder, report *distance.Report, opts *TextOptions) {
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

		opts.writeNoteBlock(&noteIn{buf: builder, pkg: &report.Packages[i], notes: notes})
	}
}

func writeNotesHeader(builder *strings.Builder, opts *TextOptions) {
	writeBuilder(builder, newline)
	writeBuilder(builder, paintWith(opts, notesTitle, ansiDim))
	writeBuilder(builder, newline)
}

func (opts *TextOptions) writeNoteBlock(input *noteIn) {
	writeBuilder(input.buf, notePkgIndent)
	writeBuilder(input.buf, paintWith(opts, input.pkg.Path, ansiDim))
	writeBuilder(input.buf, newline)

	for i := range input.notes {
		writeNoteLine(input.buf, input.notes[i], opts)
	}
}

func writeNoteLine(builder *strings.Builder, note string, opts *TextOptions) {
	writeBuilder(builder, noteLineIndent)
	writeBuilder(builder, paintWith(opts, note, ansiDim))
	writeBuilder(builder, newline)
}

func packageNotes(pkg *distance.PackageReport) []string {
	var notes []string

	for i := range pkg.Metrics {
		if pkg.Metrics[i].Reason != emptyString {
			notes = append(notes, pkg.Metrics[i].Name+": "+pkg.Metrics[i].Reason)
		}
	}

	return notes
}

func formatCell(value float64) string {
	return strconv.FormatFloat(value, 'f', two, floatBitSize)
}

func boundedColor(name string, value float64) string {
	direction, bounded, ok := qualityFor(name)

	if !ok || !bounded {
		return emptyString
	}

	return thresholdColor(direction, value)
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

// CSVHeader returns the fixed CSV column names.
func CSVHeader() []string {
	return []string{
		string(DocScopePackage),
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
func CSVRecords(report *distance.Report) [][]string {
	records := make([][]string, zero, len(report.Packages))

	for i := range report.Packages {
		pkg := &report.Packages[i]

		records = append(records, metricCSVRows(pkg.Path, emptyString, pkg.Metrics)...)
	}

	return records
}

func metricCSVRows(pkgPath, typeName string, results []distance.MetricResult) [][]string {
	rows := make([][]string, zero, len(results))

	for i := range results {
		rows = append(rows, metricCSVRow(pkgPath, typeName, &results[i]))
	}

	return rows
}

func metricCSVRow(pkgPath, typeName string, result *distance.MetricResult) []string {
	value := emptyString

	if result.Applicable {
		value = FormatValue(result.Value)
	}

	return []string{
		pkgPath, typeName, result.Name, result.Scope, value,
		strconv.FormatBool(result.Applicable), result.Reason, result.Definition,
	}
}
