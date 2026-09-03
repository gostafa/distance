// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package wire

import (
	typefacts "github.com/gostafa/distance/internal/features/typefacts/application"
)

var _ typefacts.Collector = (*typefacts.Service)(nil)
