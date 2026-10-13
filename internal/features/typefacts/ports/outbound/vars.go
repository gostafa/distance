// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package outbound

import (
	"errors"
)

// ErrLoad identifies package loading failures without exposing vendor error types.
var ErrLoad = errors.New("package fact loading failed")
