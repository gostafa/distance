// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package outbound

import (
	"errors"
)

// ErrOutput identifies report encoding, writing, or closing failures.
var ErrOutput = errors.New("report output failed")
