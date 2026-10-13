// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application_test

import (
	"context"
	"testing"

	typefacts "github.com/gostafa/distance/internal/features/typefacts/application"
	tfmodel "github.com/gostafa/distance/internal/features/typefacts/domain/model"
	tfoutbound "github.com/gostafa/distance/internal/features/typefacts/ports/outbound"
)

type fakeSource struct{}

func (fakeSource) Load(
	context.Context,
	*tfoutbound.FactOptions,
) (string, []tfmodel.PackageExtract, error) {
	return "example.com/m", []tfmodel.PackageExtract{
		{Path: "example.com/m/b", InModule: true, Types: []tfmodel.TypeName{{Name: "B"}}},
		{Path: "example.com/m/a", InModule: true, Types: []tfmodel.TypeName{{Name: "A"}}},
	}, nil
}

// Black-box: the service loads through the port and assembles deterministic,
// sorted facts with dense IDs.
func TestServiceCollect(t *testing.T) {
	t.Parallel()

	svc := typefacts.NewService(fakeSource{})

	facts, err := svc.Collect(
		t.Context(),
		&tfoutbound.FactOptions{Patterns: []string{"./..."}},
	)
	if err != nil {
		t.Fatal(err)
	}

	if facts.ModulePath != "example.com/m" {
		t.Fatalf("module = %q", facts.ModulePath)
	}

	if len(facts.Packages) != 2 || facts.Packages[0].Path != "example.com/m/a" {
		t.Fatalf("packages not sorted by path: %+v", facts.Packages)
	}

	for i, p := range facts.Packages {
		if p.ID != i {
			t.Errorf("package %q ID = %d, want %d", p.Path, p.ID, i)
		}
	}
}
