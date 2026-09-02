// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package version_test

import (
	"testing"

	"github.com/gostafa/distance/internal/shared/version"
)

// Black-box: consumers read a non-empty version string.
func TestVersionExported(t *testing.T) {
	t.Parallel()

	if version.Version() == "" {
		t.Fatal("version.Version must not be empty")
	}
}
