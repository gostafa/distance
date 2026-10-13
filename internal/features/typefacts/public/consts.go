// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package public

const (
	// KindStruct marks a named type whose underlying type is a struct.
	KindStruct uint8 = 0

	// KindInterface marks a named type whose underlying type is an interface.
	KindInterface uint8 = 1

	// KindOther marks any other named type (basic, slice, func, …).
	KindOther uint8 = 2
)
