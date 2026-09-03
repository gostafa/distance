// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	"github.com/gostafa/distance/distance"
)

type (
	// PathMatcher matches import paths against a policy pattern.
	PathMatcher interface {
		Matches(importPath string) bool
	}

	// RuleGate exposes the exclusive maximum distance for a matching rule.
	RuleGate interface {
		MaxDistance() float64
	}

	// Violation records one package whose distance exceeded a matching rule.
	Violation struct {
		Package   string
		Rule      string
		Value     float64
		Threshold float64
	}

	// Rule is a package-path glob paired with an exclusive maximum distance.
	// When more than one rule matches, the most specific pattern wins; exact
	// ties use the later rule.
	Rule struct {
		// Pattern is a glob against the full import path; * matches one segment,
		// ** matches zero or more segments, using / as the separator.
		Pattern string
		// Max is the exclusive upper bound in [0, 1]: fail when distance > Max.
		Max float64
	}

	matchPos = struct {
		pi, si int
	}

	packageGate = struct {
		pkg       string
		pattern   string
		threshold float64
		value     float64
		ok        bool
	}

	gateInput = struct {
		pkg       *distance.PackageReport
		res       *distance.MetricResult
		pattern   string
		threshold float64
	}
)
