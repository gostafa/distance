// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package public

import (
	"fmt"

	"github.com/gostafa/distance/internal/features/policy/adapters/text"
	"github.com/gostafa/distance/internal/features/policy/application"
	"github.com/gostafa/distance/internal/features/policy/domain"
)

// Check validates and applies a policy, defaulting an empty list to the recommended rule.
func Check(report *Report, rules []Rule) ([]Violation, error) {
	rows, err := application.Check(report, toRules(rules))
	if err != nil {
		return nil, fmt.Errorf("check policy: %w", err)
	}

	return rows, nil
}

// DefaultRules returns the default catch-all policy.
func DefaultRules() []Rule { return fromRules(domain.DefaultRules()) }

// Evaluate applies validated rules; an empty list selects the default policy.
func Evaluate(report *Report, rules []Rule) []Violation {
	return application.Evaluate(report, toRules(rules))
}

// FormatViolations renders a diagnostic summary.
func FormatViolations(rows []Violation) string { return text.FormatViolations(rows) }

// MatchPackage matches a full import path against the policy glob.
func MatchPackage(pattern, path string) bool { return domain.MatchPackage(pattern, path) }

// Validate checks patterns and finite thresholds in the inclusive range [0,1].
func Validate(rules []Rule) error {
	err := domain.Validate(toRules(rules))
	if err != nil {
		return fmt.Errorf("validate policy: %w", err)
	}

	return nil
}

func fromRules(rules []domain.Rule) []Rule {
	result := make([]Rule, zero, len(rules))

	for i := range rules {
		result = append(result, Rule{Pattern: rules[i].Pattern, Max: rules[i].Max})
	}

	return result
}

func toRules(rules []Rule) []domain.Rule {
	result := make([]domain.Rule, zero, len(rules))

	for i := range rules {
		result = append(result, domain.Rule{Pattern: rules[i].Pattern, Max: rules[i].Max})
	}

	return result
}
