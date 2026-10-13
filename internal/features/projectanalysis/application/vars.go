// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	projectanalysis "github.com/gostafa/distance/internal/features/projectanalysis/public"
)

var _ projectanalysis.Pipeline = (*Pipeline)(nil)
