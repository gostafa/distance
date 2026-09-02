// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/gostafa/distance/distance"
	reportdomain "github.com/gostafa/distance/internal/features/reporting/domain"
)

func defaultRuntime() reportingRuntime {
	return reportingRuntime{jsonMarshal: json.Marshal}
}

func (err missingPlaceholderError) Error() string {
	return err.message
}

func (err unknownFormatError) Error() string {
	return "unknown report format " + string(err.format)
}

func closeAfter(prior error, closer io.Closer) error {
	closeErr := closer.Close()

	if prior != nil {
		return prior
	}

	if closeErr != nil {
		return fmt.Errorf(errWrapClose, closeErr)
	}

	return nil
}

func writeBuffer(buf *bytes.Buffer, payload []byte) error {
	written, writeErr := buf.Write(payload)
	if writeErr != nil {
		return fmt.Errorf(errWrapWriteBuffer, writeErr)
	}

	if written != len(payload) {
		return fmt.Errorf(errWrapWriteBuffer, errShortWrite)
	}

	return nil
}

func writeBufferByte(buf *bytes.Buffer, value byte) error {
	writeErr := buf.WriteByte(value)
	if writeErr != nil {
		return fmt.Errorf(errWrapWriteBuffer, writeErr)
	}

	return nil
}

func formatRenderers() map[reportdomain.Format]func(io.Writer, *renderOptions) error {
	return map[reportdomain.Format]func(io.Writer, *renderOptions) error{
		reportdomain.FormatText: renderText,
		reportdomain.FormatJSON: renderJSONOpts,
		reportdomain.FormatCSV:  renderCSVOpts,
		reportdomain.FormatWeb:  renderWebOpts,
	}
}

func renderJSONOpts(writer io.Writer, opts *renderOptions) error {
	err := renderJSON(writer, opts.report)
	if err != nil {
		return fmt.Errorf("application renderJSONOpts: %w", err)
	}

	return nil
}

func renderCSVOpts(writer io.Writer, opts *renderOptions) error {
	err := renderCSV(writer, opts.report)
	if err != nil {
		return fmt.Errorf("application renderCSVOpts: %w", err)
	}

	return nil
}

func renderWebOpts(writer io.Writer, opts *renderOptions) error {
	err := renderWeb(writer, opts.report)
	if err != nil {
		return fmt.Errorf("application renderWebOpts: %w", err)
	}

	return nil
}

func marshalDocsWith(runtime reportingRuntime, toolVersion string) ([]byte, error) {
	entries := reportdomain.MetricDocs()

	out := docsPayload{
		Tool: jsonTool{Name: string(distance.MetricDistance), Version: toolVersion},
		Docs: make([]jsonMetricDoc, indexZero, len(entries)),
	}

	for i := range entries {
		out.Docs = append(out.Docs, toJSONMetricDoc(&entries[i]))
	}

	data, marshalErr := runtime.jsonMarshal(out)
	if marshalErr != nil {
		return nil, fmt.Errorf("application marshal docs: %w", marshalErr)
	}

	return data, nil
}

func toJSONMetricDoc(doc *reportdomain.MetricDoc) jsonMetricDoc {
	return jsonMetricDoc{
		Name:           doc.Name,
		Label:          doc.Label,
		FullName:       doc.FullName,
		Scope:          string(doc.Scope),
		Definition:     doc.Definition,
		FormulaMathML:  doc.FormulaMathML,
		FormulaLaTeX:   doc.FormulaLaTeX,
		Summary:        doc.Summary,
		How:            doc.HowCalculated,
		Interpretation: doc.Interpretation,
		NotApplicable:  doc.NotApplicable,
		Direction:      doc.Direction,
		Bounded:        doc.Bounded,
		Example:        doc.Example,
	}
}

// renderDocs writes the standalone metrics guide page: the embedded guide
// template with the docs payload injected.
func renderDocs(writer io.Writer, toolVersion string) error {
	err := renderDocsWith(defaultRuntime(), writer, toolVersion)
	if err != nil {
		return fmt.Errorf(errWrapRenderDocs, err)
	}

	return nil
}

