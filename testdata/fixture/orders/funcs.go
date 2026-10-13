// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package orders

import (
	"fmt"

	"example.com/fixture/store"
)

func (o *Order) GrandTotal(tax float64) float64 {
	if tax > zero && o.Total > zero {
		return o.Total * (1 + tax)
	}
	return o.Total
}

func (o Order) Note() string { return fmt.Sprintf("note: %s", o.notes) }

func (o *Order) Save(s store.Store) error { return s.Put(o.ID, o) }
