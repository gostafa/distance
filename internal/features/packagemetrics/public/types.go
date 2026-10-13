// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package public

type (
	// ProjectFacts is the assembled type graph for one analysis run.
	ProjectFacts = struct {
		ModulePath string
		Packages   []PackageFacts
		Types      []TypeFacts
	}

	// PackageFacts holds one package's assembled facts and type IDs.
	PackageFacts = struct {
		Path     string
		Imports  []string
		TypeIDs  []int
		ID       int
		InModule bool
	}

	// TypeFacts holds one named type's assembled facts.
	TypeFacts = struct {
		Name      string
		ID        int
		PackageID int
		Kind      uint8
	}

	// MetricEntry is one computed metric with applicability and formula metadata.
	MetricEntry = struct {
		Name       string
		Scope      string
		Reason     string
		Definition string
		Value      float64
		Applicable bool
	}

	// PackageMetrics contains ordered metrics and coupling counts for one package.
	PackageMetrics = struct {
		Metrics  []MetricEntry
		Afferent int
		Efferent int
	}
)