func renderDocsWith(runtime reportingRuntime, writer io.Writer, toolVersion string) error {
	payload, err := marshalDocsWith(runtime, toolVersion)
	if err != nil {
		return fmt.Errorf(errWrapRenderDocs, err)
	}

	pageErr := writeDocsPage(writer, payload)
	if pageErr != nil {
		return fmt.Errorf(errWrapRenderDocs, pageErr)
	}

	return nil
}

func writeDocsPage(writer io.Writer, payload []byte) error {
	page, err := replacePlaceholder(&placeholderSwap{
		page:        docsTemplate,
		placeholder: docsDataPlaceholder,
		value:       string(payload),
		errMsg:      errMissingDocsPlaceholder,
	})
	if err != nil {
		return fmt.Errorf(errWrapRenderDocs, err)
	}

	writeErr := writeAll(writer, page)
	if writeErr != nil {
		return fmt.Errorf(errWrapRenderDocs, writeErr)
	}

	return nil
}

// WriteDocs writes the standalone metrics guide page to w.
func WriteDocs(w io.WriteCloser, toolVersion string) error {
	closeErr := closeAfter(renderDocs(w, toolVersion), w)
	if closeErr != nil {
		return fmt.Errorf(errWrapWriteDocs, closeErr)
	}

	return nil
}

// Write renders the report into w using opts.Format. Text options are
// read only by the text format.
func Write(report *distance.Report, w io.WriteCloser, opts *WriteOptions) error {
	closeErr := closeAfter(render(w, &renderOptions{
		report: report,
		format: opts.Format,
		text:   &opts.Text,
	}), w)
	if closeErr != nil {
		return fmt.Errorf(errWrapWrite, closeErr)
	}

	return nil
}

func render(writer io.Writer, opts *renderOptions) error {
	renderFn := renderNonText

	if opts.format == reportdomain.FormatText {
		renderFn = renderText
	}

	err := renderFn(writer, opts)
	if err != nil {
		return fmt.Errorf(errWrapRender, err)
	}

	return nil
}

func renderText(writer io.Writer, opts *renderOptions) error {
	writeErr := writeAll(writer, reportdomain.Text(opts.report, opts.text))
	if writeErr != nil {
		return fmt.Errorf("application renderText: %w", writeErr)
	}

	return nil
}

func writeAll(writer io.Writer, text string) error {
	written, writeErr := io.WriteString(writer, text)
	if writeErr != nil {
		return fmt.Errorf(errWrapWriteAll, writeErr)
	}

	if written != len(text) {
		return fmt.Errorf(errWrapWriteAll, errShortWrite)
	}

	return nil
}

func renderNonText(writer io.Writer, opts *renderOptions) error {
	renderer, ok := formatRenderers()[opts.format]

	if !ok {
		return unknownFormatError{format: opts.format}
	}

	renderErr := renderer(writer, opts)
	if renderErr != nil {
		return fmt.Errorf(errWrapRender, renderErr)
	}

	return nil
}

func renderCSV(w io.Writer, report *distance.Report) error {
	// WriteAll flushes; a separate header Write cannot surface bufio
	// errors until Flush, so header and records go through one call.
	rows := append([][]string{reportdomain.CSVHeader()}, reportdomain.CSVRecords(report)...)

	err := csv.NewWriter(w).WriteAll(rows)
	if err != nil {
		return fmt.Errorf("application csv write: %w", err)
	}

	return nil
}

// String summarizes the report envelope for debugging.
func (r jsonReport) String() string {
	return fmt.Sprintf("schema %s, tool %v, %d packages", r.SchemaVersion, r.Tool, len(r.Packages))
}

// String summarizes one package entry for debugging.
func (p jsonPackage) String() string {
	return fmt.Sprintf("%s: %d metrics", p.Path, len(p.Metrics))
}

// MarshalJSON writes the object with keys in the fixed metric order.
func (m *orderedMetrics) MarshalJSON() ([]byte, error) {
	data, err := encodeOrderedMetrics([]distance.MetricResult(*m))
	if err != nil {
		return nil, fmt.Errorf("application marshal metrics: %w", err)
	}

	return data, nil
}

// encodeOrderedMetrics assembles the ordered JSON object one name→metric
// pair at a time.
func encodeOrderedMetrics(results []distance.MetricResult) ([]byte, error) {
	data, err := encodeOrderedMetricsWith(defaultRuntime(), results)
	if err != nil {
		return nil, fmt.Errorf(errWrapEncodeOrdered, err)
	}

	return data, nil
}

