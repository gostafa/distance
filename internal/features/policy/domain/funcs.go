// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/gostafa/distance/distance"
	"github.com/gostafa/distance/internal/shared/metrics"
)

// Evaluate returns policy violations for packages in report.
func Evaluate(report *distance.Report, policy Policy) []Violation {
	violations := make([]Violation, zero, len(report.Packages))

	for i := range report.Packages {
		violations = append(
			violations,
			packageViolations(&report.Packages[i], policy, report.Module)...,
		)
	}

	return violations
}

// FormatViolations renders violations as a multi-line policy failure message.
func FormatViolations(violations []Violation) string {
	if len(violations) == zero {
		return emptyString
	}

	return formatViolationList(violations)
}

// Limit returns the exclusive distance bound.
func (rule PackageRule) Limit() float64 {
	return rule.MaxDistance
}

// LoadPatterns returns unique non-empty patterns from rules, preserving order.
func LoadPatterns(rules []PackageRule) []string {
	patterns := make([]string, zero, len(rules))
	seen := make(map[string]bool, len(rules))

	return appendUniquePatterns(patterns, seen, rules)
}

// MatchPattern reports whether importPath matches pattern under modulePath.
func MatchPattern(pattern, importPath, modulePath string) bool {
	resolved := resolvePattern(pattern, modulePath)

	if resolved == allPackagesPattern {
		return true
	}

	return matchResolved(resolved, importPath)
}

// PolicyFromPatterns builds a Policy with one rule per pattern.
func PolicyFromPatterns(patterns []string, maxDistance float64) (Policy, error) {
	if len(patterns) == zero {
		patterns = []string{allPackagesPattern}
	}

	policy, err := validatedPolicy(Policy{Packages: rulesFromPatterns(patterns, maxDistance)})
	if err != nil {
		return Policy{}, fmt.Errorf("policy PolicyFromPatterns: %w", err)
	}

	return policy, nil
}

// RulePattern returns the package match pattern.
func (rule PackageRule) RulePattern() string {
	return rule.Pattern
}

// Validate checks that every package rule is well-formed.
func Validate(policy Policy) error {
	for i := range policy.Packages {
		checkErr := checkRule(policy.Packages[i], i)
		if checkErr != nil {
			return fmt.Errorf("policy Validate: %w", checkErr)
		}
	}

	return nil
}

// Error returns the validation message for a package rule.
func (err ruleError) Error() string {
	return fmt.Sprintf("packages[%d]: %s", err.index, err.reason)
}

func appendUniquePatterns(patterns []string, seen map[string]bool, rules []PackageRule) []string {
	for i := range rules {
		pattern := rules[i].RulePattern()

		if pattern == emptyString || seen[pattern] {
			continue
		}

		seen[pattern] = true
		patterns = append(patterns, pattern)
	}

	return patterns
}

func checkRule(rule PackageRule, index int) error {
	if strings.TrimSpace(rule.RulePattern()) == emptyString {
		return ruleError{index: index, reason: "pattern must not be empty"}
	}

	if !finite(rule.Limit()) {
		return ruleError{index: index, reason: "max-distance must be a finite number"}
	}

	return nil
}

func comparisonTolerance(value, threshold float64) float64 {
	return comparisonEpsilon * max(one, math.Abs(value), math.Abs(threshold))
}

func distanceViolation(pkg string, res *metrics.MetricResult, rule PackageRule) (Violation, bool) {
	if res.Name != metrics.MetricDistance || !res.Applicable {
		return Violation{}, false
	}

	if res.Value-rule.Limit() <= comparisonTolerance(res.Value, rule.Limit()) {
		return Violation{}, false
	}

	return Violation{
		Package:    pkg,
		Key:        res.Name,
		Value:      res.Value,
		Comparator: ComparatorMax,
		Threshold:  rule.Limit(),
	}, true
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, zero)
}

