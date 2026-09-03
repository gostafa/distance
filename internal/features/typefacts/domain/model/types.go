// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package model

type (
	// TypeName is one named type extract.
	TypeName = struct {
		Name string
		Kind uint8
	}

	// PackageExtract is one package's raw facts before project assembly.
	PackageExtract = struct {
		Path     string
		Imports  []string
		Types    []TypeName
		InModule bool
	}
)
