// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package goanalysis

import (
	"errors"
)

var (
	errRuleMaxRequired = errors.New("max is required")

	errDependencyScope = errors.New("invalid dependency-scope")
)
