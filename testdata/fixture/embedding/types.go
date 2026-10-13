// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package embedding

type (
	Base struct {
		Count int
	}

	Wrapper struct {
		Base
		Name string
	}
)
