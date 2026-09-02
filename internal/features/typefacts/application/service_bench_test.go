// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	"fmt"
	"testing"

	"github.com/gostafa/distance/internal/features/typefacts/domain/model"
)

func benchExtracts(pkgCount, typesPerPkg int) []model.PackageExtract {
	pkgs := make([]model.PackageExtract, pkgCount)

	for p := range pkgs {
		types := make([]model.TypeName, 0, typesPerPkg)

		for i := range typesPerPkg {
			types = append(types, model.Named(fmt.Sprintf("Type%02d", i), 0))
		}

		pkgs[p] = model.PackageExtract{
			Path:     fmt.Sprintf("example.com/m/pkg%d", p),
			InModule: true,
			Imports: []string{
				fmt.Sprintf("example.com/m/pkg%d", (p+1)%pkgCount),
				"fmt",
				"context",
			},
			Types: types,
		}
	}

	return pkgs
}

func BenchmarkAssemble(b *testing.B) {
	extracts := benchExtracts(60, 25)

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		// Assemble mutates (sorts) its input in place; copy per iteration so
		// each run sees identical work.
		cp := make([]model.PackageExtract, len(extracts))
		copy(cp, extracts)

		_ = Assemble("example.com/m", cp)
	}
}
