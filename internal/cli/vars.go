// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package cli

import (
	"errors"
)

var (
	errEmptyPattern       = errors.New("empty pattern in rule spec")
	errExpectedPatternMax = errors.New("expected pattern:max")
	errNoPolicyRules      = errors.New(
		"no policy rules configured; pass at least one -rule=pattern:max with -check",
	)
	errShortWrite = errors.New("short write")
)
