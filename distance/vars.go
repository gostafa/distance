// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package distance

import (
	"errors"
)

var (
	// ErrInvalidConfig identifies invalid analysis configuration.
	ErrInvalidConfig = errors.New("invalid distance configuration")

	// ErrAnalysis identifies analysis failures. Internal error types are not exposed.
	ErrAnalysis = errors.New("distance analysis failed")

	errEmptyPattern = errors.New("empty package pattern")

	errInvalidScope = errors.New("invalid dependency scope (want project, module, or all)")
)
