// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package model

import (
	"strings"
	"testing"
)

func TestFormatNamed(t *testing.T) {
	got := FormatNamed("E", KindOther)

	if !strings.Contains(got, `"E"`) {
		t.Fatalf("FormatNamed = %q", got)
	}
}

func TestNamedAndViews(t *testing.T) {
	extract := PackageExtract{
		Path:     "example.com/m",
		Imports:  []string{"fmt"},
		Types:    []TypeName{Named("T", KindStruct)},
		InModule: true,
	}

	if PathOf(&extract) != "example.com/m" {
		t.Fatalf("PathOf = %q", PathOf(&extract))
	}

	if !inModuleOf(&extract) {
		t.Fatal("expected in module")
	}

	if len(ImportsOf(&extract)) != 1 || len(TypesOf(&extract)) != 1 {
		t.Fatalf("imports/types = %d/%d", len(ImportsOf(&extract)), len(TypesOf(&extract)))
	}
}
