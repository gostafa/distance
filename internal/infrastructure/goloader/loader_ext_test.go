// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package goloader_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gostafa/distance/internal/features/typefacts/domain/model"
	"github.com/gostafa/distance/internal/features/typefacts/ports/outbound"
	"github.com/gostafa/distance/internal/infrastructure/goloader"
)

func fixtureDir() string { return filepath.Join("..", "..", "..", "testdata", "fixture") }

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()

	path := filepath.Join(dir, name)

	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		t.Fatal(err)
	}
}

// Black-box: the loader extracts real facts from the fixture module through
// the outbound port.
func TestLoadFixture(t *testing.T) {
	t.Parallel()

	mod, pkgs, err := goloader.New().Load(t.Context(), &outbound.FactOptions{
		Directory: fixtureDir(),
		Patterns:  []string{"./..."},
	})
	if err != nil {
		t.Fatal(err)
	}

	if mod != "example.com/fixture" {
		t.Fatalf("module = %q", mod)
	}

	var orders *model.PackageExtract

	for i := range pkgs {
		if pkgs[i].Path == "example.com/fixture/orders" {
			orders = &pkgs[i]
		}
	}

	if orders == nil {
		t.Fatal("orders package not extracted")
	}

	if !orders.InModule {
		t.Error("orders should be in-module")
	}

	var order *model.TypeName

	for i := range orders.Types {
		if orders.Types[i].Name == "Order" {
			order = &orders.Types[i]
		}
	}

	if order == nil {
		t.Fatal("Order type not extracted")
	}

	if order.Kind != model.KindStruct {
		t.Errorf("Order kind = %d, want struct", order.Kind)
	}
}

func TestLoadTypeKinds(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "module example.com/kinds\n\ngo 1.23\n")
	writeFile(t, dir, "kinds.go", `package kinds

type Widget struct { Value int }
type Closer interface { Close() }
type Alias = Widget
type Count int
`)

	_, pkgs, err := goloader.New().Load(t.Context(), &outbound.FactOptions{
		Directory: dir,
		Patterns:  []string{"."},
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(pkgs) != 1 {
		t.Fatalf("packages = %d, want 1", len(pkgs))
	}

	kinds := map[string]uint8{}

	for _, te := range pkgs[0].Types {
		kinds[te.Name] = te.Kind
	}

	if _, ok := kinds["Alias"]; ok {
		t.Fatal("type aliases should not be extracted")
	}

	if kinds["Widget"] != model.KindStruct || kinds["Closer"] != model.KindInterface ||
		kinds["Count"] != model.KindOther {

		t.Fatalf("kinds = %+v", kinds)
	}
}

// Black-box: a pattern matching nothing is an error.
func TestLoadNoMatch(t *testing.T) {
	t.Parallel()

	_, _, err := goloader.New().Load(t.Context(), &outbound.FactOptions{
		Directory: fixtureDir(),
		Patterns:  []string{"./does-not-exist"},
	})
	if err == nil {
		t.Fatal("expected error for a pattern matching no packages")
	}
}
