// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

type (

	// CouplingGraph looks up afferent and efferent counts by package ID.
	CouplingGraph interface {
		PackageCoupling(packageID int) (afferent, efferent int)
	}

	// DependencyGraph stores coupling counts indexed by package ID.
	DependencyGraph struct {
		scope     string
		Afferents []int
		Efferents []int
	}
)
