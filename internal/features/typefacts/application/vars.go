// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	typefacts "github.com/gostafa/distance/internal/features/typefacts/public"
)

var _ typefacts.Collector = (*Service)(nil)
