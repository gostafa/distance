// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package store

type (
	Store interface {
		Put(id int, v any) error
	}
)
