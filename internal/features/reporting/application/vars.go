// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	_ "embed"
	"errors"
)

var (

	// docsTemplate is the self-contained metrics guide page: inline CSS and
	// vanilla JS, native MathML formulas, no external requests.
	//
	//go:embed web_docs_template.html
	docsTemplate string

	// webTemplate is the self-contained HTML page: inline CSS and vanilla JS,
	// no external requests, usable straight from file://.
	//
	//go:embed web_template.html
	webTemplate string

	errShortWrite = errors.New("short write")
)
