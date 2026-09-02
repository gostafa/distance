// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	"errors"

	"github.com/gostafa/distance/internal/shared/version"
)

var (
	errShortWrite = errors.New("short write")
	_             = version.Version
)
