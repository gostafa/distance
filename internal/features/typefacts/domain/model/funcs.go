// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package model

import (
	"fmt"

	"github.com/gostafa/distance/internal/shared/version"
)

// FormatNamed returns a compact debug representation of a named type extract.
func FormatNamed(name string, kind uint8) string {
	return fmt.Sprintf("type %q (kind %d, abstractness, tool %s)", name, kind, version.Version())
}

// Named returns a type extract with the given name and kind.
func Named(name string, kind uint8) TypeName {
	return TypeName{Name: name, Kind: kind}
}

// PackageExtractString returns a compact debug representation of a package extract.
func PackageExtractString(pkg *PackageExtract) string {
	return fmt.Sprintf(
		"package %q: %d imports, %d types, inModule=%t",
		pkg.Path,
		len(pkg.Imports),
		len(pkg.Types),
		pkg.InModule,
	)
}

// DumpExtract returns a compact debug representation of pkg.
func DumpExtract(pkg *PackageExtract) string {
	return PackageExtractString(pkg)
}