func encodeOrderedMetricsWith(rtm reportingRuntime, rows []distance.MetricResult) ([]byte, error) {
	var buf bytes.Buffer

	err := writeMetricObject(rtm, &buf, rows)
	if err != nil {
		return nil, fmt.Errorf(errWrapEncodeOrdered, err)
	}

	return buf.Bytes(), nil
}

func writeMetricObject(rtm reportingRuntime, buf *bytes.Buffer, rows []distance.MetricResult) error {
	openErr := writeBufferByte(buf, '{')
	if openErr != nil {
		return fmt.Errorf(errWrapEncodeOrdered, openErr)
	}

	entriesErr := writeEntries(rtm, buf, rows)
	if entriesErr != nil {
		return fmt.Errorf(errWrapEncodeOrdered, entriesErr)
	}

	closeErr := writeBufferByte(buf, '}')
	if closeErr != nil {
		return fmt.Errorf(errWrapEncodeOrdered, closeErr)
	}

	return nil
}

func writeEntries(rtm reportingRuntime, buf *bytes.Buffer, rows []distance.MetricResult) error {
	for i := range rows {
		entryErr := writeOrderedEntry(rtm, buf, orderedEntry{
			index:  i,
			result: &rows[i],
		})
		if entryErr != nil {
			return fmt.Errorf(errWrapEncodeOrdered, entryErr)
		}
	}

	return nil
}

func writeOrderedEntry(runtime reportingRuntime, buf *bytes.Buffer, entry orderedEntry) error {
	commaErr := writeEntryComma(buf, entry.index)
	if commaErr != nil {
		return fmt.Errorf("application writeOrderedEntry: %w", commaErr)
	}

	encodeErr := encodeMetricEntry(runtime, buf, entry.result)
	if encodeErr != nil {
		return fmt.Errorf("application encode metric: %w", encodeErr)
	}

	return nil
}

func writeEntryComma(buf *bytes.Buffer, index int) error {
	if index <= indexZero {
		return nil
	}

	err := writeBufferByte(buf, ',')
	if err != nil {
		return fmt.Errorf("application writeEntryComma: %w", err)
	}

	return nil
}

// encodeMetricEntry writes one name→metric pair. A non-applicable metric
// carries its reason and no value — never a fake zero.
func encodeMetricEntry(rtm reportingRuntime, buf *bytes.Buffer, row *distance.MetricResult) error {
	keyErr := writeJSONKey(rtm, buf, row.Name)
	if keyErr != nil {
		return fmt.Errorf(errWrapEncodeMetric, keyErr)
	}

	encoded, marshalErr := rtm.jsonMarshal(metricJSON(row))
	if marshalErr != nil {
		return fmt.Errorf(errWrapEncodeMetric, marshalErr)
	}

	writeErr := writeBuffer(buf, encoded)
	if writeErr != nil {
		return fmt.Errorf(errWrapEncodeMetric, writeErr)
	}

	return nil
}

func writeJSONKey(runtime reportingRuntime, buf *bytes.Buffer, name string) error {
	key, marshalErr := runtime.jsonMarshal(name)
	if marshalErr != nil {
		return fmt.Errorf(errWrapWriteJSONKey, marshalErr)
	}

	writeErr := writeBuffer(buf, key)
	if writeErr != nil {
		return fmt.Errorf(errWrapWriteJSONKey, writeErr)
	}

	colonErr := writeBufferByte(buf, ':')
	if colonErr != nil {
		return fmt.Errorf(errWrapWriteJSONKey, colonErr)
	}

	return nil
}

func metricJSON(result *distance.MetricResult) jsonMetric {
	out := jsonMetric{
		Scope:      result.Scope,
		Applicable: result.Applicable,
		Reason:     result.Reason,
		Definition: result.Definition,
	}

	if result.Applicable {
		value := result.Value

		out.Value = &value
	}

	return out
}

// buildJSONReport maps the report onto the versioned JSON schema. It is
// shared by the JSON format and the web report's embedded payload.
func buildJSONReport(report *distance.Report) jsonReport {
	out := jsonReport{
		SchemaVersion: report.SchemaVersion,
		Tool:          jsonTool{Name: report.ToolName, Version: report.ToolVersion},
		Packages:      make([]jsonPackage, indexZero, len(report.Packages)),
	}

	for i := range report.Packages {
		out.Packages = append(out.Packages, toJSONPackage(&report.Packages[i]))
	}

	return out
}

