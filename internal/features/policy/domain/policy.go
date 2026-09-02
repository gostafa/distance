package domain

import (
	"fmt"
	"math"
	"strings"
)

// DefaultMaxDistance is the recommended maximum distance from the main
// sequence. Fail when a package's applicable distance exceeds this bound.
const DefaultMaxDistance = 0.5

// PackageRule is one first-match policy rule: packages whose import path
// matches Pattern are gated by MaxDistance.
type PackageRule struct {
	// Pattern is a go list package pattern (./..., ./internal/..., an
	// exact import path, or mod/pkg/...).
	Pattern string `json:"pattern"`
	// MaxDistance is the exclusive upper bound: fail when distance >
	// MaxDistance.
	MaxDistance float64 `json:"max-distance"`
}

// Policy is a complete set of package-distance rules. The first matching
// rule in list order wins. Packages that match no rule are not gated.
type Policy struct {
	Packages []PackageRule
}

// Validate checks that every rule has a non-empty pattern and a finite
// maximum distance.
func Validate(p Policy) error {
	for i, rule := range p.Packages {
		if strings.TrimSpace(rule.Pattern) == "" {
			return fmt.Errorf("packages[%d]: pattern must not be empty", i)
		}

		if !finite(rule.MaxDistance) {
			return fmt.Errorf("packages[%d]: max-distance must be a finite number", i)
		}
	}

	return nil
}

// PolicyFromPatterns builds a validated policy from positional patterns
// sharing one maximum distance. Empty patterns default to "./...".
func PolicyFromPatterns(patterns []string, maxDistance float64) (Policy, error) {
	if len(patterns) == 0 {
		patterns = []string{"./..."}
	}

	rules := make([]PackageRule, len(patterns))
	for i, pattern := range patterns {
		rules[i] = PackageRule{
			Pattern:     pattern,
			MaxDistance: maxDistance,
		}
	}

	policy := Policy{Packages: rules}
	if err := Validate(policy); err != nil {
		return Policy{}, err
	}

	return policy, nil
}

// LoadPatterns returns the deduplicated union of rule patterns in order.
func LoadPatterns(rules []PackageRule) []string {
	patterns := make([]string, 0, len(rules))
	seen := make(map[string]bool, len(rules))
	for _, rule := range rules {
		if rule.Pattern == "" || seen[rule.Pattern] {
			continue
		}

		seen[rule.Pattern] = true
		patterns = append(patterns, rule.Pattern)
	}

	return patterns
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

// MatchPattern reports whether importPath matches a go list-style pattern.
// Relative patterns (./foo, ./foo/...) are resolved against modulePath.
func MatchPattern(pattern, importPath, modulePath string) bool {
	resolved := resolvePattern(pattern, modulePath)
	if resolved == "./..." {
		return true
	}

	if strings.HasSuffix(resolved, "/...") {
		prefix := strings.TrimSuffix(resolved, "/...")
		if prefix == "" {
			return true
		}

		return importPath == prefix || strings.HasPrefix(importPath, prefix+"/")
	}

	return importPath == resolved
}

func resolvePattern(pattern, modulePath string) string {
	if pattern == "." {
		if modulePath == "" {
			return pattern
		}

		return modulePath
	}

	if pattern == "./..." {
		if modulePath == "" {
			return "./..."
		}

		return modulePath + "/..."
	}

	if after, ok := strings.CutPrefix(pattern, "./"); ok {
		if modulePath == "" {
			return pattern
		}

		if after == "" {
			return modulePath
		}

		return modulePath + "/" + after
	}

	return pattern
}

// matchingRule returns the first rule whose pattern matches importPath.
func matchingRule(rules []PackageRule, importPath, modulePath string) (PackageRule, bool) {
	for _, rule := range rules {
		if MatchPattern(rule.Pattern, importPath, modulePath) {
			return rule, true
		}
	}

	return PackageRule{}, false
}
