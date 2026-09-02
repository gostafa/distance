// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

const (
	// ComparatorMax marks a value that exceeded an upper bound.
	ComparatorMax = "max"
	// ComparatorMin marks a value that fell below a lower bound.
	ComparatorMin = "min"

	// DefaultMaxDistance is the default exclusive upper bound for distance.
	DefaultMaxDistance = 0.5

	allPackagesPattern  = "./..."
	currentPackage      = "."
	comparisonEpsilon   = 1e-12
	floatBitSize        = 64
	floatPrecFixed      = 2
	floatPrecAuto       = -1
	emptyString         = ""
	zero                = 0
	one                 = 1
	errWrapWriteBuilder = "policy writeBuilder: %w"
)
