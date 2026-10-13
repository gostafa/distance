// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	"fmt"
	"math"
	"strings"
)

// DefaultRules returns the recommended baseline when no rules are configured
// (plugin users). A single catch-all rule gates distance at 0.5.
func DefaultRules() []Rule {
	return []Rule{{Pattern: doubleStar, Max: DefaultMaxDistance}}
}

// Evaluate checks a report against rules and returns violations. When no rules
// are given, DefaultRules applies. Packages are already sorted in the report,
// so the result order is deterministic.
func Evaluate(report *Report, rules []Rule) []Violation {
	if len(rules) == zero {
		rules = DefaultRules()
	}

	violations := make([]Violation, zero, len(report.Packages))

	for index := range report.Packages {
		violations = append(violations, evaluatePackage(&report.Packages[index], rules)...)
	}

	return violations
}

// MatchPackage reports whether pattern matches importPath. * matches exactly
// one path segment; ** matches zero or more segments. Matching is against the
// full import path using / as the separator. When multiple rules match, the
// evaluator selects the most specific pattern rather than the first match.
func MatchPackage(pattern, importPath string) bool {
	return matchSegments(splitPattern(pattern), strings.Split(importPath, pathSep))
}

// Validate checks that every rule has a non-empty pattern and a finite Max in
// [0, 1].
func Validate(rules []Rule) error {
	for index := range rules {
		err := validateRule(fmt.Sprintf("rules[%d]", index), rules[index])
		if err != nil {
			return fmt.Errorf("Validate: %w", err)
		}
	}

	return nil
}

func checkFinite(key string, value float64) error {
	if math.IsNaN(value) || math.IsInf(value, zero) {
		return fmt.Errorf(errFmtKeyed, key, errNotFinite)
	}

	return nil
}

func comparisonTolerance(value, threshold float64) float64 {
	return comparisonEps * max(one, math.Abs(value), math.Abs(threshold))
}

func distanceViolation(gate *packageGate) (Violation, bool) {
	if !gate.ok {
		return Violation{Package: "", Rule: "", Value: zero, Threshold: zero}, false
	}

	if gate.value-gate.threshold <= comparisonTolerance(gate.value, gate.threshold) {
		return Violation{Package: "", Rule: "", Value: zero, Threshold: zero}, false
	}

	return Violation{
		Package:   gate.pkg,
		Value:     gate.value,
		Threshold: gate.threshold,
		Rule:      gate.pattern,
	}, true
}

func evaluatePackage(pkg *PackageReport, rules []Rule) []Violation {
	threshold, pattern := matchingRule(pkg.Path, rules)

	if pattern == emptyString {
		return nil
	}

	return scanMetrics(pkg, pattern, threshold)
}

func matchDoubleStar(pattern, path []string, pos matchPos) bool {
	pos.pi++

	if pos.pi == len(pattern) {
		return true
	}

	for pos.si <= len(path) {
		if matchFrom(pattern, path, pos) {
			return true
		}

		pos.si++
	}

	return false
}

func matchFrom(pattern, path []string, pos matchPos) bool {
	for pos.pi < len(pattern) {
		if pattern[pos.pi] == doubleStar {
			return matchDoubleStar(pattern, path, pos)
		}

		if !matchOne(pattern, path, &pos) {
			return false
		}
	}

	return pos.si == len(path)
}

func matchOne(pattern, path []string, pos *matchPos) bool {
	if pos.si >= len(path) {
		return false
	}

	if pattern[pos.pi] != star && pattern[pos.pi] != path[pos.si] {
		return false
	}

	pos.pi++

	pos.si++

	return true
}

func matchSegments(pattern, path []string) bool {
	return matchFrom(pattern, path, matchPos{pi: zero, si: zero})
}

func matchingCandidate(rule *Rule, importPath, currentPattern string) *Rule {
	if !MatchPackage(rule.Pattern, importPath) {
		return nil
	}

	if currentPattern != emptyString && !moreSpecific(rule.Pattern, currentPattern) {
		return nil
	}

	return rule
}

func matchingRule(importPath string, rules []Rule) (threshold float64, pattern string) {
	for index := range rules {
		rule := matchingCandidate(&rules[index], importPath, pattern)

		if rule == nil {
			continue
		}

		threshold = rule.Max
		pattern = rule.Pattern
	}

	return threshold, pattern
}

func metricGate(input *gateInput) *packageGate {
	if input.res.Name != MetricDistance || !input.res.Applicable {
		return &packageGate{
			pkg:       input.pkg.Path,
			pattern:   input.pattern,
			threshold: input.threshold,
			value:     zero,
			ok:        false,
		}
	}

	return &packageGate{
		pkg:       input.pkg.Path,
		pattern:   input.pattern,
		threshold: input.threshold,
		value:     input.res.Value,
		ok:        true,
	}
}

func moreSpecific(candidate, current string) bool {
	candidateLiteral, candidateWildcards, candidateSegments := patternSpecificity(candidate)
	currentLiteral, currentWildcards, currentSegments := patternSpecificity(current)

	if candidateLiteral != currentLiteral {
		return candidateLiteral > currentLiteral
	}

	if candidateWildcards != currentWildcards {
		return candidateWildcards < currentWildcards
	}

	// A longer pattern is more constrained. Equality deliberately returns true
	// so later rules override earlier rules with the same specificity.
	return candidateSegments >= currentSegments
}

func patternSpecificity(pattern string) (literal, wildcards, segments int) {
	segmentsList := splitPattern(pattern)

	for index := range segmentsList {
		segment := &segmentsList[index]
		segments++

		if *segment == star || *segment == doubleStar {
			wildcards++

			continue
		}

		literal++
	}

	return literal, wildcards, segments
}

func scanMetrics(pkg *PackageReport, pattern string, threshold float64) []Violation {
	violations := make([]Violation, zero, len(pkg.Metrics))

	for index := range pkg.Metrics {
		item, hit := distanceViolation(metricGate(&gateInput{
			pkg:       pkg,
			res:       &pkg.Metrics[index],
			pattern:   pattern,
			threshold: threshold,
		}))

		if hit {
			violations = append(violations, item)
		}
	}

	return violations
}

func splitPattern(pattern string) []string {
	if pattern == emptyString {
		return nil
	}

	return strings.Split(pattern, pathSep)
}

func validateRule(key string, rule Rule) error {
	if rule.Pattern == emptyString {
		return fmt.Errorf(errFmtKeyed, key, errPatternEmpty)
	}

	err := checkFinite(key+".max", rule.Max)
	if err != nil {
		return fmt.Errorf("validateRule: %w", err)
	}

	if rule.Max < zero || rule.Max > one {
		return fmt.Errorf("%s.max: %w (got %g)", key, errMaxOutOfRange, rule.Max)
	}

	return nil
}
