// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	"fmt"
)

func moduleOf(view moduleView) string {
	return view.Module()
}

func packageCountOf(view packageCounter) int {
	return view.PackageCount()
}

func typeCountOf(view typeCounter) int {
	return view.TypeCount()
}

// DumpFacts returns a compact debug representation of view.
func DumpFacts(view FactDumper) string {
	return view.String()
}

// ImportPath returns the package import path.
func (pkg *PackageFacts) ImportPath() string {
	return pkg.Path
}

// Module returns the analyzed module path.
func (facts *ProjectFacts) Module() string {
	return facts.ModulePath
}

// PackageCount returns how many packages were assembled.
func (facts *ProjectFacts) PackageCount() int {
	return len(facts.Packages)
}

// PathOf returns view's import path.
func PathOf(view PackagePathView) string {
	return view.ImportPath()
}

// String returns a compact debug representation of the project facts.
func (facts *ProjectFacts) String() string {
	return fmt.Sprintf(
		"module %q: %d packages, %d types",
		moduleOf(facts),
		packageCountOf(facts),
		typeCountOf(facts),
	)
}

// String returns a compact debug representation of the type facts.
func (facts *TypeFacts) String() string {
	return fmt.Sprintf(
		"type %d %q (package %d, kind %d)",
		facts.ID,
		facts.Name,
		facts.PackageID,
		facts.Kind,
	)
}

// TypeCount returns how many named types were assembled.
func (facts *ProjectFacts) TypeCount() int {
	return len(facts.Types)
}
