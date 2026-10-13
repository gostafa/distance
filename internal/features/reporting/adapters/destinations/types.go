// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package destinations

import (
	"os"
)

type (
	// directory scopes file creation and owns its handle lifecycle.
	directory = interface {
		Create(name string) (*os.File, error)
		Close() error
	}
)
