// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

type (

	// PatternHolder exposes a package rule's match pattern.
	PatternHolder interface {
		RulePattern() string
	}

	// DistanceLimit exposes a package rule's exclusive distance bound.
	DistanceLimit interface {
		Limit() float64
	}

	// ruleError is a validation failure for one packages[] entry.
	ruleError struct {
		reason string
		index  int
	}

	// Violation records one policy failure for a package metric.
	Violation struct {
		Package    string
		Key        string
		Comparator string
		Value      float64
		Threshold  float64
	}

	// PackageRule is one packages[] entry matching a pattern to a max distance.
	PackageRule struct {
		// Pattern is a go list package pattern (./..., ./internal/..., an
		// exact import path, or mod/pkg/...).
		Pattern string `json:"pattern"`
		// MaxDistance is the exclusive upper bound: fail when distance >
		// MaxDistance.
		MaxDistance float64 `json:"max_distance"`
	}

	// Policy is the set of package rules evaluated against a report.
	Policy struct {
		Packages []PackageRule
	}
)
