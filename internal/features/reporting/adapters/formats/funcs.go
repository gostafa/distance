// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package formats

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	reportdomain "github.com/gostafa/distance/internal/features/reporting/adapters/presentation"
	"github.com/gostafa/distance/internal/features/reporting/ports/outbound"
	reportmodel "github.com/gostafa/distance/internal/features/reporting/public"
)

// Write renders the report into w using opts.Format. Text options are
// read only by the text format.
func Write(report *reportmodel.Report, w io.WriteCloser, opts *WriteOptions) error {
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

// WriteDocs writes the standalone metrics guide page to w.
func WriteDocs(w io.WriteCloser, toolVersion string) error {
	closeErr := closeAfter(renderDocs(w, toolVersion), w)
	if closeErr != nil {
		return fmt.Errorf(errWrapWriteDocs, closeErr)
	}

	return nil
}

func (err missingPlaceholderError) Error() string {
	return err.message
}

func (err unknownFormatError) Error() string {
	return "unknown report format " + err.format
}

// MarshalJSON writes the object with keys in the fixed metric order.
func (m orderedMetrics) MarshalJSON() ([]byte, error) {
	data, err := encodeOrderedMetrics([]reportmodel.MetricResult(m))
	if err != nil {
		return nil, fmt.Errorf("application marshal metrics: %w", err)
	}

	return data, nil
}

// Write renders a report and closes its destination.
func (Output) Write(
	report *reportmodel.Report,
	w io.WriteCloser,
	opts *WriteOptions,
) error {
	err := Write(report, w, opts)
	if err != nil {
		return fmt.Errorf(errOutputDetails, outbound.ErrOutput, err)
	}

	return nil
}

// WriteDocs renders the guide and closes its destination.
func (Output) WriteDocs(w io.WriteCloser, version string) error {
	err := WriteDocs(w, version)
	if err != nil {
		return fmt.Errorf(errOutputDetails, outbound.ErrOutput, err)
	}

	return nil
}

// buildJSONReport maps the report onto the versioned JSON schema. It is
// shared by the JSON format and the web report's embedded payload.
func buildJSONReport(report *reportmodel.Report) jsonReport {
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

func buildWebPageWith(runtime reportingRuntime, report *reportmodel.Report) (string, error) {
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

func defaultRuntime() reportingRuntime {
	return reportingRuntime{jsonMarshal: json.Marshal}
}

// encodeMetricEntry writes one name→metric pair. A non-applicable metric
// carries its reason and no value — never a fake zero.
func encodeMetricEntry(rtm reportingRuntime, buf *[]byte, row *reportmodel.MetricResult) error {
	keyErr := writeJSONKey(rtm, buf, row.Name)
	if keyErr != nil {
		return fmt.Errorf(errWrapEncodeMetric, keyErr)
	}

	encoded, marshalErr := rtm.jsonMarshal(metricJSON(row))
	if marshalErr != nil {
		return fmt.Errorf(errWrapEncodeMetric, marshalErr)
	}

	*buf = append(*buf, encoded...)

	return nil
}

func encodeOrderedMetrics(results []reportmodel.MetricResult) ([]byte, error) {
	data, err := encodeOrderedMetricsWith(defaultRuntime(), results)
	if err != nil {
		return nil, fmt.Errorf(errWrapEncodeOrdered, err)
	}

	return data, nil
}

func encodeOrderedMetricsWith(
	rtm reportingRuntime,
	rows []reportmodel.MetricResult,
) ([]byte, error) {
	buf := []byte{}

	err := writeMetricObject(&metricObjectIn{runtime: rtm, buffer: &buf, rows: rows})
	if err != nil {
		return nil, fmt.Errorf(errWrapEncodeOrdered, err)
	}

	return buf, nil
}

func formatRenderers() map[reportdomain.Format]func(io.Writer, *renderOptions) error {
	return map[reportdomain.Format]func(io.Writer, *renderOptions) error{
		reportdomain.FormatText: renderText,
		reportdomain.FormatJSON: renderJSONOpts,
		reportdomain.FormatCSV:  renderCSVOpts,
		reportdomain.FormatWeb:  renderWebOpts,
	}
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

// jsonReportString summarizes the report envelope for debugging.
func jsonReportString(r *jsonReport) string {
	return fmt.Sprintf("schema %s, tool %v, %d packages", r.SchemaVersion, r.Tool, len(r.Packages))
}

func marshalDocsWith(runtime reportingRuntime, toolVersion string) ([]byte, error) {
	entries := reportdomain.MetricDocs()

	out := docsPayload{
		Tool: jsonTool{Name: reportmodel.MetricDistance, Version: toolVersion},
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

func marshalWebPayloadWith(runtime reportingRuntime, report *reportmodel.Report) ([]byte, error) {
	data, err := runtime.jsonMarshal(webPayload{
		Module: report.Module,
		Report: buildJSONReport(report),
	})
	if err != nil {
		return nil, fmt.Errorf("application marshal web payload: %w", err)
	}

	return data, nil
}

func metricJSON(result *reportmodel.MetricResult) jsonMetric {
	out := jsonMetric{
		Scope:      result.Scope,
		Applicable: result.Applicable,
		Reason:     result.Reason,
		Definition: result.Definition, Value: nil,
	}

	if result.Applicable {
		value := result.Value

		out.Value = &value
	}

	return out
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

func renderCSV(w io.Writer, report *reportmodel.Report) error {
	// WriteAll flushes; a separate header Write cannot surface bufio
	// errors until Flush, so header and records go through one call.
	rows := append([][]string{reportdomain.CSVHeader()}, reportdomain.CSVRecords(report)...)

	err := csv.NewWriter(w).WriteAll(rows)
	if err != nil {
		return fmt.Errorf("application csv write: %w", err)
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

func renderJSON(w io.Writer, report *reportmodel.Report) error {
	enc := json.NewEncoder(w)
	enc.SetIndent(emptyString, jsonIndent)

	err := enc.Encode(buildJSONReport(report))
	if err != nil {
		return fmt.Errorf("application encode json: %w", err)
	}

	return nil
}

func renderJSONOpts(writer io.Writer, opts *renderOptions) error {
	err := renderJSON(writer, opts.report)
	if err != nil {
		return fmt.Errorf("application renderJSONOpts: %w", err)
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

func renderText(writer io.Writer, opts *renderOptions) error {
	writeErr := writeAll(writer, reportdomain.Text(opts.report, opts.text))
	if writeErr != nil {
		return fmt.Errorf("application renderText: %w", writeErr)
	}

	return nil
}

// renderWeb writes the interactive HTML report: the embedded template with
// the JSON payloads injected. [json.Marshal] HTML-escapes <, >, and &, so the
// payloads can never terminate their <script> elements early. The trusted
// compile-time docs payload is injected before the report payload, whose
// untrusted identifiers could otherwise spoof the docs placeholder.
func renderWeb(writer io.Writer, report *reportmodel.Report) error {
	err := renderWebWith(defaultRuntime(), writer, report)
	if err != nil {
		return fmt.Errorf(errWrapRenderWeb, err)
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

func renderWebWith(runtime reportingRuntime, writer io.Writer, report *reportmodel.Report) error {
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

func replacePlaceholder(s *placeholderSwap) (string, error) {
	if !strings.Contains(s.page, s.placeholder) {
		return emptyString, missingPlaceholderError{message: s.errMsg}
	}

	return strings.Replace(s.page, s.placeholder, s.value, replaceOnce), nil
}

func swapPlaceholder(swap *placeholderSwap) (string, error) {
	out, err := replacePlaceholder(swap)
	if err != nil {
		return emptyString, fmt.Errorf("application swapPlaceholder: %w", err)
	}

	return out, nil
}

func toJSONMetricDoc(doc *reportdomain.MetricDoc) jsonMetricDoc {
	return jsonMetricDoc{
		Name:           doc.Name,
		Label:          doc.Label,
		FullName:       doc.FullName,
		Scope:          doc.Scope,
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

func toJSONPackage(pkg *reportmodel.PackageReport) jsonPackage {
	return jsonPackage{
		Path:     pkg.Path,
		Afferent: pkg.Afferent,
		Efferent: pkg.Efferent,
		Metrics:  orderedMetrics(pkg.Metrics),
	}
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

func writeEntries(rtm reportingRuntime, buf *[]byte, rows []reportmodel.MetricResult) error {
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

func writeEntryComma(buf *[]byte, index int) {
	if index > indexZero {
		*buf = append(*buf, ',')
	}
}

func writeJSONKey(runtime reportingRuntime, buf *[]byte, name string) error {
	key, marshalErr := runtime.jsonMarshal(name)
	if marshalErr != nil {
		return fmt.Errorf(errWrapWriteJSONKey, marshalErr)
	}

	*buf = append(*buf, key...)
	*buf = append(*buf, ':')

	return nil
}

func writeMetricObject(input *metricObjectIn) error {
	*input.buffer = append(*input.buffer, '{')

	entriesErr := writeEntries(input.runtime, input.buffer, input.rows)
	if entriesErr != nil {
		return fmt.Errorf(errWrapEncodeOrdered, entriesErr)
	}

	*input.buffer = append(*input.buffer, '}')

	return nil
}

func writeOrderedEntry(runtime reportingRuntime, buf *[]byte, entry orderedEntry) error {
	writeEntryComma(buf, entry.index)

	encodeErr := encodeMetricEntry(runtime, buf, entry.result)
	if encodeErr != nil {
		return fmt.Errorf("encode metric entry: %w", encodeErr)
	}

	return nil
}
