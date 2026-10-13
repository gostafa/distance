// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package multifile

func (c *Counter) Hit() { c.hits++ }

func (c Counter) Total() int { return c.total() }

func (c Counter) total() int { return c.hits + c.misses }