func formatNumber(value float64) string {
	if value == math.Trunc(value) && !math.IsInf(value, zero) {
		return strconv.FormatFloat(value, 'f', floatPrecAuto, floatBitSize)
	}

	return strconv.FormatFloat(value, 'f', floatPrecFixed, floatBitSize)
}

func formatViolationList(violations []Violation) string {
	var builder strings.Builder

	writeErr := writeBuilder(
		&builder,
		fmt.Sprintf("policy: %d %s\n", len(violations), violationNoun(len(violations))),
	)
	if writeErr != nil {
		return emptyString
	}

	return writeViolationLines(&builder, violations)
}

func matchResolved(resolved, importPath string) bool {
	before, ok := strings.CutSuffix(resolved, "/...")

	if !ok {
		return importPath == resolved
	}

	if before == emptyString {
		return true
	}

	return importPath == before || strings.HasPrefix(importPath, before+"/")
}

func matchingRule(rules []PackageRule, importPath, modulePath string) (PackageRule, bool) {
	for i := range rules {
		if MatchPattern(rules[i].RulePattern(), importPath, modulePath) {
			return rules[i], true
		}
	}

	return PackageRule{}, false
}

func packageViolations(pkg *distance.PackageReport, policy Policy, module string) []Violation {
	rule, ok := matchingRule(policy.Packages, pkg.Path, module)

	if !ok {
		return nil
	}

	return scanMetrics(pkg, rule)
}

func resolveAll(modulePath string) string {
	if modulePath == emptyString {
		return allPackagesPattern
	}

	return modulePath + "/..."
}

func resolveDot(modulePath string) string {
	if modulePath == emptyString {
		return currentPackage
	}

	return modulePath
}

func resolvePattern(pattern, modulePath string) string {
	if pattern == currentPackage {
		return resolveDot(modulePath)
	}

	if pattern == allPackagesPattern {
		return resolveAll(modulePath)
	}

	return resolveRelative(pattern, modulePath)
}

func resolveRelative(pattern, modulePath string) string {
	after, ok := strings.CutPrefix(pattern, "./")

	if !ok || modulePath == emptyString {
		return pattern
	}

	if after == emptyString {
		return modulePath
	}

	return modulePath + "/" + after
}

func rulesFromPatterns(patterns []string, maxDistance float64) []PackageRule {
	rules := make([]PackageRule, zero, len(patterns))

	for i := range patterns {
		rules = append(rules, PackageRule{Pattern: patterns[i], MaxDistance: maxDistance})
	}

	return rules
}

func scanMetrics(pkg *distance.PackageReport, rule PackageRule) []Violation {
	violations := make([]Violation, zero, len(pkg.Metrics))

	for i := range pkg.Metrics {
		item, hit := distanceViolation(pkg.Path, &pkg.Metrics[i], rule)

		if hit {
			violations = append(violations, item)
		}
	}

	return violations
}

func validatedPolicy(policy Policy) (Policy, error) {
	validateErr := Validate(policy)
	if validateErr != nil {
		return Policy{}, fmt.Errorf("policy from patterns: %w", validateErr)
	}

	return policy, nil
}

func violationNoun(count int) string {
	if count == one {
		return "violation"
	}

	return "violations"
}

func writeBuilder(builder *strings.Builder, text string) error {
	written, writeErr := builder.WriteString(text)
	if writeErr != nil {
		return fmt.Errorf(errWrapWriteBuilder, writeErr)
	}

	if written != len(text) {
		return fmt.Errorf(errWrapWriteBuilder, errShortWrite)
	}

	return nil
}

func writeViolationLines(builder *strings.Builder, violations []Violation) string {
	for i := range violations {
		item := &violations[i]
		line := fmt.Sprintf(
			"  %s (package): %s %s exceeds max %s\n",
			item.Package,
			item.Key,
			formatNumber(item.Value),
			formatNumber(item.Threshold),
		)

		writeErr := writeBuilder(builder, line)
		if writeErr != nil {
			return builder.String()
		}
	}

	return builder.String()
}
