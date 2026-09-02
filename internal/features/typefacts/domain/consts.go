// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	factmodel "github.com/gostafa/distance/internal/features/typefacts/domain/model"
)

const (
	// KindStruct marks a named type whose underlying type is a struct.
	KindStruct = factmodel.KindStruct
	// KindInterface marks a named type whose underlying type is an interface.
	KindInterface = factmodel.KindInterface
	// KindOther marks any other named type (basic, slice, func, …).
	KindOther = factmodel.KindOther
)
