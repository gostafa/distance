// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package public

import (
	"github.com/gostafa/distance/internal/features/reporting/domain"
)

// ParseFormat recognizes the supported report representations.
func ParseFormat(name string) (string, bool) { return domain.ParseFormat(name) }
