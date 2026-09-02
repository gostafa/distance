// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package outbound

import (
	"context"
	"testing"

	"github.com/gostafa/distance/internal/features/typefacts/domain/model"
)

type stubSource struct{ mod string }

func (s stubSource) Load(context.Context, *FactOptions) (string, []model.PackageExtract, error) {
	return s.mod, []model.PackageExtract{{Path: "p"}}, nil
}

var _ FactSource = stubSource{}

// White-box: the port is satisfiable and FactOptions carries the load config.
func TestFactSourceContract(t *testing.T) {
	t.Parallel()

	var src FactSource = stubSource{mod: "example.com/m"}

	mod, pkgs, err := src.Load(
		t.Context(), &FactOptions{Patterns: []string{"./..."}, Workers: 2},
	)
	if err != nil {
		t.Fatal(err)
	}

	if mod != "example.com/m" || len(pkgs) != 1 {
		t.Fatalf("mod=%q pkgs=%d", mod, len(pkgs))
	}
}
