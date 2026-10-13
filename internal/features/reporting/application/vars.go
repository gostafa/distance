// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	reporting "github.com/gostafa/distance/internal/features/reporting/public"
)

var _ reporting.Output = (*Service)(nil)
