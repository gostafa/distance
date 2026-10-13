// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package destinations

import (
	"io"
	"os"
)

var (
	_ directory = (*os.Root)(nil)

	_ io.WriteCloser = (*os.File)(nil)
)
