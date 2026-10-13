// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package orders

type (
	Order struct {
		// ID identifies the order.
		ID    int
		Total float64
		notes string
	}
)
