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

func TestNamedAndExtractFields(t *testing.T) {
	extract := PackageExtract{
		Path:     "example.com/m",
		Imports:  []string{"fmt"},
		Types:    []TypeName{Named("T", KindStruct)},
		InModule: true,
	}

	if extract.Path != "example.com/m" {
		t.Fatalf("Path = %q", extract.Path)
	}

	if !extract.InModule {
		t.Fatal("expected in module")
	}

	if len(extract.Imports) != 1 || len(extract.Types) != 1 {
		t.Fatalf("imports/types = %d/%d", len(extract.Imports), len(extract.Types))
	}
}