func toJSONPackage(pkg *distance.PackageReport) jsonPackage {
	return jsonPackage{
		Path:     pkg.Path,
		Afferent: pkg.Afferent,
		Efferent: pkg.Efferent,
		Metrics:  orderedMetrics(pkg.Metrics),
	}
}

func renderJSON(w io.Writer, report *distance.Report) error {
	enc := json.NewEncoder(w)
	enc.SetIndent(emptyString, jsonIndent)

	err := enc.Encode(buildJSONReport(report))
	if err != nil {
		return fmt.Errorf("application encode json: %w", err)
	}

	return nil
}

// renderWeb writes the interactive HTML report: the embedded template with
// the JSON payloads injected. [json.Marshal] HTML-escapes <, >, and &, so the
// payloads can never terminate their <script> elements early. The trusted
// compile-time docs payload is injected before the report payload, whose
// untrusted identifiers could otherwise spoof the docs placeholder.
func renderWeb(writer io.Writer, report *distance.Report) error {
	err := renderWebWith(defaultRuntime(), writer, report)
	if err != nil {
		return fmt.Errorf(errWrapRenderWeb, err)
	}

	return nil
}

func renderWebWith(runtime reportingRuntime, writer io.Writer, report *distance.Report) error {
	page, err := buildWebPageWith(runtime, report)
	if err != nil {
		return fmt.Errorf(errWrapRenderWeb, err)
	}

	writeErr := writeAll(writer, page)
	if writeErr != nil {
		return fmt.Errorf(errWrapRenderWeb, writeErr)
	}

	return nil
}

func buildWebPageWith(runtime reportingRuntime, report *distance.Report) (string, error) {
	payload, err := marshalWebPayloadWith(runtime, report)
	if err != nil {
		return emptyString, fmt.Errorf(errWrapBuildWeb, err)
	}

	docs, err := marshalDocsWith(runtime, report.ToolVersion)
	if err != nil {
		return emptyString, fmt.Errorf(errWrapBuildWeb, err)
	}

	page, err := injectWebPayloads(string(docs), string(payload))
	if err != nil {
		return emptyString, fmt.Errorf(errWrapBuildWeb, err)
	}

	return page, nil
}

func marshalWebPayloadWith(runtime reportingRuntime, report *distance.Report) ([]byte, error) {
	data, err := runtime.jsonMarshal(webPayload{
		Module: report.Module,
		Report: buildJSONReport(report),
	})
	if err != nil {
		return nil, fmt.Errorf("application marshal web payload: %w", err)
	}

	return data, nil
}

func injectWebPayloads(docs, payload string) (string, error) {
	page, err := injectDocs(docs)
	if err != nil {
		return emptyString, fmt.Errorf("application injectWebPayloads: %w", err)
	}

	out, swapErr := swapPlaceholder(&placeholderSwap{
		page:        page,
		placeholder: webDataPlaceholder,
		value:       payload,
		errMsg:      errMissingWebReport,
	})
	if swapErr != nil {
		return emptyString, fmt.Errorf("application inject report payload: %w", swapErr)
	}

	return out, nil
}

func injectDocs(docs string) (string, error) {
	page, err := swapPlaceholder(&placeholderSwap{
		page:        webTemplate,
		placeholder: docsDataPlaceholder,
		value:       docs,
		errMsg:      errMissingWebDocs,
	})
	if err != nil {
		return emptyString, fmt.Errorf("application injectDocs: %w", err)
	}

	return page, nil
}

func swapPlaceholder(swap *placeholderSwap) (string, error) {
	out, err := replacePlaceholder(swap)
	if err != nil {
		return emptyString, fmt.Errorf("application swapPlaceholder: %w", err)
	}

	return out, nil
}

func replacePlaceholder(s *placeholderSwap) (string, error) {
	if !strings.Contains(s.page, s.placeholder) {
		return emptyString, missingPlaceholderError{message: s.errMsg}
	}

	return strings.Replace(s.page, s.placeholder, s.value, replaceOnce), nil
}
