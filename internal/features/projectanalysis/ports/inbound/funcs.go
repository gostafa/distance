// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package inbound

import (
	"fmt"
)

// String returns a compact debug representation of the package result.
func (p PackageResult) String() string {
	return fmt.Sprintf("%s: %d package metrics", p.Path, len(p.Metrics))
}
