// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

const (

	// docsDataPlaceholder marks where the guide JSON lands. The report template
	// carries the same placeholder so its info sheets share this payload.
	docsDataPlaceholder = "__DOCS_DATA__"

	// webDataPlaceholder marks where the JSON payload lands in the template.
	webDataPlaceholder = "__REPORT_DATA__"

	emptyString               = ""
	jsonIndent                = "  "
	replaceOnce               = 1
	indexZero                 = 0
	errMissingDocsPlaceholder = "docs template is missing the docs data placeholder"
	errMissingWebDocs         = "web template is missing the docs data placeholder"
	errMissingWebReport       = "web template is missing the report data placeholder"
	errWrapRenderDocs         = "application renderDocs: %w"
	errWrapRender             = "application render: %w"
	errWrapEncodeMetric       = "application encodeMetricEntry: %w"
	errWrapRenderWeb          = "application renderWeb: %w"
	errWrapBuildWeb           = "application buildWebPage: %w"
	errWrapWriteBuffer        = "application writeBuffer: %w"
	errWrapClose              = "application close: %w"
	errWrapWriteDocs          = "application WriteDocs: %w"
	errWrapWrite              = "application Write: %w"
	errWrapEncodeOrdered      = "application encodeOrderedMetrics: %w"
	errWrapWriteJSONKey       = "application writeJSONKey: %w"
	errWrapWriteAll           = "application writeAll: %w"
)
